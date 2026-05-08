package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	goredis "github.com/redis/go-redis/v9"
)

var (
	wsConnectionsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_connections_total",
		Help: "Current WebSocket connections",
	})
	wsMessagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_total",
		Help: "Total WebSocket messages processed",
	})
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
	shards []*HubShard
	rdb    *goredis.Client
}

// HubShard handles a subset of clients to reduce mutex contention
type HubShard struct {
	clients    map[string]*Client             // userID -> client
	rooms      map[string]map[string]struct{} // roomID -> set(userID)
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

const numShards = 256

func (h *Hub) getShard(userID string) *HubShard {
	if len(userID) == 0 {
		return h.shards[0]
	}
	// Use int to avoid byte overflow
	return h.shards[int(userID[0])%numShards]
}

func NewHub(rdb *goredis.Client) *Hub {
	shards := make([]*HubShard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &HubShard{
			clients:    make(map[string]*Client),
			rooms:      make(map[string]map[string]struct{}),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	}
	return &Hub{
		shards: shards,
		rdb:    rdb,
	}
}

func (h *Hub) Run(ctx context.Context) {
	if h.rdb != nil {
		go h.listenRedis(ctx)
	}

	// Start a goroutine for each shard
	for _, shard := range h.shards {
		go h.runShard(ctx, shard)
	}

	// Wait for context cancellation
	<-ctx.Done()
}

func (h *Hub) runShard(ctx context.Context, shard *HubShard) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-shard.register:
			shard.mu.Lock()
			shard.clients[client.UserID] = client
			if _, ok := shard.rooms[client.RoomID]; !ok {
				shard.rooms[client.RoomID] = make(map[string]struct{})
			}
			shard.rooms[client.RoomID][client.UserID] = struct{}{}
			shard.mu.Unlock()

			// Increment metrics
			wsConnectionsTotal.Inc()

			// Use Redis to track global room occupancy across shards in a goroutine
			// to avoid blocking the shard's main loop.
			if h.rdb != nil {
				go func(rid, uid string) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					key := "room:occupancy:" + rid
					h.rdb.SAdd(ctx, key, uid)
					h.rdb.Expire(ctx, key, 1*time.Hour)
					count, _ := h.rdb.SCard(ctx, key).Result()

					if count >= 2 {
						h.BroadcastToRoom(rid, "", ServerMessage{
							Type:      "partner_connected",
							Timestamp: time.Now().Unix(),
						})
					}
				}(client.RoomID, client.UserID)
			}

		case client := <-shard.unregister:
			roomID := client.RoomID
			userID := client.UserID

			shard.mu.Lock()
			if _, ok := shard.clients[userID]; ok {
				delete(shard.clients, userID)
				select {
				case <-client.done:
					// уже закрыт
				default:
					close(client.done)
				}
			}

			if users, ok := shard.rooms[roomID]; ok {
				delete(users, userID)
				if len(users) == 0 {
					delete(shard.rooms, roomID)
				}
			}
			shard.mu.Unlock()

			// Decrement metrics
			wsConnectionsTotal.Dec()

			if h.rdb != nil {
				go func(rid, uid string) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					key := "room:occupancy:" + rid
					h.rdb.SRem(ctx, key, uid)
				}(roomID, userID)
			}

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
		case "partner_typing", "rtc:offer", "rtc:answer", "rtc:ice-candidate", "call:start", "call:end":
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
	shard := h.getShard(userID)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if client, ok := shard.clients[userID]; ok {
		client.conn.Close()
		// Shard will handle unregister via readPump failure
	}
}

func (h *Hub) DisconnectRoom(roomID string) {
	// Since we don't know the shards for all users in a room without scanning,
	// we broadcast a disconnect message to all shards
	for _, shard := range h.shards {
		shard.mu.RLock()
		users := shard.rooms[roomID]
		clients := make([]*Client, 0, len(users))
		for userID := range users {
			if c, ok := shard.clients[userID]; ok {
				clients = append(clients, c)
			}
		}
		shard.mu.RUnlock()

		for _, c := range clients {
			c.conn.Close()
		}
	}
}

func (h *Hub) sendToRoomLocal(roomID string, build func(recipientID string) ServerMessage) {
	// Scan all shards for users in this room
	// This is necessary because we don't centralize room membership
	for _, shard := range h.shards {
		shard.mu.RLock()
		users := shard.rooms[roomID]
		clients := make([]*Client, 0, len(users))
		for userID := range users {
			if c, ok := shard.clients[userID]; ok {
				clients = append(clients, c)
			}
		}
		shard.mu.RUnlock()

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
				// Buffer full - client is not reading fast enough, close connection
				go h.Unregister(c)
			}
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
	shard := h.getShard(client.UserID)
	shard.register <- client
}

func (h *Hub) Unregister(client *Client) {
	shard := h.getShard(client.UserID)
	shard.unregister <- client
}
