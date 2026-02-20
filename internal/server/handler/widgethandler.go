package handler

import (
	"encoding/json"
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/widgetservice"
)

type WidgetHandler struct {
	service widgetservice.WidgetService
}

func NewWidgetHandler(service widgetservice.WidgetService) *WidgetHandler {
	return &WidgetHandler{service: service}
}

func (h *WidgetHandler) InitWidget(w http.ResponseWriter, r *http.Request) {
	widgetCtx, ok := middleware.GetWidgetContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.InitWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	session, err := h.service.InitSession(r.Context(), req.VisitorID, widgetCtx.UserID, widgetCtx.Domain)
	if err != nil {
		http.Error(w, "failed to initialize session", http.StatusInternalServerError)
		return
	}

	response := model.InitWidgetResponse{
		Token:     session.Token,
		VisitorID: session.VisitorID,
	}

	if req.ChatID != nil {
		chat, err := h.service.ResumeChat(r.Context(), *req.ChatID, session.VisitorID, widgetCtx.UserID)
		if err == nil {
			response.Chat = &model.ChatResponse{
				ID:     chat.ID,
				UUID:   chat.UUID,
				Status: string(chat.Status),
				Origin: chat.Origin,
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
