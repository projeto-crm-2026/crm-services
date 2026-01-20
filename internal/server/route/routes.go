package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
)

func New(
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	contentJSONMiddleware func(http.Handler) http.Handler,
	jwtMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler.Health)

	r.Group(func(r chi.Router) {
		r.Use(contentJSONMiddleware)
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(contentJSONMiddleware)
		r.Use(jwtMiddleware)

		// rotas que serão protegidas, dps discutimos
	})

	return r
}
