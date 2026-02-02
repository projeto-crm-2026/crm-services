package organizationservice

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type OrganizationService interface {
	Create(ctx context.Context, org *entity.Organization) (*entity.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Organization, error)
	Update(ctx context.Context, organization *entity.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
}

type organizationService struct {
	repo   repo.OrganizationRepo
	logger *slog.Logger
}

func NewOrganizationService(repo repo.OrganizationRepo, logger *slog.Logger) OrganizationService {
	return &organizationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *organizationService) Create(ctx context.Context, organization *entity.Organization) (*entity.Organization, error) {
	organization, err := s.repo.Create(ctx, organization)

	if err != nil {
		s.logger.Error("failed to create new organization", "error", err)
		return nil, err
	}

	s.logger.Info("organization created successfully", "id", organization.ID)
	return organization, nil
}

func (s *organizationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	organization := &entity.Organization{}

	organization, err := s.repo.GetByID(ctx, id)

	if err != nil {
		s.logger.Error("no organization exists with this ID", "error", err)
		return nil, err
	}

	return organization, nil
}

func (s *organizationService) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	organization := &entity.Organization{}

	organization, err := s.repo.GetBySlug(ctx, slug)

	if err != nil {
		s.logger.Error("no organization exists with this Slug", "error", err)
		return nil, err
	}

	return organization, nil
}

func (s *organizationService) Update(ctx context.Context, organization *entity.Organization) error {
	err := s.repo.Update(ctx, organization)

	if err != nil {
		s.logger.Error("failed to update organization information", "error", err)
		return err
	}

	s.logger.Info("organization updated successfully", "id", organization.ID)
	return nil
}

func (s *organizationService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)

	if err != nil {
		s.logger.Error("failed to delete organization", "error", err)
		return err
	}

	s.logger.Info("organization soft deleted successfully", "id", id)
	return nil
}

func (s *organizationService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.SoftDelete(ctx, id)

	if err != nil {
		s.logger.Error("failed to soft delete organization", "error", err)
		return err
	}

	s.logger.Info("organization soft deleted successfully")
	return nil
}

func (s *organizationService)	Restore(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Restore(ctx, id)

	if err != nil {
		s.logger.Error("failed to restore organization", "error", err)
		return err
	}

	s.logger.Info("organization restored successfully")
	return nil
}
