package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerPlanRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermPlansRead)).Get("/plans", cfg.Handlers.Plan.ListPlans)
	r.With(perm(constant.PermPlansRead)).Get("/organizations/{id}/usage", cfg.Handlers.Plan.GetUsage)
}
