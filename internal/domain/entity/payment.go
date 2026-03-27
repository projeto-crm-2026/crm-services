package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentApproved PaymentStatus = "approved"
	PaymentPending  PaymentStatus = "pending"
	PaymentRejected PaymentStatus = "rejected"
	PaymentRefunded PaymentStatus = "refunded"
)

type Payment struct {
	gorm.Model

	UUID            uuid.UUID     `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID  uuid.UUID     `gorm:"type:uuid;not null;index"`
	SubscriptionID  uint          `gorm:"not null;index"`
	MpPaymentID     string        `gorm:"type:text;not null;unique;index"`
	Status          PaymentStatus `gorm:"type:text;not null"`
	StatusDetail    string        `gorm:"type:text"`
	AmountCents     int           `gorm:"not null"`
	Currency        string        `gorm:"type:text;not null;default:'BRL'"`
	PaymentMethod   string        `gorm:"type:text"`
	Description     string        `gorm:"type:text"`
	PaidAt          *time.Time    `gorm:"type:timestamptz"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Payment) TableName() string { return "payments" }
