package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChan
		logger.Info("received signal, initiating graceful shutdown", "signal", sig)
		cancel()
	}()

	cfg := config.LoadConfigs(logger)

	db, err := repo.Connect(cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	healthHandler := handler.NewHealthHandler()

	contentJsonMiddleware := middleware.JsonMiddleware()

	srv := server.NewServer(
		server.WithLogger(logger),
		server.WithConfig(cfg),
		server.WithDB(&db),
		server.WithHealthHandler(healthHandler),
		server.WithContentJSONMiddleware(contentJsonMiddleware),
	)
	srv.Start(ctx)

}
