package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	"github.com/mathalama/nektokz/api-gateway/internal/client"
	pbUser "github.com/mathalama/nektokz/proto/user/v1"
	pbMatch "github.com/mathalama/nektokz/proto/matchmaking/v1"
)

type BFFHandler struct {
	cfg *config.Config
	clients *client.GRPCClients
}

func NewBFFHandler(cfg *config.Config, clients *client.GRPCClients) *BFFHandler {
	return &BFFHandler{cfg: cfg, clients: clients}
}

type DashboardResponse struct {
	User    *pbUser.GetUserResponse `json:"user"`
	Match   *pbMatch.GetStatusResponse `json:"match"`
	Success bool                   `json:"success"`
}

func (h *BFFHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var wg sync.WaitGroup
	var user *pbUser.GetUserResponse
	var match *pbMatch.GetStatusResponse
	wg.Add(2)

	go func() {
		defer wg.Done()
		user, _ = h.clients.User.GetUser(r.Context(), &pbUser.GetUserRequest{UserId: userID})
	}()

	go func() {
		defer wg.Done()
		match, _ = h.clients.Match.GetStatus(r.Context(), &pbMatch.GetStatusRequest{UserId: userID})
	}()

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DashboardResponse{
		User:    user,
		Match:   match,
		Success: true,
	})
}
