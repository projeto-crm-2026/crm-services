package subscriptionservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/pkg/mailer"
	"github.com/projeto-crm-2026/crm-services/pkg/mercadopago"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, orgID uuid.UUID, planUUID uuid.UUID, payerEmail, cardTokenID string) (*entity.Subscription, error)
	Cancel(ctx context.Context, orgID uuid.UUID) error
	Upgrade(ctx context.Context, orgID uuid.UUID, newPlanUUID uuid.UUID, payerEmail, cardTokenID string) (*entity.Subscription, error)
	HandleWebhook(ctx context.Context, notification *mercadopago.WebhookNotification) error
	ListPayments(ctx context.Context, orgID uuid.UUID) ([]entity.Payment, error)
}

type subscriptionService struct {
	planRepo    repo.PlanRepo
	subRepo     repo.SubscriptionRepo
	paymentRepo repo.PaymentRepo
	userRepo    repo.UserRepo
	mpClient    *mercadopago.Client
	mailer      mailer.Mailer
	logger      *slog.Logger
}

func NewSubscriptionService(
	planRepo repo.PlanRepo,
	subRepo repo.SubscriptionRepo,
	paymentRepo repo.PaymentRepo,
	userRepo repo.UserRepo,
	mpClient *mercadopago.Client,
	mail mailer.Mailer,
	logger *slog.Logger,
) SubscriptionService {
	return &subscriptionService{
		planRepo:    planRepo,
		subRepo:     subRepo,
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		mpClient:    mpClient,
		mailer:      mail,
		logger:      logger,
	}
}

func (s *subscriptionService) Subscribe(ctx context.Context, orgID uuid.UUID, planUUID uuid.UUID, payerEmail, cardTokenID string) (*entity.Subscription, error) {
	plans, err := s.planRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list plans", "error", err)
		return nil, err
	}

	var plan *entity.Plan
	for i := range plans {
		if plans[i].UUID == planUUID {
			plan = &plans[i]
			break
		}
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	if plan.Name == "free" {
		return nil, ErrFreePlanNotPaid
	}

	if plan.MpPreapprovalPlanID == nil {
		return nil, fmt.Errorf("plan %s is not configured for payment", plan.Name)
	}

	existingSub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err == nil && existingSub.Status == entity.SubscriptionActive {
		existingPlan, _ := s.planRepo.GetByID(ctx, existingSub.PlanID)
		if existingPlan != nil && existingPlan.Name != "free" {
			return nil, ErrAlreadySubscribed
		}
	}

	mpSub, err := s.mpClient.CreateSubscription(ctx, &mercadopago.CreateSubscriptionRequest{
		PreapprovalPlanID: *plan.MpPreapprovalPlanID,
		Reason:            plan.DisplayName + " - CRM Services",
		PayerEmail:        payerEmail,
		CardTokenID:       cardTokenID,
		Status:            "authorized",
	})
	if err != nil {
		s.logger.Error("failed to create MP subscription", "orgID", orgID, "planID", plan.ID, "error", err)
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	s.logger.Info("MP subscription created", "mpSubID", mpSub.ID, "orgID", orgID)

	now := time.Now()
	var periodEnd *time.Time
	if mpSub.NextPaymentDate != nil {
		if t, err := time.Parse(time.RFC3339, *mpSub.NextPaymentDate); err == nil {
			periodEnd = &t
		}
	}

	if existingSub != nil {
		existingSub.PlanID = plan.ID
		existingSub.Status = entity.SubscriptionActive
		existingSub.MpSubscriptionID = &mpSub.ID
		existingSub.CurrentPeriodStart = &now
		existingSub.CurrentPeriodEnd = periodEnd
		existingSub.CancelAtPeriodEnd = false
		existingSub.CancelledAt = nil

		if err := s.subRepo.Update(ctx, existingSub); err != nil {
			s.logger.Error("failed to update local subscription", "error", err)
			return nil, err
		}
		return existingSub, nil
	}

	sub := &entity.Subscription{
		OrganizationID:     orgID,
		PlanID:             plan.ID,
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   periodEnd,
		MpSubscriptionID:   &mpSub.ID,
	}

	created, err := s.subRepo.Create(ctx, sub)
	if err != nil {
		s.logger.Error("failed to create local subscription", "error", err)
		return nil, err
	}

	s.logger.Info("subscription created", "orgID", orgID, "planID", plan.ID, "mpSubID", mpSub.ID)
	return created, nil
}

func (s *subscriptionService) Cancel(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return ErrSubscriptionNotFound
	}

	if sub.MpSubscriptionID == nil {
		return ErrNoPaidSubscription
	}

	_, err = s.mpClient.UpdateSubscription(ctx, *sub.MpSubscriptionID, &mercadopago.UpdateSubscriptionRequest{
		Status: "cancelled",
	})
	if err != nil {
		s.logger.Error("failed to cancel MP subscription", "mpSubID", *sub.MpSubscriptionID, "error", err)
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	now := time.Now()
	sub.Status = entity.SubscriptionCancelled
	sub.CancelAtPeriodEnd = true
	sub.CancelledAt = &now

	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.logger.Error("failed to update subscription after cancel", "error", err)
		return err
	}

	s.logger.Info("subscription cancelled", "orgID", orgID, "accessUntil", sub.CurrentPeriodEnd)
	return nil
}

func (s *subscriptionService) Upgrade(ctx context.Context, orgID uuid.UUID, newPlanUUID uuid.UUID, payerEmail, cardTokenID string) (*entity.Subscription, error) {
	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	currentPlan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	plans, err := s.planRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var newPlan *entity.Plan
	for i := range plans {
		if plans[i].UUID == newPlanUUID {
			newPlan = &plans[i]
			break
		}
	}
	if newPlan == nil {
		return nil, ErrPlanNotFound
	}

	if newPlan.ID == currentPlan.ID {
		return nil, ErrSamePlan
	}

	if newPlan.PriceCents < currentPlan.PriceCents {
		return nil, ErrDowngradeNotAllowed
	}

	if newPlan.Name == "free" {
		return nil, ErrFreePlanNotPaid
	}

	if newPlan.MpPreapprovalPlanID == nil {
		return nil, fmt.Errorf("plan %s is not configured for payment", newPlan.Name)
	}

	// cancela a atual
	if sub.MpSubscriptionID != nil {
		_, err = s.mpClient.UpdateSubscription(ctx, *sub.MpSubscriptionID, &mercadopago.UpdateSubscriptionRequest{
			Status: "cancelled",
		})
		if err != nil {
			s.logger.Warn("failed to cancel old MP subscription during upgrade", "mpSubID", *sub.MpSubscriptionID, "error", err)
		}
	}

	// cria nova ass
	mpSub, err := s.mpClient.CreateSubscription(ctx, &mercadopago.CreateSubscriptionRequest{
		PreapprovalPlanID: *newPlan.MpPreapprovalPlanID,
		Reason:            newPlan.DisplayName + " - CRM Services",
		PayerEmail:        payerEmail,
		CardTokenID:       cardTokenID,
		Status:            "authorized",
	})
	if err != nil {
		s.logger.Error("failed to create MP subscription for upgrade", "orgID", orgID, "error", err)
		return nil, fmt.Errorf("failed to process upgrade payment: %w", err)
	}

	// update local
	now := time.Now()
	var periodEnd *time.Time
	if mpSub.NextPaymentDate != nil {
		if t, err := time.Parse(time.RFC3339, *mpSub.NextPaymentDate); err == nil {
			periodEnd = &t
		}
	}

	sub.PlanID = newPlan.ID
	sub.Status = entity.SubscriptionActive
	sub.MpSubscriptionID = &mpSub.ID
	sub.CurrentPeriodStart = &now
	sub.CurrentPeriodEnd = periodEnd
	sub.CancelAtPeriodEnd = false
	sub.CancelledAt = nil

	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.logger.Error("failed to update subscription after upgrade", "error", err)
		return nil, err
	}

	s.logger.Info("subscription upgraded", "orgID", orgID, "from", currentPlan.Name, "to", newPlan.Name)
	return sub, nil
}

func (s *subscriptionService) HandleWebhook(ctx context.Context, notification *mercadopago.WebhookNotification) error {
	switch notification.Type {
	case "subscription_preapproval":
		return s.handleSubscriptionEvent(ctx, notification.Data.ID)
	case "payment":
		return s.handlePaymentEvent(ctx, notification.Data.ID)
	default:
		s.logger.Debug("ignoring unhandled webhook type", "type", notification.Type)
		return nil
	}
}

func (s *subscriptionService) handleSubscriptionEvent(ctx context.Context, mpSubID string) error {
	mpSub, err := s.mpClient.GetSubscription(ctx, mpSubID)
	if err != nil {
		s.logger.Error("failed to fetch MP subscription", "mpSubID", mpSubID, "error", err)
		return err
	}

	sub, err := s.subRepo.GetByMpSubscriptionID(ctx, mpSubID)
	if err != nil {
		s.logger.Warn("received webhook for unknown subscription", "mpSubID", mpSubID)
		return nil // ignore
	}

	newStatus := mapMpStatus(mpSub.Status)
	if newStatus == sub.Status {
		return nil // sem mudança de status, idempotente
	}

	if !sub.Status.CanTransitionTo(newStatus) {
		s.logger.Warn("ignoring invalid status transition from webhook", "mpSubID", mpSubID, "from", sub.Status, "to", newStatus)
		return nil
	}

	sub.Status = newStatus

	// cancelado
	if newStatus == entity.SubscriptionCancelled || newStatus == entity.SubscriptionExpired {
		if sub.CurrentPeriodEnd != nil && time.Now().After(*sub.CurrentPeriodEnd) {
			return s.downgradeToFree(ctx, sub)
		}
	}

	if err := s.subRepo.UpdateStatus(ctx, sub.ID, newStatus); err != nil {
		s.logger.Error("failed to update subscription status from webhook", "error", err)
		return err
	}

	s.logger.Info("subscription status updated from webhook", "mpSubID", mpSubID, "status", newStatus)
	return nil
}

func (s *subscriptionService) handlePaymentEvent(ctx context.Context, mpPaymentID string) error {
	// idempotencia
	existing, err := s.paymentRepo.GetByMpPaymentID(ctx, mpPaymentID)
	if err == nil && existing != nil {
		// update status
		mpPayment, err := s.mpClient.GetPayment(ctx, mpPaymentID)
		if err != nil {
			s.logger.Error("failed to fetch MP payment", "mpPaymentID", mpPaymentID, "error", err)
			return err
		}

		paymentStatus := mapPaymentStatus(mpPayment.Status)
		if paymentStatus != existing.Status {
			if err := s.paymentRepo.UpdateStatus(ctx, mpPaymentID, paymentStatus, mpPayment.StatusDetail); err != nil {
				s.logger.Error("failed to update payment status", "error", err)
				return err
			}
			s.notifyPaymentByOrg(ctx, existing.OrganizationID, paymentStatus, int(mpPayment.TransactionAmount*100), mpPayment.CurrencyID, mpPayment.StatusDetail)
		}
		return nil
	}

	if !errors.Is(err, repo.ErrPaymentNotFound) {
		return err
	}

	// novo pagamento, criar registro
	mpPayment, err := s.mpClient.GetPayment(ctx, mpPaymentID)
	if err != nil {
		s.logger.Error("failed to fetch MP payment", "mpPaymentID", mpPaymentID, "error", err)
		return err
	}

	amountCents := int(mpPayment.TransactionAmount * 100)

	s.logger.Info("payment webhook received",
		"mpPaymentID", mpPaymentID,
		"status", mpPayment.Status,
		"amount", amountCents,
	)

	return nil
}

func (s *subscriptionService) downgradeToFree(ctx context.Context, sub *entity.Subscription) error {
	freePlan, err := s.planRepo.GetByName(ctx, "free")
	if err != nil {
		s.logger.Error("failed to get free plan for downgrade", "error", err)
		return err
	}

	sub.PlanID = freePlan.ID
	sub.Status = entity.SubscriptionActive
	sub.MpSubscriptionID = nil
	sub.CancelAtPeriodEnd = false
	sub.CancelledAt = nil

	now := time.Now()
	sub.CurrentPeriodStart = &now
	sub.CurrentPeriodEnd = nil

	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.logger.Error("failed to downgrade subscription to free", "error", err)
		return err
	}

	s.logger.Info("subscription downgraded to free plan", "orgID", sub.OrganizationID)
	return nil
}

func (s *subscriptionService) ListPayments(ctx context.Context, orgID uuid.UUID) ([]entity.Payment, error) {
	payments, err := s.paymentRepo.ListByOrganization(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to list payments", "orgID", orgID, "error", err)
		return nil, err
	}
	return payments, nil
}

func mapMpStatus(mpStatus string) entity.SubscriptionStatus {
	switch mpStatus {
	case "authorized":
		return entity.SubscriptionActive
	case "paused", "pending":
		return entity.SubscriptionPastDue
	case "cancelled":
		return entity.SubscriptionCancelled
	default:
		return entity.SubscriptionPastDue
	}
}

func (s *subscriptionService) notifyPaymentByOrg(ctx context.Context, orgID uuid.UUID, status entity.PaymentStatus, amountCents int, currency, statusDetail string) {
	name, email, err := s.userRepo.GetOrgOwnerEmail(ctx, orgID)
	if err != nil {
		s.logger.Warn("could not find org owner for payment notification", "orgID", orgID, "error", err)
		return
	}

	sub, err := s.subRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return
	}
	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return
	}

	switch status {
	case entity.PaymentApproved:
		if err := s.mailer.SendPaymentSuccessEmail(email, name, plan.DisplayName, amountCents, currency); err != nil {
			s.logger.Warn("failed to send payment success email", "email", email, "error", err)
		}
	case entity.PaymentRejected:
		if err := s.mailer.SendPaymentFailedEmail(email, name, plan.DisplayName, statusDetail); err != nil {
			s.logger.Warn("failed to send payment failed email", "email", email, "error", err)
		}
	}
}

func mapPaymentStatus(mpStatus string) entity.PaymentStatus {
	switch mpStatus {
	case "approved":
		return entity.PaymentApproved
	case "pending", "in_process", "authorized":
		return entity.PaymentPending
	case "rejected":
		return entity.PaymentRejected
	case "refunded", "charged_back":
		return entity.PaymentRefunded
	default:
		return entity.PaymentPending
	}
}
