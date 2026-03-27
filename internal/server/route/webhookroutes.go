package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerWebhookPublicRoutes(r chi.Router, cfg Config) {
	r.With(cfg.RateLimiters.Webhook).
		Post("/webhook/incoming", cfg.Handlers.Webhook.HandleIncomingWebhook)

	r.Post("/webhook/mercadopago", cfg.Handlers.Subscription.HandleWebhook)
}

func registerWebhookProtectedRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	// outgoing webhooks
	r.With(perm(constant.PermWebhooksRead)).Get("/webhooks/events", cfg.Handlers.Webhook.GetAvailableEvents)
	r.With(perm(constant.PermWebhooksCreate)).Post("/webhooks", cfg.Handlers.Webhook.CreateWebhook)
	r.With(perm(constant.PermWebhooksRead)).Get("/webhooks", cfg.Handlers.Webhook.ListWebhooks)
	r.With(perm(constant.PermWebhooksRead)).Get("/webhooks/{webhookID}", cfg.Handlers.Webhook.GetWebhook)
	r.With(perm(constant.PermWebhooksUpdate)).Put("/webhooks/{webhookID}", cfg.Handlers.Webhook.UpdateWebhook)
	r.With(perm(constant.PermWebhooksDelete)).Delete("/webhooks/{webhookID}", cfg.Handlers.Webhook.DeleteWebhook)
	r.With(perm(constant.PermWebhooksRead)).Get("/webhooks/{webhookID}/logs", cfg.Handlers.Webhook.GetWebhookLogs)

	// incoming webhook tokens
	r.With(perm(constant.PermWebhooksTokensCreate)).Post("/webhooks/tokens", cfg.Handlers.Webhook.CreateIncomingToken)
	r.With(perm(constant.PermWebhooksTokensRead)).Get("/webhooks/tokens", cfg.Handlers.Webhook.ListIncomingTokens)
	r.With(perm(constant.PermWebhooksTokensDelete)).Delete("/webhooks/tokens/{tokenID}", cfg.Handlers.Webhook.DeleteIncomingToken)
}
