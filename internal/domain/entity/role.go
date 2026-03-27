package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	UUID           uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name           string         `gorm:"type:text;not null"`
	Description    string         `gorm:"type:text"`
	IsSystem       bool           `gorm:"default:false"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Permissions  []Permission  `gorm:"many2many:role_permissions;"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:UUID"`
}

func (Role) TableName() string { return "roles" }

const (
	SystemRoleOwner  = "owner"
	SystemRoleAdmin  = "admin"
	SystemRoleMember = "member"
)
