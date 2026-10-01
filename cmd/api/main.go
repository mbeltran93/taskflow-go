// Command api runs the taskflow-go HTTP server: it loads configuration,
// connects to PostgreSQL, applies migrations, wires repositories, services
// and handlers together by hand (constructor functions - no DI framework),
// and serves until it receives a shutdown signal.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	"github.com/mbeltran93/taskflow-go/internal/auth"
	"github.com/mbeltran93/taskflow-go/internal/config"
	"github.com/mbeltran93/taskflow-go/internal/db"
	"github.com/mbeltran93/taskflow-go/internal/handler"
	"github.com/mbeltran93/taskflow-go/internal/repository/postgres"
	"github.com/mbeltran93/taskflow-go/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}

	sqlDB, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	userRepo := postgres.NewUserRepository(sqlDB)
	projectRepo := postgres.NewProjectRepository(sqlDB)
	taskRepo := postgres.NewTaskRepository(sqlDB)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)

	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(userRepo, jwtManager)
	projectSvc := service.NewProjectService(projectRepo, userRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, userRepo)

	router := handler.NewRouter(handler.Handlers{
		Auth:     handler.NewAuthHandler(authSvc),
		Users:    handler.NewUserHandler(userSvc),
		Projects: handler.NewProjectHandler(projectSvc),
		Tasks:    handler.NewTaskHandler(taskSvc),
	}, jwtManager)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
