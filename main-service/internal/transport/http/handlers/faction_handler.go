package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	factionapp "github.com/story-engine/main-service/internal/application/world/faction"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// FactionHandler handles HTTP requests for factions
type FactionHandler struct {
	createFactionUseCase   *factionapp.CreateFactionUseCase
	getFactionUseCase      *factionapp.GetFactionUseCase
	listFactionsUseCase    *factionapp.ListFactionsUseCase
	updateFactionUseCase   *factionapp.UpdateFactionUseCase
	deleteFactionUseCase   *factionapp.DeleteFactionUseCase
	getChildrenUseCase     *factionapp.GetChildrenUseCase
	addReferenceUseCase    *factionapp.AddReferenceUseCase
	removeReferenceUseCase *factionapp.RemoveReferenceUseCase
	getReferencesUseCase   *factionapp.GetReferencesUseCase
	updateReferenceUseCase *factionapp.UpdateReferenceUseCase
	logger                 logger.Logger
}

// NewFactionHandler creates a new FactionHandler
func NewFactionHandler(
	createFactionUseCase *factionapp.CreateFactionUseCase,
	getFactionUseCase *factionapp.GetFactionUseCase,
	listFactionsUseCase *factionapp.ListFactionsUseCase,
	updateFactionUseCase *factionapp.UpdateFactionUseCase,
	deleteFactionUseCase *factionapp.DeleteFactionUseCase,
	getChildrenUseCase *factionapp.GetChildrenUseCase,
	addReferenceUseCase *factionapp.AddReferenceUseCase,
	removeReferenceUseCase *factionapp.RemoveReferenceUseCase,
	getReferencesUseCase *factionapp.GetReferencesUseCase,
	updateReferenceUseCase *factionapp.UpdateReferenceUseCase,
	logger logger.Logger,
) *FactionHandler {
	return &FactionHandler{
		createFactionUseCase:   createFactionUseCase,
		getFactionUseCase:      getFactionUseCase,
		listFactionsUseCase:    listFactionsUseCase,
		updateFactionUseCase:   updateFactionUseCase,
		deleteFactionUseCase:   deleteFactionUseCase,
		getChildrenUseCase:     getChildrenUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/worlds/{world_id}/factions
func (h *FactionHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		ParentID    *string `json:"parent_id,omitempty"`
		Name        string  `json:"name"`
		Type        *string `json:"type,omitempty"`
		Description string  `json:"description"`
		Beliefs     string  `json:"beliefs"`
		Structure   string  `json:"structure"`
		Symbols     string  `json:"symbols"`
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

	output, err := h.createFactionUseCase.Execute(r.Context(), factionapp.CreateFactionInput{
		TenantID:    tenantID,
		WorldID:     worldID,
		ParentID:    parentID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Beliefs:     req.Beliefs,
		Structure:   req.Structure,
		Symbols:     req.Symbols,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"faction": output.Faction,
	})
}

// Get handles GET /api/v1/factions/{id}
func (h *FactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getFactionUseCase.Execute(r.Context(), factionapp.GetFactionInput{
		TenantID: tenantID,
		ID:       factionID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"faction": output.Faction,
	})
}

// List handles GET /api/v1/worlds/{world_id}/factions
func (h *FactionHandler) List(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listFactionsUseCase.Execute(r.Context(), factionapp.ListFactionsInput{
		TenantID: tenantID,
		WorldID:  worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factions": output.Factions,
		"total":    len(output.Factions),
	})
}

// Update handles PUT /api/v1/factions/{id}
func (h *FactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Type        *string `json:"type,omitempty"`
		Description *string `json:"description,omitempty"`
		Beliefs     *string `json:"beliefs,omitempty"`
		Structure   *string `json:"structure,omitempty"`
		Symbols     *string `json:"symbols,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateFactionUseCase.Execute(r.Context(), factionapp.UpdateFactionInput{
		TenantID:    tenantID,
		ID:          factionID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Beliefs:     req.Beliefs,
		Structure:   req.Structure,
		Symbols:     req.Symbols,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"faction": output.Faction,
	})
}

// Delete handles DELETE /api/v1/factions/{id}
func (h *FactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteFactionUseCase.Execute(r.Context(), factionapp.DeleteFactionInput{
		TenantID: tenantID,
		ID:       factionID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetChildren handles GET /api/v1/factions/{id}/children
func (h *FactionHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getChildrenUseCase.Execute(r.Context(), factionapp.GetChildrenInput{
		TenantID:  tenantID,
		FactionID: factionID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"children": output.Children,
		"total":    len(output.Children),
	})
}

// AddReference handles POST /api/v1/factions/{id}/references
func (h *FactionHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType string  `json:"entity_type"`
		EntityID   string  `json:"entity_id"`
		Role       *string `json:"role,omitempty"`
		Notes      string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addReferenceUseCase.Execute(r.Context(), factionapp.AddReferenceInput{
		TenantID:   tenantID,
		FactionID:  factionID,
		EntityType: req.EntityType,
		EntityID:   entityID,
		Role:       req.Role,
		Notes:      req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetReferences handles GET /api/v1/factions/{id}/references
func (h *FactionHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUseCase.Execute(r.Context(), factionapp.GetReferencesInput{
		TenantID:  tenantID,
		FactionID: factionID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": output.References,
		"total":      len(output.References),
	})
}

// RemoveReference handles DELETE /api/v1/factions/{id}/references/{entity_type}/{entity_id}
func (h *FactionHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	factionID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeReferenceUseCase.Execute(r.Context(), factionapp.RemoveReferenceInput{
		TenantID:   tenantID,
		FactionID:  factionID,
		EntityType: entityType,
		EntityID:   entityID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateReference handles PUT /api/v1/faction-references/{id}
func (h *FactionHandler) UpdateReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	referenceID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Role  *string `json:"role,omitempty"`
		Notes *string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if err := h.updateReferenceUseCase.Execute(r.Context(), factionapp.UpdateReferenceInput{
		TenantID: tenantID,
		ID:       referenceID,
		Role:     req.Role,
		Notes:    req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

