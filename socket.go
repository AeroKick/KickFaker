package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	channels  map[string]bool
	sessionID string
}

type ClientManager struct {
	clients    map[*Client]bool
	sessions   map[string][]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	sync.Mutex
}

var manager = ClientManager{
	clients:    make(map[*Client]bool),
	sessions:   make(map[string][]*Client),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (manager *ClientManager) run() {
	for {
		select {
		case client := <-manager.register:
			manager.Lock()
			manager.clients[client] = true
			manager.sessions[client.sessionID] = append(manager.sessions[client.sessionID], client)
			manager.Unlock()
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				close(client.send)
				manager.Lock()
				delete(manager.clients, client)
				// Remove client from sessions
				if clients, ok := manager.sessions[client.sessionID]; ok {
					newClients := make([]*Client, 0)
					for _, c := range clients {
						if c != client {
							newClients = append(newClients, c)
						}
					}
					if len(newClients) == 0 {
						delete(manager.sessions, client.sessionID)
						// Clean up tickers for this session
						if ticker, exists := chatIntervals[client.sessionID]; exists {
							ticker.Stop()
							delete(chatIntervals, client.sessionID)
						}
						if ticker, exists := allEventsIntervals[client.sessionID]; exists {
							ticker.Stop()
							delete(allEventsIntervals, client.sessionID)
						}
					} else {
						manager.sessions[client.sessionID] = newClients
					}
				}
				manager.Unlock()
			}
		case message := <-manager.broadcast:
			// Broadcast to all clients regardless of session
			manager.Lock()
			for client := range manager.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(manager.clients, client)
				}
			}
			manager.Unlock()
		}
	}
}

var messageRate = 1
var chatIntervals = make(map[string]*time.Ticker)
var allEventsIntervals = make(map[string]*time.Ticker)

func updateIntervals() {
	for sessionID, ticker := range chatIntervals {
		if ticker != nil {
			ticker.Stop()
			chatIntervals[sessionID] = time.NewTicker(time.Second / time.Duration(messageRate))
		}
	}
	for sessionID, ticker := range allEventsIntervals {
		if ticker != nil {
			ticker.Stop()
			allEventsIntervals[sessionID] = time.NewTicker(time.Second / time.Duration(messageRate))
		}
	}
}

func sendEventToSubscribers(channelType string, eventGenerator func(string) PusherMessage, sessionID string) {
	manager.Lock()
	defer manager.Unlock()

	if clients, ok := manager.sessions[sessionID]; ok {
		for _, client := range clients {
			for channel := range client.channels {
				if strings.HasPrefix(channel, channelType) {
					event := eventGenerator(channel)
					eventJSON, _ := json.Marshal(event)
					client.send <- eventJSON
				}
			}
		}
	}
}

func triggerEvent(eventType string, sessionID string) {
	var eventGenerator func(string) PusherMessage
	var channelType string

	switch eventType {
	case "chat":
		eventGenerator = GenerateRandomChatMessage
		channelType = "chatroom"
	case "chat_celebration":
		eventGenerator = GenerateRandomCelebrationChatMessage
		channelType = "chatroom"
	case "subscription":
		eventGenerator = GenerateRandomSubscriptionEvent
		channelType = "chatroom"
	case "gifted_subscriptions":
		eventGenerator = GenerateRandomGiftedSubscriptionsEvent
		channelType = "chatroom"
	case "raid":
		eventGenerator = GenerateRandomRaidEvent
		channelType = "chatroom"
	case "live":
		eventGenerator = func(channel string) PusherMessage {
			return GenerateRandomLiveEvent(channel, true)
		}
		channelType = "channel"
	case "stop_broadcast":
		eventGenerator = func(channel string) PusherMessage {
			return GenerateRandomLiveEvent(channel, false)
		}
		channelType = "channel"
	default:
		log.Printf("Unknown event type: %s", eventType)
		return
	}

	sendEventToSubscribers(channelType, eventGenerator, sessionID)
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		manager.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incomingMessage map[string]interface{}
		if err := json.Unmarshal(message, &incomingMessage); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		event, ok := incomingMessage["event"].(string)
		if !ok {
			log.Printf("error: event is not a string")
			continue
		}

		switch event {
		case "pusher:subscribe":
			data, ok := incomingMessage["data"].(map[string]interface{})
			if !ok {
				log.Printf("error: data is not a map")
				continue
			}
			channel, ok := data["channel"].(string)
			if !ok {
				log.Printf("error: channel is not a string")
				continue
			}

			// Subscribe to the channel
			c.channels[channel] = true

			// Send subscription succeeded message
			successMessage := PusherMessage{
				Event:   "pusher_internal:subscription_succeeded",
				Data:    "{}",
				Channel: channel,
			}
			successMessageJSON, _ := json.Marshal(successMessage)
			c.send <- successMessageJSON

		case "set_message_rate":
			data, ok := incomingMessage["data"].(map[string]interface{})
			if !ok {
				log.Printf("error: data is not a map")
				continue
			}
			rate, ok := data["rate"].(float64)
			if !ok {
				log.Printf("error: rate is not a number")
				continue
			}

			messageRate = int(rate)
			log.Printf("Message rate set to %d messages per second", messageRate)
			updateIntervals()

		case "trigger_event":
			data, ok := incomingMessage["data"].(map[string]interface{})
			if !ok {
				log.Printf("error: data is not a map")
				continue
			}
			eventType, ok := data["type"].(string)
			if !ok {
				log.Printf("error: event type is not a string")
				continue
			}
			triggerEvent(eventType, c.sessionID)

		case "toggle_chat_interval":
			if ticker, exists := chatIntervals[c.sessionID]; exists && ticker != nil {
				ticker.Stop()
				delete(chatIntervals, c.sessionID)
			} else {
				ticker := time.NewTicker(time.Second / time.Duration(messageRate))
				chatIntervals[c.sessionID] = ticker
				go func() {
					for range ticker.C {
						sendEventToSubscribers("chatroom", GenerateRandomChatMessage, c.sessionID)
					}
				}()
			}

		case "toggle_all_events_interval":
			if ticker, exists := allEventsIntervals[c.sessionID]; exists && ticker != nil {
				ticker.Stop()
				delete(allEventsIntervals, c.sessionID)
			} else {
				ticker := time.NewTicker(time.Second / time.Duration(messageRate))
				allEventsIntervals[c.sessionID] = ticker
				go func() {
					for range ticker.C {
						sendEventToSubscribers("chatroom", GenerateRandomEvent, c.sessionID)
						sendEventToSubscribers("channel", GenerateRandomEvent, c.sessionID)
					}
				}()
			}
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		channels:  make(map[string]bool),
		sessionID: sessionID,
	}
	manager.register <- client

	// Send connection established message with session ID
	establishedMessage := PusherMessage{
		Event: "pusher:connection_established",
		Data:  fmt.Sprintf("{\"socket_id\":\"400952.30449\",\"activity_timeout\":120,\"session_id\":\"%s\"}", sessionID),
	}
	establishedMessageJSON, _ := json.Marshal(establishedMessage)
	client.send <- establishedMessageJSON

	go client.writePump()
	go client.readPump()
}
