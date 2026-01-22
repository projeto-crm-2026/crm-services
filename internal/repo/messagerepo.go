package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type MessageRepo interface {
	Insert(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) (*entity.Message, error)
	GetByChatID(ctx context.Context, chatID uint) ([]entity.Message, error)
}

type messageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) MessageRepo {
	return &messageRepo{pool: pool}
}

func (r *messageRepo) Insert(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) (*entity.Message, error) {
	query := `
        INSERT INTO message (chat_id, sender_id, visitor_id, content, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, 'text', NOW(), NOW())
        RETURNING id, uuid, chat_id, sender_id, visitor_id, content, type, created_at
    `

	var msg entity.Message
	err := r.pool.QueryRow(ctx, query, chatID, senderID, visitorID, content).Scan(
		&msg.ID, &msg.UUID, &msg.ChatID, &msg.SenderID, &msg.VisitorID, &msg.Content, &msg.Type, &msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepo) GetByChatID(ctx context.Context, chatID uint) ([]entity.Message, error) {
	query := `
        SELECT id, uuid, chat_id, sender_id, visitor_id, content, type, created_at
        FROM message WHERE chat_id = $1 AND deleted_at IS NULL
        ORDER BY created_at ASC
    `

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		if err := rows.Scan(&msg.ID, &msg.UUID, &msg.ChatID, &msg.SenderID, &msg.VisitorID, &msg.Content, &msg.Type, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
