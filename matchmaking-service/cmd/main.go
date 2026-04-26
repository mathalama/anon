package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathalama/nektokz/matchmaking-service/internal/config"
	"github.com/mathalama/nektokz/matchmaking-service/internal/client"
	delivery "github.com/mathalama/nektokz/matchmaking-service/internal/delivery/http"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
	"github.com/mathalama/nektokz/matchmaking-service/internal/repository/redis"
	"github.com/mathalama/nektokz/matchmaking-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	var repo domain.MatchRepository = redis.NewInMemoryMatchRepository()
	var redisRepo *redis.RedisMatchRepository
	if cfg.RepoDriver == "redis" {
		var err error
		redisRepo, err = redis.NewRedisMatchRepository(cfg.RedisURL)
		if err != nil {
			log.Fatalf("failed to init redis repo: %v", err)
		}
		defer redisRepo.Close()
		repo = redisRepo
	}
	userCli := client.NewUserClient(cfg.UserServiceURL, cfg.InternalToken)
	chatCli := client.NewChatClient(cfg.ChatServiceURL, cfg.InternalToken)
	uc := usecase.NewMatchUsecase(repo, userCli, chatCli, cfg.MatchTimeoutSec, cfg.MatchFilterDropSec)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewMatchHandler(r, uc)

	log.Printf("matchmaking-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
