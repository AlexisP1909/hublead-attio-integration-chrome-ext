package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"hublead-attio-integration/backend/internal/attio"
	"hublead-attio-integration/backend/internal/config"
	"hublead-attio-integration/backend/internal/httpapi"
	"hublead-attio-integration/backend/internal/migrations"
	"hublead-attio-integration/backend/internal/repository"
	"hublead-attio-integration/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Up(ctx, db); err != nil {
		logger.Error("run database migrations", "error", err)
		os.Exit(1)
	}

	store := repository.NewPostgresStore(db)
	attioClient := attio.NewClient(cfg.AttioBaseURL, cfg.AttioAPIToken, http.DefaultClient)
	app := service.New(store, attioClient)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewServer(app, store, cfg.AllowedExtensionOrigins, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("backend listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
	}
}
