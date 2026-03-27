package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionPastDue   SubscriptionStatus = "past_due"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionExpired   SubscriptionStatus = "expired"
)

var ValidTransitions = map[SubscriptionStatus][]SubscriptionStatus{
	SubscriptionActive:    {SubscriptionPastDue, SubscriptionCancelled},
	SubscriptionPastDue:   {SubscriptionActive, SubscriptionCancelled, SubscriptionExpired},
	SubscriptionCancelled: {SubscriptionExpired, SubscriptionActive},
	SubscriptionExpired:   {SubscriptionActive},
}

func (s SubscriptionStatus) CanTransitionTo(target SubscriptionStatus) bool {
	for _, valid := range ValidTransitions[s] {
		if valid == target {
			return true
		}
	}
	return false
}

type Subscription struct {
	gorm.Model

	UUID               uuid.UUID          `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID     uuid.UUID          `gorm:"type:uuid;not null;index"`
	PlanID             uint               `gorm:"not null;index"`
	Status             SubscriptionStatus `gorm:"type:text;not null;default:'active'"`
	CurrentPeriodStart *time.Time         `gorm:"type:timestamptz"`
	CurrentPeriodEnd   *time.Time         `gorm:"type:timestamptz"`
	CancelAtPeriodEnd  bool               `gorm:"default:false"`
	MpSubscriptionID   *string            `gorm:"type:text;unique;index"`
	CancelledAt        *time.Time         `gorm:"type:timestamptz"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt     `gorm:"index"`
}

func (Subscription) TableName() string { return "subscriptions" }
