package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mathalama/nektokz/notification-service/internal/domain"
)

type Handler struct {
	uc            domain.Usecase
	internalToken string
}

func New(r chi.Router, uc domain.Usecase, internalToken string) {
	h := &Handler{uc: uc, internalToken: internalToken}
	r.Post("/notify", h.Notify)
	r.Get("/health", h.Health)
}

func (h *Handler) Notify(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var n domain.Notification
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if n.Type == "" || n.UserID == "" {
		http.Error(w, "type and user_id are required", http.StatusBadRequest)
		return
	}

	if err := h.uc.Notify(r.Context(), n); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "queued"})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (h *Handler) isInternalAuthorized(r *http.Request) bool {
	if h.internalToken == "" {
		return false
	}
	return r.Header.Get("X-Internal-Token") == h.internalToken
}

