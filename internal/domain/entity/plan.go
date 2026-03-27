package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan struct {
	gorm.Model

	UUID              uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	Name              string         `gorm:"type:text;not null;unique"`
	DisplayName       string         `gorm:"type:text;not null"`
	PriceCents        int            `gorm:"not null;default:0"`
	Currency          string         `gorm:"type:text;not null;default:'BRL'"`
	MaxContacts       int            `gorm:"not null"`
	MaxMembers        int            `gorm:"not null"`
	MaxChatResponders    int            `gorm:"not null"`
	MpPreapprovalPlanID *string        `gorm:"type:text;unique;index"`
	IsActive             bool           `gorm:"default:true;index"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (Plan) TableName() string { return "plans" }
