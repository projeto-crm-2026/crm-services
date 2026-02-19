package userservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/pkg/jwt"
	"github.com/projeto-crm-2026/crm-services/pkg/mailer"
	passwordHashing "github.com/projeto-crm-2026/crm-services/pkg/passwordhashing"
	"github.com/projeto-crm-2026/crm-services/pkg/slug"
)

type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string, organizationName string) (string, *entity.User, error)
	LoginUser(ctx context.Context, email, password string) (string, *entity.User, error)
	InviteUser(ctx context.Context, adminUserID uint, name, email string) (*entity.User, error)
	AcceptInvite(ctx context.Context, token, password string) (*entity.User, error)
	ListOrganizationMembers(ctx context.Context, userID uint) ([]entity.User, error)
}

type userService struct {
	repo      repo.UserRepo
	orgRepo   repo.OrganizationRepo
	jwtConfig *config.JWTConfig
	mailer    mailer.Mailer
	logger    *slog.Logger
}

func NewUserService(
	repo repo.UserRepo,
	orgRepo repo.OrganizationRepo,
	jwtConfig *config.JWTConfig,
	mailer mailer.Mailer,
	logger *slog.Logger,
) UserService {
	return &userService{
		repo:      repo,
		orgRepo:   orgRepo,
		jwtConfig: jwtConfig,
		mailer:    mailer,
		logger:    logger,
	}
}

func (s *userService) RegisterUser(ctx context.Context, name, email, password string, organizationName string) (string, *entity.User, error) {
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

	org, err := s.orgRepo.Create(ctx, &entity.Organization{
		Name:     organizationName,
		Slug:     slug.GenerateSlug(organizationName),
		IsActive: true,
	})
	if err != nil {
		s.logger.Error("failed to create organization", "error", err)
		return "", nil, err
	}

	user, err := s.repo.Insert(ctx, name, email, passwordHash, org.UUID)
	if err != nil {
		s.logger.Error("failed to insert user", "error", err)
		return "", nil, err
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, &org.UUID, s.jwtConfig.JWTSecret)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return "", nil, err
	}

	s.logger.Info("user registered successfully", "userID", user.ID, "orgID", org.UUID)
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
		return "", nil, fmt.Errorf("invalid password or email")
	}

	if user.IsPending() {
		return "", nil, fmt.Errorf("account is pending activation, check your email")
	}

	if !passwordHashing.VerifyPassword(password, user.PasswordHash) {
		s.logger.Warn("invalid password", "email", email)
		return "", nil, fmt.Errorf("invalid password or email")
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.OrganizationID, s.jwtConfig.JWTSecret)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return "", nil, err
	}

	s.logger.Info("user logged in successfully", "userID", user.ID)
	return token, user, nil
}

func (s *userService) InviteUser(ctx context.Context, adminUserID uint, name, email string) (*entity.User, error) {
	admin, err := s.repo.GetByID(ctx, adminUserID)
	if err != nil {
		s.logger.Error("failed to get admin user", "error", err)
		return nil, err
	}

	if !admin.IsAdmin() {
		return nil, fmt.Errorf("only admins can invite users")
	}

	if admin.OrganizationID == nil {
		return nil, fmt.Errorf("admin is not associated with any organization")
	}

	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	inviteToken := "inv_" + uuid.New().String()
	inviteExpiry := time.Now().Add(72 * time.Hour)

	user, err := s.repo.InsertPending(ctx, name, email, inviteToken, inviteExpiry, *admin.OrganizationID, adminUserID)
	if err != nil {
		s.logger.Error("failed to create pending user", "error", err)
		return nil, err
	}

	// nao pode falhar pro cara mandar dps caso de erro aqui, por isso nao dou return
	if err := s.mailer.SendInviteEmail(email, name, inviteToken); err != nil {
		s.logger.Error("failed to send invite email", "error", err, "email", email)
	}

	s.logger.Info("user invited", "email", email, "invitedBy", adminUserID, "inviteToken", inviteToken, "inviteExpiry", inviteExpiry)
	return user, nil
}

func (s *userService) AcceptInvite(ctx context.Context, token, password string) (*entity.User, error) {
	if len(password) < 8 {
		s.logger.Error("password too short", "error", fmt.Errorf("password must be at least 8 characters"))
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	user, err := s.repo.GetByInviteToken(ctx, token)
	if err != nil {
		s.logger.Error("failed to get user by invite token", "error", err)
		return nil, fmt.Errorf("invalid invite token")
	}

	if user.Status != entity.StatusPending {
		return nil, fmt.Errorf("account is already active")
	}

	if user.InviteExpiry != nil && time.Now().After(*user.InviteExpiry) {
		return nil, fmt.Errorf("invite token has expired")
	}

	passwordHash, err := passwordHashing.HashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, err
	}

	if err := s.repo.ActivateUser(ctx, user.ID, passwordHash); err != nil {
		s.logger.Error("failed to activate user", "error", err)
		return nil, err
	}

	s.logger.Info("user activated via invite", "userID", user.ID, "email", user.Email)

	user.Status = entity.StatusActive
	user.InviteToken = nil
	user.InviteExpiry = nil
	return user, nil
}

func (s *userService) ListOrganizationMembers(ctx context.Context, userID uint) ([]entity.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.OrganizationID == nil {
		return nil, fmt.Errorf("user is not associated with any organization")
	}

	return s.repo.ListByOrganization(ctx, *user.OrganizationID)
}
