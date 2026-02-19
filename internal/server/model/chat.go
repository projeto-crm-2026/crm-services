package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type CreateChatRequest struct {
	Origin    string `json:"origin"`
	VisitorID string `json:"visitor_id"`
}

type CreateWidgetChatRequest struct {
	VisitorID string `json:"visitor_id"`
}

type ChatResponse struct {
	ID     uint      `json:"id"`
	UUID   uuid.UUID `json:"uuid"`
	Status string    `json:"status"`
	Origin string    `json:"origin"`
}

type MessageSenderResponse struct {
	ID   uint      `json:"id"`
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type MessageResponse struct {
	ID        uint                   `json:"id"`
	UUID      uuid.UUID              `json:"uuid"`
	ChatID    uint                   `json:"chat_id"`
	SenderID  *uint                  `json:"sender_id,omitempty"`
	VisitorID string                 `json:"visitor_id,omitempty"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Sender    *MessageSenderResponse `json:"sender,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func ToMessageResponse(msg entity.Message) MessageResponse {
	resp := MessageResponse{
		ID:        msg.ID,
		UUID:      msg.UUID,
		ChatID:    msg.ChatID,
		SenderID:  msg.SenderID,
		VisitorID: msg.VisitorID,
		Content:   msg.Content,
		Type:      string(msg.Type),
		CreatedAt: msg.CreatedAt,
	}
	if msg.Sender != nil {
		resp.Sender = &MessageSenderResponse{
			ID:   msg.Sender.ID,
			UUID: msg.Sender.UUID,
			Name: msg.Sender.Name,
		}
	}
	return resp
}

func ToMessageResponseList(messages []entity.Message) []MessageResponse {
	response := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = ToMessageResponse(msg)
	}
	return response
}

type WebSocketMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	SenderID  *uint  `json:"sender_id,omitempty"`
	VisitorID string `json:"visitor_id,omitempty"`
	ChatID    uint   `json:"chat_id"`
}
