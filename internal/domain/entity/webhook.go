package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEventType string

// eventos
const (
	EventMessageReceived     WebhookEventType = "message.received"
	EventMessageSent         WebhookEventType = "message.sent"
	EventChatCreated         WebhookEventType = "chat.created"
	EventChatClosed          WebhookEventType = "chat.closed"
	EventVisitorConnected    WebhookEventType = "visitor.connected"
	EventVisitorDisconnected WebhookEventType = "visitor.disconnected"
)

type Webhook struct {
	gorm.Model

	UUID       uuid.UUID `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	UserID     uint      `gorm:"not null;index"`
	Name       string    `gorm:"type:text;not null"`
	URL        string    `gorm:"type:text;not null"`
	Secret     string    `gorm:"type:text;not null"` // sign HMAC
	Events     string    `gorm:"type:text;not null"` // JSON de eventos
	IsActive   bool      `gorm:"default:true"`
	LastUsedAt *time.Time
	FailCount  int            `gorm:"default:0"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}

func (Webhook) TableName() string { return "webhook" }

// tentativas de envio
type WebhookLog struct {
	gorm.Model

	UUID         uuid.UUID `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	WebhookID    uint      `gorm:"not null;index"`
	EventType    string    `gorm:"type:text;not null"`
	Payload      string    `gorm:"type:text;not null"`
	ResponseCode int
	ResponseBody string    `gorm:"type:text"`
	Error        string    `gorm:"type:text"`
	Duration     int64     // milliseconds
	CreatedAt    time.Time `gorm:"autoCreateTime"`

	Webhook Webhook `gorm:"foreignKey:WebhookID"`
}

func (WebhookLog) TableName() string { return "webhook_log" }

// autenticação de webhooks de entrada
type IncomingWebhookToken struct {
	gorm.Model

	UUID       uuid.UUID `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	UserID     uint      `gorm:"not null;index"`
	Token      string    `gorm:"type:text;not null;unique;index"`
	Name       string    `gorm:"type:text"`
	IsActive   bool      `gorm:"default:true"`
	LastUsedAt *time.Time
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}

func (IncomingWebhookToken) TableName() string { return "incoming_webhook_token" }
