package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKey struct {
	gorm.Model

	UUID       uuid.UUID `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	UserID     uint      `gorm:"not null;index"`
	PublicKey  string    `gorm:"type:text;not null;unique;index"` // on embed
	SecretKey  string    `gorm:"type:text;not null;unique"`       // server-to-server
	Name       string    `gorm:"type:text"`                       // ex: "Website do pai"
	Domain     string    `gorm:"type:text;not null"`              // domain authorized
	IsActive   bool      `gorm:"default:true"`
	LastUsedAt *time.Time
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}

func (APIKey) TableName() string { return "api_key" }
