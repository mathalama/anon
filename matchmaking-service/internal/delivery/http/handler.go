package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
)

type MatchHandler struct {
	usecase domain.MatchUsecase
}

func NewMatchHandler(r chi.Router, usecase domain.MatchUsecase) {
	handler := &MatchHandler{usecase: usecase}

	r.Route("/match", func(r chi.Router) {
		r.Post("/search", handler.Search)
		r.Delete("/search", handler.Cancel)
		r.Get("/status", handler.GetStatus)
		r.Post("/next", handler.Next)
	})
	
	r.Get("/health", handler.Health)
}

func (h *MatchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Filter domain.Filter `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.usecase.Search(r.Context(), userID, req.Filter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *MatchHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if err := h.usecase.Cancel(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MatchHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	room, err := h.usecase.GetStatus(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if room == nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "searching"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "matched",
		"room_id": room.ID,
	})
}

func (h *MatchHandler) Next(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if err := h.usecase.Next(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *MatchHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
