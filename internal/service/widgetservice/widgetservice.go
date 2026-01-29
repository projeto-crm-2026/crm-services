package widgetservice

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	visitorjwt "github.com/projeto-crm-2026/crm-services/pkg/visitorjwt"
)

type WidgetService interface {
	InitSession(ctx context.Context, visitorID string, ownerUserID uint, domain string) (*WidgetSession, error)
	ValidateAPIKey(ctx context.Context, publicKey, origin string) (*APIKeyInfo, error)
	CreateAPIKey(ctx context.Context, userID uint, name, domain string) (*entity.APIKey, error)
	ListAPIKeys(ctx context.Context, userID uint) ([]entity.APIKey, error)
	DeleteAPIKey(ctx context.Context, userID uint, keyID uint) error
}

type WidgetSession struct {
	Token     string
	VisitorID string
}

type APIKeyInfo struct {
	UserID    uint
	PublicKey string
	Domain    string
}

type widgetService struct {
	apiKeyRepo repo.APIKeyRepo
	jwtConfig  *config.JWTConfig
	logger     *slog.Logger
}

func NewWidgetService(apiKeyRepo repo.APIKeyRepo, jwtConfig *config.JWTConfig, logger *slog.Logger) WidgetService {
	return &widgetService{
		apiKeyRepo: apiKeyRepo,
		jwtConfig:  jwtConfig,
		logger:     logger,
	}
}

func (s *widgetService) InitSession(ctx context.Context, visitorID string, ownerUserID uint, domain string) (*WidgetSession, error) {
	if visitorID == "" {
		visitorID = uuid.New().String()
	}

	token, err := visitorjwt.GenerateVisitorToken(
		visitorID,
		ownerUserID,
		domain,
		s.jwtConfig.JWTSecret,
	)
	if err != nil {
		s.logger.Error("failed to generate visitor token", "error", err)
		return nil, err
	}

	s.logger.Info("widget session initialized", "visitorID", visitorID, "ownerUserID", ownerUserID)

	return &WidgetSession{
		Token:     token,
		VisitorID: visitorID,
	}, nil
}

func (s *widgetService) ValidateAPIKey(ctx context.Context, publicKey, origin string) (*APIKeyInfo, error) {
	apiKey, err := s.apiKeyRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		s.logger.Error("failed to get API key", "error", err)
		return nil, err
	}

	if !apiKey.IsActive {
		s.logger.Warn("API key is inactive", "publicKey", publicKey)
		return nil, ErrAPIKeyInactive
	}

	if !s.isOriginAllowed(origin, apiKey.Domain) {
		s.logger.Warn("origin not allowed", "origin", origin, "allowedDomain", apiKey.Domain)
		return nil, ErrOriginNotAllowed
	}

	_ = s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)

	return &APIKeyInfo{
		UserID:    apiKey.UserID,
		PublicKey: apiKey.PublicKey,
		Domain:    apiKey.Domain,
	}, nil
}

func (s *widgetService) CreateAPIKey(ctx context.Context, userID uint, name, domain string) (*entity.APIKey, error) {
	publicKey := "pk_" + uuid.New().String()
	secretKey := "sk_" + uuid.New().String()

	apiKey, err := s.apiKeyRepo.Insert(ctx, userID, publicKey, secretKey, name, domain)
	if err != nil {
		s.logger.Error("failed to create API key", "error", err)
		return nil, err
	}

	s.logger.Info("API key created", "userID", userID, "name", name, "domain", domain)
	return apiKey, nil
}

func (s *widgetService) ListAPIKeys(ctx context.Context, userID uint) ([]entity.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(ctx, userID)
}

func (s *widgetService) DeleteAPIKey(ctx context.Context, userID uint, keyID uint) error {
	return s.apiKeyRepo.Delete(ctx, userID, keyID)
}

func (s *widgetService) isOriginAllowed(origin, allowedDomain string) bool {
	if allowedDomain == "*" {
		return true
	}

	return origin == allowedDomain ||
		origin == "https://"+allowedDomain ||
		origin == "http://"+allowedDomain
}
