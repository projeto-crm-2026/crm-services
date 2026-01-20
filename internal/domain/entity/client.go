package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model

	UUID uuid.UUID `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	Name string    `gorm:"type:text;not null"`
}

func (Client) TableName() string { return "client" }
