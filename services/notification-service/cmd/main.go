package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mathalama/nektokz/notification-service/internal/channel/noop"
	"github.com/mathalama/nektokz/notification-service/internal/config"
	delivery "github.com/mathalama/nektokz/notification-service/internal/delivery/http"
	"github.com/mathalama/nektokz/notification-service/internal/usecase"
	"github.com/mathalama/nektokz/pkg/mq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.AppEnv == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	}

	validateConfig(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	js, err := mq.NewJetStream(cfg.NATSURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to init NATS JetStream")
	} else {
		defer js.Close()
		worker := usecase.NewNotificationWorker(js)
		go worker.Start(ctx)
	}

	uc := usecase.New(noop.New())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.New(r, uc, cfg.InternalToken)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	log.Info().Str("port", cfg.Port).Msg("notification-service starting")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}

func validateConfig(cfg *config.Config) {
	if cfg.AppEnv != "development" {
		if cfg.InternalToken == "" || cfg.InternalToken == "dev-internal-token" {
			log.Fatal().Msg("INTERNAL_TOKEN must be set to a strong random value in non-development environments")
		}
	}
}
