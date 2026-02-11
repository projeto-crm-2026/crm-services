package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusPending UserStatus = "pending"
)

type User struct {
	gorm.Model

	UUID           uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;index"`
	Name           string         `gorm:"type:text;not null"`
	Email          string         `gorm:"type:text;not null;unique"`
	PasswordHash   string         `gorm:"type:text"`
	Role           UserRole       `gorm:"type:text;not null;default:'admin'"`
	Status         UserStatus     `gorm:"type:text;not null;default:'active'"`
	InviteToken    *string        `gorm:"type:text;unique"`
	InviteExpiry   *time.Time     `gorm:"type:timestamptz"`
	InvitedBy      *uint          `gorm:"index"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:UUID"`
}

func (User) TableName() string { return "user" }

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsPending() bool {
	return u.Status == StatusPending
}
