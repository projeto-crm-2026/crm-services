package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/server/websocket"
	"github.com/projeto-crm-2026/crm-services/internal/service/chatservice"
)

type ChatHandler struct {
	hub     *websocket.Hub
	service chatservice.ChatService
}

type messageSaverAdapter struct {
	service chatservice.ChatService
}

func NewChatHandler(hub *websocket.Hub, service chatservice.ChatService) *ChatHandler {
	return &ChatHandler{hub: hub, service: service}
}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidChatId, http.StatusBadRequest)
		return
	}

	var userID *uint
	visitorID := ""

	if claims, ok := middleware.GetUserFromContext(r.Context()); ok {
		userID = &claims.UserID
	} else {
		visitorID = r.URL.Query().Get("visitor_id")
		if visitorID == "" {
			http.Error(w, "visitor_id required", http.StatusBadRequest)
			return
		}
	}

	saver := &messageSaverAdapter{service: h.service}
	websocket.ServeWs(h.hub, w, r, uint(chatID), userID, visitorID, saver)
}

func (h *ChatHandler) CreateWidgetChat(w http.ResponseWriter, r *http.Request) {
	widgetCtx, ok := middleware.GetWidgetContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateWidgetChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.VisitorID == "" {
		http.Error(w, "visitor_id is required", http.StatusBadRequest)
		return
	}

	chat, err := h.service.CreateChat(r.Context(), widgetCtx.Domain, widgetCtx.UserID, req.VisitorID)
	if err != nil {
		http.Error(w, "failed to create chat", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.ChatResponse{
		ID:     chat.ID,
		UUID:   chat.UUID,
		Status: string(chat.Status),
		Origin: chat.Origin,
	})
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	chat, err := h.service.CreateChat(r.Context(), req.Origin, claims.UserID, req.VisitorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.ChatResponse{
		ID:     chat.ID,
		UUID:   chat.UUID,
		Status: string(chat.Status),
		Origin: chat.Origin,
	})
}

func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidChatId, http.StatusBadRequest)
		return
	}

	chat, err := h.service.GetChat(r.Context(), uint(chatID))
	if err != nil {
		http.Error(w, "chat not found", http.StatusNotFound)
		return
	}

	if chat.OwnerUserID != claims.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(model.ChatResponse{
		ID:     chat.ID,
		UUID:   chat.UUID,
		Status: string(chat.Status),
		Origin: chat.Origin,
	})
}

func (h *ChatHandler) ListChats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	chats, err := h.service.ListChats(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to list chats", http.StatusInternalServerError)
		return
	}

	response := make([]model.ChatResponse, len(chats))
	for i, chat := range chats {
		response[i] = model.ChatResponse{
			ID:     chat.ID,
			UUID:   chat.UUID,
			Status: string(chat.Status),
			Origin: chat.Origin,
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		http.Error(w, constant.InvalidChatId, http.StatusBadRequest)
		return
	}

	messages, err := h.service.GetMessages(r.Context(), uint(chatID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(messages)
}

func (a *messageSaverAdapter) SaveMessage(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) error {
	_, err := a.service.SaveMessage(ctx, chatID, senderID, visitorID, content)
	return err
}
