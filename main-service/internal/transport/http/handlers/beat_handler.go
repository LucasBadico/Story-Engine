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

// BeatHandler handles HTTP requests for beats
type BeatHandler struct {
	beatRepo  repositories.BeatRepository
	sceneRepo repositories.SceneRepository
	logger    logger.Logger
}

// NewBeatHandler creates a new BeatHandler
func NewBeatHandler(
	beatRepo repositories.BeatRepository,
	sceneRepo repositories.SceneRepository,
	logger logger.Logger,
) *BeatHandler {
	return &BeatHandler{
		beatRepo:  beatRepo,
		sceneRepo: sceneRepo,
		logger:    logger,
	}
}

// Create handles POST /api/v1/beats
func (h *BeatHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// Validate scene exists
	_, err = h.sceneRepo.GetByID(r.Context(), sceneID)
	if err != nil {
		if err.Error() == "scene not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       req.SceneID,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if req.OrderNum < 1 {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}, http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "type",
			Message: "type is required",
		}, http.StatusBadRequest)
		return
	}

	beat, err := story.NewBeat(sceneID, req.OrderNum, story.BeatType(req.Type))
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "beat",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	if req.Intent != "" {
		beat.UpdateIntent(req.Intent)
	}
	if req.Outcome != "" {
		beat.UpdateOutcome(req.Outcome)
	}

	if err := h.beatRepo.Create(r.Context(), beat); err != nil {
		h.logger.Error("failed to create beat", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": beat,
	})
}

// Get handles GET /api/v1/beats/{id}
func (h *BeatHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	beat, err := h.beatRepo.GetByID(r.Context(), beatID)
	if err != nil {
		if err.Error() == "beat not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "beat",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": beat,
	})
}

// Update handles PUT /api/v1/beats/{id}
func (h *BeatHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing beat
	beat, err := h.beatRepo.GetByID(r.Context(), beatID)
	if err != nil {
		if err.Error() == "beat not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "beat",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
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

	// Update fields if provided
	if req.OrderNum != nil {
		if *req.OrderNum < 1 {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "order_num",
				Message: "must be greater than 0",
			}, http.StatusBadRequest)
			return
		}
		beat.OrderNum = *req.OrderNum
	}

	if req.Type != nil {
		beat.Type = story.BeatType(*req.Type)
		if !isValidBeatType(beat.Type) {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "type",
				Message: "invalid beat type",
			}, http.StatusBadRequest)
			return
		}
	}

	if req.Intent != nil {
		beat.UpdateIntent(*req.Intent)
	}

	if req.Outcome != nil {
		beat.UpdateOutcome(*req.Outcome)
	}

	if err := h.beatRepo.Update(r.Context(), beat); err != nil {
		h.logger.Error("failed to update beat", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beat": beat,
	})
}

// List handles GET /api/v1/scenes/{id}/beats
func (h *BeatHandler) List(w http.ResponseWriter, r *http.Request) {
	sceneIDStr := r.PathValue("id")

	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	beats, err := h.beatRepo.ListByScene(r.Context(), sceneID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"beats": beats,
		"total": len(beats),
	})
}

// Delete handles DELETE /api/v1/beats/{id}
func (h *BeatHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	beatID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Check if beat exists
	_, err = h.beatRepo.GetByID(r.Context(), beatID)
	if err != nil {
		if err.Error() == "beat not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "beat",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err := h.beatRepo.Delete(r.Context(), beatID); err != nil {
		h.logger.Error("failed to delete beat", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isValidBeatType checks if a beat type is valid
func isValidBeatType(bt story.BeatType) bool {
	return bt == story.BeatTypeSetup ||
		bt == story.BeatTypeTurn ||
		bt == story.BeatTypeReveal ||
		bt == story.BeatTypeConflict ||
		bt == story.BeatTypeClimax ||
		bt == story.BeatTypeResolution ||
		bt == story.BeatTypeHook ||
		bt == story.BeatTypeTransition
}

