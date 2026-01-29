package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationPlan string

const (
	OrganizationPlanFree         OrganizationPlan = "free"
	OrganizationPlanStarter      OrganizationPlan = "starter"
	OrganizationPlanProfessional OrganizationPlan = "professional"
)

type Organization struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string                 `gorm:"type:text;not null;index"`
	Slug       string                 `gorm:"type:text;not null;unique;index"`
	Email      string                 `gorm:"type:text;index"`
	Phone      string                 `gorm:"type:varchar(32)"`
	Website    string                 `gorm:"type:text"`
	DocumentID string                 `gorm:"type:text;index"` // cnpj / cpf / etc...
	Industry   string                 `gorm:"type:text"`
	Plan       OrganizationPlan       `gorm:"type:text;not null;default:'free'"`
	Settings   map[string]interface{} `gorm:"type:jsonb;default:'{}'"`
	IsActive   bool                   `gorm:"default:true;index"`
	CreatedAt  time.Time              `gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt         `gorm:"index"`

	Users    []User    `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Contacts []Contact `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
}

func (Organization) TableName() string {
	return "organizations"
}

func (o *Organization) IsPremium() bool {
	return o.Plan != OrganizationPlanFree
}
