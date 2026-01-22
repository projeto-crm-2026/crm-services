package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server/handler"
	"github.com/projeto-crm-2026/crm-services/internal/server/route"
)

type Server struct {
	logger                *slog.Logger
	cfg                   *config.Config
	db                    *repo.Conn
	healthHandler         *handler.HealthHandler
	userHandler           *handler.UserHandler
	chatHandler           *handler.ChatHandler
	widgetHandler         *handler.WidgetHandler
	webhookHandler        *handler.WebhookHandler
	contentJSONMiddleware func(http.Handler) http.Handler
	jwtMiddleware         func(http.Handler) http.Handler
	corsMiddleware        func(http.Handler) http.Handler
	widgetAuthMiddleware  func(http.Handler) http.Handler
	httpSrv               *http.Server
}

type Option func(*Server)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) { s.logger = logger }
}

func WithConfig(cfg *config.Config) Option {
	return func(s *Server) { s.cfg = cfg }
}

func WithDB(db *repo.Conn) Option {
	return func(s *Server) { s.db = db }
}

func WithHealthHandler(h *handler.HealthHandler) Option {
	return func(s *Server) { s.healthHandler = h }
}

func WithUserHandler(h *handler.UserHandler) Option {
	return func(s *Server) { s.userHandler = h }
}

func WithChatHandler(h *handler.ChatHandler) Option {
	return func(s *Server) { s.chatHandler = h }
}

// WithWidgetHandler returns an Option that sets the Server's widgetHandler to h.
func WithWidgetHandler(h *handler.WidgetHandler) Option {
	return func(s *Server) { s.widgetHandler = h }
}

// WithWebhookHandler returns an Option that sets the Server's webhookHandler to the provided handler.
func WithWebhookHandler(h *handler.WebhookHandler) Option {
	return func(s *Server) { s.webhookHandler = h }
}

// WithContentJSONMiddleware returns an Option that sets the server's content-JSON middleware.
// The mw parameter is an HTTP middleware that wraps handlers to enforce or set JSON content headers for requests and responses.
func WithContentJSONMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.contentJSONMiddleware = mw
	}
}

func WithJWTMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.jwtMiddleware = mw
	}
}

func WithCorsMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.corsMiddleware = mw
	}
}

func WithWidgetAuthMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.widgetAuthMiddleware = mw
	}
}

// NewServer constructs a Server configured by the provided options.
// It applies each Option to the Server and initializes the bundled http.Server,
// setting its address from the server configuration and creating the HTTP handler
// via the router using the configured handlers and middleware.
// The configured Server is returned.
func NewServer(opts ...Option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	s.httpSrv = &http.Server{
		Addr:    s.cfg.Server.Address,
		Handler: route.New(s.healthHandler, s.userHandler, s.chatHandler, s.widgetHandler, s.webhookHandler, s.contentJSONMiddleware, s.jwtMiddleware, s.corsMiddleware, s.widgetAuthMiddleware),
	}

	return s
}

func (s *Server) Start(ctx context.Context) {
	go func() {
		s.logger.Info("Starting server", "address", s.cfg.Server.Address)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Could not start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down server...")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		s.logger.Error("Could not gracefully shutdown the server", "error", err)
		os.Exit(1)
	}

	s.logger.Info("Server stopped gracefully")
}