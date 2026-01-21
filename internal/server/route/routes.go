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
	})

	return r
}
