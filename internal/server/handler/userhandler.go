package handler

import (
	"encoding/json"
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/userservice"
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
		defer r.Body.Close()
		return
	}
	defer r.Body.Close()

	if err := request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.service.LoginUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.AuthResponse{
		User: model.NewUserResponse(user),
	})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		defer r.Body.Close()
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.service.RegisterUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(model.AuthResponse{
		User: model.NewUserResponse(user),
	})
}
