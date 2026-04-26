package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathalama/nektokz/matchmaking-service/internal/config"
	"github.com/mathalama/nektokz/matchmaking-service/internal/client"
	delivery "github.com/mathalama/nektokz/matchmaking-service/internal/delivery/http"
	"github.com/mathalama/nektokz/matchmaking-service/internal/repository/redis"
	"github.com/mathalama/nektokz/matchmaking-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	repo := redis.NewInMemoryMatchRepository()
	userCli := client.NewUserClient(cfg.UserServiceURL)
	chatCli := client.NewChatClient(cfg.ChatServiceURL)
	uc := usecase.NewMatchUsecase(repo, userCli, chatCli)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewMatchHandler(r, uc)

	log.Printf("matchmaking-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
