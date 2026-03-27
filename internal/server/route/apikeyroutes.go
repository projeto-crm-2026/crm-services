package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerAPIKeyRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermApiKeysCreate)).Post("/api-keys", cfg.Handlers.Widget.CreateAPIKey)
	r.With(perm(constant.PermApiKeysRead)).Get("/api-keys", cfg.Handlers.Widget.ListAPIKeys)
	r.With(perm(constant.PermApiKeysDelete)).Delete("/api-keys/{keyID}", cfg.Handlers.Widget.DeleteAPIKey)
}
