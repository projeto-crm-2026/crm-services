package entity

import (
	"time"

	"gorm.io/gorm"
)

type ParticipantRole string

const (
	ParticipantRoleAgent   ParticipantRole = "agent"
	ParticipantRoleVisitor ParticipantRole = "visitor"
)

type ChatParticipant struct {
	gorm.Model

	ChatID    uint            `gorm:"not null;index"`
	UserID    *uint           `gorm:"index"`
	VisitorID string          `gorm:"type:text"`
	Role      ParticipantRole `gorm:"type:text;not null"`
	JoinedAt  time.Time       `gorm:"autoCreateTime"`
	LeftAt    *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Chat Chat  `gorm:"foreignKey:ChatID"`
	User *User `gorm:"foreignKey:UserID"`
}

func (ChatParticipant) TableName() string { return "chat_participant" }
