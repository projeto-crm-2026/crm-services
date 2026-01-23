package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
)

func (h *WidgetHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if req.Domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}

	apiKey, err := h.service.CreateAPIKey(r.Context(), claims.UserID, req.Name, req.Domain)
	if err != nil {
		http.Error(w, "failed to create API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.APIKeyResponse{
		ID:        apiKey.ID,
		PublicKey: apiKey.PublicKey,
		SecretKey: apiKey.SecretKey, // only on creation
		Name:      apiKey.Name,
		Domain:    apiKey.Domain,
		IsActive:  apiKey.IsActive,
	})
}

func (h *WidgetHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	apiKeys, err := h.service.ListAPIKeys(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to list API keys", http.StatusInternalServerError)
		return
	}

	response := make([]model.APIKeyResponse, len(apiKeys))
	for i, key := range apiKeys {
		response[i] = model.APIKeyResponse{
			ID:        key.ID,
			PublicKey: key.PublicKey,
			Name:      key.Name,
			Domain:    key.Domain,
			IsActive:  key.IsActive,
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *WidgetHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	keyIDStr := chi.URLParam(r, "keyID")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid key ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteAPIKey(r.Context(), claims.UserID, uint(keyID)); err != nil {
		http.Error(w, "failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
