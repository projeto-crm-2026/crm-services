package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type UserRepo interface {
	Insert(ctx context.Context, name string, email string, passwordHash string, organizationID uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByInviteToken(ctx context.Context, token string) (*entity.User, error)
	InsertPending(ctx context.Context, name, email, inviteToken string, inviteExpiry time.Time, organizationID uuid.UUID, invitedBy uint) (*entity.User, error)
	ActivateUser(ctx context.Context, userID uint, passwordHash string) error
	ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error)
}

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Insert(ctx context.Context, name string, email string, passwordHash string, organizationID uuid.UUID) (*entity.User, error) {
	query := `
        INSERT INTO "user" (name, email, password_hash, role, status, organization_id, created_at, updated_at)
        VALUES ($1, $2, $3, 'admin', 'active', $4, NOW(), NOW())
        RETURNING id, uuid, organization_id, name, email, password_hash, role, status, created_at, updated_at
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, name, email, passwordHash, organizationID).Scan(
		&user.ID,
		&user.UUID,
		&user.OrganizationID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
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
        SELECT id, uuid, organization_id, name, email, password_hash
        FROM "user"
        WHERE email = $1 AND deleted_at IS NULL
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.UUID,
		&user.OrganizationID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	query := `
        SELECT id, uuid, organization_id, name, email, password_hash, role, status, 
                invite_token, invite_expiry, invited_by, created_at, updated_at
        FROM "user"
        WHERE id = $1 AND deleted_at IS NULL
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.UUID, &user.OrganizationID, &user.Name, &user.Email,
		&user.PasswordHash, &user.Role, &user.Status,
		&user.InviteToken, &user.InviteExpiry, &user.InvitedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByInviteToken(ctx context.Context, token string) (*entity.User, error) {
	query := `
        SELECT id, uuid, organization_id, name, email, password_hash, role, status, 
                invite_token, invite_expiry, invited_by, created_at, updated_at
        FROM "user"
        WHERE invite_token = $1 AND deleted_at IS NULL
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&user.ID, &user.UUID, &user.OrganizationID, &user.Name, &user.Email,
		&user.PasswordHash, &user.Role, &user.Status,
		&user.InviteToken, &user.InviteExpiry, &user.InvitedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) InsertPending(ctx context.Context, name, email, inviteToken string, inviteExpiry time.Time, organizationID uuid.UUID, invitedBy uint) (*entity.User, error) {
	query := `
        INSERT INTO "user" (name, email, password_hash, role, status, organization_id, invite_token, invite_expiry, invited_by, created_at, updated_at)
        VALUES ($1, $2, '', 'member', 'pending', $3, $4, $5, $6, NOW(), NOW())
        RETURNING id, uuid, organization_id, name, email, role, status, invite_token, invite_expiry, invited_by, created_at, updated_at
    `

	var user entity.User
	err := r.pool.QueryRow(ctx, query, name, email, organizationID, inviteToken, inviteExpiry, invitedBy).Scan(
		&user.ID, &user.UUID, &user.OrganizationID, &user.Name, &user.Email,
		&user.Role, &user.Status, &user.InviteToken, &user.InviteExpiry, &user.InvitedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) ActivateUser(ctx context.Context, userID uint, passwordHash string) error {
	query := `
        UPDATE "user"
        SET password_hash = $1, status = 'active', invite_token = NULL, invite_expiry = NULL, updated_at = NOW()
        WHERE id = $2
    `
	_, err := r.pool.Exec(ctx, query, passwordHash, userID)
	return err
}

func (r *userRepo) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error) {
	query := `
        SELECT id, uuid, organization_id, name, email, role, status, created_at, updated_at
        FROM "user"
        WHERE organization_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID, &u.UUID, &u.OrganizationID, &u.Name, &u.Email,
			&u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
