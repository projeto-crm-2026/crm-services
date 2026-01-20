package repo

import (
	"context"
	"fmt"

	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"gorm.io/gorm"
)

type ClientRepo interface {
	GetAll(ctx context.Context) ([]entity.Client, int64, error)
}

type clientRepo struct {
	db *gorm.DB
}

func NewClientRepo(db *gorm.DB) ClientRepo {
	return &clientRepo{db: db}
}

// só exemplo
func (r *clientRepo) GetAll(ctx context.Context) ([]entity.Client, int64, error) {
	var clients []entity.Client
	var total int64

	if err := r.db.WithContext(ctx).
		Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get clients: %w", err)
	}

	return clients, total, nil
}
