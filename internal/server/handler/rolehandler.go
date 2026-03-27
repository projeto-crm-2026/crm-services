package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/roleservice"
)

type RoleHandler struct {
	service roleservice.RoleService
}

func NewRoleHandler(svc roleservice.RoleService) *RoleHandler {
	return &RoleHandler{service: svc}
}

func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	roles, err := h.service.ListRoles(r.Context(), *claims.OrganizationID)
	if err != nil {
		http.Error(w, "failed to list roles", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Roles retrieved successfully", model.NewRoleListResponse(roles)))
}

func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	var req model.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := h.service.CreateRole(r.Context(), *claims.OrganizationID, req.Name, req.Description, req.Permissions)
	if err != nil {
		if errors.Is(err, roleservice.ErrRoleNameTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Role created successfully", model.NewRoleResponse(role)))
}

func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	roleID, err := strconv.ParseUint(chi.URLParam(r, "roleID"), 10, 32)
	if err != nil {
		http.Error(w, "invalid role ID", http.StatusBadRequest)
		return
	}

	var req model.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := h.service.UpdateRole(r.Context(), uint(roleID), *claims.OrganizationID, req.Name, req.Description, req.Permissions)
	if err != nil {
		if errors.Is(err, roleservice.ErrRoleNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, roleservice.ErrCannotModifySystemRole) || errors.Is(err, roleservice.ErrRoleNameTaken) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Role updated successfully", model.NewRoleResponse(role)))
}

func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	roleID, err := strconv.ParseUint(chi.URLParam(r, "roleID"), 10, 32)
	if err != nil {
		http.Error(w, "invalid role ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteRole(r.Context(), uint(roleID), *claims.OrganizationID); err != nil {
		if errors.Is(err, roleservice.ErrRoleNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, roleservice.ErrCannotDeleteSystemRole) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	targetUserID, err := strconv.ParseUint(chi.URLParam(r, "userID"), 10, 32)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req model.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.AssignRole(r.Context(), *claims.OrganizationID, uint(targetUserID), req.RoleID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Role assigned successfully", nil))
}

func (h *RoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms := h.service.ListPermissions(r.Context())

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Permissions retrieved successfully", model.NewPermissionCatalogResponse(perms)))
}
