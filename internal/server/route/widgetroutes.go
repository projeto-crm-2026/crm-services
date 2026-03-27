package route

import "github.com/go-chi/chi/v5"

func registerWidgetRoutes(r chi.Router, cfg Config) {
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.WidgetAuth)
		r.Use(cfg.Middlewares.ContentJSON)
		r.Use(cfg.RateLimiters.Widget)

		r.Post("/widget/init", cfg.Handlers.Widget.InitWidget)
		r.Post("/widget/chat", cfg.Handlers.Chat.CreateWidgetChat)
		r.Get("/widget/chat/{chatID}/messages", cfg.Handlers.Chat.GetMessages)
	})
}
