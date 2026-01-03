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

// ContentBlockHandler handles HTTP requests for content blocks
type ContentBlockHandler struct {
	createContentBlockUseCase *contentblockapp.CreateContentBlockUseCase
	getContentBlockUseCase    *contentblockapp.GetContentBlockUseCase
	updateContentBlockUseCase *contentblockapp.UpdateContentBlockUseCase
	deleteContentBlockUseCase *contentblockapp.DeleteContentBlockUseCase
	listContentBlocksUseCase  *contentblockapp.ListContentBlocksUseCase
	logger                    logger.Logger
}

// NewContentBlockHandler creates a new ContentBlockHandler
func NewContentBlockHandler(
	createContentBlockUseCase *contentblockapp.CreateContentBlockUseCase,
	getContentBlockUseCase *contentblockapp.GetContentBlockUseCase,
	updateContentBlockUseCase *contentblockapp.UpdateContentBlockUseCase,
	deleteContentBlockUseCase *contentblockapp.DeleteContentBlockUseCase,
	listContentBlocksUseCase *contentblockapp.ListContentBlocksUseCase,
	logger logger.Logger,
) *ContentBlockHandler {
	return &ContentBlockHandler{
		createContentBlockUseCase: createContentBlockUseCase,
		getContentBlockUseCase:    getContentBlockUseCase,
		updateContentBlockUseCase: updateContentBlockUseCase,
		deleteContentBlockUseCase: deleteContentBlockUseCase,
		listContentBlocksUseCase:  listContentBlocksUseCase,
		logger:                    logger,
	}
}

// ListByChapter handles GET /api/v1/chapters/{id}/content-blocks
func (h *ContentBlockHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	chapterIDStr := r.PathValue("id")

	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listContentBlocksUseCase.Execute(r.Context(), contentblockapp.ListContentBlocksInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
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

// Create handles POST /api/v1/chapters/{id}/content-blocks
func (h *ContentBlockHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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
		chapterID = &parsedChapterID
	}

	var req struct {
		OrderNum *int                    `json:"order_num,omitempty"`
		Type     string                  `json:"type"`
		Kind     string                  `json:"kind"`
		Content  string                  `json:"content"`
		Metadata *story.ContentMetadata  `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	metadata := story.ContentMetadata{}
	if req.Metadata != nil {
		metadata = *req.Metadata
	}

	output, err := h.createContentBlockUseCase.Execute(r.Context(), contentblockapp.CreateContentBlockInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  req.OrderNum,
		Type:      story.ContentType(req.Type),
		Kind:      story.ContentKind(req.Kind),
		Content:   req.Content,
		Metadata:  metadata,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content_block": output.ContentBlock,
	})
}

// Get handles GET /api/v1/content-blocks/{id}
func (h *ContentBlockHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	contentBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getContentBlockUseCase.Execute(r.Context(), contentblockapp.GetContentBlockInput{
		TenantID: tenantID,
		ID:       contentBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content_block": output.ContentBlock,
	})
}

// Update handles PUT /api/v1/content-blocks/{id}
func (h *ContentBlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	contentBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		OrderNum *int                   `json:"order_num,omitempty"`
		Type     *string                `json:"type,omitempty"`
		Kind     *string                `json:"kind,omitempty"`
		Content  *string                `json:"content,omitempty"`
		Metadata *story.ContentMetadata `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	input := contentblockapp.UpdateContentBlockInput{
		TenantID: tenantID,
		ID:       contentBlockID,
		OrderNum: req.OrderNum,
		Content:  req.Content,
		Metadata: req.Metadata,
	}
	if req.Type != nil {
		contentType := story.ContentType(*req.Type)
		input.Type = &contentType
	}
	if req.Kind != nil {
		kind := story.ContentKind(*req.Kind)
		input.Kind = &kind
	}

	output, err := h.updateContentBlockUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content_block": output.ContentBlock,
	})
}

// Delete handles DELETE /api/v1/content-blocks/{id}
func (h *ContentBlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	contentBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteContentBlockUseCase.Execute(r.Context(), contentblockapp.DeleteContentBlockInput{
		TenantID: tenantID,
		ID:       contentBlockID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
