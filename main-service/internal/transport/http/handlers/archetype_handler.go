package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ArchetypeHandler handles HTTP requests for archetypes
type ArchetypeHandler struct {
	createArchetypeUseCase     *archetypeapp.CreateArchetypeUseCase
	getArchetypeUseCase        *archetypeapp.GetArchetypeUseCase
	listArchetypesUseCase      *archetypeapp.ListArchetypesUseCase
	updateArchetypeUseCase     *archetypeapp.UpdateArchetypeUseCase
	deleteArchetypeUseCase     *archetypeapp.DeleteArchetypeUseCase
	addTraitUseCase            *archetypeapp.AddTraitToArchetypeUseCase
	removeTraitUseCase         *archetypeapp.RemoveTraitFromArchetypeUseCase
	logger                     logger.Logger
}

// NewArchetypeHandler creates a new ArchetypeHandler
func NewArchetypeHandler(
	createArchetypeUseCase *archetypeapp.CreateArchetypeUseCase,
	getArchetypeUseCase *archetypeapp.GetArchetypeUseCase,
	listArchetypesUseCase *archetypeapp.ListArchetypesUseCase,
	updateArchetypeUseCase *archetypeapp.UpdateArchetypeUseCase,
	deleteArchetypeUseCase *archetypeapp.DeleteArchetypeUseCase,
	addTraitUseCase *archetypeapp.AddTraitToArchetypeUseCase,
	removeTraitUseCase *archetypeapp.RemoveTraitFromArchetypeUseCase,
	logger logger.Logger,
) *ArchetypeHandler {
	return &ArchetypeHandler{
		createArchetypeUseCase: createArchetypeUseCase,
		getArchetypeUseCase:    getArchetypeUseCase,
		listArchetypesUseCase:  listArchetypesUseCase,
		updateArchetypeUseCase: updateArchetypeUseCase,
		deleteArchetypeUseCase: deleteArchetypeUseCase,
		addTraitUseCase:        addTraitUseCase,
		removeTraitUseCase:     removeTraitUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/archetypes
func (h *ArchetypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createArchetypeUseCase.Execute(r.Context(), archetypeapp.CreateArchetypeInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"archetype": output.Archetype,
	})
}

// Get handles GET /api/v1/archetypes/{id}
func (h *ArchetypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	archetypeID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getArchetypeUseCase.Execute(r.Context(), archetypeapp.GetArchetypeInput{
		TenantID: tenantID,
		ID:       archetypeID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"archetype": output.Archetype,
	})
}

// List handles GET /api/v1/archetypes?limit=20&offset=0
func (h *ArchetypeHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	output, err := h.listArchetypesUseCase.Execute(r.Context(), archetypeapp.ListArchetypesInput{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"archetypes": output.Archetypes,
		"total":      output.Total,
	})
}

// Update handles PUT /api/v1/archetypes/{id}
func (h *ArchetypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	archetypeID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateArchetypeUseCase.Execute(r.Context(), archetypeapp.UpdateArchetypeInput{
		TenantID:    tenantID,
		ID:          archetypeID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"archetype": output.Archetype,
	})
}

// Delete handles DELETE /api/v1/archetypes/{id}
func (h *ArchetypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	archetypeID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteArchetypeUseCase.Execute(r.Context(), archetypeapp.DeleteArchetypeInput{
		TenantID: tenantID,
		ID:       archetypeID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddTrait handles POST /api/v1/archetypes/{id}/traits
func (h *ArchetypeHandler) AddTrait(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	archetypeID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		TraitID      string `json:"trait_id"`
		DefaultValue string `json:"default_value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	traitID, err := uuid.Parse(req.TraitID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.addTraitUseCase.Execute(r.Context(), archetypeapp.AddTraitToArchetypeInput{
		TenantID:     tenantID,
		ArchetypeID:  archetypeID,
		TraitID:      traitID,
		DefaultValue: req.DefaultValue,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveTrait handles DELETE /api/v1/archetypes/{id}/traits/{trait_id}
func (h *ArchetypeHandler) RemoveTrait(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	archetypeID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	traitIDStr := r.PathValue("trait_id")
	traitID, err := uuid.Parse(traitIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.removeTraitUseCase.Execute(r.Context(), archetypeapp.RemoveTraitFromArchetypeInput{
		TenantID:    tenantID,
		ArchetypeID: archetypeID,
		TraitID:     traitID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


