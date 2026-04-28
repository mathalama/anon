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

	r.Route("/chat", func(r chi.Router) {
		r.HandleFunc("/ws", h.ServeWS)
		r.Post("/internal/rooms", h.CreateRoom)
		r.Post("/internal/disconnect", h.DisconnectUser)
	})
	
	r.HandleFunc("/ws", h.ServeWS) // Fallback for direct gateway /ws mapping
	r.Get("/health", h.Health)
}

func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}
	if roomID == "" {
		RespondWithError(w, r, http.StatusBadRequest, "BAD_REQUEST", "room_id is required")
		return
	}

	room, err := h.repo.GetRoom(r.Context(), roomID)
	if err != nil || room == nil {
		RespondWithError(w, r, http.StatusNotFound, "NOT_FOUND", "room not found")
		return
	}
	if userID != room.UserA && userID != room.UserB {
		RespondWithError(w, r, http.StatusForbidden, "FORBIDDEN", "access denied to this room")
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
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "internal token missing or invalid")
		return
	}
	var req domain.Room
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	req.Status = "active"
	if err := h.repo.CreateRoom(r.Context(), &req); err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) DisconnectUser(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "internal token missing or invalid")
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	h.hub.DisconnectUser(req.UserID)
	w.WriteHeader(http.StatusOK)
}

func (h *ChatHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("ERROR: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *ChatHandler) isInternalAuthorized(r *http.Request) bool {
	if h.internalToken == "" {
		return false
	}
	return r.Header.Get("X-Internal-Token") == h.internalToken
}
