package planservice

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/pkg/mailer"
	"github.com/projeto-crm-2026/crm-services/pkg/utils"
)

type PlanService interface {
	CheckContactLimit(ctx context.Context, orgID uuid.UUID) error
	CheckMemberLimit(ctx context.Context, orgID uuid.UUID) error
	CheckChatResponderLimit(ctx context.Context, orgID uuid.UUID) error
	ListPlans(ctx context.Context) ([]entity.Plan, error)
	GetOrganizationUsage(ctx context.Context, orgID uuid.UUID) (*model.UsageResponse, error)
	GetFreePlan(ctx context.Context) (*entity.Plan, error)
	CreateSubscription(ctx context.Context, orgID uuid.UUID, planID uint) (*entity.Subscription, error)
	TransitionSubscription(ctx context.Context, orgID uuid.UUID, newStatus entity.SubscriptionStatus) error
	NotifyUsageWarnings(ctx context.Context, orgID uuid.UUID, resource string, current, limit int)
}

type planService struct {
	planRepo repo.PlanRepo
	subRepo  repo.SubscriptionRepo
	userRepo repo.UserRepo
	mailer   mailer.Mailer
	logger   *slog.Logger
}

func NewPlanService(planRepo repo.PlanRepo, subRepo repo.SubscriptionRepo, userRepo repo.UserRepo, mail mailer.Mailer, logger *slog.Logger) PlanService {
	return &planService{planRepo: planRepo, subRepo: subRepo, userRepo: userRepo, mailer: mail, logger: logger}
}

func (s *planService) ListPlans(ctx context.Context) ([]entity.Plan, error) {
	plans, err := s.planRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list plans", "error", err)
		return nil, err
	}
	return plans, nil
}

func (s *planService) GetOrganizationUsage(ctx context.Context, orgID uuid.UUID) (*model.UsageResponse, error) {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get subscription for organization", "orgID", orgID, "error", err)
		return nil, ErrOrganizationNotFound
	}

	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		s.logger.Error("failed to get plan for subscription", "planID", sub.PlanID, "error", err)
		return nil, err
	}

	contactCount, err := s.planRepo.CountContacts(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count contacts", "orgID", orgID, "error", err)
		return nil, err
	}

	memberCount, err := s.planRepo.CountMembers(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count members", "orgID", orgID, "error", err)
		return nil, err
	}

	responderCount, err := s.planRepo.CountChatResponders(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count chat responders", "orgID", orgID, "error", err)
		return nil, err
	}

	usage := &model.UsageResponse{
		Plan: model.NewPlanResponse(plan),
		Subscription: model.SubscriptionInfo{
			Status:             string(sub.Status),
			CurrentPeriodStart: utils.FormatTimePtr(sub.CurrentPeriodStart),
			CurrentPeriodEnd:   utils.FormatTimePtr(sub.CurrentPeriodEnd),
		},
		Usage: model.UsageData{
			Contacts:       model.NewUsageResource(contactCount, plan.MaxContacts),
			Members:        model.NewUsageResource(memberCount, plan.MaxMembers),
			ChatResponders: model.NewUsageResource(responderCount, plan.MaxChatResponders),
		},
	}

	return usage, nil
}

func (s *planService) GetFreePlan(ctx context.Context) (*entity.Plan, error) {
	plan, err := s.planRepo.GetByName(ctx, "free")
	if err != nil {
		s.logger.Error("failed to get free plan", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrFreePlanNotFound, err)
	}
	return plan, nil
}

func (s *planService) CreateSubscription(ctx context.Context, orgID uuid.UUID, planID uint) (*entity.Subscription, error) {
	now := time.Now()
	sub := &entity.Subscription{
		OrganizationID:     orgID,
		PlanID:             planID,
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: &now,
	}

	created, err := s.subRepo.Create(ctx, sub)
	if err != nil {
		s.logger.Error("failed to create subscription", "orgID", orgID, "planID", planID, "error", err)
		return nil, err
	}

	s.logger.Info("subscription created", "orgID", orgID, "planID", planID)
	return created, nil
}

func (s *planService) getOrgPlan(ctx context.Context, orgID uuid.UUID) (*entity.Plan, error) {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *planService) CheckContactLimit(ctx context.Context, orgID uuid.UUID) error {
	plan, err := s.getOrgPlan(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get plan for contact limit check", "orgID", orgID, "error", err)
		return err
	}

	count, err := s.planRepo.CountContacts(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count contacts for limit check", "orgID", orgID, "error", err)
		return err
	}

	if count >= plan.MaxContacts {
		s.logger.Warn("contact limit reached", "orgID", orgID, "current", count, "limit", plan.MaxContacts)
		return constant.ErrContactLimitReached
	}

	return nil
}

func (s *planService) CheckMemberLimit(ctx context.Context, orgID uuid.UUID) error {
	plan, err := s.getOrgPlan(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get plan for member limit check", "orgID", orgID, "error", err)
		return err
	}

	count, err := s.planRepo.CountMembers(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count members for limit check", "orgID", orgID, "error", err)
		return err
	}

	if count >= plan.MaxMembers {
		s.logger.Warn("member limit reached", "orgID", orgID, "current", count, "limit", plan.MaxMembers)
		return constant.ErrMemberLimitReached
	}

	return nil
}

func (s *planService) CheckChatResponderLimit(ctx context.Context, orgID uuid.UUID) error {
	plan, err := s.getOrgPlan(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get plan for chat responder limit check", "orgID", orgID, "error", err)
		return err
	}

	count, err := s.planRepo.CountChatResponders(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to count chat responders for limit check", "orgID", orgID, "error", err)
		return err
	}

	if count >= plan.MaxChatResponders {
		s.logger.Warn("chat responder limit reached", "orgID", orgID, "current", count, "limit", plan.MaxChatResponders)
		return constant.ErrChatResponderLimitReached
	}

	return nil
}

func (s *planService) TransitionSubscription(ctx context.Context, orgID uuid.UUID, newStatus entity.SubscriptionStatus) error {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get subscription for transition", "orgID", orgID, "error", err)
		return ErrOrganizationNotFound
	}

	if !sub.Status.CanTransitionTo(newStatus) {
		s.logger.Warn("invalid subscription transition attempted", "orgID", orgID, "from", sub.Status, "to", newStatus)
		return ErrInvalidTransition
	}

	if err := s.subRepo.UpdateStatus(ctx, sub.ID, newStatus); err != nil {
		s.logger.Error("failed to update subscription status", "orgID", orgID, "from", sub.Status, "to", newStatus, "error", err)
		return err
	}

	s.logger.Info("subscription status transitioned", "orgID", orgID, "from", sub.Status, "to", newStatus)
	return nil
}

func (s *planService) NotifyUsageWarnings(ctx context.Context, orgID uuid.UUID, resource string, current, limit int) {
	if limit <= 0 {
		return
	}

	threshold80 := int(math.Ceil(float64(limit) * 0.8))
	threshold90 := int(math.Ceil(float64(limit) * 0.9))
	prev := current - 1

	var pct int
	switch {
	case prev < threshold90 && current >= threshold90:
		pct = 90
	case prev < threshold80 && current >= threshold80:
		pct = 80
	default:
		return
	}

	name, email, err := s.userRepo.GetOrgOwnerEmail(ctx, orgID)
	if err != nil {
		s.logger.Warn("could not find org owner for usage warning", "orgID", orgID, "error", err)
		return
	}

	go func() {
		if err := s.mailer.SendUsageWarningEmail(email, name, resource, current, limit, pct); err != nil {
			s.logger.Warn("failed to send usage warning email", "email", email, "resource", resource, "error", err)
		}
	}()
}
