package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	storyapp "github.com/story-engine/main-service/internal/application/story"
	storycore "github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// StoryHandler handles HTTP requests for stories
type StoryHandler struct {
	createStoryUseCase *storyapp.CreateStoryUseCase
	cloneStoryUseCase  *storyapp.CloneStoryUseCase
	storyRepo          repositories.StoryRepository
	logger             logger.Logger
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(
	createStoryUseCase *storyapp.CreateStoryUseCase,
	cloneStoryUseCase *storyapp.CloneStoryUseCase,
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
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		Title   string     `json:"title"`
		WorldID *uuid.UUID `json:"world_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
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

	output, err := h.createStoryUseCase.Execute(r.Context(), storyapp.CreateStoryInput{
		TenantID: tenantID,
		Title:    req.Title,
		WorldID:  req.WorldID,
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
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	storyID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	story, err := h.storyRepo.GetByID(r.Context(), tenantID, storyID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": story,
	})
}

// List handles GET /api/v1/stories?limit=20&offset=0
func (h *StoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

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

// Update handles PUT /api/v1/stories/{id}
func (h *StoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	storyID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing story
	s, err := h.storyRepo.GetByID(r.Context(), tenantID, storyID)
	if err != nil {
		if err.Error() == "story not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "story",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	var req struct {
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
	if req.Title != nil {
		if err := s.UpdateTitle(*req.Title); err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "title",
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}
	}

	if req.Status != nil {
		if err := s.UpdateStatus(storycore.StoryStatus(*req.Status)); err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "status",
				Message: "invalid status",
			}, http.StatusBadRequest)
			return
		}
	}

	if err := h.storyRepo.Update(r.Context(), s); err != nil {
		h.logger.Error("failed to update story", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": s,
	})
}

// Clone handles POST /api/v1/stories/{id}/clone
func (h *StoryHandler) Clone(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	sourceStoryID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.cloneStoryUseCase.Execute(r.Context(), storyapp.CloneStoryInput{
		SourceStoryID: sourceStoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Fetch the new story to get full details
	newStory, err := h.storyRepo.GetByID(r.Context(), tenantID, output.NewStoryID)
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

