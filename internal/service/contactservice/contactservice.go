package contactservice

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type ContactService interface {
	Create(ctx context.Context, contact *entity.Contact) (*entity.Contact, error)
	GetByID(ctx context.Context, id uuid.UUID, organization_id uuid.UUID) (*entity.Contact, error)
	GetByEmail(ctx context.Context, email string, organization_id uuid.UUID) (*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, filters repo.ContactFilters) ([]*entity.Contact, error)
	ListPaginated(ctx context.Context, filters repo.ContactFilters, page, pageSize int) (*repo.PaginatedResult[entity.Contact], error)

	Search(ctx context.Context, query string, filters repo.ContactFilters, organization_id uuid.UUID) ([]*entity.Contact, error)
}

type contactService struct {
	repo   repo.ContactRepo
	logger *slog.Logger
}

func NewContactService(repo repo.ContactRepo, logger *slog.Logger) ContactService {
	return &contactService{
		repo:   repo,
		logger: logger,
	}
}

func (s *contactService) Create(ctx context.Context, contact *entity.Contact) (*entity.Contact, error) {
	contact, err := s.repo.Create(ctx, contact)

	if err != nil {
		s.logger.Error("failed to create new contact", "error", err)
		return nil, err
	}

	s.logger.Info("contact created successfully", "id", contact.ID)
	return contact, nil
}

func (s *contactService) GetByID(ctx context.Context, id uuid.UUID, organization_id uuid.UUID) (*entity.Contact, error) {
	contact := &entity.Contact{}

	contact, err := s.repo.GetByID(ctx, id, organization_id)

	if err != nil {
		s.logger.Error("no contact exists with this ID", "error", err)
		return nil, err
	}

	return contact, nil
}

func (s *contactService) GetByEmail(ctx context.Context, email string, organization_id uuid.UUID) (*entity.Contact, error) {
	contact := &entity.Contact{}

	contact, err := s.repo.GetByEmail(ctx, email, organization_id)

	if err != nil {
		s.logger.Error("no contact exists with this email", "error", err)
		return nil, err
	}

	return contact, nil
}

func (s *contactService) Update(ctx context.Context, contact *entity.Contact) error {
	err := s.repo.Update(ctx, contact)

	if err != nil {
		s.logger.Error("failed to update contact information", "error", err)
		return err
	}

	s.logger.Info("contact updated successfully", "id", contact.ID)
	return nil
}

func (s *contactService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)

	if err != nil {
		s.logger.Error("failed to delete contact", "error", err)
		return err
	}

	s.logger.Info("contact soft deleted successfully", "id", id)
	return nil
}

func (s *contactService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.SoftDelete(ctx, id)

	if err != nil {
		s.logger.Error("failed to soft delete contact", "error", err)
		return err
	}

	s.logger.Info("contact soft deleted successfully")
	return nil
}

func (s *contactService) List(ctx context.Context, filters repo.ContactFilters) ([]*entity.Contact, error) {
	contacts, err := s.repo.List(ctx, filters)
	if err != nil {
		s.logger.Error("failed to list contacts", "error", err)
		return nil, err
	}

	s.logger.Debug("contacts listed successfully", "count", len(contacts))
	return contacts, nil
}

func (s *contactService) ListPaginated(ctx context.Context, filters repo.ContactFilters, page, pageSize int) (*repo.PaginatedResult[entity.Contact], error) {
	result, err := s.repo.ListPaginated(ctx, filters, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list paginated contacts", "page", page, "pageSize", pageSize, "error", err)
		return nil, err
	}

	s.logger.Debug("contacts paginated successfully",
		"page", result.Page,
		"total", result.Total,
		"count", len(result.Data))
	return result, nil
}

func (s *contactService) Search(ctx context.Context, query string, filters repo.ContactFilters, organization_id uuid.UUID) ([]*entity.Contact, error) {
	contacts, err := s.repo.Search(ctx, query, filters, organization_id)
	if err != nil {
		s.logger.Error("failed to search contacts", "query", query, "error", err)
		return nil, err
	}

	s.logger.Debug("contacts searched successfully", "query", query, "count", len(contacts))
	return contacts, nil
}
