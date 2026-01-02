package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ProseBlockReferenceHandler handles HTTP requests for prose block references
type ProseBlockReferenceHandler struct {
	createReferenceUC  *proseblockapp.CreateProseBlockReferenceUseCase
	listByProseBlockUC *proseblockapp.ListProseBlockReferencesByProseBlockUseCase
	listByEntityUC     *proseblockapp.ListProseBlocksByEntityUseCase
	deleteReferenceUC  *proseblockapp.DeleteProseBlockReferenceUseCase
	logger             logger.Logger
}

// NewProseBlockReferenceHandler creates a new ProseBlockReferenceHandler
func NewProseBlockReferenceHandler(
	createReferenceUC *proseblockapp.CreateProseBlockReferenceUseCase,
	listByProseBlockUC *proseblockapp.ListProseBlockReferencesByProseBlockUseCase,
	listByEntityUC *proseblockapp.ListProseBlocksByEntityUseCase,
	deleteReferenceUC *proseblockapp.DeleteProseBlockReferenceUseCase,
	logger logger.Logger,
) *ProseBlockReferenceHandler {
	return &ProseBlockReferenceHandler{
		createReferenceUC:  createReferenceUC,
		listByProseBlockUC: listByProseBlockUC,
		listByEntityUC:     listByEntityUC,
		deleteReferenceUC:  deleteReferenceUC,
		logger:             logger,
	}
}

// Create handles POST /api/v1/prose-blocks/{id}/references
func (h *ProseBlockReferenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	proseBlockIDStr := r.PathValue("id")

	proseBlockID, err := uuid.Parse(proseBlockIDStr)
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

	output, err := h.createReferenceUC.Execute(r.Context(), proseblockapp.CreateProseBlockReferenceInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
		EntityType:   story.EntityType(req.EntityType),
		EntityID:     entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reference": output.Reference,
	})
}

// ListByProseBlock handles GET /api/v1/prose-blocks/{id}/references
func (h *ProseBlockReferenceHandler) ListByProseBlock(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	proseBlockIDStr := r.PathValue("id")

	proseBlockID, err := uuid.Parse(proseBlockIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listByProseBlockUC.Execute(r.Context(), proseblockapp.ListProseBlockReferencesByProseBlockInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": output.References,
		"total":      output.Total,
	})
}

// ListByScene handles GET /api/v1/scenes/{id}/prose-blocks
func (h *ProseBlockReferenceHandler) ListByScene(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listByEntityUC.Execute(r.Context(), proseblockapp.ListProseBlocksByEntityInput{
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
		"prose_blocks": output.ProseBlocks,
		"total":        output.Total,
	})
}

// ListByBeat handles GET /api/v1/beats/{id}/prose-blocks
func (h *ProseBlockReferenceHandler) ListByBeat(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listByEntityUC.Execute(r.Context(), proseblockapp.ListProseBlocksByEntityInput{
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
		"prose_blocks": output.ProseBlocks,
		"total":        output.Total,
	})
}

// Delete handles DELETE /api/v1/prose-block-references/{id}
func (h *ProseBlockReferenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.deleteReferenceUC.Execute(r.Context(), proseblockapp.DeleteProseBlockReferenceInput{
		TenantID: tenantID,
		ID:       refID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
