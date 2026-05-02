package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/mathalama/nektokz/api-gateway/docs"
	"github.com/mathalama/nektokz/api-gateway/internal/config"
	gwMiddleware "github.com/mathalama/nektokz/api-gateway/internal/middleware"
	"github.com/mathalama/nektokz/api-gateway/internal/proxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gwMiddleware.Metrics)

	// Security Headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")
			next.ServeHTTP(w, r)
		})
	})

	// CORS
	origins := cfg.AllowedOrigins
	if cfg.AppEnv == "development" {
		origins = append(origins, cfg.DevAllowedOrigins...)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID", "Last-Event-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
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
	p.AddTarget("/api/v1/report", cfg.ModerationServiceURL)
	p.AddTarget("/api/v1/chat", cfg.ChatServiceURL)
	p.AddTarget("/ws", cfg.ChatServiceURL)

	// Health check (gateway level)
	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger & Docs
	r.Get("/api/v1/docs/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.yaml")
	})
	r.Get("/api/v1/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/docs/swagger.yaml"),
	))

	// Proxy all requests starting with /api/v1
	r.Group(func(r chi.Router) {
		r.Use(gwMiddleware.DenyInternal)
		r.HandleFunc("/api/v1/*", p.Handler)
	})

	// WebSocket handler
	r.HandleFunc("/ws", p.Handler)
	r.HandleFunc("/ws/*", p.Handler)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// Fallback for any other requests
	r.NotFound(p.Handler)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // Protect against Slowloris
		ReadTimeout:       0,                // Allow long-lived connections (WS/SSE)
		WriteTimeout:      0,                // Allow long-lived connections (WS/SSE)
		IdleTimeout:       1 * time.Hour,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().
			Str("port", cfg.Port).
			Str("env", cfg.AppEnv).
			Msg("api-gateway starting...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down api-gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}
	log.Info().Msg("api-gateway stopped")
}
