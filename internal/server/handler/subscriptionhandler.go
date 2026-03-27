package handler

import (
	"encoding/json"
	"net/http"

	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/subscriptionservice"
	"github.com/projeto-crm-2026/crm-services/pkg/mercadopago"
)

type SubscriptionHandler struct {
	service  subscriptionservice.SubscriptionService
	planRepo repo.PlanRepo
}

func NewSubscriptionHandler(svc subscriptionservice.SubscriptionService, planRepo repo.PlanRepo) *SubscriptionHandler {
	return &SubscriptionHandler{service: svc, planRepo: planRepo}
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	var req model.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := h.service.Subscribe(r.Context(), *claims.OrganizationID, req.PlanID, req.PayerEmail, req.CardTokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, _ := h.planRepo.GetByID(r.Context(), sub.PlanID)
	planUUID := req.PlanID
	if plan != nil {
		planUUID = plan.UUID
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Subscription created successfully", model.NewSubscriptionResponse(sub, planUUID)))
}

func (h *SubscriptionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	if err := h.service.Cancel(r.Context(), *claims.OrganizationID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Subscription cancelled. Access retained until end of billing period.", nil))
}

func (h *SubscriptionHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	var req model.UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := h.service.Upgrade(r.Context(), *claims.OrganizationID, req.PlanID, req.PayerEmail, req.CardTokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, _ := h.planRepo.GetByID(r.Context(), sub.PlanID)
	planUUID := req.PlanID
	if plan != nil {
		planUUID = plan.UUID
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Plan upgraded successfully", model.NewSubscriptionResponse(sub, planUUID)))
}

func (h *SubscriptionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var notification mercadopago.WebhookNotification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		w.WriteHeader(http.StatusOK) // 200 to MP
		return
	}

	// async
	w.WriteHeader(http.StatusOK)

	go func() {
		if err := h.service.HandleWebhook(r.Context(), &notification); err != nil {
			// log no service
		}
	}()
}

func (h *SubscriptionHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, constant.Unauthorized, http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	payments, err := h.service.ListPayments(r.Context(), *claims.OrganizationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.NewSuccessResponse("Payments retrieved", model.NewPaymentListResponse(payments)))
}
