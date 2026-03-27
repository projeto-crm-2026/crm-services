package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerRoleRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermRolesRead)).Get("/roles", cfg.Handlers.Role.ListRoles)
	r.With(perm(constant.PermRolesManage)).Post("/roles", cfg.Handlers.Role.CreateRole)
	r.With(perm(constant.PermRolesManage)).Patch("/roles/{roleID}", cfg.Handlers.Role.UpdateRole)
	r.With(perm(constant.PermRolesManage)).Delete("/roles/{roleID}", cfg.Handlers.Role.DeleteRole)
	r.Get("/permissions", cfg.Handlers.Role.ListPermissions)
}
