package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type UserRepo interface {
	Insert(ctx context.Context, name string, email string, passwordHash string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Insert(ctx context.Context, name string, email string, passwordHash string) (*entity.User, error) {
	query := `
        INSERT INTO "user" (name, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, uuid, name, email, password_hash, created_at, updated_at
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, name, email, passwordHash).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
        SELECT id, uuid, name, email, password_hash, created_at, updated_at
        FROM "user"
        WHERE email = $1 AND deleted_at IS NULL
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
