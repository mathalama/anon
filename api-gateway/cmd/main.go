package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	gwMiddleware "github.com/mathalama/nektokz/api-gateway/internal/middleware"
	"github.com/mathalama/nektokz/api-gateway/internal/proxy"
)

func main() {
	cfg := config.Load()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.AppEnv == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting stub
	r.Use(gwMiddleware.InMemoryRateLimiter(100))

	// Auth middleware (validation only)
	r.Use(gwMiddleware.Auth(cfg.JWTSecret))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup Proxy
	p := proxy.NewProxy()
	p.AddTarget("/api/v1/users", cfg.UserServiceURL)
	p.AddTarget("/api/v1/match", cfg.MatchmakingServiceURL)
	p.AddTarget("/api/v1/chat", cfg.ChatServiceURL)
	p.AddTarget("/api/v1/report", cfg.ModerationServiceURL)
	p.AddTarget("/ws", cfg.ChatServiceURL)

	// Proxy all requests
	r.HandleFunc("/api/v1/*", p.Handler)
	r.HandleFunc("/ws", p.Handler)

	log.Info().
		Str("port", cfg.Port).
		Str("env", cfg.AppEnv).
		Msg("api-gateway starting...")

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
