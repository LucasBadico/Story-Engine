package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ContentAnchorHandler handles HTTP requests for content anchors
type ContentAnchorHandler struct {
	createAnchorUC       *contentblockapp.CreateContentAnchorUseCase
	listByContentBlockUC *contentblockapp.ListContentAnchorsByContentBlockUseCase
	listByEntityUC       *contentblockapp.ListContentBlocksByEntityUseCase
	deleteAnchorUC       *contentblockapp.DeleteContentAnchorUseCase
	logger               logger.Logger
}

// NewContentAnchorHandler creates a new ContentAnchorHandler
func NewContentAnchorHandler(
	createAnchorUC *contentblockapp.CreateContentAnchorUseCase,
	listByContentBlockUC *contentblockapp.ListContentAnchorsByContentBlockUseCase,
	listByEntityUC *contentblockapp.ListContentBlocksByEntityUseCase,
	deleteAnchorUC *contentblockapp.DeleteContentAnchorUseCase,
	logger logger.Logger,
) *ContentAnchorHandler {
	return &ContentAnchorHandler{
		createAnchorUC:       createAnchorUC,
		listByContentBlockUC: listByContentBlockUC,
		listByEntityUC:       listByEntityUC,
		deleteAnchorUC:       deleteAnchorUC,
		logger:               logger,
	}
}

// Create handles POST /api/v1/content-blocks/{id}/anchors
func (h *ContentAnchorHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	contentBlockIDStr := r.PathValue("id")

	contentBlockID, err := uuid.Parse(contentBlockIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
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

	output, err := h.createAnchorUC.Execute(r.Context(), contentblockapp.CreateContentAnchorInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
		EntityType:     story.EntityType(req.EntityType),
		EntityID:       entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anchor": output.Anchor,
	})
}

// ListByContentBlock handles GET /api/v1/content-blocks/{id}/anchors
func (h *ContentAnchorHandler) ListByContentBlock(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	contentBlockIDStr := r.PathValue("id")

	contentBlockID, err := uuid.Parse(contentBlockIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listByContentBlockUC.Execute(r.Context(), contentblockapp.ListContentAnchorsByContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anchors": output.Anchors,
		"total":   output.Total,
	})
}

// ListByScene handles GET /api/v1/scenes/{id}/content-blocks
func (h *ContentAnchorHandler) ListByScene(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	entityIDStr := r.PathValue("id")
	entityTypeStr := "scene"

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listByEntityUC.Execute(r.Context(), contentblockapp.ListContentBlocksByEntityInput{
		TenantID:   tenantID,
		EntityType: story.EntityType(entityTypeStr),
		EntityID:   entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content_blocks": output.ContentBlocks,
		"total":          output.Total,
	})
}

// ListByBeat handles GET /api/v1/beats/{id}/content-blocks
func (h *ContentAnchorHandler) ListByBeat(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	entityIDStr := r.PathValue("id")
	entityTypeStr := "beat"

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listByEntityUC.Execute(r.Context(), contentblockapp.ListContentBlocksByEntityInput{
		TenantID:   tenantID,
		EntityType: story.EntityType(entityTypeStr),
		EntityID:   entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content_blocks": output.ContentBlocks,
		"total":          output.Total,
	})
}

// Delete handles DELETE /api/v1/content-anchors/{id}
func (h *ContentAnchorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	refID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteAnchorUC.Execute(r.Context(), contentblockapp.DeleteContentAnchorInput{
		TenantID: tenantID,
		ID:       refID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


