package model

import (
	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type PlanResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	DisplayName       string    `json:"display_name"`
	PriceCents        int       `json:"price_cents"`
	Currency          string    `json:"currency"`
	MaxContacts       int       `json:"max_contacts"`
	MaxMembers        int       `json:"max_members"`
	MaxChatResponders int       `json:"max_chat_responders"`
}

func NewPlanResponse(p *entity.Plan) PlanResponse {
	return PlanResponse{
		ID:                p.UUID,
		Name:              p.Name,
		DisplayName:       p.DisplayName,
		PriceCents:        p.PriceCents,
		Currency:          p.Currency,
		MaxContacts:       p.MaxContacts,
		MaxMembers:        p.MaxMembers,
		MaxChatResponders: p.MaxChatResponders,
	}
}

func NewPlanListResponse(plans []entity.Plan) []PlanResponse {
	result := make([]PlanResponse, len(plans))
	for i, p := range plans {
		result[i] = NewPlanResponse(&p)
	}
	return result
}

type UsageWarning string

const (
	WarningNone     UsageWarning = ""
	WarningAt80     UsageWarning = "approaching_limit"
	WarningAt90     UsageWarning = "near_limit"
	WarningAtLimit  UsageWarning = "limit_reached"
)

type UsageResource struct {
	Current int          `json:"current"`
	Limit   int          `json:"limit"`
	Warning UsageWarning `json:"warning,omitempty"`
}

func NewUsageResource(current, limit int) UsageResource {
	r := UsageResource{Current: current, Limit: limit}
	if limit > 0 {
		pct := float64(current) / float64(limit) * 100
		switch {
		case current >= limit:
			r.Warning = WarningAtLimit
		case pct >= 90:
			r.Warning = WarningAt90
		case pct >= 80:
			r.Warning = WarningAt80
		}
	}
	return r
}

type SubscriptionInfo struct {
	Status             string  `json:"status"`
	CurrentPeriodStart *string `json:"current_period_start"`
	CurrentPeriodEnd   *string `json:"current_period_end"`
}

type UsageResponse struct {
	Plan         PlanResponse     `json:"plan"`
	Subscription SubscriptionInfo `json:"subscription"`
	Usage        UsageData        `json:"usage"`
}

type UsageData struct {
	Contacts       UsageResource `json:"contacts"`
	Members        UsageResource `json:"members"`
	ChatResponders UsageResource `json:"chat_responders"`
}
