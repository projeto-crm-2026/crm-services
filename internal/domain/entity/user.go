package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	UUID           uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Name           string         `gorm:"type:text;not null"`
	Email          string         `gorm:"type:text;not null;unique"`
	PasswordHash   string         `gorm:"type:text;not null"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string { return "user" }
