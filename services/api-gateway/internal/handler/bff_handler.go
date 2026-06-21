package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
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
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		res, err := h.clients.UserBreaker.Execute(func() (interface{}, error) {
			return h.clients.User.GetUser(ctx, &pbUser.GetUserRequest{UserId: userID})
		})
		if err == nil {
			user = res.(*pbUser.GetUserResponse)
		}
	}()

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		res, err := h.clients.MatchBreaker.Execute(func() (interface{}, error) {
			return h.clients.Match.GetStatus(ctx, &pbMatch.GetStatusRequest{UserId: userID})
		})
		if err == nil {
			match = res.(*pbMatch.GetStatusResponse)
		}
	}()

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DashboardResponse{
		User:    user,
		Match:   match,
		Success: true,
	})
}
