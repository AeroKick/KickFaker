package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	conn     *websocket.Conn
	send     chan []byte
	channels map[string]bool
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	sync.Mutex
}

var manager = ClientManager{
	clients:    make(map[*Client]bool),
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
			manager.Unlock()
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				close(client.send)
				manager.Lock()
				delete(manager.clients, client)
				manager.Unlock()
			}
		case message := <-manager.broadcast:
			for client := range manager.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					manager.Lock()
					delete(manager.clients, client)
					manager.Unlock()
				}
			}
		}
	}
}

var messageRate = 1
var chatInterval *time.Ticker
var allEventsInterval *time.Ticker

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
			triggerEvent(eventType)

		case "toggle_chat_interval":
			toggleChatInterval()

		case "toggle_all_events_interval":
			toggleAllEventsInterval()
		}
	}
}

func updateIntervals() {
	if chatInterval != nil {
		chatInterval.Stop()
		chatInterval = time.NewTicker(time.Second / time.Duration(messageRate))
	}
	if allEventsInterval != nil {
		allEventsInterval.Stop()
		allEventsInterval = time.NewTicker(time.Second / time.Duration(messageRate))
	}
}

func toggleChatInterval() {
	if chatInterval == nil {
		chatInterval = time.NewTicker(time.Second / time.Duration(messageRate))
		go func() {
			for range chatInterval.C {
				sendEventToSubscribers("chatroom", GenerateRandomChatMessage)
			}
		}()
	} else {
		chatInterval.Stop()
		chatInterval = nil
	}
}

func toggleAllEventsInterval() {
	if allEventsInterval == nil {
		allEventsInterval = time.NewTicker(time.Second / time.Duration(messageRate))
		go func() {
			for range allEventsInterval.C {
				sendEventToSubscribers("chatroom", GenerateRandomEvent)
				sendEventToSubscribers("channel", GenerateRandomEvent)
			}
		}()
	} else {
		allEventsInterval.Stop()
		allEventsInterval = nil
	}
}

func sendEventToSubscribers(channelType string, eventGenerator func(string) PusherMessage) {
	for client := range manager.clients {
		for channel := range client.channels {
			if strings.HasPrefix(channel, channelType) {
				event := eventGenerator(channel)
				eventJSON, _ := json.Marshal(event)
				client.send <- eventJSON
			}
		}
	}
}

func triggerEvent(eventType string) {
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

	sendEventToSubscribers(channelType, eventGenerator)
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

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		channels: make(map[string]bool),
	}
	manager.register <- client

	// Send connection established message
	establishedMessage := PusherMessage{
		Event: "pusher:connection_established",
		Data:  "{\"socket_id\":\"400952.30449\",\"activity_timeout\":120}",
	}
	establishedMessageJSON, _ := json.Marshal(establishedMessage)
	client.send <- establishedMessageJSON

	go client.writePump()
	go client.readPump()
}
