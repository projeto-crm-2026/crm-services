package handler

import (
	"encoding/json"
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
	"github.com/projeto-crm-2026/crm-services/pkg/https"
)

type UserHandler struct {
	service userservice.UserService
}

func NewUserHandler(svc userservice.UserService) *UserHandler {
	return &UserHandler{service: svc}
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

	user, err := h.service.InviteUser(r.Context(), claims.UserID, req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	members, err := h.service.ListOrganizationMembers(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewMemberListResponse(members))
}
