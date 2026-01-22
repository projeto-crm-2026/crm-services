package repo

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type WebhookRepo interface {
	// outgoing webhooks
	Insert(ctx context.Context, userID uint, name, url, secret string, events []string) (*entity.Webhook, error)
	GetByID(ctx context.Context, id uint) (*entity.Webhook, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.Webhook, error)
	GetActiveByUserAndEvent(ctx context.Context, userID uint, eventType string) ([]entity.Webhook, error)
	Update(ctx context.Context, id uint, name, url string, events []string, isActive bool) error
	Delete(ctx context.Context, userID, webhookID uint) error
	UpdateLastUsed(ctx context.Context, webhookID uint) error
	IncrementFailCount(ctx context.Context, webhookID uint) error
	ResetFailCount(ctx context.Context, webhookID uint) error

	// logs
	InsertLog(ctx context.Context, log *entity.WebhookLog) error
	GetLogsByWebhookID(ctx context.Context, webhookID uint, limit int) ([]entity.WebhookLog, error)

	// incoming webhook
	InsertToken(ctx context.Context, userID uint, token, name string) (*entity.IncomingWebhookToken, error)
	GetTokenByValue(ctx context.Context, token string) (*entity.IncomingWebhookToken, error)
	GetTokensByUserID(ctx context.Context, userID uint) ([]entity.IncomingWebhookToken, error)
	DeleteToken(ctx context.Context, userID, tokenID uint) error
	UpdateTokenLastUsed(ctx context.Context, tokenID uint) error
}

type webhookRepo struct {
	pool *pgxpool.Pool
}

// NewWebhookRepo creates a WebhookRepo backed by the provided pgxpool.Pool.
// The pool must be non-nil.
func NewWebhookRepo(pool *pgxpool.Pool) WebhookRepo {
	return &webhookRepo{pool: pool}
}

func (r *webhookRepo) Insert(ctx context.Context, userID uint, name, url, secret string, events []string) (*entity.Webhook, error) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return nil, err
	}

	query := `
        INSERT INTO webhook (user_id, name, url, secret, events, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
        RETURNING id, uuid, user_id, name, url, secret, events, is_active, created_at, updated_at
    `

	var webhook entity.Webhook
	err = r.pool.QueryRow(ctx, query, userID, name, url, secret, string(eventsJSON)).Scan(
		&webhook.ID, &webhook.UUID, &webhook.UserID, &webhook.Name, &webhook.URL,
		&webhook.Secret, &webhook.Events, &webhook.IsActive, &webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepo) GetByID(ctx context.Context, id uint) (*entity.Webhook, error) {
	query := `
        SELECT id, uuid, user_id, name, url, secret, events, is_active, last_used_at, fail_count, created_at, updated_at
        FROM webhook
        WHERE id = $1 AND deleted_at IS NULL
    `

	var webhook entity.Webhook
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID, &webhook.UUID, &webhook.UserID, &webhook.Name, &webhook.URL,
		&webhook.Secret, &webhook.Events, &webhook.IsActive, &webhook.LastUsedAt,
		&webhook.FailCount, &webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepo) GetByUserID(ctx context.Context, userID uint) ([]entity.Webhook, error) {
	query := `
        SELECT id, uuid, user_id, name, url, events, is_active, last_used_at, fail_count, created_at, updated_at
        FROM webhook
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []entity.Webhook
	for rows.Next() {
		var webhook entity.Webhook
		if err := rows.Scan(
			&webhook.ID, &webhook.UUID, &webhook.UserID, &webhook.Name, &webhook.URL,
			&webhook.Events, &webhook.IsActive, &webhook.LastUsedAt, &webhook.FailCount,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

func (r *webhookRepo) GetActiveByUserAndEvent(ctx context.Context, userID uint, eventType string) ([]entity.Webhook, error) {
	query := `
        SELECT id, uuid, user_id, name, url, secret, events, is_active, fail_count, created_at, updated_at
        FROM webhook
        WHERE user_id = $1 
          AND is_active = true 
          AND deleted_at IS NULL
          AND fail_count < 10
          AND events LIKE '%' || $2 || '%'
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, userID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []entity.Webhook
	for rows.Next() {
		var webhook entity.Webhook
		if err := rows.Scan(
			&webhook.ID, &webhook.UUID, &webhook.UserID, &webhook.Name, &webhook.URL,
			&webhook.Secret, &webhook.Events, &webhook.IsActive, &webhook.FailCount,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

func (r *webhookRepo) Update(ctx context.Context, id uint, name, url string, events []string, isActive bool) error {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}

	query := `
        UPDATE webhook 
        SET name = $1, url = $2, events = $3, is_active = $4, updated_at = NOW()
        WHERE id = $5 AND deleted_at IS NULL
    `
	_, err = r.pool.Exec(ctx, query, name, url, string(eventsJSON), isActive, id)

	return err
}

func (r *webhookRepo) Delete(ctx context.Context, userID, webhookID uint) error {
	query := `UPDATE webhook SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, webhookID, userID)
	return err
}

func (r *webhookRepo) UpdateLastUsed(ctx context.Context, webhookID uint) error {
	query := `UPDATE webhook SET last_used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, webhookID)
	return err
}

func (r *webhookRepo) IncrementFailCount(ctx context.Context, webhookID uint) error {
	query := `UPDATE webhook SET fail_count = fail_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, webhookID)
	return err
}

func (r *webhookRepo) ResetFailCount(ctx context.Context, webhookID uint) error {
	query := `UPDATE webhook SET fail_count = 0 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, webhookID)
	return err
}

func (r *webhookRepo) InsertLog(ctx context.Context, log *entity.WebhookLog) error {
	query := `
        INSERT INTO webhook_log (webhook_id, event_type, payload, response_code, response_body, error, duration, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
    `
	_, err := r.pool.Exec(ctx, query, log.WebhookID, log.EventType, log.Payload, log.ResponseCode, log.ResponseBody, log.Error, log.Duration)
	return err
}

func (r *webhookRepo) GetLogsByWebhookID(ctx context.Context, webhookID uint, limit int) ([]entity.WebhookLog, error) {
	query := `
        SELECT id, uuid, webhook_id, event_type, payload, response_code, response_body, error, duration, created_at
        FROM webhook_log
        WHERE webhook_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `

	rows, err := r.pool.Query(ctx, query, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.WebhookLog
	for rows.Next() {
		var log entity.WebhookLog
		if err := rows.Scan(
			&log.ID, &log.UUID, &log.WebhookID, &log.EventType, &log.Payload,
			&log.ResponseCode, &log.ResponseBody, &log.Error, &log.Duration, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// incoming
func (r *webhookRepo) InsertToken(ctx context.Context, userID uint, token, name string) (*entity.IncomingWebhookToken, error) {
	query := `
        INSERT INTO incoming_webhook_token (user_id, token, name, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, true, NOW(), NOW())
        RETURNING id, uuid, user_id, token, name, is_active, created_at, updated_at
    `

	var t entity.IncomingWebhookToken
	err := r.pool.QueryRow(ctx, query, userID, token, name).Scan(
		&t.ID, &t.UUID, &t.UserID, &t.Token, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *webhookRepo) GetTokenByValue(ctx context.Context, token string) (*entity.IncomingWebhookToken, error) {
	query := `
        SELECT id, uuid, user_id, token, name, is_active, last_used_at, created_at, updated_at
        FROM incoming_webhook_token
        WHERE token = $1 AND is_active = true AND deleted_at IS NULL
    `

	var t entity.IncomingWebhookToken
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.UUID, &t.UserID, &t.Token, &t.Name, &t.IsActive, &t.LastUsedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *webhookRepo) GetTokensByUserID(ctx context.Context, userID uint) ([]entity.IncomingWebhookToken, error) {
	query := `
        SELECT id, uuid, user_id, token, name, is_active, last_used_at, created_at, updated_at
        FROM incoming_webhook_token
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []entity.IncomingWebhookToken
	for rows.Next() {
		var t entity.IncomingWebhookToken
		if err := rows.Scan(
			&t.ID, &t.UUID, &t.UserID, &t.Token, &t.Name, &t.IsActive, &t.LastUsedAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	return tokens, nil
}

func (r *webhookRepo) DeleteToken(ctx context.Context, userID, tokenID uint) error {
	query := `UPDATE incoming_webhook_token SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, tokenID, userID)
	return err
}

func (r *webhookRepo) UpdateTokenLastUsed(ctx context.Context, tokenID uint) error {
	query := `UPDATE incoming_webhook_token SET last_used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, tokenID)
	return err
}