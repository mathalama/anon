package main

import (
	"context"
	"log"
	"net"
	"net/http"
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
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
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
			log.Printf("Waiting for database... (%d/10)", i+1)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			log.Fatalf("failed to init postgres pool after retries: %v", err)
		}
		defer pool.Close()

		// Run migrations
		if err := runMigrations(cfg.DBURL); err != nil {
			log.Printf("Migration warning: %v", err)
		}

		repo = postgres.NewPGUserRepository(pool)
	}

	tm := usecase.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	uc := usecase.NewUserUsecase(repo, tm)

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
		log.Printf("user-service HTTP starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen for gRPC: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, grpcDelivery.NewUserHandler(uc))

	go func() {
		log.Printf("user-service gRPC starting on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to start gRPC server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down user-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}
	grpcServer.GracefulStop()
	log.Println("user-service stopped")
}

func validateConfig(cfg *config.Config) {
	if cfg.AppEnv != "development" {
		if cfg.JWTSecret == "" || cfg.JWTSecret == "very-secret-key" {
			log.Fatalf("JWT_SECRET must be set to a strong random value in non-development environments")
		}
		if cfg.InternalToken == "" || cfg.InternalToken == "dev-internal-token" {
			log.Fatalf("INTERNAL_TOKEN must be set to a strong random value in non-development environments")
		}
		if cfg.RepoDriver == "postgres" && strings.Contains(cfg.DBURL, "user:pass@") {
			log.Fatalf("DB_URL must be set (no default credentials) when REPO_DRIVER=postgres in non-development environments")
		}
	}
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
