package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type CreateContactRequest struct {
	Type           string   `json:"type"`
	OrganizationID string   `json:"organization_id"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	CompanyName    string   `json:"company_name"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone"`
	MobilePhone    string   `json:"mobile_phone"`
	JobTitle       string   `json:"job_title"`
	Department     string   `json:"department"`
	Street         string   `json:"street"`
	Number         string   `json:"number"`
	Complement     string   `json:"complement"`
	District       string   `json:"district"`
	City           string   `json:"city"`
	State          string   `json:"state"`
	ZipCode        string   `json:"zip_code"`
	Country        string   `json:"country"`
	Status         string   `json:"status"`
	Source         string   `json:"source"`
	Tags           []string `json:"tags"`
	Notes          string   `json:"notes"`
	AssignedToID   *uint    `json:"assigned_to_id"`
}

type UpdateContactRequest struct {
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	CompanyName  string   `json:"company_name"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone"`
	MobilePhone  string   `json:"mobile_phone"`
	JobTitle     string   `json:"job_title"`
	Department   string   `json:"department"`
	Street       string   `json:"street"`
	Number       string   `json:"number"`
	Complement   string   `json:"complement"`
	District     string   `json:"district"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	ZipCode      string   `json:"zip_code"`
	Country      string   `json:"country"`
	Status       string   `json:"status"`
	Source       string   `json:"source"`
	Tags         []string `json:"tags"`
	Notes        string   `json:"notes"`
	AssignedToID *uint    `json:"assigned_to_id"`
}

type ContactResponse struct {
	ID             uuid.UUID `json:"id"`
	Type           string    `json:"type"`
	FullName       string    `json:"full_name"`
	FirstName      string    `json:"first_name,omitempty"`
	LastName       string    `json:"last_name,omitempty"`
	Email          *string   `json:"email,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	MobilePhone    *string   `json:"mobile_phone,omitempty"`
	AlternateEmail *string   `json:"alternate_email,omitempty"`
	CompanyName    *string   `json:"company_name,omitempty"`
	JobTitle       *string   `json:"job_title,omitempty"`
	Department     *string   `json:"department,omitempty"`
	Address        Address   `json:"address"`
	Status         string    `json:"status"`
	Source         string    `json:"source,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AssignedToID   *uint     `json:"assigned_to_id,omitempty"`
	CreatedByID    *uint     `json:"created_by_id,omitempty"`
}

type Address struct {
	Street     *string `json:"street,omitempty"`
	Number     *string `json:"number,omitempty"`
	Complement *string `json:"complement,omitempty"`
	District   *string `json:"district,omitempty"`
	City       *string `json:"city,omitempty"`
	State      *string `json:"state,omitempty"`
	ZipCode    *string `json:"zip_code,omitempty"`
	Country    *string `json:"country,omitempty"`
}

type PaginatedContactResponse struct {
	Data       []ContactResponse `json:"data"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"total_pages"`
}

func (r CreateContactRequest) Validate() error {
	if r.Type == "" {
		return fmt.Errorf("contact type is required")
	}

	if entity.ContactType(r.Type) == entity.ContactTypePerson {
		if r.FirstName == "" {
			return fmt.Errorf("first name is required for person contacts")
		}
	} else if entity.ContactType(r.Type) == entity.ContactTypeCompany {
		if r.CompanyName == "" {
			return fmt.Errorf("company name is required for company contacts")
		}
	} else {
		return fmt.Errorf("invalid contact type: %s", r.Type)
	}

	return nil
}

func (r CreateContactRequest) ToEntity() *entity.Contact {
	orgID, _ := uuid.Parse(r.OrganizationID)

	c := &entity.Contact{
		OrganizationID: orgID,
		Type:           entity.ContactType(r.Type),
		FirstName:      r.FirstName,
		LastName:       r.LastName,
		Status:         entity.ContactStatus(r.Status),
		Source:         entity.ContactSource(r.Source),
		Tags:           r.Tags,
		AssignedToID:   r.AssignedToID,
	}
	if c.Status == "" {
		c.Status = entity.ContactStatusLead
	}
	c.CompanyName = toNullString(r.CompanyName)
	c.Email = toNullString(r.Email)
	c.Phone = toNullString(r.Phone)
	c.MobilePhone = toNullString(r.MobilePhone)
	c.JobTitle = toNullString(r.JobTitle)
	c.Department = toNullString(r.Department)
	c.Notes = toNullString(r.Notes)
	c.Street = toNullString(r.Street)
	c.Number = toNullString(r.Number)
	c.Complement = toNullString(r.Complement)
	c.District = toNullString(r.District)
	c.City = toNullString(r.City)
	c.State = toNullString(r.State)
	c.ZipCode = toNullString(r.ZipCode)
	c.Country = toNullString(r.Country)

	return c
}

func (r UpdateContactRequest) UpdateEntity(c *entity.Contact) {
	if r.FirstName != "" {
		c.FirstName = r.FirstName
	}
	if r.LastName != "" {
		c.LastName = r.LastName
	}
	if r.Status != "" {
		c.Status = entity.ContactStatus(r.Status)
	}
	if r.Tags != nil {
		c.Tags = r.Tags
	} // se passar um array vazio limpa as tags nesse caso
	if r.AssignedToID != nil {
		c.AssignedToID = r.AssignedToID
	}

	if r.Email != "" {
		c.Email = toNullString(r.Email)
	}
	if r.CompanyName != "" {
		c.CompanyName = toNullString(r.CompanyName)
	}
}

func NewContactResponse(c *entity.Contact) ContactResponse {
	return ContactResponse{
		ID:             c.UUID,
		Type:           string(c.Type),
		FullName:       c.FullName,
		FirstName:      c.FirstName,
		LastName:       c.LastName,
		Email:          fromNullString(c.Email),
		Phone:          fromNullString(c.Phone),
		MobilePhone:    fromNullString(c.MobilePhone),
		AlternateEmail: fromNullString(c.AlternateEmail),
		CompanyName:    fromNullString(c.CompanyName),
		JobTitle:       fromNullString(c.JobTitle),
		Department:     fromNullString(c.Department),
		Status:         string(c.Status),
		Source:         string(c.Source),
		Tags:           c.Tags,
		Notes:          fromNullString(c.Notes),
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		AssignedToID:   c.AssignedToID,
		CreatedByID:    c.CreatedByID,
		Address: Address{
			Street:     fromNullString(c.Street),
			Number:     fromNullString(c.Number),
			Complement: fromNullString(c.Complement),
			District:   fromNullString(c.District),
			City:       fromNullString(c.City),
			State:      fromNullString(c.State),
			ZipCode:    fromNullString(c.ZipCode),
			Country:    fromNullString(c.Country),
		},
	}
}

func NewContactListResponse(contacts []*entity.Contact) []ContactResponse {
	list := make([]ContactResponse, len(contacts))
	for i, contact := range contacts {
		list[i] = NewContactResponse(contact)
	}
	return list
}

func NewPaginatedContactResponse(result *repo.PaginatedResult[entity.Contact]) PaginatedContactResponse {
	responseList := NewContactListResponse(result.Data)

	return PaginatedContactResponse{
		Data:       responseList,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false} // ou sql.NullString
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}
