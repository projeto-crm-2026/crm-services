package route

import (
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
)

type Handlers struct {
	Health  *handler.HealthHandler
	User    *handler.UserHandler
	Chat    *handler.ChatHandler
	Widget  *handler.WidgetHandler
	Webhook *handler.WebhookHandler
}

type Middlewares struct {
	ContentJSON func(http.Handler) http.Handler
	JWT         func(http.Handler) http.Handler
	CORS        func(http.Handler) http.Handler
	WidgetAuth  func(http.Handler) http.Handler
}

type RateLimiters struct {
	Auth    func(http.Handler) http.Handler
	Widget  func(http.Handler) http.Handler
	Webhook func(http.Handler) http.Handler
	API     func(http.Handler) http.Handler
}

type Config struct {
	Handlers     Handlers
	Middlewares  Middlewares
	RateLimiters RateLimiters
}
