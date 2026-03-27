package middleware

import (
	"context"
	"net/http"
)

const PermissionsContextKey contextKey = "permissions"

type PermissionChecker interface {
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
}

type PermissionSet map[string]struct{}

func (ps PermissionSet) Has(perm string) bool {
	_, ok := ps[perm]
	return ok
}

// run after JWTMiddleware in the middleware chain.
func LoadPermissions(checker PermissionChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			perms, err := checker.GetUserPermissions(r.Context(), claims.UserID)
			if err != nil {
				http.Error(w, "failed to load permissions", http.StatusInternalServerError)
				return
			}

			ps := make(PermissionSet, len(perms))
			for _, p := range perms {
				ps[p] = struct{}{}
			}

			ctx := context.WithValue(r.Context(), PermissionsContextKey, ps)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ps := GetPermissionsFromContext(r.Context())
			if ps == nil || !ps.Has(perm) {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetPermissionsFromContext(ctx context.Context) PermissionSet {
	ps, _ := ctx.Value(PermissionsContextKey).(PermissionSet)
	return ps
}
