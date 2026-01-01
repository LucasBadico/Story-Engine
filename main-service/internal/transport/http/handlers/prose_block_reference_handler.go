package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ProseBlockReferenceHandler handles HTTP requests for prose block references
type ProseBlockReferenceHandler struct {
	refRepo      repositories.ProseBlockReferenceRepository
	proseBlockRepo repositories.ProseBlockRepository
	logger       logger.Logger
}

// NewProseBlockReferenceHandler creates a new ProseBlockReferenceHandler
func NewProseBlockReferenceHandler(
	refRepo repositories.ProseBlockReferenceRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *ProseBlockReferenceHandler {
	return &ProseBlockReferenceHandler{
		refRepo:        refRepo,
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// Create handles POST /api/v1/prose-blocks/{id}/references
func (h *ProseBlockReferenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	proseBlockIDStr := r.PathValue("id")

	proseBlockID, err := uuid.Parse(proseBlockIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Validate prose block exists
	_, err = h.proseBlockRepo.GetByID(r.Context(), proseBlockID)
	if err != nil {
		if err.Error() == "prose block not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       proseBlockIDStr,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
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

	if req.EntityType == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}, http.StatusBadRequest)
		return
	}

	if req.EntityID == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "entity_id is required",
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

	ref, err := story.NewProseBlockReference(proseBlockID, story.EntityType(req.EntityType), entityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "reference",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	if err := h.refRepo.Create(r.Context(), ref); err != nil {
		h.logger.Error("failed to create prose block reference", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reference": ref,
	})
}

// ListByProseBlock handles GET /api/v1/prose-blocks/{id}/references
func (h *ProseBlockReferenceHandler) ListByProseBlock(w http.ResponseWriter, r *http.Request) {
	proseBlockIDStr := r.PathValue("id")

	proseBlockID, err := uuid.Parse(proseBlockIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Validate prose block exists
	_, err = h.proseBlockRepo.GetByID(r.Context(), proseBlockID)
	if err != nil {
		if err.Error() == "prose block not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       proseBlockIDStr,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	references, err := h.refRepo.ListByProseBlock(r.Context(), proseBlockID)
	if err != nil {
		h.logger.Error("failed to list prose block references", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": references,
		"total":      len(references),
	})
}

// ListByScene handles GET /api/v1/scenes/{id}/prose-blocks
func (h *ProseBlockReferenceHandler) ListByScene(w http.ResponseWriter, r *http.Request) {
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

	if entityTypeStr == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}, http.StatusBadRequest)
		return
	}

	references, err := h.refRepo.ListByEntity(r.Context(), story.EntityType(entityTypeStr), entityID)
	if err != nil {
		h.logger.Error("failed to list prose block references by entity", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Get prose blocks for each reference
	proseBlocks := make([]*story.ProseBlock, 0, len(references))
	for _, ref := range references {
		proseBlock, err := h.proseBlockRepo.GetByID(r.Context(), ref.ProseBlockID)
		if err != nil {
			h.logger.Error("failed to get prose block", "prose_block_id", ref.ProseBlockID, "error", err)
			continue
		}
		proseBlocks = append(proseBlocks, proseBlock)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_blocks": proseBlocks,
		"total":        len(proseBlocks),
	})
}

// ListByBeat handles GET /api/v1/beats/{id}/prose-blocks
func (h *ProseBlockReferenceHandler) ListByBeat(w http.ResponseWriter, r *http.Request) {
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

	references, err := h.refRepo.ListByEntity(r.Context(), story.EntityType(entityTypeStr), entityID)
	if err != nil {
		h.logger.Error("failed to list prose block references by entity", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Get prose blocks for each reference
	proseBlocks := make([]*story.ProseBlock, 0, len(references))
	for _, ref := range references {
		proseBlock, err := h.proseBlockRepo.GetByID(r.Context(), ref.ProseBlockID)
		if err != nil {
			h.logger.Error("failed to get prose block", "prose_block_id", ref.ProseBlockID, "error", err)
			continue
		}
		proseBlocks = append(proseBlocks, proseBlock)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_blocks": proseBlocks,
		"total":        len(proseBlocks),
	})
}

// Delete handles DELETE /api/v1/prose-block-references/{id}
func (h *ProseBlockReferenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	refID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Check if reference exists
	_, err = h.refRepo.GetByID(r.Context(), refID)
	if err != nil {
		if err.Error() == "prose block reference not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block_reference",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err := h.refRepo.Delete(r.Context(), refID); err != nil {
		h.logger.Error("failed to delete prose block reference", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

