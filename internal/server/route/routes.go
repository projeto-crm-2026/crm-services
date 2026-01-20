package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
)

func New(
	healthHandler *handler.HealthHandler,
	contentJSONMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler.Health)

	r.Group(func(r chi.Router) {
		r.Use(contentJSONMiddleware)
		// outras rotas
	})

	return r
}
