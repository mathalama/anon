package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	repo       domain.ChatRepository
	moderation domain.ModerationClient

	UserID string
	RoomID string
	done   chan struct{}
}

type Hub struct {
	clients    map[string]*Client             // userID -> client
	rooms      map[string]map[string]struct{} // roomID -> set(userID)
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	rdb        *goredis.Client
}

func NewHub(rdb *goredis.Client) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
	}
}

func (h *Hub) Run(ctx context.Context) {
	if h.rdb != nil {
		go h.listenRedis(ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[string]struct{})
			}
			h.rooms[client.RoomID][client.UserID] = struct{}{}
			count := len(h.rooms[client.RoomID])
			h.mu.Unlock()

			// When the second user connects to the room, notify both.
			if count >= 2 {
				h.BroadcastToRoom(client.RoomID, "", ServerMessage{Type: "partner_connected", Timestamp: time.Now().Unix()})
			}

		case client := <-h.unregister:
			roomID := client.RoomID
			userID := client.UserID

			h.mu.Lock()
			if _, ok := h.clients[userID]; ok {
				delete(h.clients, userID)
				select {
				case <-client.done:
					// уже закрыт
				default:
					close(client.done)
				}
			}

			if users, ok := h.rooms[roomID]; ok {
				delete(users, userID)
				if len(users) == 0 {
					delete(h.rooms, roomID)
				}
			}
			h.mu.Unlock()

			h.BroadcastToRoom(roomID, "", ServerMessage{Type: "partner_disconnected", Timestamp: time.Now().Unix()})
		}
	}
}

type redisMsg struct {
	RoomID           string        `json:"room_id"`
	OriginalSenderID string        `json:"original_sender_id"`
	Payload          ServerMessage `json:"payload"`
}

func (h *Hub) BroadcastToRoom(roomID string, senderID string, msg ServerMessage) {
	if h.rdb == nil {
		h.sendToRoomLocal(roomID, func(recipientID string) ServerMessage {
			return h.personalize(msg, senderID, recipientID)
		})
		return
	}
	b, _ := json.Marshal(redisMsg{RoomID: roomID, OriginalSenderID: senderID, Payload: msg})
	h.rdb.Publish(context.Background(), "chat:rooms", b)
}

func (h *Hub) personalize(msg ServerMessage, senderID, recipientID string) ServerMessage {
	// Don't send typing or RTC signals back to the sender
	if senderID != "" && recipientID == senderID {
		switch msg.Type {
		case "partner_typing", "rtc:offer", "rtc:answer", "rtc:ice-candidate":
			return ServerMessage{Type: ""}
		}
	}

	if msg.Type != "message" {
		return msg
	}

	// For chat messages, set "me" or "partner"
	res := msg
	if recipientID == senderID {
		res.Sender = "me"
	} else {
		res.Sender = "partner"
	}
	return res
}

func (h *Hub) listenRedis(ctx context.Context) {
	pubsub := h.rdb.Subscribe(ctx, "chat:rooms")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			var rm redisMsg
			if err := json.Unmarshal([]byte(msg.Payload), &rm); err == nil {
				h.sendToRoomLocal(rm.RoomID, func(recipientID string) ServerMessage {
					return h.personalize(rm.Payload, rm.OriginalSenderID, recipientID)
				})
			}
		}
	}
}

func (h *Hub) DisconnectUser(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[userID]; ok {
		client.conn.Close()
		// Hub will handle unregister via readPump failure
	}
}

func (h *Hub) DisconnectRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	users := h.rooms[roomID]
	if users == nil {
		return
	}
	for userID := range users {
		if c, ok := h.clients[userID]; ok {
			c.conn.Close()
		}
	}
}

func (h *Hub) sendToRoomLocal(roomID string, build func(recipientID string) ServerMessage) {
	h.mu.RLock()
	users := h.rooms[roomID]
	clients := make([]*Client, 0, len(users))
	for userID := range users {
		if c, ok := h.clients[userID]; ok {
			clients = append(clients, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range clients {
		msg := build(c.UserID)
		if msg.Type == "" {
			continue
		}
		data, _ := json.Marshal(msg)
		select {
		case <-c.done:
			continue
		case c.send <- data:
		default:
		}
	}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, roomID string, repo domain.ChatRepository, moderation domain.ModerationClient) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		repo:       repo,
		moderation: moderation,
		UserID:     userID,
		RoomID:     roomID,
		done:       make(chan struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}
