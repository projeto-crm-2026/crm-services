package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatStatus string

const (
	ChatStatusOpen   ChatStatus = "open"
	ChatStatusClosed ChatStatus = "closed"
)

type Chat struct {
	gorm.Model

	UUID        uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	Status      ChatStatus     `gorm:"type:text;not null;default:'open'"`
	Origin      string         `gorm:"type:text;not null"`
	OwnerUserID uint           `gorm:"not null;index"` // CRM account
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Owner User `gorm:"foreignKey:OwnerUserID"`
}

func (Chat) TableName() string { return "chat" }
