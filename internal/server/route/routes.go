package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func New(cfg Config) http.Handler {
	r := chi.NewRouter()

	r.Use(cfg.Middlewares.CORS)
	r.Get("/health", cfg.Handlers.Health.Health)

	// auth
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.RateLimiters.Auth)

		r.Post("/register", cfg.Handlers.User.Register)
		r.Post("/login", cfg.Handlers.User.Login)
		r.Post("/logout", cfg.Handlers.User.Logout)
		r.Post("/invite/accept", cfg.Handlers.User.AcceptInvite)
	})

	// widget
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.WidgetAuth)
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.RateLimiters.Widget)

		r.Post("/widget/init", cfg.Handlers.Widget.InitWidget)
		r.Post("/widget/chat", cfg.Handlers.Chat.CreateWidgetChat)
		r.Get("/widget/chat/{chatID}/messages", cfg.Handlers.Chat.GetMessages)
	})

	// webSocket for authenticated CRM agents and widgets
	r.Get("/ws/chat/{chatID}", cfg.Handlers.Chat.HandleWebSocket)
	r.Get("/ws/widget/{chatID}", cfg.Handlers.Chat.HandleWebSocket)

	// incoming webhook
	r.With(cfg.RateLimiters.Webhook).
		Post("/webhook/incoming", cfg.Handlers.Webhook.HandleIncomingWebhook)

	// protected
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.Middlewares.JWT)
		r.Use(cfg.RateLimiters.API)

		r.Post("/members/invite", cfg.Handlers.User.InviteUser)
		r.Get("/members", cfg.Handlers.User.ListMembers)

		// API Key
		r.Post("/api-keys", cfg.Handlers.Widget.CreateAPIKey)
		r.Get("/api-keys", cfg.Handlers.Widget.ListAPIKeys)
		r.Delete("/api-keys/{keyID}", cfg.Handlers.Widget.DeleteAPIKey)

		// CRM agent
		r.Get("/chats", cfg.Handlers.Chat.ListChats)
		r.Get("/chats/{chatID}", cfg.Handlers.Chat.GetChat)
		r.Get("/chats/{chatID}/messages", cfg.Handlers.Chat.GetMessages)

		// outgoing webhooks
		r.Get("/webhooks/events", cfg.Handlers.Webhook.GetAvailableEvents)
		r.Post("/webhooks", cfg.Handlers.Webhook.CreateWebhook)
		r.Get("/webhooks", cfg.Handlers.Webhook.ListWebhooks)
		r.Get("/webhooks/{webhookID}", cfg.Handlers.Webhook.GetWebhook)
		r.Put("/webhooks/{webhookID}", cfg.Handlers.Webhook.UpdateWebhook)
		r.Delete("/webhooks/{webhookID}", cfg.Handlers.Webhook.DeleteWebhook)
		r.Get("/webhooks/{webhookID}/logs", cfg.Handlers.Webhook.GetWebhookLogs)

		// incoming webhook tokens
		r.Post("/webhooks/tokens", cfg.Handlers.Webhook.CreateIncomingToken)
		r.Get("/webhooks/tokens", cfg.Handlers.Webhook.ListIncomingTokens)
		r.Delete("/webhooks/tokens/{tokenID}", cfg.Handlers.Webhook.DeleteIncomingToken)
	})

	// contacts
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.Middlewares.JWT)
		r.Use(cfg.RateLimiters.API)

		r.Post("/contacts", cfg.Handlers.Contact.Create)

		r.Get("/contacts", cfg.Handlers.Contact.List)
		r.Get("/contacts/search", cfg.Handlers.Contact.Search)
		r.Get("/contacts/{id}", cfg.Handlers.Contact.GetByID)
		r.Get("/contacts/email/{email}", cfg.Handlers.Contact.GetByEmail)

		r.Patch("/contacts/{id}", cfg.Handlers.Contact.Update)

		r.Delete("/contacts/{id}", cfg.Handlers.Contact.SoftDelete)
		r.Delete("/contacts/{id}/permanent", cfg.Handlers.Contact.Delete)
	})

	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.Middlewares.JWT)
		r.Use(cfg.RateLimiters.API)

		r.Post("/organizations", cfg.Handlers.Organization.Create)
		r.Post("/organizations/{id}/restore", cfg.Handlers.Organization.Restore)

		r.Get("/organizations/{id}", cfg.Handlers.Organization.GetByID)
		r.Get("/organizations/slug/{slug}", cfg.Handlers.Organization.GetBySlug)

		r.Patch("/organizations/{id}", cfg.Handlers.Organization.Update)

		r.Delete("/organizations/{id}", cfg.Handlers.Organization.SoftDelete)
		r.Delete("/organizations/{id}/permanent", cfg.Handlers.Organization.Delete)
	})

	return r
}
