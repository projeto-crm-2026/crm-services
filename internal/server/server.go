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
	db                    *repo.Repo
	healthHandle          *handler.HealthHandler
	contentJSONMiddleware func(http.Handler) http.Handler
	httpSrv               *http.Server
}

type Option func(*Server)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) { s.logger = logger }
}

func WithConfig(cfg *config.Config) Option {
	return func(s *Server) { s.cfg = cfg }
}

func WithDB(db *repo.Repo) Option {
	return func(s *Server) { s.db = db }
}

func WithHealthHandler(h *handler.HealthHandler) Option {
	return func(s *Server) { s.healthHandle = h }
}

func WithContentJSONMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.contentJSONMiddleware = mw
	}
}

func NewServer(opts ...Option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	s.httpSrv = &http.Server{
		Addr:    s.cfg.Server.Address,
		Handler: route.New(s.healthHandle, s.contentJSONMiddleware),
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
