package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionRepo interface {
	Create(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error)
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*entity.Subscription, error)
	GetByID(ctx context.Context, id uint) (*entity.Subscription, error)
	GetByMpSubscriptionID(ctx context.Context, mpSubID string) (*entity.Subscription, error)
	UpdateStatus(ctx context.Context, id uint, status entity.SubscriptionStatus) error
	Update(ctx context.Context, sub *entity.Subscription) error
}

type subscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) SubscriptionRepo {
	return &subscriptionRepo{pool: pool}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error) {
	query := `
		INSERT INTO subscriptions (
			uuid, organization_id, plan_id, status, current_period_start, current_period_end,
			cancel_at_period_end, mp_subscription_id, cancelled_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
		)
		RETURNING id, uuid, created_at, updated_at`

	result := &entity.Subscription{
		OrganizationID:     sub.OrganizationID,
		PlanID:             sub.PlanID,
		Status:             sub.Status,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		MpSubscriptionID:   sub.MpSubscriptionID,
		CancelledAt:        sub.CancelledAt,
	}

	err := r.pool.QueryRow(ctx, query,
		sub.OrganizationID, sub.PlanID, sub.Status,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.MpSubscriptionID, sub.CancelledAt,
	).Scan(&result.ID, &result.UUID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *subscriptionRepo) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*entity.Subscription, error) {
	query := `
		SELECT id, uuid, organization_id, plan_id, status, current_period_start, current_period_end,
			cancel_at_period_end, mp_subscription_id, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE organization_id = $1 AND deleted_at IS NULL`

	sub := &entity.Subscription{}
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&sub.ID, &sub.UUID, &sub.OrganizationID, &sub.PlanID, &sub.Status,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.MpSubscriptionID, &sub.CancelledAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}

	return sub, err
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uint) (*entity.Subscription, error) {
	query := `
		SELECT id, uuid, organization_id, plan_id, status, current_period_start, current_period_end,
			cancel_at_period_end, mp_subscription_id, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1 AND deleted_at IS NULL`

	sub := &entity.Subscription{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.UUID, &sub.OrganizationID, &sub.PlanID, &sub.Status,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.MpSubscriptionID, &sub.CancelledAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}

	return sub, err
}

func (r *subscriptionRepo) GetByMpSubscriptionID(ctx context.Context, mpSubID string) (*entity.Subscription, error) {
	query := `
		SELECT id, uuid, organization_id, plan_id, status, current_period_start, current_period_end,
			cancel_at_period_end, mp_subscription_id, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE mp_subscription_id = $1 AND deleted_at IS NULL`

	sub := &entity.Subscription{}
	err := r.pool.QueryRow(ctx, query, mpSubID).Scan(
		&sub.ID, &sub.UUID, &sub.OrganizationID, &sub.PlanID, &sub.Status,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.MpSubscriptionID, &sub.CancelledAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}

	return sub, err
}

func (r *subscriptionRepo) UpdateStatus(ctx context.Context, id uint, status entity.SubscriptionStatus) error {
	query := `UPDATE subscriptions SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (r *subscriptionRepo) Update(ctx context.Context, sub *entity.Subscription) error {
	query := `
		UPDATE subscriptions SET
			plan_id = $2, status = $3, current_period_start = $4, current_period_end = $5,
			cancel_at_period_end = $6, mp_subscription_id = $7, cancelled_at = $8, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		sub.ID, sub.PlanID, sub.Status,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.MpSubscriptionID, sub.CancelledAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}
