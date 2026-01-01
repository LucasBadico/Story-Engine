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

// SceneHandler handles HTTP requests for scenes
type SceneHandler struct {
	sceneRepo    repositories.SceneRepository
	chapterRepo  repositories.ChapterRepository
	storyRepo    repositories.StoryRepository
	logger       logger.Logger
}

// NewSceneHandler creates a new SceneHandler
func NewSceneHandler(
	sceneRepo repositories.SceneRepository,
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *SceneHandler {
	return &SceneHandler{
		sceneRepo:   sceneRepo,
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// Create handles POST /api/v1/scenes
func (h *SceneHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StoryID        string  `json:"story_id"`
		ChapterID      *string `json:"chapter_id,omitempty"`
		OrderNum       int     `json:"order_num"`
		POVCharacterID *string `json:"pov_character_id,omitempty"`
		LocationID     *string `json:"location_id,omitempty"`
		TimeRef        string  `json:"time_ref,omitempty"`
		Goal           string  `json:"goal,omitempty"`
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
	_, err = h.storyRepo.GetByID(r.Context(), storyID)
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	var chapterID *uuid.UUID
	if req.ChapterID != nil && *req.ChapterID != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "chapter_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}

		// Validate chapter exists
		_, err = h.chapterRepo.GetByID(r.Context(), parsedChapterID)
		if err != nil {
			if err.Error() == "chapter not found" {
				WriteError(w, &platformerrors.NotFoundError{
					Resource: "chapter",
					ID:       *req.ChapterID,
				}, http.StatusNotFound)
			} else {
				WriteError(w, err, http.StatusInternalServerError)
			}
			return
		}
		chapterID = &parsedChapterID
	}

	if req.OrderNum < 1 {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "order_num",
			Message: "must be greater than 0",
		}, http.StatusBadRequest)
		return
	}

	scene, err := story.NewScene(storyID, chapterID, req.OrderNum)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "scene",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	if req.Goal != "" {
		scene.UpdateGoal(req.Goal)
	}
	if req.TimeRef != "" {
		scene.TimeRef = req.TimeRef
	}

	if req.POVCharacterID != nil && *req.POVCharacterID != "" {
		charID, err := uuid.Parse(*req.POVCharacterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "pov_character_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		scene.UpdatePOV(&charID)
	}

	if req.LocationID != nil && *req.LocationID != "" {
		locID, err := uuid.Parse(*req.LocationID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "location_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		scene.UpdateLocation(&locID)
	}

	if err := h.sceneRepo.Create(r.Context(), scene); err != nil {
		h.logger.Error("failed to create scene", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": scene,
	})
}

// Get handles GET /api/v1/scenes/{id}
func (h *SceneHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	scene, err := h.sceneRepo.GetByID(r.Context(), sceneID)
	if err != nil {
		if err.Error() == "scene not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": scene,
	})
}

// Update handles PUT /api/v1/scenes/{id}
func (h *SceneHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing scene
	scene, err := h.sceneRepo.GetByID(r.Context(), sceneID)
	if err != nil {
		if err.Error() == "scene not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		OrderNum       *int    `json:"order_num,omitempty"`
		POVCharacterID *string `json:"pov_character_id,omitempty"`
		LocationID     *string `json:"location_id,omitempty"`
		TimeRef        *string `json:"time_ref,omitempty"`
		Goal           *string `json:"goal,omitempty"`
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
		scene.OrderNum = *req.OrderNum
	}

	if req.Goal != nil {
		scene.UpdateGoal(*req.Goal)
	}

	if req.TimeRef != nil {
		scene.TimeRef = *req.TimeRef
	}

	if req.POVCharacterID != nil {
		if *req.POVCharacterID == "" {
			scene.UpdatePOV(nil)
		} else {
			charID, err := uuid.Parse(*req.POVCharacterID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "pov_character_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			scene.UpdatePOV(&charID)
		}
	}

	if req.LocationID != nil {
		if *req.LocationID == "" {
			scene.UpdateLocation(nil)
		} else {
			locID, err := uuid.Parse(*req.LocationID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "location_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			scene.UpdateLocation(&locID)
		}
	}

	if err := h.sceneRepo.Update(r.Context(), scene); err != nil {
		h.logger.Error("failed to update scene", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": scene,
	})
}

// List handles GET /api/v1/chapters/{id}/scenes
func (h *SceneHandler) List(w http.ResponseWriter, r *http.Request) {
	chapterIDStr := r.PathValue("id")

	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	scenes, err := h.sceneRepo.ListByChapter(r.Context(), chapterID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": scenes,
		"total":  len(scenes),
	})
}

// Delete handles DELETE /api/v1/scenes/{id}
func (h *SceneHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Check if scene exists
	_, err = h.sceneRepo.GetByID(r.Context(), sceneID)
	if err != nil {
		if err.Error() == "scene not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	if err := h.sceneRepo.Delete(r.Context(), sceneID); err != nil {
		h.logger.Error("failed to delete scene", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByStory handles GET /api/v1/stories/{id}/scenes
func (h *SceneHandler) ListByStory(w http.ResponseWriter, r *http.Request) {
	storyIDStr := r.PathValue("id")

	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Validate story exists
	_, err = h.storyRepo.GetByID(r.Context(), storyID)
	if err != nil {
		if err.Error() == "story not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "story",
				ID:       storyIDStr,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	scenes, err := h.sceneRepo.ListByStory(r.Context(), storyID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": scenes,
		"total":  len(scenes),
	})
}

// Move handles PUT /api/v1/scenes/{id}/move
func (h *SceneHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Get existing scene
	scene, err := h.sceneRepo.GetByID(r.Context(), sceneID)
	if err != nil {
		if err.Error() == "scene not found" {
			WriteError(w, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       id,
			}, http.StatusNotFound)
		} else {
			WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		ChapterID *string `json:"chapter_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var newChapterID *uuid.UUID
	if req.ChapterID != nil && *req.ChapterID != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "chapter_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}

		// Validate chapter exists
		_, err = h.chapterRepo.GetByID(r.Context(), parsedChapterID)
		if err != nil {
			if err.Error() == "chapter not found" {
				WriteError(w, &platformerrors.NotFoundError{
					Resource: "chapter",
					ID:       *req.ChapterID,
				}, http.StatusNotFound)
			} else {
				WriteError(w, err, http.StatusInternalServerError)
			}
			return
		}
		newChapterID = &parsedChapterID
	}

	scene.UpdateChapter(newChapterID)

	if err := h.sceneRepo.Update(r.Context(), scene); err != nil {
		h.logger.Error("failed to move scene", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": scene,
	})
}

