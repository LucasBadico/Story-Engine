package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// LocationHandler handles HTTP requests for locations
type LocationHandler struct {
	createLocationUseCase   *locationapp.CreateLocationUseCase
	getLocationUseCase      *locationapp.GetLocationUseCase
	listLocationsUseCase    *locationapp.ListLocationsUseCase
	updateLocationUseCase   *locationapp.UpdateLocationUseCase
	deleteLocationUseCase   *locationapp.DeleteLocationUseCase
	getChildrenUseCase      *locationapp.GetChildrenUseCase
	getAncestorsUseCase     *locationapp.GetAncestorsUseCase
	getDescendantsUseCase   *locationapp.GetDescendantsUseCase
	moveLocationUseCase     *locationapp.MoveLocationUseCase
	logger                  logger.Logger
}

// NewLocationHandler creates a new LocationHandler
func NewLocationHandler(
	createLocationUseCase *locationapp.CreateLocationUseCase,
	getLocationUseCase *locationapp.GetLocationUseCase,
	listLocationsUseCase *locationapp.ListLocationsUseCase,
	updateLocationUseCase *locationapp.UpdateLocationUseCase,
	deleteLocationUseCase *locationapp.DeleteLocationUseCase,
	getChildrenUseCase *locationapp.GetChildrenUseCase,
	getAncestorsUseCase *locationapp.GetAncestorsUseCase,
	getDescendantsUseCase *locationapp.GetDescendantsUseCase,
	moveLocationUseCase *locationapp.MoveLocationUseCase,
	logger logger.Logger,
) *LocationHandler {
	return &LocationHandler{
		createLocationUseCase: createLocationUseCase,
		getLocationUseCase:    getLocationUseCase,
		listLocationsUseCase:  listLocationsUseCase,
		updateLocationUseCase: updateLocationUseCase,
		deleteLocationUseCase: deleteLocationUseCase,
		getChildrenUseCase:    getChildrenUseCase,
		getAncestorsUseCase:   getAncestorsUseCase,
		getDescendantsUseCase: getDescendantsUseCase,
		moveLocationUseCase:   moveLocationUseCase,
		logger:                logger,
	}
}

// Create handles POST /api/v1/worlds/:world_id/locations
func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ParentID    *string `json:"parent_id"`
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentID = &pid
	}

	output, err := h.createLocationUseCase.Execute(r.Context(), locationapp.CreateLocationInput{
		TenantID:    tenantID,
		WorldID:     worldID,
		ParentID:    parentID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"location": output.Location,
	})
}

// Get handles GET /api/v1/locations/{id}
func (h *LocationHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getLocationUseCase.Execute(r.Context(), locationapp.GetLocationInput{
		TenantID: tenantID,
		ID:       locationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"location": output.Location,
	})
}

// List handles GET /api/v1/worlds/:world_id/locations?format=tree&limit=20&offset=0
func (h *LocationHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
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

	output, err := h.listLocationsUseCase.Execute(r.Context(), locationapp.ListLocationsInput{
		TenantID: tenantID,
		WorldID:  worldID,
		Format:   format,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locations": output.Locations,
		"total":     output.Total,
	})
}

// Update handles PUT /api/v1/locations/{id}
func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Type        *string `json:"type"`
		Description *string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateLocationUseCase.Execute(r.Context(), locationapp.UpdateLocationInput{
		TenantID:    tenantID,
		ID:          locationID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"location": output.Location,
	})
}

// Delete handles DELETE /api/v1/locations/{id}
func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteLocationUseCase.Execute(r.Context(), locationapp.DeleteLocationInput{
		TenantID: tenantID,
		ID:       locationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetChildren handles GET /api/v1/locations/{id}/children
func (h *LocationHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getChildrenUseCase.Execute(r.Context(), locationapp.GetChildrenInput{
		TenantID:   tenantID,
		LocationID: locationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"children": output.Children,
	})
}

// GetAncestors handles GET /api/v1/locations/{id}/ancestors
func (h *LocationHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getAncestorsUseCase.Execute(r.Context(), locationapp.GetAncestorsInput{
		TenantID:   tenantID,
		LocationID: locationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ancestors": output.Ancestors,
	})
}

// GetDescendants handles GET /api/v1/locations/{id}/descendants
func (h *LocationHandler) GetDescendants(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getDescendantsUseCase.Execute(r.Context(), locationapp.GetDescendantsInput{
		TenantID:   tenantID,
		LocationID: locationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"descendants": output.Descendants,
	})
}

// Move handles PUT /api/v1/locations/{id}/move
func (h *LocationHandler) Move(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ParentID *string `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentID = &pid
	}

	output, err := h.moveLocationUseCase.Execute(r.Context(), locationapp.MoveLocationInput{
		TenantID:   tenantID,
		LocationID: locationID,
		NewParentID: parentID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"location": output.Location,
	})
}


