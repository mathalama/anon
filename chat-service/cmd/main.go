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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/mathalama/nektokz/chat-service/internal/client"
	"github.com/mathalama/nektokz/chat-service/internal/config"
	delivery "github.com/mathalama/nektokz/chat-service/internal/delivery/http"
	"github.com/mathalama/nektokz/chat-service/internal/delivery/ws"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
	"github.com/mathalama/nektokz/chat-service/internal/repository/postgres"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.Load()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: cfg.RedisURL,
	})
	defer rdb.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := ws.NewHub(rdb)
	go hub.Run(ctx)

	var repo domain.ChatRepository = postgres.NewInMemoryChatRepository()
	var pool *pgxpool.Pool
	if cfg.RepoDriver == "postgres" {
		var err error
		for i := 0; i < 10; i++ {
			pool, err = pgxpool.New(ctx, cfg.DBURL)
			if err == nil {
				err = pool.Ping(ctx)
				if err == nil {
					break
				}
			}
			log.Info().Int("attempt", i+1).Msg("Waiting for database...")
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init postgres pool after retries")
		}
		defer pool.Close()

		if err := runMigrations(cfg.DBURL); err != nil {
			log.Warn().Err(err).Msg("Migration warning")
		}

		repo = postgres.NewPGChatRepository(pool)
	}

	modCli := client.NewModerationClient(cfg.ModerationServiceURL, cfg.InternalToken)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewChatHandler(r, hub, repo, modCli, cfg.InternalToken)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("chat-service starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down chat-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}
	log.Info().Msg("chat-service stopped")
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
