package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	loreapp "github.com/story-engine/main-service/internal/application/world/lore"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// LoreHandler handles HTTP requests for lores
type LoreHandler struct {
	createLoreUseCase      *loreapp.CreateLoreUseCase
	getLoreUseCase         *loreapp.GetLoreUseCase
	listLoresUseCase       *loreapp.ListLoresUseCase
	updateLoreUseCase      *loreapp.UpdateLoreUseCase
	deleteLoreUseCase      *loreapp.DeleteLoreUseCase
	getChildrenUseCase     *loreapp.GetChildrenUseCase
	addReferenceUseCase    *loreapp.AddReferenceUseCase
	removeReferenceUseCase *loreapp.RemoveReferenceUseCase
	getReferencesUseCase   *loreapp.GetReferencesUseCase
	updateReferenceUseCase *loreapp.UpdateReferenceUseCase
	logger                 logger.Logger
}

// NewLoreHandler creates a new LoreHandler
func NewLoreHandler(
	createLoreUseCase *loreapp.CreateLoreUseCase,
	getLoreUseCase *loreapp.GetLoreUseCase,
	listLoresUseCase *loreapp.ListLoresUseCase,
	updateLoreUseCase *loreapp.UpdateLoreUseCase,
	deleteLoreUseCase *loreapp.DeleteLoreUseCase,
	getChildrenUseCase *loreapp.GetChildrenUseCase,
	addReferenceUseCase *loreapp.AddReferenceUseCase,
	removeReferenceUseCase *loreapp.RemoveReferenceUseCase,
	getReferencesUseCase *loreapp.GetReferencesUseCase,
	updateReferenceUseCase *loreapp.UpdateReferenceUseCase,
	logger logger.Logger,
) *LoreHandler {
	return &LoreHandler{
		createLoreUseCase:      createLoreUseCase,
		getLoreUseCase:         getLoreUseCase,
		listLoresUseCase:       listLoresUseCase,
		updateLoreUseCase:      updateLoreUseCase,
		deleteLoreUseCase:      deleteLoreUseCase,
		getChildrenUseCase:     getChildrenUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/worlds/{world_id}/lores
func (h *LoreHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		ParentID     *string `json:"parent_id,omitempty"`
		Name         string  `json:"name"`
		Category     *string `json:"category,omitempty"`
		Description  string  `json:"description"`
		Rules        string  `json:"rules"`
		Limitations  string  `json:"limitations"`
		Requirements string  `json:"requirements"`
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

	output, err := h.createLoreUseCase.Execute(r.Context(), loreapp.CreateLoreInput{
		TenantID:     tenantID,
		WorldID:      worldID,
		ParentID:     parentID,
		Name:         req.Name,
		Category:     req.Category,
		Description:  req.Description,
		Rules:        req.Rules,
		Limitations:  req.Limitations,
		Requirements: req.Requirements,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lore": output.Lore,
	})
}

// Get handles GET /api/v1/lores/{id}
func (h *LoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getLoreUseCase.Execute(r.Context(), loreapp.GetLoreInput{
		TenantID: tenantID,
		ID:       loreID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lore": output.Lore,
	})
}

// List handles GET /api/v1/worlds/{world_id}/lores
func (h *LoreHandler) List(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listLoresUseCase.Execute(r.Context(), loreapp.ListLoresInput{
		TenantID: tenantID,
		WorldID:  worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lores": output.Lores,
		"total": len(output.Lores),
	})
}

// Update handles PUT /api/v1/lores/{id}
func (h *LoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name         *string `json:"name,omitempty"`
		Category     *string `json:"category,omitempty"`
		Description  *string `json:"description,omitempty"`
		Rules        *string `json:"rules,omitempty"`
		Limitations  *string `json:"limitations,omitempty"`
		Requirements *string `json:"requirements,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateLoreUseCase.Execute(r.Context(), loreapp.UpdateLoreInput{
		TenantID:     tenantID,
		ID:           loreID,
		Name:         req.Name,
		Category:     req.Category,
		Description:  req.Description,
		Rules:        req.Rules,
		Limitations:  req.Limitations,
		Requirements: req.Requirements,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lore": output.Lore,
	})
}

// Delete handles DELETE /api/v1/lores/{id}
func (h *LoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteLoreUseCase.Execute(r.Context(), loreapp.DeleteLoreInput{
		TenantID: tenantID,
		ID:       loreID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetChildren handles GET /api/v1/lores/{id}/children
func (h *LoreHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getChildrenUseCase.Execute(r.Context(), loreapp.GetChildrenInput{
		TenantID: tenantID,
		LoreID:   loreID,
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

// AddReference handles POST /api/v1/lores/{id}/references
func (h *LoreHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType       string  `json:"entity_type"`
		EntityID         string  `json:"entity_id"`
		RelationshipType *string `json:"relationship_type,omitempty"`
		Notes            string  `json:"notes"`
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

	if err := h.addReferenceUseCase.Execute(r.Context(), loreapp.AddReferenceInput{
		TenantID:         tenantID,
		LoreID:           loreID,
		EntityType:       req.EntityType,
		EntityID:         entityID,
		RelationshipType: req.RelationshipType,
		Notes:            req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetReferences handles GET /api/v1/lores/{id}/references
func (h *LoreHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	loreID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUseCase.Execute(r.Context(), loreapp.GetReferencesInput{
		TenantID: tenantID,
		LoreID:   loreID,
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

// RemoveReference handles DELETE /api/v1/lores/{id}/references/{entity_type}/{entity_id}
func (h *LoreHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	loreID, err := uuid.Parse(id)
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

	if err := h.removeReferenceUseCase.Execute(r.Context(), loreapp.RemoveReferenceInput{
		TenantID:   tenantID,
		LoreID:     loreID,
		EntityType: entityType,
		EntityID:   entityID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateReference handles PUT /api/v1/lore-references/{id}
func (h *LoreHandler) UpdateReference(w http.ResponseWriter, r *http.Request) {
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
		RelationshipType *string `json:"relationship_type,omitempty"`
		Notes            *string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if err := h.updateReferenceUseCase.Execute(r.Context(), loreapp.UpdateReferenceInput{
		TenantID:         tenantID,
		ID:               referenceID,
		RelationshipType: req.RelationshipType,
		Notes:            req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

