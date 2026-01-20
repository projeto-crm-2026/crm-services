package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    uint      `json:"id"`
	UUID  uuid.UUID `json:"uuid"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

func (r RegisterRequest) Validate() error {
	if r.Name == "" || r.Email == "" || r.Password == "" {
		return fmt.Errorf("name, email and password are required")
	}
	return nil
}

func (r LoginRequest) Validate() error {
	if r.Email == "" || r.Password == "" {
		return fmt.Errorf("email and password are required")
	}
	return nil
}

func NewUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		UUID:  u.UUID,
		Name:  u.Name,
		Email: u.Email,
	}
}
