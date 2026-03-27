package roleservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type RoleService interface {
	CreateRole(ctx context.Context, orgID uuid.UUID, name, description string, permissionNames []string) (*entity.Role, error)
	UpdateRole(ctx context.Context, roleID uint, orgID uuid.UUID, name, description string, permissionNames []string) (*entity.Role, error)
	DeleteRole(ctx context.Context, roleID uint, orgID uuid.UUID) error
	ListRoles(ctx context.Context, orgID uuid.UUID) ([]entity.Role, error)
	GetRole(ctx context.Context, roleID uint) (*entity.Role, error)
	AssignRole(ctx context.Context, orgID uuid.UUID, targetUserID uint, roleID uint) error
	HasPermission(ctx context.Context, userID uint, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	CreateSystemRoles(ctx context.Context, orgID uuid.UUID) (ownerRoleID uint, err error)
	GetMemberRole(ctx context.Context, orgID uuid.UUID) (uint, error)
	IsOwner(ctx context.Context, orgID uuid.UUID, userID uint) (bool, error)
	ListPermissions(ctx context.Context) []constant.PermissionDef
}

type roleService struct {
	roleRepo repo.RoleRepo
	userRepo repo.UserRepo
	logger   *slog.Logger
}

func NewRoleService(roleRepo repo.RoleRepo, userRepo repo.UserRepo, logger *slog.Logger) RoleService {
	return &roleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *roleService) CreateRole(ctx context.Context, orgID uuid.UUID, name, description string, permissionNames []string) (*entity.Role, error) {
	// check name doesn't conflict with system roles
	existing, err := s.roleRepo.GetByOrgAndName(ctx, orgID, name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}
	if existing != nil {
		return nil, ErrRoleNameTaken
	}

	role, err := s.roleRepo.Create(ctx, &entity.Role{
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		IsSystem:       false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	if err := s.setPermissionsByName(ctx, role.ID, permissionNames); err != nil {
		return nil, err
	}

	perms, _ := s.roleRepo.GetRolePermissions(ctx, role.ID)
	role.Permissions = perms

	s.logger.Info("custom role created", "roleID", role.ID, "orgID", orgID, "name", name)
	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, roleID uint, orgID uuid.UUID, name, description string, permissionNames []string) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	if role.OrganizationID != orgID {
		return nil, ErrRoleNotFound
	}
	if role.IsSystem {
		return nil, ErrCannotModifySystemRole
	}

	if role.Name != name {
		existing, err := s.roleRepo.GetByOrgAndName(ctx, orgID, name)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		if existing != nil {
			return nil, ErrRoleNameTaken
		}
	}

	role.Name = name
	role.Description = description
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	if err := s.setPermissionsByName(ctx, role.ID, permissionNames); err != nil {
		return nil, err
	}

	perms, _ := s.roleRepo.GetRolePermissions(ctx, role.ID)
	role.Permissions = perms

	s.logger.Info("custom role updated", "roleID", roleID, "name", name)
	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID uint, orgID uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}
	if role.OrganizationID != orgID {
		return ErrRoleNotFound
	}
	if role.IsSystem {
		return ErrCannotDeleteSystemRole
	}

	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logger.Info("custom role deleted", "roleID", roleID)
	return nil
}

func (s *roleService) ListRoles(ctx context.Context, orgID uuid.UUID) ([]entity.Role, error) {
	roles, err := s.roleRepo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	for i := range roles {
		perms, _ := s.roleRepo.GetRolePermissions(ctx, roles[i].ID)
		roles[i].Permissions = perms
	}

	return roles, nil
}

func (s *roleService) GetRole(ctx context.Context, roleID uint) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	perms, _ := s.roleRepo.GetRolePermissions(ctx, role.ID)
	role.Permissions = perms
	return role, nil
}

func (s *roleService) AssignRole(ctx context.Context, orgID uuid.UUID, targetUserID uint, roleID uint) error {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}
	if role.OrganizationID != orgID {
		return ErrRoleNotFound
	}

	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.OrganizationID == nil || *user.OrganizationID != orgID {
		return fmt.Errorf("user does not belong to this organization")
	}

	return s.userRepo.UpdateRoleID(ctx, targetUserID, roleID)
}

func (s *roleService) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	perms, err := s.roleRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}

func (s *roleService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	return s.roleRepo.GetUserPermissions(ctx, userID)
}

func (s *roleService) CreateSystemRoles(ctx context.Context, orgID uuid.UUID) (uint, error) {
	systemRoles := []struct {
		name        string
		description string
	}{
		{entity.SystemRoleOwner, "Full access to all organization resources"},
		{entity.SystemRoleAdmin, "Administrative access (cannot delete organization)"},
		{entity.SystemRoleMember, "Basic member access"},
	}

	var ownerRoleID uint
	for _, sr := range systemRoles {
		role, err := s.roleRepo.Create(ctx, &entity.Role{
			OrganizationID: orgID,
			Name:           sr.name,
			Description:    sr.description,
			IsSystem:       true,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create system role %s: %w", sr.name, err)
		}

		permNames := constant.SystemRolePermissions[sr.name]
		if err := s.setPermissionsByName(ctx, role.ID, permNames); err != nil {
			return 0, fmt.Errorf("failed to set permissions for %s: %w", sr.name, err)
		}

		if sr.name == entity.SystemRoleOwner {
			ownerRoleID = role.ID
		}
	}

	s.logger.Info("system roles created for organization", "orgID", orgID)
	return ownerRoleID, nil
}

func (s *roleService) IsOwner(ctx context.Context, orgID uuid.UUID, userID uint) (bool, error) {
	ownerRole, err := s.roleRepo.GetByOrgAndName(ctx, orgID, entity.SystemRoleOwner)
	if err != nil {
		return false, err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.RoleID != nil && *user.RoleID == ownerRole.ID, nil
}

func (s *roleService) GetMemberRole(ctx context.Context, orgID uuid.UUID) (uint, error) {
	role, err := s.roleRepo.GetByOrgAndName(ctx, orgID, entity.SystemRoleMember)
	if err != nil {
		return 0, fmt.Errorf("member role not found: %w", err)
	}
	return role.ID, nil
}

func (s *roleService) ListPermissions(_ context.Context) []constant.PermissionDef {
	return constant.PermissionCatalog
}

func (s *roleService) setPermissionsByName(ctx context.Context, roleID uint, names []string) error {
	if len(names) == 0 {
		return s.roleRepo.SetRolePermissions(ctx, roleID, nil)
	}

	ids, err := s.resolvePermissionIDs(ctx, names)
	if err != nil {
		return err
	}

	return s.roleRepo.SetRolePermissions(ctx, roleID, ids)
}

func (s *roleService) resolvePermissionIDs(ctx context.Context, names []string) ([]uint, error) {
	return s.roleRepo.GetPermissionIDsByNames(ctx, names)
}
