package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ChapterHandler handles HTTP requests for chapters
type ChapterHandler struct {
	chapterRepo repositories.ChapterRepository
	storyRepo   repositories.StoryRepository
	logger      logger.Logger
}

// NewChapterHandler creates a new ChapterHandler
func NewChapterHandler(
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *ChapterHandler {
	return &ChapterHandler{
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// Create handles POST /api/v1/chapters
func (h *ChapterHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		StoryID string `json:"story_id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Status  string `json:"status,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "story_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Validate story exists
	_, err = h.storyRepo.GetByID(r.Context(), tenantID, storyID)
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	if req.Number < 1 {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "number",
			Message: "must be greater than 0",
		}, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "title",
			Message: "title is required",
		}, http.StatusBadRequest)
		return
	}

	chapter, err := story.NewChapter(tenantID, storyID, req.Number, req.Title)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "chapter",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Set status if provided
	if req.Status != "" {
		if err := chapter.UpdateStatus(story.ChapterStatus(req.Status)); err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "status",
				Message: "invalid status",
			}, http.StatusBadRequest)
			return
		}
	}

	if err := h.chapterRepo.Create(r.Context(), chapter); err != nil {
		h.logger.Error("failed to create chapter", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": chapter,
	})
}

// Get handles GET /api/v1/chapters/{id}
func (h *ChapterHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	chapter, err := h.chapterRepo.GetByID(r.Context(), tenantID, chapterID)
	if err != nil {
		if err.Error() == "chapter not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": chapter,
	})
}

// Update handles PUT /api/v1/chapters/{id}
func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing chapter
	chapter, err := h.chapterRepo.GetByID(r.Context(), tenantID, chapterID)
	if err != nil {
		if err.Error() == "chapter not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		Number *int    `json:"number,omitempty"`
		Title  *string `json:"title,omitempty"`
		Status *string `json:"status,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.Number != nil {
		if *req.Number < 1 {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "number",
				Message: "must be greater than 0",
			}, http.StatusBadRequest)
			return
		}
		chapter.Number = *req.Number
	}

	if req.Title != nil {
		chapter.UpdateTitle(*req.Title)
	}

	if req.Status != nil {
		if err := chapter.UpdateStatus(story.ChapterStatus(*req.Status)); err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "status",
				Message: "invalid status",
			}, http.StatusBadRequest)
			return
		}
	}

	if err := h.chapterRepo.Update(r.Context(), chapter); err != nil {
		h.logger.Error("failed to update chapter", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": chapter,
	})
}

// List handles GET /api/v1/stories/{id}/chapters
func (h *ChapterHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	storyIDStr := r.PathValue("id")

	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	chapters, err := h.chapterRepo.ListByStory(r.Context(), tenantID, storyID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapters": chapters,
		"total":    len(chapters),
	})
}

// Delete handles DELETE /api/v1/chapters/{id}
func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Check if chapter exists
	_, err = h.chapterRepo.GetByID(r.Context(), tenantID, chapterID)
	if err != nil {
		if err.Error() == "chapter not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err := h.chapterRepo.Delete(r.Context(), tenantID, chapterID); err != nil {
		h.logger.Error("failed to delete chapter", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

