package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerContactRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.Middlewares.JWT)
		r.Use(cfg.Middlewares.LoadPermissions)
		r.Use(cfg.RateLimiters.API)

		r.With(perm(constant.PermContactsCreate)).Post("/contacts", cfg.Handlers.Contact.Create)

		r.With(perm(constant.PermContactsRead)).Get("/contacts", cfg.Handlers.Contact.List)
		r.With(perm(constant.PermContactsRead)).Get("/contacts/search", cfg.Handlers.Contact.Search)
		r.With(perm(constant.PermContactsRead)).Get("/contacts/{id}", cfg.Handlers.Contact.GetByID)
		r.With(perm(constant.PermContactsRead)).Get("/contacts/email/{email}", cfg.Handlers.Contact.GetByEmail)

		r.With(perm(constant.PermContactsUpdate)).Patch("/contacts/{id}", cfg.Handlers.Contact.Update)

		r.With(perm(constant.PermContactsDelete)).Delete("/contacts/{id}", cfg.Handlers.Contact.SoftDelete)
		r.With(perm(constant.PermContactsDelete)).Delete("/contacts/{id}/permanent", cfg.Handlers.Contact.Delete)
	})
}
