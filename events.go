package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type PusherMessage struct {
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
}

type IncomingPusherMessage struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}

type parsedChatData struct {
	AeroKickChannelId uuid.UUID
	ID                string      `json:"id"`
	ChatroomID        json.Number `json:"chatroom_id"`
	Content           string      `json:"content"`
	Type              string      `json:"type"`
	CreatedAt         string      `json:"created_at"`
	Sender            struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Slug     string `json:"slug"`
		Identity struct {
			Color  string `json:"color"`
			Badges []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"badges"`
		} `json:"identity"`
	} `json:"sender"`
	MetaData struct {
		Celebration struct {
			ID          uint        `json:"id"`
			Type        string      `json:"type"`
			TotalMonths json.Number `json:"total_months"`
			CreatedAt   string      `json:"created_at"`
		} `json:"celebration"`
	} `json:"metadata"`
}

type ParsedSubscriberData struct {
	ChatroomID int     `json:"chatroom_id"`
	Username   *string `json:"username"`
	Months     int     `json:"months"`
}

type ParsedGiftedSubscriptionsData struct {
	ChatroomID  *uint    `json:"chatroom_id"`
	Username    *string  `json:"gifter_username"`
	Gifts       []string `json:"gifted_usernames"`
	GifterTotal int      `json:"gifter_total"`
}

type ParsedRaidData struct {
	ChatroomID *uint `json:"chatroom_id"`
	Message    struct {
		NumberofViewers *uint `json:"numberOfViewers"`
	} `json:"message"`
	User struct {
		ID       *uint   `json:"id"`
		Username *string `json:"username"`
	} `json:"user"`
}

type IsLiveData struct {
	Livestream struct {
		ID    int    `json:"id"`
		Title string `json:"session_title"`
	} `json:"livestream"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateRandomChatMessage(channelID string) PusherMessage {
	var user = GetRandomSlug()
	chatData := parsedChatData{
		AeroKickChannelId: uuid.New(),
		ID:                uuid.New().String(),
		ChatroomID:        json.Number(fmt.Sprintf("%d", rand.Intn(10000))),
		Content:           fmt.Sprintf(GetRandomMessage()),
		Type:              "message",
		CreatedAt:         time.Now().Format(time.RFC3339),
		Sender: struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Slug     string `json:"slug"`
			Identity struct {
				Color  string `json:"color"`
				Badges []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"badges"`
			} `json:"identity"`
		}{
			ID:       uint(rand.Intn(10000)),
			Username: fmt.Sprintf(user),
			Slug:     fmt.Sprintf(user),
			Identity: struct {
				Color  string `json:"color"`
				Badges []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"badges"`
			}{
				Color:  fmt.Sprintf("#%06x", rand.Intn(0xFFFFFF)),
				Badges: GetRandomBadges(),
			},
		},
	}

	chatDataJSON, _ := json.Marshal(chatData)
	return PusherMessage{
		Event:   "App\\Events\\ChatMessageEvent",
		Data:    string(chatDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomCelebrationChatMessage(channelID string) PusherMessage {
	var user = GetRandomSlug()
	chatData := parsedChatData{
		AeroKickChannelId: uuid.New(),
		ID:                uuid.New().String(),
		ChatroomID:        json.Number(fmt.Sprintf("%d", rand.Intn(10000))),
		Content:           fmt.Sprintf(GetRandomMessage()),
		Type:              "celebration",
		CreatedAt:         time.Now().Format(time.RFC3339),
		Sender: struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Slug     string `json:"slug"`
			Identity struct {
				Color  string `json:"color"`
				Badges []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"badges"`
			} `json:"identity"`
		}{
			ID:       uint(rand.Intn(10000)),
			Username: fmt.Sprintf(user),
			Slug:     fmt.Sprintf(user),
			Identity: struct {
				Color  string `json:"color"`
				Badges []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"badges"`
			}{
				Color:  fmt.Sprintf("#%06x", rand.Intn(0xFFFFFF)),
				Badges: GetRandomBadges(),
			},
		},
		MetaData: struct {
			Celebration struct {
				ID          uint        `json:"id"`
				Type        string      `json:"type"`
				TotalMonths json.Number `json:"total_months"`
				CreatedAt   string      `json:"created_at"`
			} `json:"celebration"`
		}(struct {
			Celebration struct {
				ID          uint        `json:"id"`
				Type        string      `json:"type"`
				TotalMonths json.Number `json:"total_months"`
				CreatedAt   string      `json:"created_at"`
			}
		}{
			Celebration: struct {
				ID          uint        `json:"id"`
				Type        string      `json:"type"`
				TotalMonths json.Number `json:"total_months"`
				CreatedAt   string      `json:"created_at"`
			}{
				ID:          uint(rand.Intn(10000)),
				Type:        "subscription_renewed",
				TotalMonths: json.Number(fmt.Sprintf("%d", rand.Intn(12))),
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
		}),
	}

	chatDataJSON, _ := json.Marshal(chatData)
	return PusherMessage{
		Event:   "App\\Events\\ChatMessageEvent",
		Data:    string(chatDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomSubscriptionEvent(channelID string) PusherMessage {
	//chatroomID := uint(rand.Intn(10000))
	username := fmt.Sprintf("user%d", rand.Intn(1000))

	subData := ParsedSubscriberData{
		ChatroomID: 2271287,
		Username:   &username,
		Months:     1,
	}

	subDataJSON, _ := json.Marshal(subData)
	return PusherMessage{
		Event:   "App\\Events\\SubscriptionEvent",
		Data:    string(subDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomGiftedSubscriptionsEvent(channelID string) PusherMessage {
	chatroomID := uint(rand.Intn(10000))
	gifterUsername := fmt.Sprintf("gifter%d", rand.Intn(1000))
	gifterTotal := rand.Intn(10) + 1
	gifts := make([]string, gifterTotal)
	for i := 0; i < gifterTotal; i++ {
		gifts[i] = fmt.Sprintf("user%d", rand.Intn(1000))
	}

	giftData := ParsedGiftedSubscriptionsData{
		ChatroomID:  &chatroomID,
		Username:    &gifterUsername,
		Gifts:       gifts,
		GifterTotal: gifterTotal,
	}

	giftDataJSON, _ := json.Marshal(giftData)
	return PusherMessage{
		Event:   "App\\Events\\GiftedSubscriptionsEvent",
		Data:    string(giftDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomRaidEvent(channelID string) PusherMessage {
	chatroomID := uint(rand.Intn(10000))
	hostUsername := fmt.Sprintf("host%d", rand.Intn(1000))
	viewers := uint(rand.Intn(1000) + 1)
	userid := uint(rand.Intn(10000))

	raidData := ParsedRaidData{
		ChatroomID: &chatroomID,
		Message: struct {
			NumberofViewers *uint `json:"numberOfViewers"`
		}{
			NumberofViewers: &viewers,
		},
		User: struct {
			ID       *uint   `json:"id"`
			Username *string `json:"username"`
		}{ID: &userid, Username: &hostUsername},
	}

	raidDataJSON, _ := json.Marshal(raidData)
	return PusherMessage{
		Event:   "App\\Events\\StreamHostedEvent",
		Data:    string(raidDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomLiveEvent(channelID string, isLive bool) PusherMessage {
	liveData := IsLiveData{
		Livestream: struct {
			ID    int    `json:"id"`
			Title string `json:"session_title"`
		}{
			ID:    rand.Intn(10000),
			Title: fmt.Sprintf("Stream %d", rand.Intn(1000)),
		},
	}

	liveDataJSON, _ := json.Marshal(liveData)
	event := "App\\Events\\StreamerIsLive"
	if !isLive {
		event = "App\\Events\\StopStreamBroadcast"
	}

	return PusherMessage{
		Event:   event,
		Data:    string(liveDataJSON),
		Channel: channelID,
	}
}

func GenerateRandomEvent(channelID string) PusherMessage {
	eventType := rand.Intn(5)
	switch eventType {
	case 0:
		return GenerateRandomChatMessage(channelID)
	case 1:
		return GenerateRandomSubscriptionEvent(channelID)
	case 2:
		return GenerateRandomGiftedSubscriptionsEvent(channelID)
	case 3:
		return GenerateRandomRaidEvent(channelID)
	case 4:
		return GenerateRandomLiveEvent(channelID, rand.Intn(2) == 0)
	default:
		return GenerateRandomChatMessage(channelID)
	}
}
