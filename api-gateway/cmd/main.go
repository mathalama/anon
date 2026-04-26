package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/mathalama/nektokz/api-gateway/docs"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	gwMiddleware "github.com/mathalama/nektokz/api-gateway/internal/middleware"
	"github.com/mathalama/nektokz/api-gateway/internal/proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
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
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Auth middleware (validation only)
	r.Use(gwMiddleware.Auth(cfg.JWTSecret))

	// Rate limiting (Redis): 600 req/min anon, 1000 req/min auth
	rl, err := gwMiddleware.NewRedisRateLimiter(cfg.RedisURL, 600, 1000)
	if err != nil {
		log.Warn().Err(err).Msg("failed to init redis rate limiter, falling back to in-memory")
		r.Use(gwMiddleware.InMemoryRateLimiter(100))
	} else {
		defer rl.Close()
		r.Use(rl.Middleware())
	}

	// Setup Proxy
	p := proxy.NewProxy()
	p.AddTarget("/api/v1/users", cfg.UserServiceURL)
	p.AddTarget("/api/v1/match", cfg.MatchmakingServiceURL)
	p.AddTarget("/api/v1/chat", cfg.ChatServiceURL)
	p.AddTarget("/api/v1/report", cfg.ModerationServiceURL)
	p.AddTarget("/ws", cfg.ChatServiceURL)

	// Proxy all requests
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(gwMiddleware.DenyInternal)

		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Swagger
		r.Get("/docs/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.yaml")
		})
		r.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/docs/swagger.yaml"),
		))

		r.HandleFunc("/*", p.Handler)
	})
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
