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

// ProseBlockHandler handles HTTP requests for prose blocks
type ProseBlockHandler struct {
	createProseBlockUseCase *proseblockapp.CreateProseBlockUseCase
	getProseBlockUseCase    *proseblockapp.GetProseBlockUseCase
	updateProseBlockUseCase *proseblockapp.UpdateProseBlockUseCase
	deleteProseBlockUseCase *proseblockapp.DeleteProseBlockUseCase
	listProseBlocksUseCase  *proseblockapp.ListProseBlocksUseCase
	logger                   logger.Logger
}

// NewProseBlockHandler creates a new ProseBlockHandler
func NewProseBlockHandler(
	createProseBlockUseCase *proseblockapp.CreateProseBlockUseCase,
	getProseBlockUseCase *proseblockapp.GetProseBlockUseCase,
	updateProseBlockUseCase *proseblockapp.UpdateProseBlockUseCase,
	deleteProseBlockUseCase *proseblockapp.DeleteProseBlockUseCase,
	listProseBlocksUseCase *proseblockapp.ListProseBlocksUseCase,
	logger logger.Logger,
) *ProseBlockHandler {
	return &ProseBlockHandler{
		createProseBlockUseCase: createProseBlockUseCase,
		getProseBlockUseCase:    getProseBlockUseCase,
		updateProseBlockUseCase: updateProseBlockUseCase,
		deleteProseBlockUseCase: deleteProseBlockUseCase,
		listProseBlocksUseCase:  listProseBlocksUseCase,
		logger:                  logger,
	}
}

// ListByChapter handles GET /api/v1/chapters/{id}/prose-blocks
func (h *ProseBlockHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listProseBlocksUseCase.Execute(r.Context(), proseblockapp.ListProseBlocksInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
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

// Create handles POST /api/v1/chapters/{id}/prose-blocks
func (h *ProseBlockHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.createProseBlockUseCase.Execute(r.Context(), proseblockapp.CreateProseBlockInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  req.OrderNum,
		Kind:      story.ProseKind(req.Kind),
		Content:   req.Content,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": output.ProseBlock,
	})
}

// Get handles GET /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getProseBlockUseCase.Execute(r.Context(), proseblockapp.GetProseBlockInput{
		TenantID: tenantID,
		ID:       proseBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": output.ProseBlock,
	})
}

// Update handles PUT /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
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

	input := proseblockapp.UpdateProseBlockInput{
		TenantID: tenantID,
		ID:       proseBlockID,
		OrderNum: req.OrderNum,
		Content:  req.Content,
	}
	if req.Kind != nil {
		kind := story.ProseKind(*req.Kind)
		input.Kind = &kind
	}

	output, err := h.updateProseBlockUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prose_block": output.ProseBlock,
	})
}

// Delete handles DELETE /api/v1/prose-blocks/{id}
func (h *ProseBlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	proseBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteProseBlockUseCase.Execute(r.Context(), proseblockapp.DeleteProseBlockInput{
		TenantID: tenantID,
		ID:       proseBlockID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

