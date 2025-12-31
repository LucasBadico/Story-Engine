package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// StoryHandler handles HTTP requests for stories
type StoryHandler struct {
	createStoryUseCase *story.CreateStoryUseCase
	cloneStoryUseCase  *story.CloneStoryUseCase
	storyRepo          repositories.StoryRepository
	logger             logger.Logger
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(
	createStoryUseCase *story.CreateStoryUseCase,
	cloneStoryUseCase *story.CloneStoryUseCase,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *StoryHandler {
	return &StoryHandler{
		createStoryUseCase: createStoryUseCase,
		cloneStoryUseCase:  cloneStoryUseCase,
		storyRepo:          storyRepo,
		logger:             logger,
	}
}

// Create handles POST /api/v1/stories
func (h *StoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
		Title    string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "tenant_id",
			Message: "invalid UUID format",
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

	output, err := h.createStoryUseCase.Execute(r.Context(), story.CreateStoryInput{
		TenantID: tenantID,
		Title:    req.Title,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": output.Story,
	})
}

// Get handles GET /api/v1/stories/{id}
func (h *StoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	storyID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	story, err := h.storyRepo.GetByID(r.Context(), storyID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": story,
	})
}

// List handles GET /api/v1/stories?tenant_id=xxx&limit=20&offset=0
func (h *StoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if tenantIDStr == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "tenant_id",
			Message: "required",
		}, http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "tenant_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	stories, err := h.storyRepo.ListByTenant(r.Context(), tenantID, limit, offset)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stories": stories,
		"total":   len(stories),
	})
}

// Clone handles POST /api/v1/stories/{id}/clone
func (h *StoryHandler) Clone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sourceStoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.cloneStoryUseCase.Execute(r.Context(), story.CloneStoryInput{
		SourceStoryID: sourceStoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Fetch the new story to get full details
	newStory, err := h.storyRepo.GetByID(r.Context(), output.NewStoryID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story":          newStory,
		"version_number": newStory.VersionNumber,
	})
}

