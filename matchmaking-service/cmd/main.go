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
	"github.com/mathalama/nektokz/matchmaking-service/internal/client"
	"github.com/mathalama/nektokz/matchmaking-service/internal/config"
	delivery "github.com/mathalama/nektokz/matchmaking-service/internal/delivery/http"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
	"github.com/mathalama/nektokz/matchmaking-service/internal/repository/redis"
	"github.com/mathalama/nektokz/matchmaking-service/internal/usecase"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.Load()
	validateConfig(cfg)

	var repo domain.MatchRepository = redis.NewInMemoryMatchRepository()
	var redisRepo *redis.RedisMatchRepository
	if cfg.RepoDriver == "redis" {
		var err error
		redisRepo, err = redis.NewRedisMatchRepository(cfg.RedisURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init redis repo")
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

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go uc.StartWorker(ctx)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("matchmaking-service starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down matchmaking-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}
	log.Info().Msg("matchmaking-service stopped")
}

func validateConfig(cfg *config.Config) {
	if cfg.AppEnv != "development" {
		if cfg.InternalToken == "" || cfg.InternalToken == "dev-internal-token" {
			log.Fatal().Msg("INTERNAL_TOKEN must be set to a strong random value in non-development environments")
		}
	}
}
