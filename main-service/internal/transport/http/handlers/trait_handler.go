package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// TraitHandler handles HTTP requests for traits
type TraitHandler struct {
	createTraitUseCase *traitapp.CreateTraitUseCase
	getTraitUseCase    *traitapp.GetTraitUseCase
	listTraitsUseCase  *traitapp.ListTraitsUseCase
	updateTraitUseCase *traitapp.UpdateTraitUseCase
	deleteTraitUseCase *traitapp.DeleteTraitUseCase
	logger             logger.Logger
}

// NewTraitHandler creates a new TraitHandler
func NewTraitHandler(
	createTraitUseCase *traitapp.CreateTraitUseCase,
	getTraitUseCase *traitapp.GetTraitUseCase,
	listTraitsUseCase *traitapp.ListTraitsUseCase,
	updateTraitUseCase *traitapp.UpdateTraitUseCase,
	deleteTraitUseCase *traitapp.DeleteTraitUseCase,
	logger logger.Logger,
) *TraitHandler {
	return &TraitHandler{
		createTraitUseCase: createTraitUseCase,
		getTraitUseCase:    getTraitUseCase,
		listTraitsUseCase:  listTraitsUseCase,
		updateTraitUseCase: updateTraitUseCase,
		deleteTraitUseCase: deleteTraitUseCase,
		logger:             logger,
	}
}

// Create handles POST /api/v1/traits
func (h *TraitHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createTraitUseCase.Execute(r.Context(), traitapp.CreateTraitInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trait": output.Trait,
	})
}

// Get handles GET /api/v1/traits/{id}
func (h *TraitHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	traitID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getTraitUseCase.Execute(r.Context(), traitapp.GetTraitInput{
		ID: traitID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trait": output.Trait,
	})
}

// List handles GET /api/v1/traits?limit=20&offset=0
func (h *TraitHandler) List(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listTraitsUseCase.Execute(r.Context(), traitapp.ListTraitsInput{
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
		"traits": output.Traits,
		"total":  output.Total,
	})
}

// Update handles PUT /api/v1/traits/{id}
func (h *TraitHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	traitID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Category    *string `json:"category"`
		Description *string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateTraitUseCase.Execute(r.Context(), traitapp.UpdateTraitInput{
		ID: traitID,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trait": output.Trait,
	})
}

// Delete handles DELETE /api/v1/traits/{id}
func (h *TraitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	traitID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteTraitUseCase.Execute(r.Context(), traitapp.DeleteTraitInput{
		ID: traitID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


