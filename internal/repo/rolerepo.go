package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type RoleRepo interface {
	Create(ctx context.Context, role *entity.Role) (*entity.Role, error)
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	GetByOrgAndName(ctx context.Context, orgID uuid.UUID, name string) (*entity.Role, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetRolePermissions(ctx context.Context, roleID uint) ([]entity.Permission, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	GetPermissionIDsByNames(ctx context.Context, names []string) ([]uint, error)
}

type roleRepo struct {
	pool *pgxpool.Pool
}

func NewRoleRepo(pool *pgxpool.Pool) RoleRepo {
	return &roleRepo{pool: pool}
}

func (r *roleRepo) Create(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	query := `
		INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, uuid, organization_id, name, description, is_system, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		role.OrganizationID, role.Name, role.Description, role.IsSystem,
	).Scan(
		&role.ID, &role.UUID, &role.OrganizationID, &role.Name,
		&role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return role, nil
}

func (r *roleRepo) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	query := `
		SELECT id, uuid, organization_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1 AND deleted_at IS NULL
	`
	var role entity.Role
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.UUID, &role.OrganizationID, &role.Name,
		&role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetByOrgAndName(ctx context.Context, orgID uuid.UUID, name string) (*entity.Role, error) {
	query := `
		SELECT id, uuid, organization_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL
	`
	var role entity.Role
	err := r.pool.QueryRow(ctx, query, orgID, name).Scan(
		&role.ID, &role.UUID, &role.OrganizationID, &role.Name,
		&role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.Role, error) {
	query := `
		SELECT id, uuid, organization_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY is_system DESC, name ASC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(
			&role.ID, &role.UUID, &role.OrganizationID, &role.Name,
			&role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, role.Name, role.Description, role.ID)
	return err
}

func (r *roleRepo) Delete(ctx context.Context, id uint) error {
	query := `UPDATE roles SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *roleRepo) SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return err
	}

	if len(permissionIDs) > 0 {
		batch := &pgx.Batch{}
		for _, pid := range permissionIDs {
			batch.Queue(`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roleID, pid)
		}
		br := tx.SendBatch(ctx, batch)
		for range permissionIDs {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return err
			}
		}
		br.Close()
	}

	return tx.Commit(ctx)
}

func (r *roleRepo) GetRolePermissions(ctx context.Context, roleID uint) ([]entity.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.category
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.name
	`
	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []entity.Permission
	for rows.Next() {
		var p entity.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *roleRepo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN "user" u ON u.role_id = rp.role_id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

func (r *roleRepo) GetPermissionIDsByNames(ctx context.Context, names []string) ([]uint, error) {
	query := `SELECT id FROM permissions WHERE name = ANY($1)`
	rows, err := r.pool.Query(ctx, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
