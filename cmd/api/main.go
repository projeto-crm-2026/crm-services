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
	"github.com/projeto-crm-2026/crm-services/internal/server/adapters"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/websocket"
	"github.com/projeto-crm-2026/crm-services/internal/service/chatservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/webhookservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/widgetservice"
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
	dbConn, err := repo.Connect(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	gormDB := dbConn.GetDB()
	pgxPool := dbConn.GetPool()

	// migrations
	logger.Info("running database migrations...")
	if err := gormDB.AutoMigrate(
		&entity.User{},
		&entity.Chat{},
		&entity.Message{},
		&entity.ChatParticipant{},
		&entity.APIKey{},
		&entity.Webhook{},
		&entity.WebhookLog{},
		&entity.IncomingWebhookToken{},
		// adicionar as entidades que forem criadas
	); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations completed successfully")

	// repositories
	userRepo := repo.NewUserRepo(pgxPool)
	chatRepo := repo.NewChatRepo(pgxPool)
	messageRepo := repo.NewMessageRepo(pgxPool)
	apiKeyRepo := repo.NewAPIKeyRepo(pgxPool)
	webhookRepo := repo.NewWebhookRepo(pgxPool)

	// websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// services
	userSvc := userservice.NewUserService(userRepo, &cfg.JWT, logger)
	widgetSvc := widgetservice.NewWidgetService(apiKeyRepo, &cfg.JWT, logger)
	chatSvc := chatservice.NewChatService(chatRepo, messageRepo, logger)
	webhookSvc := webhookservice.NewWebhookService(webhookRepo, chatSvc, hub, logger)

	chatSvc.SetMessageHandler(webhookSvc)

	// handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userSvc)
	chatHandler := handler.NewChatHandler(hub, chatSvc)
	widgetHandler := handler.NewWidgetHandler(widgetSvc)
	webhookHandler := handler.NewWebhookHandler(webhookSvc)

	// adapters
	widgetAdapter := adapters.NewWidgetValidator(widgetSvc)

	// middlewares
	contentJsonMiddleware := middleware.JsonMiddleware()
	jwtMiddleware := middleware.JWTMiddleware(&cfg.JWT)
	corsMiddleware := middleware.CORSMiddleware()
	widgetAuthMiddleware := middleware.WidgetAuthMiddleware(widgetAdapter)

	srv := server.NewServer(
		server.WithLogger(logger),
		server.WithConfig(cfg),
		server.WithDB(dbConn),
		server.WithHealthHandler(healthHandler),
		server.WithUserHandler(userHandler),
		server.WithChatHandler(chatHandler),
		server.WithWidgetHandler(widgetHandler),
		server.WithWebhookHandler(webhookHandler),
		server.WithContentJSONMiddleware(contentJsonMiddleware),
		server.WithJWTMiddleware(jwtMiddleware),
		server.WithCorsMiddleware(corsMiddleware),
		server.WithWidgetAuthMiddleware(widgetAuthMiddleware),
	)
	srv.Start(ctx)

}
