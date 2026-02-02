package chatservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type ChatService interface {
	CreateChat(ctx context.Context, origin string, ownerUserID uint, visitorID string) (*entity.Chat, error)
	GetChat(ctx context.Context, chatID uint) (*entity.Chat, error)
	ListChats(ctx context.Context, ownerUserID uint) ([]entity.Chat, error)
	GetMessages(ctx context.Context, chatID uint) ([]entity.Message, error)
	SaveMessage(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) (*entity.Message, error)
	SetMessageHandler(handler MessageEventHandler)
}

type MessageEventHandler interface {
	OnMessageReceived(ctx context.Context, ownerUserID uint, chatID uint, message *entity.Message)
}

type chatService struct {
	chatRepo       repo.ChatRepo
	messageRepo    repo.MessageRepo
	logger         *slog.Logger
	messageHandler MessageEventHandler
}

func NewChatService(chatRepo repo.ChatRepo, messageRepo repo.MessageRepo, logger *slog.Logger) ChatService {
	return &chatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		logger:      logger,
	}
}

func (s *chatService) SetMessageHandler(handler MessageEventHandler) {
	s.messageHandler = handler
}

func (s *chatService) CreateChat(ctx context.Context, origin string, ownerUserID uint, visitorID string) (*entity.Chat, error) {
	chat, err := s.chatRepo.Insert(ctx, origin, ownerUserID)
	if err != nil {
		s.logger.Error("failed to create chat", "error", err)
		return nil, err
	}

	if err := s.chatRepo.AddParticipant(ctx, chat.ID, &ownerUserID, "", entity.ParticipantRoleAgent); err != nil {
		s.logger.Error("failed to add owner as participant", "error", err)
		return nil, err
	}

	if visitorID != "" {
		if err := s.chatRepo.AddParticipant(ctx, chat.ID, nil, visitorID, entity.ParticipantRoleVisitor); err != nil {
			s.logger.Error("failed to add visitor as participant", "error", err)
			return nil, err
		}
	}

	s.logger.Info("chat created", "chatID", chat.ID, "ownerUserID", ownerUserID)
	return chat, nil
}

func (s *chatService) GetChat(ctx context.Context, chatID uint) (*entity.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		s.logger.Error("failed to get chat", "error", err, "chatID", chatID)
		return nil, err
	}
	return chat, nil
}

func (s *chatService) ListChats(ctx context.Context, ownerUserID uint) ([]entity.Chat, error) {
	chats, err := s.chatRepo.GetByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		s.logger.Error("failed to list chats", "error", err, "ownerUserID", ownerUserID)
		return nil, err
	}
	return chats, nil
}

func (s *chatService) GetMessages(ctx context.Context, chatID uint) ([]entity.Message, error) {
	messages, err := s.messageRepo.GetByChatID(ctx, chatID)
	if err != nil {
		s.logger.Error("failed to get messages", "error", err, "chatID", chatID)
		return nil, err
	}

	return messages, nil
}

func (s *chatService) SaveMessage(ctx context.Context, chatID uint, senderID *uint, visitorID, content string) (*entity.Message, error) {
	message, err := s.messageRepo.Insert(ctx, chatID, senderID, visitorID, content)
	if err != nil {
		s.logger.Error("failed to save message", "error", err, "chatID", chatID)
		return nil, err
	}

	if s.messageHandler != nil && visitorID != "" && senderID == nil {
		chat, err := s.chatRepo.GetByID(ctx, chatID)
		if err != nil {
			s.logger.Error("failed to load chat for webhook", "error", err, "chatID", chatID)
		} else {
			s.logger.Info("triggering webhook for message", "chatID", chatID, "messageID", message.ID)
			go func() {
				webhookCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // obrigado coderabbit, essa go func podia travar infinitamente sem context
				defer cancel()
				s.messageHandler.OnMessageReceived(webhookCtx, chat.OwnerUserID, chatID, message)
			}()
		}
	}

	s.logger.Info("message saved", "messageID", message.ID, "chatID", chatID)

	return message, nil
}
