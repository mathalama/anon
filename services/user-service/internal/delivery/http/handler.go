package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mathalama/nektokz/user-service/internal/domain"
)

type UserHandler struct {
	usecase domain.UserUsecase
	internalToken string
}

func NewUserHandler(r chi.Router, usecase domain.UserUsecase, internalToken string) {
	handler := &UserHandler{
		usecase:        usecase,
		internalToken:  internalToken,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/anonymous", handler.CreateAnonymous)
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/refresh", handler.Refresh)
		
		r.Group(func(r chi.Router) {
			// Middleware for JWT would go here
			r.Get("/me", handler.GetMe)
			r.Put("/me", handler.UpdateMe)
		})

		r.Get("/{id}/status", handler.GetBanStatus)
	})

	r.Post("/internal/ban", handler.BanUser)
	r.Get("/health", handler.Health)
}

func (h *UserHandler) CreateAnonymous(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	access, refresh, err := h.usecase.CreateAnonymous(r.Context(), req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.usecase.Register(r.Context(), req.Email, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	access, refresh, err := h.usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}


func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	access, refresh, err := h.usecase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// In a real app, userID would come from JWT middleware
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.usecase.GetMe(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Gender    string   `json:"gender"`
		Interests []string `json:"interests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.usecase.UpdateMe(r.Context(), userID, req.Gender, req.Interests); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) GetBanStatus(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	ban, active, err := h.usecase.GetBanStatus(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_banned": active,
		"ban":       ban,
	})
}

func (h *UserHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	if !h.isInternalAuthorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		UserID   string `json:"user_id"`
		Reason   string `json:"reason"`
		BannedBy string `json:"banned_by"`
		Hours    int    `json:"hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration := time.Duration(req.Hours) * time.Hour
	if err := h.usecase.BanUser(r.Context(), req.UserID, req.Reason, req.BannedBy, duration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *UserHandler) isInternalAuthorized(r *http.Request) bool {
	if h.internalToken == "" {
		return false
	}
	return r.Header.Get("X-Internal-Token") == h.internalToken
}
