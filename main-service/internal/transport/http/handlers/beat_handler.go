package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// BeatHandler handles HTTP requests for beats
type BeatHandler struct {
	createBeatUseCase *beatapp.CreateBeatUseCase
	getBeatUseCase    *beatapp.GetBeatUseCase
	updateBeatUseCase *beatapp.UpdateBeatUseCase
	deleteBeatUseCase *beatapp.DeleteBeatUseCase
	listBeatsUseCase  *beatapp.ListBeatsUseCase
	moveBeatUseCase   *beatapp.MoveBeatUseCase
	logger            logger.Logger
}

// NewBeatHandler creates a new BeatHandler
func NewBeatHandler(
	createBeatUseCase *beatapp.CreateBeatUseCase,
	getBeatUseCase *beatapp.GetBeatUseCase,
	updateBeatUseCase *beatapp.UpdateBeatUseCase,
	deleteBeatUseCase *beatapp.DeleteBeatUseCase,
	listBeatsUseCase *beatapp.ListBeatsUseCase,
	moveBeatUseCase *beatapp.MoveBeatUseCase,
	logger logger.Logger,
) *BeatHandler {
	return &BeatHandler{
		createBeatUseCase: createBeatUseCase,
		getBeatUseCase:    getBeatUseCase,
		updateBeatUseCase: updateBeatUseCase,
		deleteBeatUseCase: deleteBeatUseCase,
		listBeatsUseCase:  listBeatsUseCase,
		moveBeatUseCase:   moveBeatUseCase,
		logger:            logger,
	}
}

// Create handles POST /api/v1/beats
func (h *BeatHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		SceneID  string `json:"scene_id"`
		OrderNum int    `json:"order_num"`
		Type     string `json:"type"`
		Intent   string `json:"intent,omitempty"`
		Outcome  string `json:"outcome,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	sceneID, err := uuid.Parse(req.SceneID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "scene_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createBeatUseCase.Execute(r.Context(), beatapp.CreateBeatInput{
		TenantID: tenantID,
		SceneID:  sceneID,
		OrderNum: req.OrderNum,
		Type:     story.BeatType(req.Type),
		Intent:   req.Intent,
		Outcome:  req.Outcome,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": output.Beat,
	})
}

// Get handles GET /api/v1/beats/{id}
func (h *BeatHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getBeatUseCase.Execute(r.Context(), beatapp.GetBeatInput{
		TenantID: tenantID,
		ID:       beatID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": output.Beat,
	})
}

// Update handles PUT /api/v1/beats/{id}
func (h *BeatHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		OrderNum *int    `json:"order_num,omitempty"`
		Type     *string `json:"type,omitempty"`
		Intent   *string `json:"intent,omitempty"`
		Outcome  *string `json:"outcome,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	input := beatapp.UpdateBeatInput{
		TenantID: tenantID,
		ID:       beatID,
		OrderNum: req.OrderNum,
		Intent:   req.Intent,
		Outcome:  req.Outcome,
	}
	if req.Type != nil {
		beatType := story.BeatType(*req.Type)
		input.Type = &beatType
	}

	output, err := h.updateBeatUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": output.Beat,
	})
}

// List handles GET /api/v1/scenes/{id}/beats
func (h *BeatHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sceneIDStr := r.PathValue("id")

	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listBeatsUseCase.Execute(r.Context(), beatapp.ListBeatsInput{
		TenantID: tenantID,
		SceneID:  sceneID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beats": output.Beats,
		"total": output.Total,
	})
}

// Delete handles DELETE /api/v1/beats/{id}
func (h *BeatHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteBeatUseCase.Execute(r.Context(), beatapp.DeleteBeatInput{
		TenantID: tenantID,
		ID:       beatID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByStory handles GET /api/v1/stories/{id}/beats
func (h *BeatHandler) ListByStory(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listBeatsUseCase.Execute(r.Context(), beatapp.ListBeatsInput{
		TenantID: tenantID,
		StoryID:  &storyID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beats": output.Beats,
		"total": output.Total,
	})
}

// Move handles PUT /api/v1/beats/{id}/move
func (h *BeatHandler) Move(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		SceneID string `json:"scene_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	newSceneID, err := uuid.Parse(req.SceneID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "scene_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.moveBeatUseCase.Execute(r.Context(), beatapp.MoveBeatInput{
		TenantID:   tenantID,
		BeatID:     beatID,
		NewSceneID: newSceneID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": output.Beat,
	})
}

