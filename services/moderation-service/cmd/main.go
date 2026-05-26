package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
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
			log.Info().Msgf("Waiting for database... (%d/10)", i+1)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init postgres pool after retries")
		}
		defer pool.Close()

		if err := runMigrations(cfg.DBURL); err != nil {
			log.Warn().Err(err).Msg("Migration warning")
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

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	log.Info().Str("port", cfg.Port).Msg("moderation-service starting")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}

func validateConfig(cfg *config.Config) {
	if cfg.AppEnv != "development" {
		if cfg.InternalToken == "" || cfg.InternalToken == "dev-internal-token" {
			log.Fatal().Msg("INTERNAL_TOKEN must be set to a strong random value in non-development environments")
		}
		if cfg.RepoDriver == "postgres" && strings.Contains(cfg.DBURL, "user:pass@") {
			log.Fatal().Msg("DB_URL must be set (no default credentials) when REPO_DRIVER=postgres in non-development environments")
		}
	}
}

func runMigrations(dbURL string) error {
	baseAddr := dbURL
	if u, err := url.Parse(dbURL); err == nil && u.Scheme != "" {
		sslmode := u.Query().Get("sslmode")
		u.RawQuery = ""
		if sslmode != "" {
			q := url.Values{}
			q.Set("sslmode", sslmode)
			u.RawQuery = q.Encode()
		}
		baseAddr = u.String()
	} else if idx := strings.Index(baseAddr, "?"); idx != -1 {
		noQuery := baseAddr[:idx]
		ssl := ""
		for _, part := range strings.Split(baseAddr[idx+1:], "&") {
			if strings.HasPrefix(part, "sslmode=") {
				ssl = part
				break
			}
		}
		if ssl != "" {
			baseAddr = noQuery + "?" + ssl
		} else {
			baseAddr = noQuery
		}
	}

	m, err := migrate.New("file://migrations", baseAddr)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
