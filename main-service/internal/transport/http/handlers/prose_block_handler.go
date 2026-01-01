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

// ProseBlockHandler handles HTTP requests for prose blocks
type ProseBlockHandler struct {
	proseBlockRepo repositories.ProseBlockRepository
	chapterRepo    repositories.ChapterRepository
	logger         logger.Logger
}

// NewProseBlockHandler creates a new ProseBlockHandler
func NewProseBlockHandler(
	proseBlockRepo repositories.ProseBlockRepository,
	chapterRepo repositories.ChapterRepository,
	logger logger.Logger,
) *ProseBlockHandler {
	return &ProseBlockHandler{
		proseBlockRepo: proseBlockRepo,
		chapterRepo:    chapterRepo,
		logger:         logger,
	}
}

// ListByChapter handles GET /api/v1/chapters/{id}/prose-blocks
func (h *ProseBlockHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	chapterIDStr := r.PathValue("id")

	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Validate chapter exists
	_, err = h.chapterRepo.GetByID(r.Context(), chapterID)
	if err != nil {
		if err.Error() == "chapter not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       chapterIDStr,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	proseBlocks, err := h.proseBlockRepo.ListByChapter(r.Context(), chapterID)
	if err != nil {
		h.logger.Error("failed to list prose blocks", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_blocks": proseBlocks,
		"total":        len(proseBlocks),
	})
}

// Create handles POST /api/v1/chapters/{id}/prose-blocks
func (h *ProseBlockHandler) Create(w http.ResponseWriter, r *http.Request) {
	chapterIDStr := r.PathValue("id")

	var chapterID *uuid.UUID
	if chapterIDStr != "" {
		parsedChapterID, err := uuid.Parse(chapterIDStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}

		// Validate chapter exists if provided
		_, err = h.chapterRepo.GetByID(r.Context(), parsedChapterID)
		if err != nil {
			if err.Error() == "chapter not found" {
				WriteError(w, &platformerrors.NotFoundError{
					Resource: "chapter",
					ID:       chapterIDStr,
				}, http.StatusNotFound)
			} else {
				WriteError(w, err, http.StatusInternalServerError)
			}
			return
		}
		chapterID = &parsedChapterID
	}

	var req struct {
		OrderNum *int   `json:"order_num,omitempty"`
		Kind     string `json:"kind"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if req.OrderNum != nil && *req.OrderNum < 1 {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}, http.StatusBadRequest)
		return
	}

	if req.Kind == "" {
		req.Kind = "final"
	}

	if req.Content == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "content",
			Message: "content is required",
		}, http.StatusBadRequest)
		return
	}

	proseBlock, err := story.NewProseBlock(chapterID, req.OrderNum, story.ProseKind(req.Kind), req.Content)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "prose_block",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	if err := h.proseBlockRepo.Create(r.Context(), proseBlock); err != nil {
		h.logger.Error("failed to create prose block", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": proseBlock,
	})
}

// Get handles GET /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	proseBlock, err := h.proseBlockRepo.GetByID(r.Context(), proseBlockID)
	if err != nil {
		if err.Error() == "prose block not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": proseBlock,
	})
}

// Update handles PUT /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing prose block
	proseBlock, err := h.proseBlockRepo.GetByID(r.Context(), proseBlockID)
	if err != nil {
		if err.Error() == "prose block not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		OrderNum *int    `json:"order_num,omitempty"`
		Kind     *string `json:"kind,omitempty"`
		Content  *string `json:"content,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.OrderNum != nil {
		if *req.OrderNum < 1 {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "order_num",
				Message: "must be greater than 0",
			}, http.StatusBadRequest)
			return
		}
		proseBlock.OrderNum = *req.OrderNum
	}

	if req.Kind != nil {
		proseBlock.Kind = story.ProseKind(*req.Kind)
	}

	if req.Content != nil {
		proseBlock.UpdateContent(*req.Content)
	}

	if err := proseBlock.Validate(); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "prose_block",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	if err := h.proseBlockRepo.Update(r.Context(), proseBlock); err != nil {
		h.logger.Error("failed to update prose block", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": proseBlock,
	})
}

// Delete handles DELETE /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Check if prose block exists
	_, err = h.proseBlockRepo.GetByID(r.Context(), proseBlockID)
	if err != nil {
		if err.Error() == "prose block not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "prose_block",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err := h.proseBlockRepo.Delete(r.Context(), proseBlockID); err != nil {
		h.logger.Error("failed to delete prose block", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

