package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathalama/nektokz/user-service/internal/config"
	delivery "github.com/mathalama/nektokz/user-service/internal/delivery/http"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	"github.com/mathalama/nektokz/user-service/internal/repository/postgres"
	"github.com/mathalama/nektokz/user-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	var repo domain.UserRepository = postgres.NewInMemoryUserRepository()
	var pool *pgxpool.Pool
	if cfg.RepoDriver == "postgres" {
		var err error
		pool, err = pgxpool.New(context.Background(), cfg.DBURL)
		if err != nil {
			log.Fatalf("failed to init postgres pool: %v", err)
		}
		defer pool.Close()
		repo = postgres.NewPGUserRepository(pool)
	}

	tm := usecase.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	uc := usecase.NewUserUsecase(repo, tm, cfg.TelegramBotToken)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewUserHandler(r, uc, cfg.InternalToken)

	log.Printf("user-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
