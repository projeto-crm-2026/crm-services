package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/planservice"
)

type PlanHandler struct {
	service planservice.PlanService
}

func NewPlanHandler(svc planservice.PlanService) *PlanHandler {
	return &PlanHandler{service: svc}
}

func (h *PlanHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context())
	if err != nil {
		http.Error(w, "failed to retrieve plans", http.StatusInternalServerError)
		return
	}

	planResponses := model.NewPlanListResponse(plans)
	apiResponse := model.NewSuccessResponse("Plans retrieved successfully", planResponses)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *PlanHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rawID := r.PathValue("id")
	orgID, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid organization id", http.StatusBadRequest)
		return
	}

	if claims.OrganizationID == nil || *claims.OrganizationID != orgID {
		http.Error(w, "access denied - you can only view usage for your own organization", http.StatusForbidden)
		return
	}

	usage, err := h.service.GetOrganizationUsage(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, planservice.ErrOrganizationNotFound) {
			http.Error(w, "organization not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve usage", http.StatusInternalServerError)
		return
	}

	apiResponse := model.NewSuccessResponse("Usage retrieved successfully", usage)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}
