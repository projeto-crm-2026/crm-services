package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerChatWebSocketRoutes(r chi.Router, cfg Config) {
	// websocket for widgets (no auth)
	r.Get("/ws/widget/{chatID}", cfg.Handlers.Chat.HandleWebSocket)

	// websocket for CRM agents (JWT auth)
	r.Group(func(r chi.Router) {
		r.Use(cfg.Middlewares.JWT)
		r.Get("/ws/chat/{chatID}", cfg.Handlers.Chat.HandleWebSocket)
	})
}

func registerChatRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermChatsRead)).Get("/chats", cfg.Handlers.Chat.ListChats)
	r.With(perm(constant.PermChatsRead)).Get("/chats/{chatID}", cfg.Handlers.Chat.GetChat)
	r.With(perm(constant.PermChatsRead)).Get("/chats/{chatID}/messages", cfg.Handlers.Chat.GetMessages)
}
