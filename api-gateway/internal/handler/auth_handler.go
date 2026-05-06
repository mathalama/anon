package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	"github.com/mathalama/nektokz/api-gateway/internal/client"
	pbUser "github.com/mathalama/nektokz/proto/user/v1"
	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

type AuthHandler struct {
	cfg     *config.Config
	clients *client.GRPCClients
}

func NewAuthHandler(cfg *config.Config, clients *client.GRPCClients) *AuthHandler {
	return &AuthHandler{cfg: cfg, clients: clients}
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	result, err := h.clients.UserBreaker.Execute(func() (interface{}, error) {
		return h.clients.User.Login(r.Context(), &pbUser.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			log.Warn().Msg("user-service breaker is OPEN")
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Error().Err(err).Msg("gRPC login failed")
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := result.(*pbUser.AuthResponse)
	h.setAuthCookies(w, resp.AccessToken, resp.RefreshToken)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"access_token": resp.AccessToken,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.clients.UserBreaker.Execute(func() (interface{}, error) {
		return h.clients.User.Register(r.Context(), &pbUser.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		})
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Error().Err(err).Msg("gRPC registration failed")
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, access, refresh string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.AppEnv != "development",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.AppEnv != "development",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}
