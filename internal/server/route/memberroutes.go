package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
)

func registerMemberRoutes(r chi.Router, cfg Config, perm RequirePermissionFunc) {
	r.With(perm(constant.PermMembersInvite)).Post("/members/invite", cfg.Handlers.User.InviteUser)
	r.With(perm(constant.PermMembersList)).Get("/members", cfg.Handlers.User.ListMembers)
	r.With(perm(constant.PermMembersManageRole)).Patch("/members/{userID}/role", cfg.Handlers.Role.AssignRole)
	r.With(perm(constant.PermMembersRemove)).Delete("/members/{userID}", cfg.Handlers.User.RemoveMember)
	r.With(perm(constant.PermMembersDeactivate)).Post("/members/{userID}/deactivate", cfg.Handlers.User.DeactivateMember)
	r.With(perm(constant.PermMembersDeactivate)).Post("/members/{userID}/reactivate", cfg.Handlers.User.ReactivateMember)
}
