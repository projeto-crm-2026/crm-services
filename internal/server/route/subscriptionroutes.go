package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerSubscriptionRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermSubscriptionsManage)).Post("/subscriptions", cfg.Handlers.Subscription.Subscribe)
	r.With(perm(constant.PermSubscriptionsManage)).Post("/subscriptions/cancel", cfg.Handlers.Subscription.Cancel)
	r.With(perm(constant.PermSubscriptionsManage)).Post("/subscriptions/upgrade", cfg.Handlers.Subscription.Upgrade)
	r.With(perm(constant.PermPaymentsRead)).Get("/payments", cfg.Handlers.Subscription.ListPayments)
}
