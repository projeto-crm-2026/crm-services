package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

var ErrOrganizationNotFound = errors.New("organization not found")

type OrganizationRepo interface {
	Create(ctx context.Context, org *entity.Organization) (*entity.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Organization, error)
	Update(ctx context.Context, org *entity.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
}

type organizationRepo struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepo(pool *pgxpool.Pool) OrganizationRepo {
	return &organizationRepo{pool: pool}
}

func (r *organizationRepo) Create(ctx context.Context, org *entity.Organization) (*entity.Organization, error) {
	query := `
		INSERT INTO organizations (
			name, slug, email, phone, website, document_id, industry, plan, settings, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			NOW(), NOW()
		)
		RETURNING uuid, created_at, updated_at`

	result := &entity.Organization{
		Name:       org.Name,
		Slug:       org.Slug,
		Email:      org.Email,
		Phone:      org.Phone,
		Website:    org.Website,
		DocumentID: org.DocumentID,
		Industry:   org.Industry,
		Plan:       org.Plan,
		Settings:   org.Settings,
		IsActive:   org.IsActive,
	}

	err := r.pool.QueryRow(ctx, query,
		org.Name, org.Slug, org.Email, org.Phone, org.Website,
		org.DocumentID, org.Industry, org.Plan, org.Settings, org.IsActive,
	).Scan(&result.UUID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *organizationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	query := `
		SELECT uuid, name, slug, email, phone, website, document_id, industry, plan, settings, is_active, created_at, updated_at, deleted_at
		FROM organizations
		WHERE uuid = $1 AND deleted_at IS NULL`

	org := &entity.Organization{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.UUID, &org.Name, &org.Slug, &org.Email, &org.Phone, &org.Website, &org.DocumentID, &org.Industry, &org.Plan,
		&org.Settings, &org.IsActive, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrganizationNotFound
	}

	return org, err
}

func (r *organizationRepo) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	query := `
		SELECT uuid, name, slug, email, phone, website, document_id, industry, plan, settings, is_active, created_at, updated_at, deleted_at
		FROM organizations
		WHERE slug = $1 AND deleted_at IS NULL`

	org := &entity.Organization{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.UUID, &org.Name, &org.Slug, &org.Email, &org.Phone, &org.Website, &org.DocumentID, &org.Industry, &org.Plan,
		&org.Settings, &org.IsActive, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrganizationNotFound
	}

	return org, err
}

func (r *organizationRepo) Update(ctx context.Context, org *entity.Organization) error {
	query := `
		UPDATE organizations SET
			name = $2, slug = $3, email = $4, phone = $5, website = $6, document_id = $7, industry = $8, settings = $9,
			updated_at = NOW()
		WHERE uuid = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		org.UUID, org.Name, org.Slug, org.Email, org.Phone, org.Website, org.DocumentID, org.Industry, org.Settings,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

func (r *organizationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE uuid = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE organizations SET deleted_at = NOW() WHERE uuid = $1 AND deleted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepo) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE organizations SET deleted_at = NULL WHERE uuid = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}
