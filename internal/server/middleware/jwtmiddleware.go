package middleware

import (
	"context"
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/pkg/jwt"
)

type contextKey string

const UserContextKey contextKey = "user"

func JWTMiddleware(jwtConfig *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateToken(cookie.Value, jwtConfig.JWTSecret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	return claims, ok
}
