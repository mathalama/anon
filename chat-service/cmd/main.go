package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathalama/nektokz/chat-service/internal/client"
	"github.com/mathalama/nektokz/chat-service/internal/config"
	delivery "github.com/mathalama/nektokz/chat-service/internal/delivery/http"
	"github.com/mathalama/nektokz/chat-service/internal/delivery/ws"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
	"github.com/mathalama/nektokz/chat-service/internal/repository/postgres"
)

func main() {
	cfg := config.Load()

	hub := ws.NewHub()
	go hub.Run()

	var repo domain.ChatRepository = postgres.NewInMemoryChatRepository()
	var pool *pgxpool.Pool
	if cfg.RepoDriver == "postgres" {
		var err error
		pool, err = pgxpool.New(context.Background(), cfg.DBURL)
		if err != nil {
			log.Fatalf("failed to init postgres pool: %v", err)
		}
		defer pool.Close()
		repo = postgres.NewPGChatRepository(pool)
	}

	modCli := client.NewModerationClient(cfg.ModerationServiceURL, cfg.InternalToken)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewChatHandler(r, hub, repo, modCli, cfg.InternalToken)

	log.Printf("chat-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
