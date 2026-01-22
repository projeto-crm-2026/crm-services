package middleware

import "net/http"

// CORSMiddleware returns an HTTP middleware that sets Cross-Origin Resource Sharing (CORS)
// response headers and handles preflight (OPTIONS) requests.
// 
// The middleware sets Access-Control-Allow-Origin to the request's Origin header when present,
// and unconditionally sets Access-Control-Allow-Methods to "GET, POST, PUT, DELETE, OPTIONS",
// Access-Control-Allow-Headers to "Content-Type, Authorization, X-Widget-Key, X-Visitor-Token, X-Webhook-Token",
// Access-Control-Allow-Credentials to "true", and Access-Control-Max-Age to "86400".
// For OPTIONS requests it writes HTTP 200 OK and does not invoke the next handler; for other
// methods it forwards the request to the next handler in the chain.
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Widget-Key, X-Visitor-Token, X-Webhook-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}