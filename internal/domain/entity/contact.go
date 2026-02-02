package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContactType string
type ContactStatus string
type ContactSource string

const (
	ContactTypePerson  ContactType = "person"
	ContactTypeCompany ContactType = "company"
)

const (
	ContactStatusLead      ContactStatus = "lead"
	ContactStatusQualified ContactStatus = "qualified"
	ContactStatusCustomer  ContactStatus = "customer"
	ContactStatusInactive  ContactStatus = "inactive"
	ContactStatusLost      ContactStatus = "lost"
)

const (
	ContactSourceWebsite       ContactSource = "website"
	ContactSourceReferral      ContactSource = "referral"
	ContactSourceEmail         ContactSource = "email"
	ContactSourcePaidAds       ContactSource = "paid_ads"
	ContactSourceOrganicSearch ContactSource = "organic_search"
	ContactSourceSocialMedia   ContactSource = "social_media"
	ContactSourceEvent         ContactSource = "event"
	ContactSourceOther         ContactSource = "other"
)

type Contact struct {
	gorm.Model

	UUID           uuid.UUID      `gorm:"type:uuid;not null;unique;default:gen_random_uuid()"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Type           ContactType    `gorm:"type:text"`
	FirstName      string         `gorm:"type:text"`
	LastName       string         `gorm:"type:text"`
	FullName       string         `gorm:"type:text;not null;index"`
	Email          sql.NullString `gorm:"type:text;index"`
	Phone          sql.NullString `gorm:"type:varchar(32)"`
	MobilePhone    sql.NullString `gorm:"type:varchar(32)"`
	AlternateEmail sql.NullString `gorm:"type:text"`
	CompanyName    sql.NullString `gorm:"type:text;index"`
	JobTitle       sql.NullString `gorm:"type:text"`
	Department     sql.NullString `gorm:"type:text"`
	Street         sql.NullString `gorm:"type:text"`
	Number         sql.NullString `gorm:"type:varchar(20)"`
	Complement     sql.NullString `gorm:"type:text"`
	District       sql.NullString `gorm:"type:text"`
	City           sql.NullString `gorm:"type:text"`
	State          sql.NullString `gorm:"type:text"`
	ZipCode        sql.NullString `gorm:"type:text"`
	Country        sql.NullString `gorm:"type:varchar(2);default:'BR'"`
	Status         ContactStatus  `gorm:"type:text;not null;default:'lead';index"`
	Source         ContactSource  `gorm:"type:text;index"`
	Tags           []string       `gorm:"type:text[]"`
	Notes          sql.NullString `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	AssignedToID *uint `gorm:"index"`
	AssignedTo   *User `gorm:"foreignKey:AssignedToID;references:ID;constraint:OnDelete:SET NULL"`

	CreatedByID  *uint `gorm:"index"`
	CreatedBy    *User `gorm:"foreignKey:CreatedByID;references:ID;constraint:OnDelete:SET NULL"`

	UpdatedByID  *uint `gorm:"index"`
	UpdatedBy    *User `gorm:"foreignKey:UpdatedByID;references:ID;constraint:OnDelete:SET NULL"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:UUID"`
}

func (Contact) TableName() string {
	return "contacts"
}

func (c *Contact) BeforeSave(tx *gorm.DB) error {
	if c.FullName == "" {
		if c.Type == ContactTypePerson {
			c.FullName = c.FirstName + " " + c.LastName
		} else {
			c.FullName = c.CompanyName.String
		}
	}
	return nil
}

func (c *Contact) IsLead() bool {
	return c.Status == ContactStatusLead || c.Status == ContactStatusQualified
}

func (c *Contact) IsCustomer() bool {
	return c.Status == ContactStatusCustomer
}

func (c *Contact) GetPrimaryEmail() string {
	if c.Email.Valid {
		return c.Email.String
	}
	return ""
}

func (c *Contact) GetPrimaryPhone() string {
	if c.MobilePhone.Valid {
		return c.MobilePhone.String
	}
	if c.Phone.Valid {
		return c.Phone.String
	}
	return ""
}
