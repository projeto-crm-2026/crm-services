package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/constant"
	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/server/model"
	"github.com/projeto-crm-2026/crm-services/internal/service/contactservice"
)

const (
	errContactNotFound = "contact not found"
	errInvalidID       = "invalid id provided"
)

type ContactHandler struct {
	service contactservice.ContactService
}

func NewContactHandler(svc contactservice.ContactService) *ContactHandler {
	return &ContactHandler{service: svc}
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	var request model.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contact := request.Parse()
	contact.OrganizationID = *claims.OrganizationID
	contact.CreatedByID = &claims.UserID

	created, err := h.service.Create(r.Context(), contact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(model.NewContactResponse(created))
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	var request model.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByID(r.Context(), id, *claims.OrganizationID)
	if err != nil {
		http.Error(w, errContactNotFound, http.StatusNotFound)
		return
	}

	request.Apply(contact)
	contact.UpdatedByID = &claims.UserID

	if err := h.service.Update(r.Context(), contact); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.NewContactResponse(contact))
}

func (h *ContactHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByID(r.Context(), id, *claims.OrganizationID)
	if err != nil {
		http.Error(w, errContactNotFound, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.NewContactResponse(contact))
}

func (h *ContactHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	email := r.PathValue("email")
	if email == "" {
		http.Error(w, "invalid email provided", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetByEmail(r.Context(), email, *claims.OrganizationID)
	if err != nil {
		http.Error(w, errContactNotFound, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.NewContactResponse(contact))
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	if _, err := h.service.GetByID(r.Context(), id, *claims.OrganizationID); err != nil {
		http.Error(w, errContactNotFound, http.StatusNotFound)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, "error while deleting the contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	if _, err := h.service.GetByID(r.Context(), id, *claims.OrganizationID); err != nil {
		http.Error(w, errContactNotFound, http.StatusNotFound)
		return
	}

	if err := h.service.SoftDelete(r.Context(), id); err != nil {
		http.Error(w, "error while deleting the contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	filters := model.ParseContactFilters(r)
	filters.OrganizationID = *claims.OrganizationID

	q := r.URL.Query()
	pageStr := q.Get("page")

	if pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(q.Get("page_size"))
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

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(model.NewPaginatedContactResponse(result))
		return
	}

	contacts, err := h.service.List(r.Context(), filters)
	if err != nil {
		http.Error(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.NewContactListResponse(contacts))
}

func (h *ContactHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.OrganizationID == nil {
		http.Error(w, constant.UserNotInOrganization, http.StatusForbidden)
		return
	}

	queryTerm := r.URL.Query().Get("q")
	if queryTerm == "" {
		http.Error(w, "search query 'q' is required", http.StatusBadRequest)
		return
	}

	filters := model.ParseContactFilters(r)
	filters.OrganizationID = *claims.OrganizationID

	contacts, err := h.service.Search(r.Context(), queryTerm, filters)
	if err != nil {
		http.Error(w, "failed to search contacts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.NewContactListResponse(contacts))
}
