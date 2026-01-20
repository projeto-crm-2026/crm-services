package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
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

	// configs
	cfg := config.LoadConfigs(logger)

	// database
	dbConn, err := repo.Connect(cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	gormDB := dbConn.GetDB()

	// migrations
	logger.Info("running database migrations...")
	if err := gormDB.AutoMigrate(
		&entity.User{},
		// adicionar as entidades que forem criadas
	); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations completed successfully")

	// repositories
	userRepo := repo.NewUserRepo(gormDB)

	// services
	userSvc := userservice.NewUserService(userRepo, &cfg.JWT, logger)

	//handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userSvc)

	// middlewares
	contentJsonMiddleware := middleware.JsonMiddleware()
	jwtMiddleware := middleware.JWTMiddleware(&cfg.JWT)

	srv := server.NewServer(
		server.WithLogger(logger),
		server.WithConfig(cfg),
		server.WithDB(dbConn),
		server.WithHealthHandler(healthHandler),
		server.WithUserHandler(userHandler),
		server.WithContentJSONMiddleware(contentJsonMiddleware),
		server.WithJWTMiddleware(jwtMiddleware),
	)
	srv.Start(ctx)

}
