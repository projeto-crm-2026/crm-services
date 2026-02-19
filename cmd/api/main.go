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
	"github.com/projeto-crm-2026/crm-services/internal/service/contactservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/organizationservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/webhookservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/widgetservice"
	"github.com/projeto-crm-2026/crm-services/pkg/mailer"
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
		&entity.Contact{},
		&entity.Organization{},
		// adicionar as entidades que forem criadas
	); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := repo.RunCustomMigrations(gormDB); err != nil {
		logger.Error("failed to run custom migrations", "error", err)
		os.Exit(1)
	}

	if err := repo.RunSeeds(gormDB); err != nil {
		logger.Error("failed to run seeds", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations completed successfully")

	// repositories
	userRepo := repo.NewUserRepo(pgxPool)
	chatRepo := repo.NewChatRepo(pgxPool)
	messageRepo := repo.NewMessageRepo(pgxPool)
	apiKeyRepo := repo.NewAPIKeyRepo(pgxPool)
	webhookRepo := repo.NewWebhookRepo(pgxPool)
	contactRepo := repo.NewContactRepo(pgxPool)
	organizationRepo := repo.NewOrganizationRepo(pgxPool)

	// websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// mailer - isso aqui vamos ver dps como vai enviar o email, qual serviço vamos utilzar
	mailClient := mailer.NewSMTPMailer(mailer.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		BaseURL:  cfg.SMTP.BaseURL,
	})

	// services
	userSvc := userservice.NewUserService(userRepo, organizationRepo, &cfg.JWT, mailClient, logger)
	widgetSvc := widgetservice.NewWidgetService(apiKeyRepo, &cfg.JWT, logger)
	chatSvc := chatservice.NewChatService(chatRepo, messageRepo, logger)
	webhookSvc := webhookservice.NewWebhookService(webhookRepo, chatSvc, hub, cfg.Crypto.AESKey, logger)
	contactSvc := contactservice.NewContactService(contactRepo, logger)
	organizationSvc := organizationservice.NewOrganizationService(organizationRepo, logger)

	chatSvc.SetMessageHandler(webhookSvc)

	// handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userSvc)
	chatHandler := handler.NewChatHandler(hub, chatSvc)
	widgetHandler := handler.NewWidgetHandler(widgetSvc)
	webhookHandler := handler.NewWebhookHandler(webhookSvc)
	contactHandler := handler.NewContactHandler(contactSvc)
	organizationHandler := handler.NewOrganizationHandler(organizationSvc)

	// adapters
	widgetAdapter := adapters.NewWidgetValidator(widgetSvc)

	// middlewares
	contentJsonMiddleware := middleware.JsonMiddleware()
	jwtMiddleware := middleware.JWTMiddleware(&cfg.JWT)
	corsMiddleware := middleware.CORSMiddleware()
	widgetAuthMiddleware := middleware.WidgetAuthMiddleware(widgetAdapter)

	//rate limiters
	authRateLimiter := middleware.RateLimitMiddleware(middleware.AuthLimit)
	widgetRateLimiter := middleware.RateLimitMiddleware(middleware.WidgetLimit)
	webhookRateLimiter := middleware.RateLimitMiddleware(middleware.WebhookLimit)
	apiRateLimiter := middleware.RateLimitMiddleware(middleware.APILimit)

	srv := server.NewServer(
		server.WithLogger(logger),
		server.WithConfig(cfg),
		server.WithDB(dbConn),
		server.WithHealthHandler(healthHandler),
		server.WithUserHandler(userHandler),
		server.WithChatHandler(chatHandler),
		server.WithContactHandler(contactHandler),
		server.WithOrganizationHandler(organizationHandler),
		server.WithWidgetHandler(widgetHandler),
		server.WithWebhookHandler(webhookHandler),
		server.WithContentJSONMiddleware(contentJsonMiddleware),
		server.WithJWTMiddleware(jwtMiddleware),
		server.WithCorsMiddleware(corsMiddleware),
		server.WithWidgetAuthMiddleware(widgetAuthMiddleware),
		server.WithAuthRateLimiter(authRateLimiter),
		server.WithWidgetRateLimiter(widgetRateLimiter),
		server.WithWebhookRateLimiter(webhookRateLimiter),
		server.WithAPIRateLimiter(apiRateLimiter),
	)
	srv.Start(ctx)
}
