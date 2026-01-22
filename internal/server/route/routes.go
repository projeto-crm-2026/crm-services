package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
)

func New(
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	chatHandler *handler.ChatHandler,
	widgetHandler *handler.WidgetHandler,
	webhookHandler *handler.WebhookHandler,
	contentJSONMiddleware func(http.Handler) http.Handler,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	widgetAuthMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(corsMiddleware)
	r.Get("/health", healthHandler.Health)

	r.Group(func(r chi.Router) {
		r.Use(contentJSONMiddleware)
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
		r.Post("/logout", userHandler.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(widgetAuthMiddleware)
		r.Use(contentJSONMiddleware)

		r.Post("/widget/init", widgetHandler.InitWidget)
		r.Post("/widget/chat", chatHandler.CreateWidgetChat)
		r.Get("/widget/chat/{chatID}/messages", chatHandler.GetMessages)
	})

	// webSocket for authenticated CRM agents and widgets
	r.Get("/ws/chat/{chatID}", chatHandler.HandleWebSocket)
	r.Get("/ws/widget/{chatID}", chatHandler.HandleWebSocket)

	// incoming webhooks
	r.Post("/webhook/incoming", webhookHandler.HandleIncomingWebhook)

	r.Group(func(r chi.Router) {
		r.Use(contentJSONMiddleware)
		r.Use(jwtMiddleware)

		// API Key
		r.Post("/api-keys", widgetHandler.CreateAPIKey)
		r.Get("/api-keys", widgetHandler.ListAPIKeys)
		r.Delete("/api-keys/{keyID}", widgetHandler.DeleteAPIKey)

		// CRM agent
		r.Get("/chats", chatHandler.ListChats)
		r.Get("/chats/{chatID}", chatHandler.GetChat)
		r.Get("/chats/{chatID}/messages", chatHandler.GetMessages)

		// outgoing webhooks
		r.Get("/webhooks/events", webhookHandler.GetAvailableEvents)
		r.Post("/webhooks", webhookHandler.CreateWebhook)
		r.Get("/webhooks", webhookHandler.ListWebhooks)
		r.Get("/webhooks/{webhookID}", webhookHandler.GetWebhook)
		r.Put("/webhooks/{webhookID}", webhookHandler.UpdateWebhook)
		r.Delete("/webhooks/{webhookID}", webhookHandler.DeleteWebhook)
		r.Get("/webhooks/{webhookID}/logs", webhookHandler.GetWebhookLogs)

		// incoming webhook tokens
		r.Post("/webhooks/tokens", webhookHandler.CreateIncomingToken)
		r.Get("/webhooks/tokens", webhookHandler.ListIncomingTokens)
		r.Delete("/webhooks/tokens/{tokenID}", webhookHandler.DeleteIncomingToken)
	})

	return r
}
