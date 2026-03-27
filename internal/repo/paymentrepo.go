package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepo interface {
	Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)
	GetByMpPaymentID(ctx context.Context, mpPaymentID string) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, mpPaymentID string, status entity.PaymentStatus, statusDetail string) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.Payment, error)
}

type paymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) PaymentRepo {
	return &paymentRepo{pool: pool}
}

func (r *paymentRepo) Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error) {
	query := `
		INSERT INTO payments (uuid, organization_id, subscription_id, mp_payment_id, status, status_detail, amount_cents, currency, payment_method, description, paid_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, uuid, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		payment.OrganizationID, payment.SubscriptionID, payment.MpPaymentID,
		payment.Status, payment.StatusDetail, payment.AmountCents, payment.Currency,
		payment.PaymentMethod, payment.Description, payment.PaidAt,
	).Scan(&payment.ID, &payment.UUID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *paymentRepo) GetByMpPaymentID(ctx context.Context, mpPaymentID string) (*entity.Payment, error) {
	query := `
		SELECT id, uuid, organization_id, subscription_id, mp_payment_id, status, status_detail, amount_cents, currency, payment_method, description, paid_at, created_at, updated_at
		FROM payments
		WHERE mp_payment_id = $1 AND deleted_at IS NULL`

	p := &entity.Payment{}
	err := r.pool.QueryRow(ctx, query, mpPaymentID).Scan(
		&p.ID, &p.UUID, &p.OrganizationID, &p.SubscriptionID, &p.MpPaymentID,
		&p.Status, &p.StatusDetail, &p.AmountCents, &p.Currency,
		&p.PaymentMethod, &p.Description, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	return p, err
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, mpPaymentID string, status entity.PaymentStatus, statusDetail string) error {
	query := `UPDATE payments SET status = $2, status_detail = $3, updated_at = NOW() WHERE mp_payment_id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, mpPaymentID, status, statusDetail)
	return err
}

func (r *paymentRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.Payment, error) {
	query := `
		SELECT id, uuid, organization_id, subscription_id, mp_payment_id, status, status_detail, amount_cents, currency, payment_method, description, paid_at, created_at, updated_at
		FROM payments
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []entity.Payment
	for rows.Next() {
		var p entity.Payment
		if err := rows.Scan(
			&p.ID, &p.UUID, &p.OrganizationID, &p.SubscriptionID, &p.MpPaymentID,
			&p.Status, &p.StatusDetail, &p.AmountCents, &p.Currency,
			&p.PaymentMethod, &p.Description, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, rows.Err()
}
