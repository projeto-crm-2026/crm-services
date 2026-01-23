package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterType string

const (
	AuthLimit    RateLimiterType = "auth"
	WebhookLimit RateLimiterType = "webhook"
	WidgetLimit  RateLimiterType = "widget"
	APILimit     RateLimiterType = "api"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterPool struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiterPool(r rate.Limit, b int) *RateLimiterPool {
	pool := &RateLimiterPool{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}

	go pool.cleanupVisitors()

	return pool
}

func (p *RateLimiterPool) getLimiter(key string) *rate.Limiter {
	p.mu.Lock()
	defer p.mu.Unlock()

	v, exists := p.visitors[key]
	if !exists {
		limiter := rate.NewLimiter(p.rate, p.burst)
		p.visitors[key] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (p *RateLimiterPool) cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)
		p.mu.Lock()
		for key, v := range p.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(p.visitors, key)
			}
		}
		p.mu.Unlock()
	}
}

type KeyExtractor func(*http.Request) string

type RateLimiterStrategy struct {
	Pool         *RateLimiterPool
	KeyExtractor KeyExtractor
}

var (
	// auth: 5 req/min por IP
	authPool = NewRateLimiterPool(rate.Every(12*time.Second), 5)

	// webhook: 60 req/min por token
	webhookPool = NewRateLimiterPool(rate.Every(1*time.Second), 60)

	// widget: 60 req/min por chave de API
	widgetPool = NewRateLimiterPool(rate.Every(1*time.Second), 60)

	// api: 100 req/min por user
	apiPool = NewRateLimiterPool(rate.Every(600*time.Millisecond), 100)
)

var (
	extractIPKey = func(r *http.Request) string {
		return getIP(r)
	}

	extractWebhookKey = func(r *http.Request) string {
		if token := r.Header.Get("X-Webhook-Token"); token != "" {
			return token
		}
		return getIP(r)
	}

	extractWidgetKey = func(r *http.Request) string {
		if key := r.Header.Get("X-Widget-Key"); key != "" {
			return key
		}
		return getIP(r)
	}

	extractAPIKey = func(r *http.Request) string {
		if claims, ok := GetUserFromContext(r.Context()); ok {
			return fmt.Sprintf("user:%d", claims.UserID)
		}
		return getIP(r)
	}
)

var rateLimiterStrategies = map[RateLimiterType]*RateLimiterStrategy{
	AuthLimit: {
		Pool:         authPool,
		KeyExtractor: extractIPKey,
	},
	WebhookLimit: {
		Pool:         webhookPool,
		KeyExtractor: extractWebhookKey,
	},
	WidgetLimit: {
		Pool:         widgetPool,
		KeyExtractor: extractWidgetKey,
	},
	APILimit: {
		Pool:         apiPool,
		KeyExtractor: extractAPIKey,
	},
}

func RateLimitMiddleware(limiterType RateLimiterType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			strategy, exists := rateLimiterStrategies[limiterType]
			if !exists {
				// fallback para api limit pq nao achou nenhum
				strategy = rateLimiterStrategies[APILimit]
			}

			key := strategy.KeyExtractor(r)
			limiter := strategy.Pool.getLimiter(key)

			if !limiter.Allow() {
				http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
