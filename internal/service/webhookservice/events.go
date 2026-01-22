package webhookservice

import (
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type WebhookEvent struct {
	ID        string                  `json:"id"`
	Type      entity.WebhookEventType `json:"type"`
	Timestamp time.Time               `json:"timestamp"`
	Data      interface{}             `json:"data"`
}

// the provided event type, and the given payload as Data.
func NewWebhookEvent(eventType entity.WebhookEventType, data interface{}) *WebhookEvent {
	return &WebhookEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

type MessageEventData struct {
	ChatID    uint   `json:"chat_id"`
	MessageID uint   `json:"message_id"`
	Content   string `json:"content"`
	SenderID  *uint  `json:"sender_id,omitempty"`
	VisitorID string `json:"visitor_id,omitempty"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type ChatEventData struct {
	ChatID      uint   `json:"chat_id"`
	Status      string `json:"status"`
	Origin      string `json:"origin"`
	OwnerUserID uint   `json:"owner_user_id"`
	VisitorID   string `json:"visitor_id,omitempty"`
	Timestamp   string `json:"timestamp"`
}

type VisitorEventData struct {
	ChatID    uint   `json:"chat_id"`
	VisitorID string `json:"visitor_id"`
	Timestamp string `json:"timestamp"`
}

// payload recebido de webhooks externos
type IncomingWebhookPayload struct {
	Action  string                 `json:"action"`
	ChatID  uint                   `json:"chat_id"`
	Content string                 `json:"content,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}