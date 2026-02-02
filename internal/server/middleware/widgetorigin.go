package middleware

import (
	"context"
	"net/http"
)

type WidgetContext struct {
	UserID    uint
	PublicKey string
	Domain    string
}

type widgetContextKey string

const WidgetCtxKey widgetContextKey = "widget"

type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, publicKey, origin string) (*WidgetContext, error)
}

func WidgetAuthMiddleware(validator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			publicKey := r.Header.Get("X-Widget-Key")
			if publicKey == "" {
				publicKey = r.URL.Query().Get("widget_key")
			}

			if publicKey == "" {
				http.Error(w, "missing widget key", http.StatusUnauthorized)
				return
			}

			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = r.Header.Get("Referer")
			}

			widgetCtx, err := validator.ValidateAPIKey(r.Context(), publicKey, origin)
			if err != nil {
				http.Error(w, "invalid widget key or origin", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), WidgetCtxKey, widgetCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetWidgetContext(ctx context.Context) (*WidgetContext, bool) {
	wc, ok := ctx.Value(WidgetCtxKey).(*WidgetContext)
	return wc, ok
}
