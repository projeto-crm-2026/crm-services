package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (r *CreateRoleRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Permissions) == 0 {
		return fmt.Errorf("at least one permission is required")
	}
	return nil
}

type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (r *UpdateRoleRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

type AssignRoleRequest struct {
	RoleID uint `json:"role_id"`
}

func (r *AssignRoleRequest) Validate() error {
	if r.RoleID == 0 {
		return fmt.Errorf("role_id is required")
	}
	return nil
}

type RoleResponse struct {
	ID          uint                 `json:"id"`
	UUID        uuid.UUID            `json:"uuid"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   string               `json:"created_at"`
}

type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type PermissionCatalogResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func NewRoleResponse(role *entity.Role) RoleResponse {
	perms := make([]PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		perms[i] = PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		}
	}
	return RoleResponse{
		ID:          role.ID,
		UUID:        role.UUID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: perms,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func NewRoleListResponse(roles []entity.Role) []RoleResponse {
	resp := make([]RoleResponse, len(roles))
	for i := range roles {
		resp[i] = NewRoleResponse(&roles[i])
	}
	return resp
}

func NewPermissionCatalogResponse(perms []constant.PermissionDef) []PermissionCatalogResponse {
	resp := make([]PermissionCatalogResponse, len(perms))
	for i, p := range perms {
		resp[i] = PermissionCatalogResponse{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		}
	}
	return resp
}
