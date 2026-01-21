package model

import "github.com/google/uuid"

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

type WebSocketMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	SenderID  *uint  `json:"sender_id,omitempty"`
	VisitorID string `json:"visitor_id,omitempty"`
	ChatID    uint   `json:"chat_id"`
}
