package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mathalama/nektokz/user-service/internal/config"
	delivery "github.com/mathalama/nektokz/user-service/internal/delivery/http"
	grpcDelivery "github.com/mathalama/nektokz/user-service/internal/delivery/grpc"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	"github.com/mathalama/nektokz/user-service/internal/repository/postgres"
	"github.com/mathalama/nektokz/user-service/internal/usecase"
	pb "github.com/mathalama/nektokz/proto/user/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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

	var repo domain.UserRepository = postgres.NewInMemoryUserRepository()
	var pool *pgxpool.Pool
	if cfg.RepoDriver == "postgres" {
		var err error
		// Retry connecting to DB (useful for docker-compose startup)
		for i := 0; i < 10; i++ {
			pool, err = pgxpool.New(ctx, cfg.DBURL)
			if err == nil {
				err = pool.Ping(ctx)
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

		// Run migrations
		if err := runMigrations(cfg.DBURL); err != nil {
			log.Warn().Err(err).Msg("Migration warning")
		}

		repo = postgres.NewPGUserRepository(pool)
	}

	tm := usecase.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	var rdb *goredis.Client
	if cfg.RedisURL != "" {
		if strings.HasPrefix(cfg.RedisURL, "redis://") {
			opt, err := goredis.ParseURL(cfg.RedisURL)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to parse redis url")
			}
			rdb = goredis.NewClient(opt)
		} else {
			rdb = goredis.NewClient(&goredis.Options{Addr: cfg.RedisURL})
		}
		defer rdb.Close()
	}

	uc := usecase.NewUserUsecase(repo, tm, rdb)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	delivery.NewUserHandler(r, uc, cfg.InternalToken)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("user-service HTTP starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start HTTP server")
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen for gRPC")
	}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, grpcDelivery.NewUserHandler(uc))

	go func() {
		log.Info().Str("port", cfg.GRPCPort).Msg("user-service gRPC starting")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to start gRPC server")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down user-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown failed")
	}
	grpcServer.GracefulStop()
	log.Info().Msg("user-service stopped")
}

func validateConfig(cfg *config.Config) {
	if cfg.AppEnv != "development" {
		if cfg.JWTSecret == "" || cfg.JWTSecret == "very-secret-key" {
			log.Fatal().Msg("JWT_SECRET must be set to a strong random value in non-development environments")
		}
		if cfg.InternalToken == "" || cfg.InternalToken == "dev-internal-token" {
			log.Fatal().Msg("INTERNAL_TOKEN must be set to a strong random value in non-development environments")
		}
		if cfg.RepoDriver == "postgres" && strings.Contains(cfg.DBURL, "user:pass@") {
			log.Fatal().Msg("DB_URL must be set (no default credentials) when REPO_DRIVER=postgres in non-development environments")
		}
	}
}

func runMigrations(dbURL string) error {
	// Migrations use lib/pq driver; pgx pool params (pool_*) can break parsing.
	// Keep only the essential query params (at least sslmode) and drop the rest.
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
		// Fallback for non-URL DSNs: preserve sslmode if present, otherwise strip query.
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
