package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerOrganizationRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.Middlewares.JWT)
		r.Use(cfg.Middlewares.LoadPermissions)
		r.Use(cfg.RateLimiters.API)

		r.With(perm(constant.PermOrganizationsManage)).Post("/organizations", cfg.Handlers.Organization.Create)
		r.With(perm(constant.PermOrganizationsManage)).Post("/organizations/{id}/restore", cfg.Handlers.Organization.Restore)

		r.With(perm(constant.PermOrganizationsRead)).Get("/organizations/{id}", cfg.Handlers.Organization.GetByID)
		r.With(perm(constant.PermOrganizationsRead)).Get("/organizations/slug/{slug}", cfg.Handlers.Organization.GetBySlug)

		r.With(perm(constant.PermOrganizationsUpdate)).Patch("/organizations/{id}", cfg.Handlers.Organization.Update)

		r.With(perm(constant.PermOrganizationsDelete)).Delete("/organizations/{id}", cfg.Handlers.Organization.SoftDelete)
		r.With(perm(constant.PermOrganizationsDelete)).Delete("/organizations/{id}/permanent", cfg.Handlers.Organization.Delete)
	})
}
