package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathalama/nektokz/notification-service/internal/channel/noop"
	"github.com/mathalama/nektokz/notification-service/internal/config"
	delivery "github.com/mathalama/nektokz/notification-service/internal/delivery/http"
	"github.com/mathalama/nektokz/notification-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	uc := usecase.New(noop.New())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.New(r, uc, cfg.InternalToken)

	log.Printf("notification-service starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
