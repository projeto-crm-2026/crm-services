package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	UUID         uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	Name         string         `gorm:"type:text;not null"`
	Email        string         `gorm:"type:text;not null;unique"`
	PasswordHash string         `gorm:"type:text;not null"`
	CreatedAt    int64          `gorm:"autoCreateTime"`
	UpdatedAt    int64          `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string { return "user" }
