package model

type CreateWebhookRequest struct {
	Name   string   `json:"name" validate:"max=255"`
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
}

type UpdateWebhookRequest struct {
	Name     string   `json:"name" validate:"max=255"`
	URL      string   `json:"url" validate:"required,url"`
	Events   []string `json:"events" validate:"required,min=1"`
	IsActive bool     `json:"is_active"`
}

type WebhookResponse struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Secret     string   `json:"secret,omitempty"`
	Events     []string `json:"events"`
	IsActive   bool     `json:"is_active"`
	FailCount  int      `json:"fail_count"`
	LastUsedAt *string  `json:"last_used_at,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

type WebhookLogResponse struct {
	ID           uint   `json:"id"`
	EventType    string `json:"event_type"`
	ResponseCode int    `json:"response_code"`
	Error        string `json:"error,omitempty"`
	Duration     int64  `json:"duration_ms"`
	CreatedAt    string `json:"created_at"`
}

type CreateIncomingTokenRequest struct {
	Name string `json:"name"`
}

type IncomingTokenResponse struct {
	ID         uint    `json:"id"`
	Token      string  `json:"token,omitempty"`
	Name       string  `json:"name"`
	IsActive   bool    `json:"is_active"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

type IncomingWebhookRequest struct {
	Action  string                 `json:"action"`
	ChatID  uint                   `json:"chat_id"`
	Content string                 `json:"content,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// events
var AvailableWebhookEvents = []string{
	"message.received",
	"message.sent",
	"chat.created",
	"chat.closed",
	"visitor.connected",
	"visitor.disconnected",
}
