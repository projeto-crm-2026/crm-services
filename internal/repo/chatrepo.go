package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type ChatRepo interface {
	Insert(ctx context.Context, origin string, ownerUserID uint) (*entity.Chat, error)
	GetByID(ctx context.Context, id uint) (*entity.Chat, error)
	GetByOwnerUserID(ctx context.Context, ownerUserID uint) ([]entity.Chat, error)
	AddParticipant(ctx context.Context, chatID uint, userID *uint, visitorID string, role entity.ParticipantRole) error
}

type chatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) ChatRepo {
	return &chatRepo{pool: pool}
}

func (r *chatRepo) Insert(ctx context.Context, origin string, ownerUserID uint) (*entity.Chat, error) {
	query := `
        INSERT INTO chat (origin, owner_user_id, status, created_at, updated_at)
        VALUES ($1, $2, 'open', NOW(), NOW())
        RETURNING id, uuid, status, origin, owner_user_id, created_at, updated_at
    `

	var chat entity.Chat
	err := r.pool.QueryRow(ctx, query, origin, ownerUserID).Scan(
		&chat.ID, &chat.UUID, &chat.Status, &chat.Origin, &chat.OwnerUserID, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepo) GetByID(ctx context.Context, id uint) (*entity.Chat, error) {
	query := `
        SELECT id, uuid, status, origin, owner_user_id, created_at, updated_at 
        FROM chat 
        WHERE id = $1 AND deleted_at IS NULL
    `

	var chat entity.Chat
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&chat.ID, &chat.UUID, &chat.Status, &chat.Origin, &chat.OwnerUserID, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepo) GetByOwnerUserID(ctx context.Context, ownerUserID uint) ([]entity.Chat, error) {
	query := `
        SELECT id, uuid, status, origin, owner_user_id, created_at, updated_at 
        FROM chat 
        WHERE owner_user_id = $1 AND deleted_at IS NULL
        ORDER BY updated_at DESC
    `

	rows, err := r.pool.Query(ctx, query, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []entity.Chat
	for rows.Next() {
		var chat entity.Chat
		if err := rows.Scan(
			&chat.ID, &chat.UUID, &chat.Status, &chat.Origin, &chat.OwnerUserID, &chat.CreatedAt, &chat.UpdatedAt,
		); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func (r *chatRepo) AddParticipant(ctx context.Context, chatID uint, userID *uint, visitorID string, role entity.ParticipantRole) error {
	query := `
        INSERT INTO chat_participant (chat_id, user_id, visitor_id, role, joined_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())
    `
	_, err := r.pool.Exec(ctx, query, chatID, userID, visitorID, role)
	return err
}
