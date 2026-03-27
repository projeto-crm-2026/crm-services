package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/pkg/utils"
)

type SubscribeRequest struct {
	PlanID      uuid.UUID `json:"plan_id"`
	PayerEmail  string    `json:"payer_email"`
	CardTokenID string    `json:"card_token_id"`
}

func (r *SubscribeRequest) Validate() error {
	if r.PlanID == uuid.Nil {
		return fmt.Errorf("plan_id is required")
	}
	if r.PayerEmail == "" {
		return fmt.Errorf("payer_email is required")
	}
	if r.CardTokenID == "" {
		return fmt.Errorf("card_token_id is required")
	}
	return nil
}

type UpgradeRequest struct {
	PlanID      uuid.UUID `json:"plan_id"`
	PayerEmail  string    `json:"payer_email"`
	CardTokenID string    `json:"card_token_id"`
}

func (r *UpgradeRequest) Validate() error {
	if r.PlanID == uuid.Nil {
		return fmt.Errorf("plan_id is required")
	}
	if r.PayerEmail == "" {
		return fmt.Errorf("payer_email is required")
	}
	if r.CardTokenID == "" {
		return fmt.Errorf("card_token_id is required")
	}
	return nil
}

type SubscriptionResponse struct {
	ID                 uuid.UUID `json:"id"`
	PlanID             uuid.UUID `json:"plan_id"`
	Status             string    `json:"status"`
	CurrentPeriodStart *string   `json:"current_period_start"`
	CurrentPeriodEnd   *string   `json:"current_period_end"`
}

func NewSubscriptionResponse(sub *entity.Subscription, planUUID uuid.UUID) SubscriptionResponse {
	return SubscriptionResponse{
		ID:                 sub.UUID,
		PlanID:             planUUID,
		Status:             string(sub.Status),
		CurrentPeriodStart: utils.FormatTimePtr(sub.CurrentPeriodStart),
		CurrentPeriodEnd:   utils.FormatTimePtr(sub.CurrentPeriodEnd),
	}
}

type PaymentResponse struct {
	ID            uuid.UUID `json:"id"`
	Status        string    `json:"status"`
	StatusDetail  string    `json:"status_detail,omitempty"`
	AmountCents   int       `json:"amount_cents"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	Description   string    `json:"description,omitempty"`
	PaidAt        *string   `json:"paid_at,omitempty"`
	CreatedAt     string    `json:"created_at"`
}

func NewPaymentResponse(p *entity.Payment) PaymentResponse {
	return PaymentResponse{
		ID:            p.UUID,
		Status:        string(p.Status),
		StatusDetail:  p.StatusDetail,
		AmountCents:   p.AmountCents,
		Currency:      p.Currency,
		PaymentMethod: p.PaymentMethod,
		Description:   p.Description,
		PaidAt:        utils.FormatTimePtr(p.PaidAt),
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func NewPaymentListResponse(payments []entity.Payment) []PaymentResponse {
	result := make([]PaymentResponse, len(payments))
	for i := range payments {
		result[i] = NewPaymentResponse(&payments[i])
	}
	return result
}
