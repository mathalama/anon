package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/mathalama/nektokz/chat-service/internal/delivery/ws"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ChatHandler struct {
	hub  *ws.Hub
	repo domain.ChatRepository
	moderation domain.ModerationClient
	internalToken string
}

func NewChatHandler(r chi.Router, hub *ws.Hub, repo domain.ChatRepository, moderation domain.ModerationClient, internalToken string) {
	h := &ChatHandler{hub: hub, repo: repo, moderation: moderation, internalToken: internalToken}

	r.Get("/ws", h.ServeWS)
	r.Post("/internal/rooms", h.CreateRoom)
	r.Post("/internal/disconnect", h.DisconnectUser)
	r.Get("/health", h.Health)
}

func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	userID := r.Header.Get("X-User-ID") // Passed by gateway
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	room, err := h.repo.GetRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	if room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	if userID != room.UserA && userID != room.UserB {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, userID, roomID, h.repo, h.moderation)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req domain.Room
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Status = "active"
	if err := h.repo.CreateRoom(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) DisconnectUser(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.hub.DisconnectUser(req.UserID)
	w.WriteHeader(http.StatusOK)
}

func (h *ChatHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *ChatHandler) isInternalAuthorized(r *http.Request) bool {
	if h.internalToken == "" {
		return false
	}
	return r.Header.Get("X-Internal-Token") == h.internalToken
}
