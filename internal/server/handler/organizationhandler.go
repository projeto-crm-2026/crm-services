package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/organizationservice"
)

type OrganizationHandler struct {
	service organizationservice.OrganizationService
}

func NewOrganizationHandler(svc organizationservice.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: svc}
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateOrganizationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	organization, err := h.service.Create(r.Context(), request.ToEntity())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.NewOrganizationResponse(organization)
	apiResponse := model.NewSuccessResponse("Organization created successfully", response)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	organization, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	response := model.NewOrganizationResponse(organization)
	apiResponse := model.NewSuccessResponse("Organization returned successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "invalid slug provided", http.StatusBadRequest)
		return
	}

	organization, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	response := model.NewOrganizationResponse(organization)
	apiResponse := model.NewSuccessResponse("Organization returned successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	var request model.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	organization, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	request.UpdateEntity(organization)

	err = h.service.Update(r.Context(), organization)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.NewOrganizationResponse(organization)
	apiResponse := model.NewSuccessResponse("Organization updated successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	_, err = h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "error while deleting the organization", http.StatusInternalServerError)
		return
	}

	apiResponse := model.NewSuccessResponse("Organization deleted successfully", nil)

	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	_, err = h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	err = h.service.SoftDelete(r.Context(), id)
	if err != nil {
		http.Error(w, "error while tried to soft delete the organization", http.StatusInternalServerError)
		return
	}

	apiResponse := model.NewSuccessResponse("Organization soft deleted successfully", nil)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *OrganizationHandler) Restore(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	err = h.service.Restore(r.Context(), id)
	if err != nil {
		http.Error(w, "error while tried to restore the organization", http.StatusInternalServerError)
		return
	}

	apiResponse := model.NewSuccessResponse("Organization restored successfully", nil)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}
