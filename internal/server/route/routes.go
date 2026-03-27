package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func New(cfg Config) http.Handler {
	r := chi.NewRouter()
	perm := cfg.RequirePermission

	r.Use(cfg.Middlewares.CORS)
	r.Get("/health", cfg.Handlers.Health.Health)

	r.Route("/v1", func(r chi.Router) {
		registerAuthRoutes(r, cfg)
		registerWidgetRoutes(r, cfg)
		registerWebhookPublicRoutes(r, cfg)
		registerChatWebSocketRoutes(r, cfg)
		registerContactRoutes(r, cfg, perm)
		registerOrganizationRoutes(r, cfg, perm)

		r.Group(func(r chi.Router) {
			r.Use(cfg.Middlewares.ContentJSON)
			r.Use(cfg.Middlewares.JWT)
			r.Use(cfg.Middlewares.LoadPermissions)
			r.Use(cfg.RateLimiters.API)

			registerMemberRoutes(r, cfg, perm)
			registerAPIKeyRoutes(r, cfg, perm)
			registerChatRoutes(r, cfg, perm)
			registerWebhookProtectedRoutes(r, cfg, perm)
			registerPlanRoutes(r, cfg, perm)
			registerSubscriptionRoutes(r, cfg, perm)
			registerRoleRoutes(r, cfg, perm)
		})
	})

	return r
}
