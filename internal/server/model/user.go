package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type RegisterRequest struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	OrganizationName string `json:"organization_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type InviteUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type AcceptInviteRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID             uint       `json:"id"`
	UUID           uuid.UUID  `json:"uuid"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	InviteToken    string     `json:"invite_token,omitempty"`
	InviteExpiry   string     `json:"invite_expiry,omitempty"`
	Status         string     `json:"status"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

type MemberListResponse struct {
	Members []UserResponse `json:"members"`
}

func (r RegisterRequest) Validate() error {
	if r.Name == "" || r.Email == "" || r.Password == "" || r.OrganizationName == "" {
		return fmt.Errorf("name, email, password and organization_name are required")
	}
	return nil
}

func (r LoginRequest) Validate() error {
	if r.Email == "" || r.Password == "" {
		return fmt.Errorf("email and password are required")
	}
	return nil
}

func (r InviteUserRequest) Validate() error {
	if r.Name == "" || r.Email == "" {
		return fmt.Errorf("name and email are required")
	}
	return nil
}

func (r AcceptInviteRequest) Validate() error {
	if r.Token == "" || r.Password == "" {
		return fmt.Errorf("token and password are required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func NewUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:             u.ID,
		UUID:           u.UUID,
		Name:           u.Name,
		Email:          u.Email,
		Role:           string(u.Role),
		Status:         string(u.Status),
		OrganizationID: u.OrganizationID,
	}
}

func NewMemberListResponse(users []entity.User) MemberListResponse {
	members := make([]UserResponse, len(users))
	for i, u := range users {
		members[i] = NewUserResponse(&u)
	}
	return MemberListResponse{Members: members}
}
