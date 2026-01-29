package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type CreateOrganizationRequest struct {
	Name       string                     `json:"name" validate:"required"`
	Slug       string                     `json:"slug" validate:"required"`
	Email      string                     `json:"email" validate:"email"`
	Phone      string                     `json:"phone"`
	Website    string                     `json:"website"`
	Street     string                     `json:"street"`
	Number     string                     `json:"number"`
	Complement string                     `json:"complement"`
	District   string                     `json:"district"`
	City       string                     `json:"city"`
	State      string                     `json:"state"`
	ZipCode    string                     `json:"zip_code"`
	Country    string                     `json:"country"`
	DocumentID      string                     `json:"tax_id"`
	Industry   string                     `json:"industry"`
	Plan       entity.OrganizationPlan    `json:"plan"`
	MaxUsers   int                        `json:"max_users"`
	MaxContacts int                       `json:"max_contacts"`
	Settings   map[string]interface{}     `json:"settings"`
}

func (r *CreateOrganizationRequest) ToEntity() *entity.Organization {
	return &entity.Organization{
		Name:        r.Name,
		Slug:        r.Slug,
		Email:       r.Email,
		Phone:       r.Phone,
		Website:     r.Website,
		DocumentID:       r.DocumentID,
		Industry:    r.Industry,
		Plan:        r.Plan,
		Settings:    r.Settings,
		IsActive:    true,
	}
}

type UpdateOrganizationRequest struct {
	Name       *string                 `json:"name"`
	Slug       *string                 `json:"slug"`
	Email      *string                 `json:"email"`
	Phone      *string                 `json:"phone"`
	Website    *string                 `json:"website"`
	Street     *string                 `json:"street"`
	Number     *string                 `json:"number"`
	Complement *string                 `json:"complement"`
	District   *string                 `json:"district"`
	City       *string                 `json:"city"`
	State      *string                 `json:"state"`
	ZipCode    *string                 `json:"zip_code"`
	Country    *string                 `json:"country"`
	DocumentID      *string                 `json:"tax_id"`
	Industry   *string                 `json:"industry"`
	Settings   map[string]interface{}  `json:"settings"`
}

func (r *UpdateOrganizationRequest) UpdateEntity(org *entity.Organization) {
	if r.Name != nil {
		org.Name = *r.Name
	}
	if r.Slug != nil {
		org.Slug = *r.Slug
	}
	if r.Email != nil {
		org.Email = *r.Email
	}
	if r.Phone != nil {
		org.Phone = *r.Phone
	}
	if r.Website != nil {
		org.Website = *r.Website
	}
	if r.DocumentID != nil {
		org.DocumentID = *r.DocumentID
	}
	if r.Industry != nil {
		org.Industry = *r.Industry
	}
	if r.Settings != nil {
		org.Settings = r.Settings
	}
}

type OrganizationResponse struct {
	ID                 uuid.UUID              `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Email              string                 `json:"email"`
	Phone              string                 `json:"phone"`
	Website            string                 `json:"website"`
	Street             string                 `json:"street"`
	Number             string                 `json:"number"`
	Complement         string                 `json:"complement"`
	District           string                 `json:"district"`
	City               string                 `json:"city"`
	State              string                 `json:"state"`
	ZipCode            string                 `json:"zip_code"`
	Country            string                 `json:"country"`
	DocumentID              string                 `json:"tax_id"`
	Industry           string                 `json:"industry"`
	Plan               entity.OrganizationPlan `json:"plan"`
	MaxUsers           int                    `json:"max_users"`
	MaxContacts        int                    `json:"max_contacts"`
	SubscriptionEndsAt *time.Time             `json:"subscription_ends_at,omitempty"`
	Settings           map[string]interface{} `json:"settings"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

func NewOrganizationResponse(org *entity.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		ID:                 org.ID,
		Name:               org.Name,
		Slug:               org.Slug,
		Email:              org.Email,
		Phone:              org.Phone,
		Website:            org.Website,
		DocumentID:              org.DocumentID,
		Industry:           org.Industry,
		Plan:               org.Plan,
		Settings:           org.Settings,
		IsActive:           org.IsActive,
		CreatedAt:          org.CreatedAt,
		UpdatedAt:          org.UpdatedAt,
	}
}
