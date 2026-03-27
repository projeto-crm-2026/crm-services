package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/planservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/roleservice"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
	"github.com/projeto-crm-2026/crm-services/pkg/https"
)

type UserHandler struct {
	service userservice.UserService
	planSvc planservice.PlanService
	roleSvc roleservice.RoleService
}

func NewUserHandler(svc userservice.UserService, planSvc planservice.PlanService, roleSvc roleservice.RoleService) *UserHandler {
	return &UserHandler{service: svc, planSvc: planSvc, roleSvc: roleSvc}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.service.LoginUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	secure := https.IsHTTPS(r)
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.AuthResponse{
		User: model.NewUserResponse(user),
	})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.service.RegisterUser(r.Context(), req.Name, req.Email, req.Password, req.OrganizationName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.OrganizationID != nil {
		ownerRoleID, err := h.roleSvc.CreateSystemRoles(r.Context(), *user.OrganizationID)
		if err == nil {
			_ = h.roleSvc.AssignRole(r.Context(), *user.OrganizationID, user.ID, ownerRoleID)
		}
	}

	secure := https.IsHTTPS(r)
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.AuthResponse{
		User: model.NewUserResponse(user),
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	secure := https.IsHTTPS(r)
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func (h *UserHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if claims.OrganizationID != nil {
		if err := h.planSvc.CheckMemberLimit(r.Context(), *claims.OrganizationID); err != nil {
			if errors.Is(err, constant.ErrMemberLimitReached) {
				http.Error(w, constant.MemberLimitReached, http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	user, err := h.service.InviteUser(r.Context(), claims.UserID, req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if claims.OrganizationID != nil {
		if memberRole, err := h.roleSvc.GetMemberRole(r.Context(), *claims.OrganizationID); err == nil {
			_ = h.roleSvc.AssignRole(r.Context(), *claims.OrganizationID, user.ID, memberRole)
		}
		if usageResp, err := h.planSvc.GetOrganizationUsage(r.Context(), *claims.OrganizationID); err == nil {
			h.planSvc.NotifyUsageWarnings(r.Context(), *claims.OrganizationID, "membros", usageResp.Usage.Members.Current, usageResp.Usage.Members.Limit)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("User invited successfully", model.NewUserResponse(user)))
}

func (h *UserHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req model.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.AcceptInvite(r.Context(), req.Token, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Account activated successfully", model.NewUserResponse(user)))
}

func (h *UserHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	members, err := h.service.ListOrganizationMembersWithRole(r.Context(), *claims.OrganizationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewMemberListResponse(members))
}

func (h *UserHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	targetUserID, err := parseUserIDParam(r)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if isOwner, _ := h.roleSvc.IsOwner(r.Context(), *claims.OrganizationID, targetUserID); isOwner {
		http.Error(w, "the organization owner cannot be removed", http.StatusForbidden)
		return
	}

	if err := h.service.RemoveMember(r.Context(), *claims.OrganizationID, targetUserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Member removed successfully", nil))
}

func (h *UserHandler) DeactivateMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	targetUserID, err := parseUserIDParam(r)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if isOwner, _ := h.roleSvc.IsOwner(r.Context(), *claims.OrganizationID, targetUserID); isOwner {
		http.Error(w, "the organization owner cannot be deactivated", http.StatusForbidden)
		return
	}

	if err := h.service.DeactivateMember(r.Context(), *claims.OrganizationID, targetUserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Member deactivated successfully", nil))
}

func (h *UserHandler) ReactivateMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	targetUserID, err := parseUserIDParam(r)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.service.ReactivateMember(r.Context(), *claims.OrganizationID, targetUserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Member reactivated successfully", nil))
}

func parseUserIDParam(r *http.Request) (uint, error) {
	idStr := r.PathValue("userID")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
