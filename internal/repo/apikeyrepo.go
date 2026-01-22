package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type APIKeyRepo interface {
	Insert(ctx context.Context, userID uint, publicKey, secretKey, name, domain string) (*entity.APIKey, error)
	GetByPublicKey(ctx context.Context, publicKey string) (*entity.APIKey, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.APIKey, error)
	Delete(ctx context.Context, userID, keyID uint) error
	UpdateLastUsed(ctx context.Context, keyID uint) error
}

type apiKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepo(pool *pgxpool.Pool) APIKeyRepo {
	return &apiKeyRepo{pool: pool}
}

func (r *apiKeyRepo) Insert(ctx context.Context, userID uint, publicKey, secretKey, name, domain string) (*entity.APIKey, error) {
	query := `
        INSERT INTO api_key (user_id, public_key, secret_key, name, domain, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
        RETURNING id, uuid, user_id, public_key, secret_key, name, domain, is_active, created_at, updated_at
    `

	var apiKey entity.APIKey
	err := r.pool.QueryRow(ctx, query, userID, publicKey, secretKey, name, domain).Scan(
		&apiKey.ID, &apiKey.UUID, &apiKey.UserID, &apiKey.PublicKey, &apiKey.SecretKey,
		&apiKey.Name, &apiKey.Domain, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepo) GetByPublicKey(ctx context.Context, publicKey string) (*entity.APIKey, error) {
	query := `
        SELECT id, uuid, user_id, public_key, secret_key, name, domain, is_active, last_used_at, created_at, updated_at
        FROM api_key
        WHERE public_key = $1 AND deleted_at IS NULL
    `

	var apiKey entity.APIKey
	err := r.pool.QueryRow(ctx, query, publicKey).Scan(
		&apiKey.ID, &apiKey.UUID, &apiKey.UserID, &apiKey.PublicKey, &apiKey.SecretKey,
		&apiKey.Name, &apiKey.Domain, &apiKey.IsActive, &apiKey.LastUsedAt, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepo) GetByUserID(ctx context.Context, userID uint) ([]entity.APIKey, error) {
	query := `
        SELECT id, uuid, user_id, public_key, name, domain, is_active, last_used_at, created_at, updated_at
        FROM api_key
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []entity.APIKey
	for rows.Next() {
		var apiKey entity.APIKey
		if err := rows.Scan(
			&apiKey.ID, &apiKey.UUID, &apiKey.UserID, &apiKey.PublicKey,
			&apiKey.Name, &apiKey.Domain, &apiKey.IsActive, &apiKey.LastUsedAt,
			&apiKey.CreatedAt, &apiKey.UpdatedAt,
		); err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}
	return apiKeys, nil
}

func (r *apiKeyRepo) Delete(ctx context.Context, userID, keyID uint) error {
	query := `UPDATE api_key SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, keyID, userID)
	return err
}

func (r *apiKeyRepo) UpdateLastUsed(ctx context.Context, keyID uint) error {
	query := `UPDATE api_key SET last_used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, keyID)
	return err
}
