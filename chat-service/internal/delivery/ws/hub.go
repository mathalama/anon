package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	repo domain.ChatRepository
	moderation domain.ModerationClient
	
	UserID string
	RoomID string
}

type Hub struct {
	clients    map[string]*Client // userID -> client
	rooms      map[string]map[string]struct{} // roomID -> set(userID)
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
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
				h.sendToRoom(client.RoomID, func(recipientID string) ServerMessage {
					return ServerMessage{Type: "partner_connected", Timestamp: time.Now().Unix()}
				})
				h.sendToRoom(client.RoomID, func(recipientID string) ServerMessage {
					return ServerMessage{Type: "match_found", RoomID: client.RoomID, Timestamp: time.Now().Unix()}
				})
			}

		case client := <-h.unregister:
			roomID := client.RoomID
			userID := client.UserID

			h.mu.Lock()
			if _, ok := h.clients[userID]; ok {
				delete(h.clients, userID)
				close(client.send)
			}

			if users, ok := h.rooms[roomID]; ok {
				delete(users, userID)
				if len(users) == 0 {
					delete(h.rooms, roomID)
				}
			}
			h.mu.Unlock()

			// Notify remaining user (if any).
			h.sendToRoom(roomID, func(recipientID string) ServerMessage {
				return ServerMessage{Type: "partner_disconnected", Timestamp: time.Now().Unix()}
			})
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

func (h *Hub) sendToRoom(roomID string, build func(recipientID string) ServerMessage) {
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
		case c.send <- data:
		default:
		}
	}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, roomID string, repo domain.ChatRepository, moderation domain.ModerationClient) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		repo:   repo,
		moderation: moderation,
		UserID: userID,
		RoomID: roomID,
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}
