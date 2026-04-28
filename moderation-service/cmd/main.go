package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
		for i := 0; i < 10; i++ {
			pool, err = pgxpool.New(context.Background(), cfg.DBURL)
			if err == nil {
				err = pool.Ping(context.Background())
				if err == nil {
					break
				}
			}
			log.Printf("Waiting for database... (%d/10)", i+1)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			log.Fatalf("failed to init postgres pool after retries: %v", err)
		}
		defer pool.Close()

		if err := runMigrations(cfg.DBURL); err != nil {
			log.Printf("Migration warning: %v", err)
		}

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

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
