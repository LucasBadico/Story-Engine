package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// EntityRelationHandler handles HTTP requests for entity relations
type EntityRelationHandler struct {
	createRelationUseCase *relationapp.CreateRelationUseCase
	getRelationUseCase    *relationapp.GetRelationUseCase
	listBySourceUseCase   *relationapp.ListRelationsBySourceUseCase
	listByTargetUseCase   *relationapp.ListRelationsByTargetUseCase
	listByWorldUseCase    *relationapp.ListRelationsByWorldUseCase
	updateRelationUseCase *relationapp.UpdateRelationUseCase
	deleteRelationUseCase *relationapp.DeleteRelationUseCase
	logger                logger.Logger
}

// NewEntityRelationHandler creates a new EntityRelationHandler
func NewEntityRelationHandler(
	createRelationUseCase *relationapp.CreateRelationUseCase,
	getRelationUseCase *relationapp.GetRelationUseCase,
	listBySourceUseCase *relationapp.ListRelationsBySourceUseCase,
	listByTargetUseCase *relationapp.ListRelationsByTargetUseCase,
	listByWorldUseCase *relationapp.ListRelationsByWorldUseCase,
	updateRelationUseCase *relationapp.UpdateRelationUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	logger logger.Logger,
) *EntityRelationHandler {
	return &EntityRelationHandler{
		createRelationUseCase: createRelationUseCase,
		getRelationUseCase:    getRelationUseCase,
		listBySourceUseCase:   listBySourceUseCase,
		listByTargetUseCase:   listByTargetUseCase,
		listByWorldUseCase:    listByWorldUseCase,
		updateRelationUseCase: updateRelationUseCase,
		deleteRelationUseCase: deleteRelationUseCase,
		logger:                logger,
	}
}

// Create handles POST /api/v1/relations
// Both source_id and target_id are required - entities must exist before creating relations
func (h *EntityRelationHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		WorldID         string                 `json:"world_id"`
		SourceType      string                 `json:"source_type"`
		SourceID        string                 `json:"source_id"`
		TargetType      string                 `json:"target_type"`
		TargetID        string                 `json:"target_id"`
		RelationType    string                 `json:"relation_type"`
		ContextType     *string                `json:"context_type,omitempty"`
		ContextID       *string                `json:"context_id,omitempty"`
		Attributes      map[string]interface{} `json:"attributes,omitempty"`
		Summary         string                 `json:"summary,omitempty"`
		CreatedByUserID *string                `json:"created_by_user_id,omitempty"`
		CreateMirror    bool                   `json:"create_mirror,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	// Parse UUIDs
	worldID, err := uuid.Parse(req.WorldID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// source_id is required
	if req.SourceID == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_id",
			Message: "source_id is required",
		}, http.StatusBadRequest)
		return
	}
	sourceID, err := uuid.Parse(req.SourceID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// target_id is required
	if req.TargetID == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_id",
			Message: "target_id is required",
		}, http.StatusBadRequest)
		return
	}
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var contextID *uuid.UUID
	if req.ContextID != nil && *req.ContextID != "" {
		parsedID, err := uuid.Parse(*req.ContextID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "context_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		contextID = &parsedID
	}

	var createdByUserID *uuid.UUID
	if req.CreatedByUserID != nil && *req.CreatedByUserID != "" {
		parsedID, err := uuid.Parse(*req.CreatedByUserID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "created_by_user_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		createdByUserID = &parsedID
	}

	output, err := h.createRelationUseCase.Execute(r.Context(), relationapp.CreateRelationInput{
		TenantID:        tenantID,
		WorldID:         worldID,
		SourceType:      req.SourceType,
		SourceID:        sourceID,
		TargetType:      req.TargetType,
		TargetID:        targetID,
		RelationType:    req.RelationType,
		ContextType:     req.ContextType,
		ContextID:       contextID,
		Attributes:      req.Attributes,
		Summary:         req.Summary,
		CreatedByUserID: createdByUserID,
		CreateMirror:    req.CreateMirror,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"relation": output.Relation,
	}
	if output.Mirror != nil {
		response["mirror"] = output.Mirror
	}
	json.NewEncoder(w).Encode(response)
}

// Get handles GET /api/v1/relations/{id}
func (h *EntityRelationHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	relationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getRelationUseCase.Execute(r.Context(), relationapp.GetRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relation": output.Relation,
	})
}

// Update handles PUT /api/v1/relations/{id}
func (h *EntityRelationHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	relationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		RelationType *string                 `json:"relation_type,omitempty"`
		Attributes   *map[string]interface{} `json:"attributes,omitempty"`
		Summary      *string                 `json:"summary,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateRelationUseCase.Execute(r.Context(), relationapp.UpdateRelationInput{
		TenantID:     tenantID,
		ID:           relationID,
		RelationType: req.RelationType,
		Attributes:   req.Attributes,
		Summary:      req.Summary,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relation": output.Relation,
	})
}

// Delete handles DELETE /api/v1/relations/{id}
func (h *EntityRelationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")
	relationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteRelationUseCase.Execute(r.Context(), relationapp.DeleteRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByWorld handles GET /api/v1/worlds/{world_id}/relations
func (h *EntityRelationHandler) ListByWorld(w http.ResponseWriter, r *http.Request) {
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

	opts := h.parseListOptions(r)

	output, err := h.listByWorldUseCase.Execute(r.Context(), relationapp.ListRelationsByWorldInput{
		TenantID: tenantID,
		WorldID:  worldID,
		Options:  opts,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	h.writePaginatedResponse(w, output.Relations)
}

// ListBySource handles GET /api/v1/relations/source?source_type=X&source_id=Y
func (h *EntityRelationHandler) ListBySource(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sourceType := r.URL.Query().Get("source_type")
	if sourceType == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_type",
			Message: "source_type is required",
		}, http.StatusBadRequest)
		return
	}

	sourceIDStr := r.URL.Query().Get("source_id")
	if sourceIDStr == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_id",
			Message: "source_id is required",
		}, http.StatusBadRequest)
		return
	}

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	opts := h.parseListOptions(r)

	output, err := h.listBySourceUseCase.Execute(r.Context(), relationapp.ListRelationsBySourceInput{
		TenantID:   tenantID,
		SourceType: sourceType,
		SourceID:   sourceID,
		Options:    opts,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	h.writePaginatedResponse(w, output.Relations)
}

// ListByTarget handles GET /api/v1/relations/target?target_type=X&target_id=Y
func (h *EntityRelationHandler) ListByTarget(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	targetType := r.URL.Query().Get("target_type")
	if targetType == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_type",
			Message: "target_type is required",
		}, http.StatusBadRequest)
		return
	}

	targetIDStr := r.URL.Query().Get("target_id")
	if targetIDStr == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_id",
			Message: "target_id is required",
		}, http.StatusBadRequest)
		return
	}

	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	opts := h.parseListOptions(r)

	output, err := h.listByTargetUseCase.Execute(r.Context(), relationapp.ListRelationsByTargetInput{
		TenantID:   tenantID,
		TargetType: targetType,
		TargetID:   targetID,
		Options:    opts,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	h.writePaginatedResponse(w, output.Relations)
}

// parseListOptions parses query parameters into ListOptions
func (h *EntityRelationHandler) parseListOptions(r *http.Request) repositories.ListOptions {
	opts := repositories.ListOptions{}

	// Cursor
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		opts.Cursor = &cursor
	}

	// Limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	// Relation type
	if relationType := r.URL.Query().Get("relation_type"); relationType != "" {
		opts.RelationType = &relationType
	}

	// Order by
	opts.OrderBy = r.URL.Query().Get("order_by")
	if opts.OrderBy == "" {
		opts.OrderBy = "created_at"
	}

	// Order direction
	opts.OrderDir = r.URL.Query().Get("order_dir")
	if opts.OrderDir == "" {
		opts.OrderDir = "asc"
	}

	// Exclude mirrors
	if excludeMirrors := r.URL.Query().Get("exclude_mirrors"); excludeMirrors == "true" {
		opts.ExcludeMirrors = true
	}

	return opts
}

// writePaginatedResponse writes a paginated response with cursor
func (h *EntityRelationHandler) writePaginatedResponse(w http.ResponseWriter, result *repositories.ListResult) {
	pagination := map[string]interface{}{
		"has_more": result.HasMore,
	}

	if result.NextCursor != nil {
		pagination["next_cursor"] = *result.NextCursor
	}

	response := map[string]interface{}{
		"data":       result.Items,
		"pagination": pagination,
	}

	json.NewEncoder(w).Encode(response)
}
