package route

import "github.com/go-chi/chi/v5"

func registerAuthRoutes(r chi.Router, cfg Config) {
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.RateLimiters.Auth)

		r.Post("/register", cfg.Handlers.User.Register)
		r.Post("/login", cfg.Handlers.User.Login)
		r.Post("/logout", cfg.Handlers.User.Logout)
		r.Post("/invite/accept", cfg.Handlers.User.AcceptInvite)
	})
}
