package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mathalama/nektokz/api-gateway/internal/client"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	cfg     *config.Config
	clients *client.GRPCClients
}

func NewAuthHandler(cfg *config.Config, clients *client.GRPCClients) *AuthHandler {
	return &AuthHandler{cfg: cfg, clients: clients}
}

type AnonymousRequest struct {
	DeviceID string `json:"device_id"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) CreateAnonymous(w http.ResponseWriter, r *http.Request) {
	var req AnonymousRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	base, err := url.Parse(h.cfg.UserServiceURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		log.Error().Err(err).Str("user_service_url", h.cfg.UserServiceURL).Msg("invalid USER_SERVICE_URL")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	base.Path = "/users/anonymous"

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("user-service anonymous request failed")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Warn().Int("status", resp.StatusCode).RawJSON("body", respBody).Msg("user-service anonymous non-2xx")
		http.Error(w, "failed to create session", http.StatusBadGateway)
		return
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil || out.AccessToken == "" || out.RefreshToken == "" {
		log.Error().Err(err).RawJSON("body", respBody).Msg("invalid user-service anonymous response")
		http.Error(w, "failed to create session", http.StatusBadGateway)
		return
	}

	h.setAuthCookies(w, out.AccessToken, out.RefreshToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":      true,
		"access_token": out.AccessToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	base, err := url.Parse(h.cfg.UserServiceURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		log.Error().Err(err).Str("user_service_url", h.cfg.UserServiceURL).Msg("invalid USER_SERVICE_URL")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	base.Path = "/users/refresh"

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("user-service refresh request failed")
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Warn().Int("status", resp.StatusCode).RawJSON("body", respBody).Msg("user-service refresh non-2xx")
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil || out.AccessToken == "" || out.RefreshToken == "" {
		log.Error().Err(err).RawJSON("body", respBody).Msg("invalid user-service refresh response")
		http.Error(w, "service unavailable", http.StatusBadGateway)
		return
	}

	h.setAuthCookies(w, out.AccessToken, out.RefreshToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":      true,
		"access_token": out.AccessToken,
	})
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
