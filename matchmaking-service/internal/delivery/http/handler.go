package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type MatchHandler struct {
	usecase  domain.MatchUsecase
	validate *validator.Validate
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
}

func NewMatchHandler(r chi.Router, usecase domain.MatchUsecase) {
	handler := &MatchHandler{
		usecase:  usecase,
		validate: validator.New(),
		limiters: make(map[string]*rate.Limiter),
	}

	r.Route("/match", func(r chi.Router) {
		r.Post("/search", handler.Search)
		r.Delete("/search", handler.Cancel)
		r.Get("/status", handler.GetStatus)
		r.Get("/status/events", handler.StatusSSE)
		r.Post("/next", handler.Next)
	})
	
	r.Get("/health", handler.Health)
}

func (h *MatchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}

	// Per-user rate limit: 1 search every 5 seconds
	if !h.allowSearch(userID) {
		RespondWithError(w, r, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "please wait 5 seconds before searching again")
		return
	}
	var req struct {
		Filter domain.Filter `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}

	if req.Filter.Gender == "" {
		req.Filter.Gender = "any"
	}

	if err := h.validate.Struct(req.Filter); err != nil {
		log.Warn().Err(err).Interface("filter", req.Filter).Msg("validation failed for search")
		RespondWithError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.Search(r.Context(), userID, req.Filter); err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *MatchHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}
	if err := h.usecase.Cancel(r.Context(), userID); err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MatchHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}
	room, err := h.usecase.GetStatus(r.Context(), userID)
	if err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if room == nil {
		RespondWithJSON(w, http.StatusOK, map[string]string{"status": "searching"})
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "matched",
		"room_id": room.ID,
		"mode":    room.Mode,
	})
}

func (h *MatchHandler) StatusSSE(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "streaming not supported")
		return
	}

	// Check if already matched before subscribing
	if room, _ := h.usecase.GetStatus(r.Context(), userID); room != nil {
		data := map[string]interface{}{
			"status":       "matched",
			"room_id":      room.ID,
			"mode":         room.Mode,
			"is_initiator": room.UserA == userID,
			"partner_gender": "unknown", // Room struct doesn't store this, but search will populate it
		}
		payload, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", payload)
		flusher.Flush()
		
		// Keep alive for a bit to let client process and close
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		return
	}

	ch, cleanup, err := h.usecase.SubscribeToMatch(r.Context(), userID)
	if err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	defer cleanup()

	for {
		select {
		case <-r.Context().Done():
			return
		case match, ok := <-ch:
			if !ok {
				return
			}
			data := map[string]interface{}{
				"status":       "matched",
				"room_id":      match.RoomID,
				"mode":         match.Mode,
				"is_initiator": match.IsInitiator,
				"partner_gender": match.PartnerGender,
			}
			payload, _ := json.Marshal(data)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
			
			// Keep alive for a bit to let client process and close
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Second):
			}
			return // End stream after match found
		}
	}
}

func (h *MatchHandler) Next(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		RespondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user id missing")
		return
	}
	if err := h.usecase.Next(r.Context(), userID); err != nil {
		RespondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *MatchHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.usecase.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("ERROR: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *MatchHandler) allowSearch(userID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	limiter, ok := h.limiters[userID]
	if !ok {
		// Relaxed: 2 requests every 5 seconds (burst of 3)
		limiter = rate.NewLimiter(rate.Every(5*time.Second/2), 3)
		h.limiters[userID] = limiter
	}

	return limiter.Allow()
}
