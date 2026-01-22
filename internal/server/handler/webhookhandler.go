package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/webhookservice"
)

type WebhookHandler struct {
	service webhookservice.WebhookService
}

// NewWebhookHandler returns a WebhookHandler initialized with the provided webhook service.
func NewWebhookHandler(service webhookservice.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// outgoing webhooks

func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	var req model.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, constant.InvalidPayload, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.URL == "" || len(req.Events) == 0 {
		http.Error(w, constant.URLAndEventsRequired, http.StatusBadRequest)
		return
	}

	webhook, err := h.service.CreateWebhook(r.Context(), claims.UserID, req.Name, req.URL, req.Events)
	if err != nil {
		http.Error(w, constant.FailedToCreateWebhook, http.StatusInternalServerError)
		return
	}

	var events []string
	json.Unmarshal([]byte(webhook.Events), &events)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.WebhookResponse{
		ID:        webhook.ID,
		Name:      webhook.Name,
		URL:       webhook.URL,
		Secret:    webhook.Secret, // apenas na criação
		Events:    events,
		IsActive:  webhook.IsActive,
		FailCount: webhook.FailCount,
		CreatedAt: webhook.CreatedAt.Format(constant.DateFormat),
	})
}

func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	webhooks, err := h.service.ListWebhooks(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, constant.FailedToListWebhooks, http.StatusInternalServerError)
		return
	}

	response := make([]model.WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		var events []string
		json.Unmarshal([]byte(wh.Events), &events)

		var lastUsed *string
		if wh.LastUsedAt != nil {
			t := wh.LastUsedAt.Format(constant.DateFormat)
			lastUsed = &t
		}

		response[i] = model.WebhookResponse{
			ID:         wh.ID,
			Name:       wh.Name,
			URL:        wh.URL,
			Events:     events,
			IsActive:   wh.IsActive,
			FailCount:  wh.FailCount,
			LastUsedAt: lastUsed,
			CreatedAt:  wh.CreatedAt.Format(constant.DateFormat),
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	webhookID, err := strconv.ParseUint(chi.URLParam(r, "webhookID"), 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidToken, http.StatusBadRequest)
		return
	}

	webhook, err := h.service.GetWebhook(r.Context(), claims.UserID, uint(webhookID))
	if err != nil {
		http.Error(w, constant.WebhookNotFound, http.StatusNotFound)
		return
	}

	var events []string
	json.Unmarshal([]byte(webhook.Events), &events)

	var lastUsed *string
	if webhook.LastUsedAt != nil {
		t := webhook.LastUsedAt.Format(constant.DateFormat)
		lastUsed = &t
	}

	json.NewEncoder(w).Encode(model.WebhookResponse{
		ID:         webhook.ID,
		Name:       webhook.Name,
		URL:        webhook.URL,
		Events:     events,
		IsActive:   webhook.IsActive,
		FailCount:  webhook.FailCount,
		LastUsedAt: lastUsed,
		CreatedAt:  webhook.CreatedAt.Format(constant.DateFormat),
	})
}

func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	webhookID, err := strconv.ParseUint(chi.URLParam(r, "webhookID"), 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidWebhookId, http.StatusBadRequest)
		return
	}

	var req model.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, constant.InvalidPayload, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.service.UpdateWebhook(r.Context(), claims.UserID, uint(webhookID), req.Name, req.URL, req.Events, req.IsActive)
	if err != nil {
		http.Error(w, constant.FailedToUpdateWebhook, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	webhookID, err := strconv.ParseUint(chi.URLParam(r, "webhookID"), 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidWebhookId, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteWebhook(r.Context(), claims.UserID, uint(webhookID)); err != nil {
		http.Error(w, constant.FailedToDeleteWebhook, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) GetWebhookLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	webhookID, err := strconv.ParseUint(chi.URLParam(r, "webhookID"), 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidWebhookId, http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	logs, err := h.service.GetWebhookLogs(r.Context(), claims.UserID, uint(webhookID), limit)
	if err != nil {
		http.Error(w, constant.FailedToGetLogs, http.StatusInternalServerError)
		return
	}

	response := make([]model.WebhookLogResponse, len(logs))
	for i, log := range logs {
		response[i] = model.WebhookLogResponse{
			ID:           log.ID,
			EventType:    log.EventType,
			ResponseCode: log.ResponseCode,
			Error:        log.Error,
			Duration:     log.Duration,
			CreatedAt:    log.CreatedAt.Format(constant.DateFormat),
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *WebhookHandler) GetAvailableEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(model.AvailableWebhookEvents)
}

// incoming webhook tokens

func (h *WebhookHandler) CreateIncomingToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	var req model.CreateIncomingTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, constant.InvalidPayload, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	token, err := h.service.CreateIncomingToken(r.Context(), claims.UserID, req.Name)
	if err != nil {
		http.Error(w, constant.FailedToCreateToken, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.IncomingTokenResponse{
		ID:        token.ID,
		Token:     token.Token, // apenas na criação
		Name:      token.Name,
		IsActive:  token.IsActive,
		CreatedAt: token.CreatedAt.Format(constant.DateFormat),
	})
}

func (h *WebhookHandler) ListIncomingTokens(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	tokens, err := h.service.ListIncomingTokens(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, constant.FailedToListTokens, http.StatusInternalServerError)
		return
	}

	response := make([]model.IncomingTokenResponse, len(tokens))
	for i, t := range tokens {
		var lastUsed *string
		if t.LastUsedAt != nil {
			lu := t.LastUsedAt.Format(constant.DateFormat)
			lastUsed = &lu
		}

		response[i] = model.IncomingTokenResponse{
			ID:         t.ID,
			Name:       t.Name,
			IsActive:   t.IsActive,
			LastUsedAt: lastUsed,
			CreatedAt:  t.CreatedAt.Format(constant.DateFormat),
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *WebhookHandler) DeleteIncomingToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}

	tokenID, err := strconv.ParseUint(chi.URLParam(r, "tokenID"), 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidTokenId, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteIncomingToken(r.Context(), claims.UserID, uint(tokenID)); err != nil {
		http.Error(w, constant.FailedToDeleteToken, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// incoming webhook

func (h *WebhookHandler) HandleIncomingWebhook(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Webhook-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	if token == "" {
		http.Error(w, constant.MissingWebhookToken, http.StatusUnauthorized)
		return
	}

	var req model.IncomingWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, constant.InvalidPayload, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	payload := &webhookservice.IncomingWebhookPayload{
		Action:  req.Action,
		ChatID:  req.ChatID,
		Content: req.Content,
		Data:    req.Data,
	}

	if err := h.service.ProcessIncomingWebhook(r.Context(), token, payload); err != nil {
		statusCode, message := mapWebhookError(err)
		http.Error(w, message, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// mapWebhookError maps known webhookservice errors to an HTTP status code and a client-facing message.
// For unmapped errors it returns HTTP 500 and the error's message.
func mapWebhookError(err error) (int, string) {
	errorMap := map[error]struct {
		code    int
		message string
	}{
		webhookservice.ErrInvalidToken:  {http.StatusUnauthorized, constant.InvalidToken},
		webhookservice.ErrInvalidAction: {http.StatusBadRequest, constant.InvalidAction},
		webhookservice.ErrChatNotFound:  {http.StatusNotFound, constant.InvalidChatId},
	}

	if mapped, ok := errorMap[err]; ok {
		return mapped.code, mapped.message
	}

	return http.StatusInternalServerError, err.Error()
}