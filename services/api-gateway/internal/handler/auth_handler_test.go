package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mathalama/nektokz/api-gateway/internal/client"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	"github.com/mathalama/nektokz/api-gateway/internal/handler"
)

func TestAuthHandler_CreateAnonymous(t *testing.T) {
	// Mock user-service
	mockUserService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/anonymous" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"access_token":  "mock-access",
				"refresh_token": "mock-refresh",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockUserService.Close()

	cfg := &config.Config{
		UserServiceURL: mockUserService.URL,
	}
	h := handler.NewAuthHandler(cfg, &client.GRPCClients{})

	reqBody := `{"device_id":"test-device"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/anonymous", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateAnonymous(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["access_token"] != "mock-access" {
		t.Errorf("expected mock-access, got %v", resp["access_token"])
	}

	// Check cookies
	cookies := w.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == "access_token" && c.Value == "mock-access" {
			hasAccess = true
		}
		if c.Name == "refresh_token" && c.Value == "mock-refresh" {
			hasRefresh = true
		}
	}
	if !hasAccess || !hasRefresh {
		t.Errorf("missing or incorrect cookies: %v", cookies)
	}
}
