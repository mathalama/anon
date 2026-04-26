package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathalama/nektokz/moderation-service/internal/client"
	"github.com/mathalama/nektokz/moderation-service/internal/config"
	delivery "github.com/mathalama/nektokz/moderation-service/internal/delivery/http"
	"github.com/mathalama/nektokz/moderation-service/internal/domain"
	"github.com/mathalama/nektokz/moderation-service/internal/repository/memory"
	"github.com/mathalama/nektokz/moderation-service/internal/repository/postgres"
	"github.com/mathalama/nektokz/moderation-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	var repo domain.ReportRepository = memory.New()
	var pool *pgxpool.Pool
	if cfg.RepoDriver == "postgres" {
		var err error
		pool, err = pgxpool.New(context.Background(), cfg.DBURL)
		if err != nil {
			log.Fatalf("failed to init postgres pool: %v", err)
		}
		defer pool.Close()
		repo = postgres.NewPGReportRepository(pool)
	}

	userCli := client.NewUserClient(cfg.UserServiceURL, cfg.InternalToken)
	chatCli := client.NewChatClient(cfg.ChatServiceURL, cfg.InternalToken)
	notifCli := client.NewNotificationClient(cfg.NotificationServiceURL, cfg.InternalToken)
	uc := usecase.New(repo, userCli, chatCli, notifCli, cfg.ToxicWords)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.New(r, uc, cfg.InternalToken)

	log.Printf("moderation-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
