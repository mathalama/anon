package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathalama/nektokz/user-service/internal/config"
	delivery "github.com/mathalama/nektokz/user-service/internal/delivery/http"
	"github.com/mathalama/nektokz/user-service/internal/repository/postgres"
	"github.com/mathalama/nektokz/user-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	repo := postgres.NewInMemoryUserRepository()
	tm := usecase.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	uc := usecase.NewUserUsecase(repo, tm)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewUserHandler(r, uc)

	log.Printf("user-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
