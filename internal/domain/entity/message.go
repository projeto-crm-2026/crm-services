package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	gorm.Model

	UUID      uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	ChatID    uint           `gorm:"not null;index"`
	SenderID  *uint          `gorm:"index"`
	VisitorID string         `gorm:"type:text"` // external visitor id
	Content   string         `gorm:"type:text;not null"`
	Type      MessageType    `gorm:"type:text;not null;default:'text'"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Chat   Chat  `gorm:"foreignKey:ChatID"`
	Sender *User `gorm:"foreignKey:SenderID"`
}

func (Message) TableName() string { return "message" }
