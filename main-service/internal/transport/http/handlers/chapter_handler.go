package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ChapterHandler handles HTTP requests for chapters
type ChapterHandler struct {
	createChapterUseCase *chapterapp.CreateChapterUseCase
	getChapterUseCase    *chapterapp.GetChapterUseCase
	updateChapterUseCase *chapterapp.UpdateChapterUseCase
	deleteChapterUseCase *chapterapp.DeleteChapterUseCase
	listChaptersUseCase  *chapterapp.ListChaptersUseCase
	logger               logger.Logger
}

// NewChapterHandler creates a new ChapterHandler
func NewChapterHandler(
	createChapterUseCase *chapterapp.CreateChapterUseCase,
	getChapterUseCase *chapterapp.GetChapterUseCase,
	updateChapterUseCase *chapterapp.UpdateChapterUseCase,
	deleteChapterUseCase *chapterapp.DeleteChapterUseCase,
	listChaptersUseCase *chapterapp.ListChaptersUseCase,
	logger logger.Logger,
) *ChapterHandler {
	return &ChapterHandler{
		createChapterUseCase: createChapterUseCase,
		getChapterUseCase:    getChapterUseCase,
		updateChapterUseCase: updateChapterUseCase,
		deleteChapterUseCase: deleteChapterUseCase,
		listChaptersUseCase:  listChaptersUseCase,
		logger:               logger,
	}
}

// Create handles POST /api/v1/chapters
func (h *ChapterHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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

	input := chapterapp.CreateChapterInput{
		TenantID: tenantID,
		StoryID:  storyID,
		Number:   req.Number,
		Title:    req.Title,
	}
	if req.Status != "" {
		status := story.ChapterStatus(req.Status)
		input.Status = &status
	}

	output, err := h.createChapterUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": output.Chapter,
	})
}

// Get handles GET /api/v1/chapters/{id}
func (h *ChapterHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getChapterUseCase.Execute(r.Context(), chapterapp.GetChapterInput{
		TenantID: tenantID,
		ID:       chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": output.Chapter,
	})
}

// Update handles PUT /api/v1/chapters/{id}
func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
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

	input := chapterapp.UpdateChapterInput{
		TenantID: tenantID,
		ID:       chapterID,
		Number:   req.Number,
		Title:    req.Title,
	}
	if req.Status != nil {
		status := story.ChapterStatus(*req.Status)
		input.Status = &status
	}

	output, err := h.updateChapterUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter": output.Chapter,
	})
}

// List handles GET /api/v1/stories/{id}/chapters
func (h *ChapterHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	storyIDStr := r.PathValue("id")

	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listChaptersUseCase.Execute(r.Context(), chapterapp.ListChaptersInput{
		TenantID: tenantID,
		StoryID:  storyID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapters": output.Chapters,
		"total":    output.Total,
	})
}

// Delete handles DELETE /api/v1/chapters/{id}
func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	chapterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteChapterUseCase.Execute(r.Context(), chapterapp.DeleteChapterInput{
		TenantID: tenantID,
		ID:       chapterID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

