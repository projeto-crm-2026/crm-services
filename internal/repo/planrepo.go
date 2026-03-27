package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

var ErrPlanNotFound = errors.New("plan not found")

type PlanRepo interface {
	GetAll(ctx context.Context) ([]entity.Plan, error)
	GetByID(ctx context.Context, id uint) (*entity.Plan, error)
	GetByName(ctx context.Context, name string) (*entity.Plan, error)
	GetByMpPlanID(ctx context.Context, mpPlanID string) (*entity.Plan, error)
	UpdateMpPlanID(ctx context.Context, id uint, mpPlanID string) error
	CountContacts(ctx context.Context, orgID uuid.UUID) (int, error)
	CountMembers(ctx context.Context, orgID uuid.UUID) (int, error)
	CountChatResponders(ctx context.Context, orgID uuid.UUID) (int, error)
}

type planRepo struct {
	pool *pgxpool.Pool
}

func NewPlanRepo(pool *pgxpool.Pool) PlanRepo {
	return &planRepo{pool: pool}
}

func (r *planRepo) GetAll(ctx context.Context) ([]entity.Plan, error) {
	query := `
		SELECT id, uuid, name, display_name, price_cents, currency, max_contacts, max_members, max_chat_responders, mp_preapproval_plan_id, is_active, created_at, updated_at
		FROM plans
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY price_cents ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []entity.Plan
	for rows.Next() {
		var plan entity.Plan
		if err := rows.Scan(
			&plan.ID, &plan.UUID, &plan.Name, &plan.DisplayName, &plan.PriceCents, &plan.Currency,
			&plan.MaxContacts, &plan.MaxMembers, &plan.MaxChatResponders, &plan.MpPreapprovalPlanID, &plan.IsActive,
			&plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *planRepo) GetByID(ctx context.Context, id uint) (*entity.Plan, error) {
	query := `
		SELECT id, uuid, name, display_name, price_cents, currency, max_contacts, max_members, max_chat_responders, mp_preapproval_plan_id, is_active, created_at, updated_at
		FROM plans
		WHERE id = $1 AND deleted_at IS NULL`

	plan := &entity.Plan{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&plan.ID, &plan.UUID, &plan.Name, &plan.DisplayName, &plan.PriceCents, &plan.Currency,
		&plan.MaxContacts, &plan.MaxMembers, &plan.MaxChatResponders, &plan.MpPreapprovalPlanID, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlanNotFound
	}

	return plan, err
}

func (r *planRepo) GetByName(ctx context.Context, name string) (*entity.Plan, error) {
	query := `
		SELECT id, uuid, name, display_name, price_cents, currency, max_contacts, max_members, max_chat_responders, mp_preapproval_plan_id, is_active, created_at, updated_at
		FROM plans
		WHERE name = $1 AND deleted_at IS NULL`

	plan := &entity.Plan{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&plan.ID, &plan.UUID, &plan.Name, &plan.DisplayName, &plan.PriceCents, &plan.Currency,
		&plan.MaxContacts, &plan.MaxMembers, &plan.MaxChatResponders, &plan.MpPreapprovalPlanID, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlanNotFound
	}

	return plan, err
}

func (r *planRepo) GetByMpPlanID(ctx context.Context, mpPlanID string) (*entity.Plan, error) {
	query := `
		SELECT id, uuid, name, display_name, price_cents, currency, max_contacts, max_members, max_chat_responders, mp_preapproval_plan_id, is_active, created_at, updated_at
		FROM plans
		WHERE mp_preapproval_plan_id = $1 AND deleted_at IS NULL`

	plan := &entity.Plan{}
	err := r.pool.QueryRow(ctx, query, mpPlanID).Scan(
		&plan.ID, &plan.UUID, &plan.Name, &plan.DisplayName, &plan.PriceCents, &plan.Currency,
		&plan.MaxContacts, &plan.MaxMembers, &plan.MaxChatResponders, &plan.MpPreapprovalPlanID, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlanNotFound
	}

	return plan, err
}

func (r *planRepo) UpdateMpPlanID(ctx context.Context, id uint, mpPlanID string) error {
	query := `UPDATE plans SET mp_preapproval_plan_id = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id, mpPlanID)
	return err
}

func (r *planRepo) CountContacts(ctx context.Context, orgID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM contacts WHERE organization_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.pool.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}

func (r *planRepo) CountMembers(ctx context.Context, orgID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM "user" WHERE organization_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.pool.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}

func (r *planRepo) CountChatResponders(ctx context.Context, orgID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM "user" u
		JOIN role_permissions rp ON rp.role_id = u.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.organization_id = $1 AND u.deleted_at IS NULL AND p.name = 'chats.read'
	`

	var count int
	err := r.pool.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}
