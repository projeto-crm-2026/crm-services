package userservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/pkg/jwt"
	passwordHashing "github.com/projeto-crm-2026/crm-services/pkg/passwordhashing"
)

type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string) (string, *entity.User, error)
	LoginUser(ctx context.Context, email, password string) (string, *entity.User, error)
}

type userService struct {
	repo      repo.UserRepo
	jwtConfig *config.JWTConfig
	logger    *slog.Logger
}

func NewUserService(repo repo.UserRepo, jwtConfig *config.JWTConfig, logger *slog.Logger) UserService {
	return &userService{
		repo:      repo,
		jwtConfig: jwtConfig,
		logger:    logger,
	}
}

func (s *userService) RegisterUser(ctx context.Context, name, email, password string) (string, *entity.User, error) {
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("failed to check existing user", "error", err)
		return "", nil, err
	}

	if existingUser != nil {
		s.logger.Warn("user with email already exists", "email", email)
		return "", nil, fmt.Errorf("user with email %s already exists", email)
	}

	passwordHash, err := passwordHashing.HashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return "", nil, err
	}

	user, err := s.repo.Insert(ctx, name, email, passwordHash)
	if err != nil {
		s.logger.Error("failed to insert user", "error", err)
		return "", nil, err
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.OrganizationID, s.jwtConfig.JWTSecret)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return "", nil, err
	}

	s.logger.Info("user registered successfully", "userID", user.ID, "token", token)
	return token, user, nil
}

func (s *userService) LoginUser(ctx context.Context, email, password string) (string, *entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("failed to get user by email", "error", err)
		return "", nil, err
	}

	if user == nil {
		s.logger.Warn("user not found", "email", email)
		return "", nil, fmt.Errorf("invalid password or email %s", email)
	}

	if !passwordHashing.VerifyPassword(password, user.PasswordHash) {
		s.logger.Warn("invalid password", "email", email)
		return "", nil, fmt.Errorf("invalid password or email %s", email)
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.OrganizationID, s.jwtConfig.JWTSecret)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return "", nil, err
	}

	s.logger.Info("user logged in successfully", "userID", user.ID, "token", token)

	return token, user, nil
}
