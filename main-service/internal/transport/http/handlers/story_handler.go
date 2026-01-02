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
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// StoryHandler handles HTTP requests for stories
type StoryHandler struct {
	createStoryUseCase *storyapp.CreateStoryUseCase
	getStoryUseCase    *storyapp.GetStoryUseCase
	updateStoryUseCase *storyapp.UpdateStoryUseCase
	listStoriesUseCase *storyapp.ListStoriesUseCase
	cloneStoryUseCase  *storyapp.CloneStoryUseCase
	logger             logger.Logger
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(
	createStoryUseCase *storyapp.CreateStoryUseCase,
	getStoryUseCase *storyapp.GetStoryUseCase,
	updateStoryUseCase *storyapp.UpdateStoryUseCase,
	listStoriesUseCase *storyapp.ListStoriesUseCase,
	cloneStoryUseCase *storyapp.CloneStoryUseCase,
	logger logger.Logger,
) *StoryHandler {
	return &StoryHandler{
		createStoryUseCase: createStoryUseCase,
		getStoryUseCase:    getStoryUseCase,
		updateStoryUseCase: updateStoryUseCase,
		listStoriesUseCase: listStoriesUseCase,
		cloneStoryUseCase:  cloneStoryUseCase,
		logger:             logger,
	}
}

// Create handles POST /api/v1/stories
func (h *StoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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

	id := r.PathValue("id")
	storyID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getStoryUseCase.Execute(r.Context(), storyapp.GetStoryInput{
		TenantID: tenantID,
		ID:       storyID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": output.Story,
	})
}

// List handles GET /api/v1/stories?limit=20&offset=0
func (h *StoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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

	output, err := h.listStoriesUseCase.Execute(r.Context(), storyapp.ListStoriesInput{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stories": output.Stories,
		"total":   output.Total,
	})
}

// Update handles PUT /api/v1/stories/{id}
func (h *StoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	storyID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
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

	input := storyapp.UpdateStoryInput{
		TenantID: tenantID,
		ID:       storyID,
		Title:    req.Title,
	}
	if req.Status != nil {
		status := storycore.StoryStatus(*req.Status)
		input.Status = &status
	}

	output, err := h.updateStoryUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": output.Story,
	})
}

// Clone handles POST /api/v1/stories/{id}/clone
func (h *StoryHandler) Clone(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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
		TenantID:      tenantID,
		SourceStoryID: sourceStoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Fetch the new story to get full details
	getOutput, err := h.getStoryUseCase.Execute(r.Context(), storyapp.GetStoryInput{
		TenantID: tenantID,
		ID:       output.NewStoryID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story":          getOutput.Story,
		"version_number": getOutput.Story.VersionNumber,
	})
}

