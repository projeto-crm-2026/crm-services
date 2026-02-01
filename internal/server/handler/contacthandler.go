package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/contactservice"
)

type ContactHandler struct {
	service contactservice.ContactService
}

func NewContactHandler(svc contactservice.ContactService) *ContactHandler {
	return &ContactHandler{service: svc}
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateContactRequest

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	entity := request.ToEntity()
	entity.OrganizationID = *claims.OrganizationID
	contact, err := h.service.Create(r.Context(), entity)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.NewContactResponse(contact)

	apiResponse := model.NewSuccessResponse("Contact created successfully", response)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	var request model.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByID(r.Context(), id, organization_id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	request.UpdateEntity(contact)

	err = h.service.Update(r.Context(), contact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.NewContactResponse(contact)
	apiResponse := model.NewSuccessResponse("Contact updated successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByID(r.Context(), id, organization_id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	response := model.NewContactResponse(contact)
	apiResponse := model.NewSuccessResponse("Contact returned successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	email := r.PathValue("email")
	if email == "" {
		http.Error(w, "invalid email provided", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByEmail(r.Context(), email, organization_id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	response := model.NewContactResponse(contact)
	apiResponse := model.NewSuccessResponse("Contact returned successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	_, err = h.service.GetByID(r.Context(), id, organization_id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "error while deleting the user", http.StatusNotFound)
		return
	}

	apiResponse := model.NewSuccessResponse("Contact deleted successfully", nil)

	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, "invalid id provided", http.StatusBadRequest)
		return
	}

	_, err = h.service.GetByID(r.Context(), id, organization_id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	err = h.service.SoftDelete(r.Context(), id)
	if err != nil {
		http.Error(w, "error while tried to soft delete the user", http.StatusNotFound)
		return
	}

	apiResponse := model.NewSuccessResponse("Contact soft deleted successfully", nil)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := h.parseFilters(r)

	query := r.URL.Query()
	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	if pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}

		result, err := h.service.ListPaginated(r.Context(), filters, page, pageSize)
		if err != nil {
			http.Error(w, "failed to list contacts", http.StatusInternalServerError)
			return
		}

		response := model.NewPaginatedContactResponse(result)
		apiResponse := model.NewSuccessResponse("Contacts listed successfully", response)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(apiResponse)
		return
	}

	contacts, err := h.service.List(r.Context(), filters)
	if err != nil {
		http.Error(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}

	response := model.NewContactListResponse(contacts)
	apiResponse := model.NewSuccessResponse("Contacts listed successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.OrganizationID == nil {
		http.Error(w, "user not associated with any organization", http.StatusForbidden)
		return
	}

	organization_id := *claims.OrganizationID

	queryTerm := r.URL.Query().Get("q")
	if queryTerm == "" {
		http.Error(w, "search query 'q' is required", http.StatusBadRequest)
		return
	}

	filters := h.parseFilters(r)

	contacts, err := h.service.Search(r.Context(), queryTerm, filters, organization_id)
	if err != nil {
		http.Error(w, "failed to search contacts", http.StatusInternalServerError)
		return
	}

	response := model.NewContactListResponse(contacts)
	apiResponse := model.NewSuccessResponse("Search completed successfully", response)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse)
}

func (h *ContactHandler) parseFilters(r *http.Request) repo.ContactFilters {
	q := r.URL.Query()
	f := repo.ContactFilters{}

	// filtros
	if v := q.Get("status"); v != "" {
		f.Status = v
	}
	if v := q.Get("type"); v != "" {
		f.Type = v
	}
	if v := q.Get("source"); v != "" {
		f.Source = v
	}
	if v := q.Get("city"); v != "" {
		f.City = &v
	}
	if v := q.Get("state"); v != "" {
		f.State = &v
	}
	if v := q.Get("country"); v != "" {
		f.Country = &v
	}

	// tags (separadas por vírgula: ?tags=cliente,vip)
	if v := q.Get("tags"); v != "" {
		f.Tags = strings.Split(v, ",")
	}

	// ids
	if v := q.Get("assigned_to_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			uid := uint(id)
			f.AssignedToID = &uid
		}
	}
	if v := q.Get("created_by_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			uid := uint(id)
			f.CreatedByID = &uid
		}
	}

	if v := q.Get("created_after"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.CreatedAfter = &t
		}
	}
	if v := q.Get("created_before"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.CreatedBefore = &t
		}
	}

	return f
}
