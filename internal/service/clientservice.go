package clientservice

import (
	"context"
	"log/slog"

	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type ClientService interface {
	GetAll(ctx context.Context) ([]entity.Client, int64, error)
}

type clientService struct {
	repo   repo.ClientRepo
	logger *slog.Logger
}

func NewClientService(repo repo.ClientRepo, logger *slog.Logger) ClientService {
	return &clientService{repo: repo,
		logger: logger}
}

func (s *clientService) GetAll(ctx context.Context) ([]entity.Client, int64, error) {
	clients, total, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to get clients", "error", err)
		return nil, 0, err
	}

	return clients, total, nil
}
