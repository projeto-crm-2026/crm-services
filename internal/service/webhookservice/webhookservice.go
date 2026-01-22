package webhookservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server/websocket"
	"github.com/projeto-crm-2026/crm-services/internal/service/chatservice"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
	ErrInvalidToken    = errors.New("invalid webhook token")
	ErrInvalidAction   = errors.New("invalid webhook action")
	ErrChatNotFound    = errors.New("chat not found")
)

type WebhookService interface {
	// outgoing webhooks
	CreateWebhook(ctx context.Context, userID uint, name, url string, events []string) (*entity.Webhook, error)
	ListWebhooks(ctx context.Context, userID uint) ([]entity.Webhook, error)
	GetWebhook(ctx context.Context, userID, webhookID uint) (*entity.Webhook, error)
	UpdateWebhook(ctx context.Context, userID, webhookID uint, name, url string, events []string, isActive bool) error
	DeleteWebhook(ctx context.Context, userID, webhookID uint) error
	GetWebhookLogs(ctx context.Context, userID, webhookID uint, limit int) ([]entity.WebhookLog, error)

	// incoming webhook tokens
	CreateIncomingToken(ctx context.Context, userID uint, name string) (*entity.IncomingWebhookToken, error)
	ListIncomingTokens(ctx context.Context, userID uint) ([]entity.IncomingWebhookToken, error)
	DeleteIncomingToken(ctx context.Context, userID, tokenID uint) error

	// event dispatching
	DispatchEvent(ctx context.Context, userID uint, event *WebhookEvent)

	// incoming webhook processing
	ProcessIncomingWebhook(ctx context.Context, token string, payload *IncomingWebhookPayload) error

	// helper
	OnMessageReceived(ctx context.Context, userID uint, chatID uint, message *entity.Message)
	OnChatCreated(ctx context.Context, userID uint, chat *entity.Chat, visitorID string)
	OnChatClosed(ctx context.Context, userID uint, chat *entity.Chat)
}

type webhookService struct {
	repo        repo.WebhookRepo
	chatService chatservice.ChatService
	hub         *websocket.Hub
	dispatcher  *Dispatcher
	logger      *slog.Logger
}

func NewWebhookService(
	webhookRepo repo.WebhookRepo,
	chatService chatservice.ChatService,
	hub *websocket.Hub,
	logger *slog.Logger,
) WebhookService {
	dispatcher := NewDispatcher(webhookRepo, logger)

	return &webhookService{
		repo:        webhookRepo,
		chatService: chatService,
		hub:         hub,
		dispatcher:  dispatcher,
		logger:      logger,
	}
}

func (s *webhookService) CreateWebhook(ctx context.Context, userID uint, name, url string, events []string) (*entity.Webhook, error) {
	secret := "whsec_" + uuid.New().String()

	webhook, err := s.repo.Insert(ctx, userID, name, url, secret, events)
	if err != nil {
		s.logger.Error("failed to create webhook", "error", err)
		return nil, err
	}

	s.logger.Info("webhook created", "webhookID", webhook.ID, "userID", userID)
	return webhook, nil
}

func (s *webhookService) ListWebhooks(ctx context.Context, userID uint) ([]entity.Webhook, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *webhookService) GetWebhook(ctx context.Context, userID, webhookID uint) (*entity.Webhook, error) {
	webhook, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrWebhookNotFound
	}

	if webhook.UserID != userID {
		return nil, ErrWebhookNotFound
	}

	return webhook, nil
}

func (s *webhookService) UpdateWebhook(ctx context.Context, userID, webhookID uint, name, url string, events []string, isActive bool) error {
	webhook, err := s.GetWebhook(ctx, userID, webhookID)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, webhook.ID, name, url, events, isActive)
}

func (s *webhookService) DeleteWebhook(ctx context.Context, userID, webhookID uint) error {
	return s.repo.Delete(ctx, userID, webhookID)
}

func (s *webhookService) GetWebhookLogs(ctx context.Context, userID, webhookID uint, limit int) ([]entity.WebhookLog, error) {
	_, err := s.GetWebhook(ctx, userID, webhookID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetLogsByWebhookID(ctx, webhookID, limit)
}

func (s *webhookService) CreateIncomingToken(ctx context.Context, userID uint, name string) (*entity.IncomingWebhookToken, error) {
	token := "whit_" + uuid.New().String()

	t, err := s.repo.InsertToken(ctx, userID, token, name)
	if err != nil {
		s.logger.Error("failed to create incoming token", "error", err)
		return nil, err
	}

	s.logger.Info("incoming webhook token created", "tokenID", t.ID, "userID", userID)
	return t, nil
}

func (s *webhookService) ListIncomingTokens(ctx context.Context, userID uint) ([]entity.IncomingWebhookToken, error) {
	return s.repo.GetTokensByUserID(ctx, userID)
}

func (s *webhookService) DeleteIncomingToken(ctx context.Context, userID, tokenID uint) error {
	return s.repo.DeleteToken(ctx, userID, tokenID)
}

// envia um evento para todos os webhooks configurados
func (s *webhookService) DispatchEvent(ctx context.Context, userID uint, event *WebhookEvent) {
	s.dispatcher.Dispatch(ctx, userID, event)
}

func (s *webhookService) ProcessIncomingWebhook(ctx context.Context, token string, payload *IncomingWebhookPayload) error {
	t, err := s.repo.GetTokenByValue(ctx, token)
	if err != nil {
		s.logger.Warn("invalid incoming webhook token", "token", token[:10]+"...")
		return ErrInvalidToken
	}

	s.repo.UpdateTokenLastUsed(ctx, t.ID)

	switch payload.Action {
	case constant.ActionSendMessage:
		return s.handleSendMessage(ctx, t.UserID, payload)
	case constant.ActionCloseChat:
		return s.handleCloseChat(ctx, t.UserID, payload)
	default:
		return ErrInvalidAction
	}
}

func (s *webhookService) handleSendMessage(ctx context.Context, userID uint, payload *IncomingWebhookPayload) error {
	if payload.ChatID == 0 || payload.Content == "" {
		return errors.New("chat_id and content are required")
	}

	chat, err := s.chatService.GetChat(ctx, payload.ChatID)
	if err != nil || chat.OwnerUserID != userID {
		return ErrChatNotFound
	}

	message, err := s.chatService.SaveMessage(ctx, payload.ChatID, &userID, "", payload.Content)
	if err != nil {
		return err
	}

	// broadcast via websocket
	wsMessage := map[string]interface{}{
		"type":       "message",
		"content":    payload.Content,
		"sender_id":  userID,
		"visitor_id": "",
		"chat_id":    payload.ChatID,
		"message_id": message.ID,
		"timestamp":  message.CreatedAt,
	}

	msgBytes, _ := json.Marshal(wsMessage)
	s.hub.BroadcastToChat(payload.ChatID, msgBytes)

	s.logger.Info("message sent via incoming webhook", "chatID", payload.ChatID, "userID", userID)
	return nil
}

func (s *webhookService) handleCloseChat(ctx context.Context, userID uint, payload *IncomingWebhookPayload) error {
	// TODO: logica de fechar chat, ainda nao vou fazer, nao sei se continua
	s.logger.Info("close chat requested via webhook", "chatID", payload.ChatID, "userID", userID)
	return nil
}

func (s *webhookService) OnMessageReceived(ctx context.Context, userID uint, chatID uint, message *entity.Message) {
	event := NewWebhookEvent(entity.EventMessageReceived, MessageEventData{
		ChatID:    chatID,
		MessageID: message.ID,
		Content:   message.Content,
		SenderID:  message.SenderID,
		VisitorID: message.VisitorID,
		Type:      string(message.Type),
		Timestamp: message.CreatedAt.Format(time.RFC3339),
	})

	s.DispatchEvent(ctx, userID, event)
}

func (s *webhookService) OnChatCreated(ctx context.Context, userID uint, chat *entity.Chat, visitorID string) {
	event := NewWebhookEvent(entity.EventChatCreated, ChatEventData{
		ChatID:      chat.ID,
		Status:      string(chat.Status),
		Origin:      chat.Origin,
		OwnerUserID: chat.OwnerUserID,
		VisitorID:   visitorID,
		Timestamp:   chat.CreatedAt.Format(time.RFC3339),
	})

	s.DispatchEvent(ctx, userID, event)
}

func (s *webhookService) OnChatClosed(ctx context.Context, userID uint, chat *entity.Chat) {
	event := NewWebhookEvent(entity.EventChatClosed, ChatEventData{
		ChatID:      chat.ID,
		Status:      string(chat.Status),
		Origin:      chat.Origin,
		OwnerUserID: chat.OwnerUserID,
		Timestamp:   time.Now().Format(time.RFC3339),
	})

	s.DispatchEvent(ctx, userID, event)
}
