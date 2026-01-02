package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// SceneHandler handles HTTP requests for scenes
type SceneHandler struct {
	createSceneUseCase    *sceneapp.CreateSceneUseCase
	getSceneUseCase       *sceneapp.GetSceneUseCase
	updateSceneUseCase    *sceneapp.UpdateSceneUseCase
	deleteSceneUseCase    *sceneapp.DeleteSceneUseCase
	listScenesUseCase     *sceneapp.ListScenesUseCase
	moveSceneUseCase      *sceneapp.MoveSceneUseCase
	addReferenceUC        *sceneapp.AddSceneReferenceUseCase
	removeReferenceUC     *sceneapp.RemoveSceneReferenceUseCase
	getReferencesUC       *sceneapp.GetSceneReferencesUseCase
	logger                 logger.Logger
}

// NewSceneHandler creates a new SceneHandler
func NewSceneHandler(
	createSceneUseCase *sceneapp.CreateSceneUseCase,
	getSceneUseCase *sceneapp.GetSceneUseCase,
	updateSceneUseCase *sceneapp.UpdateSceneUseCase,
	deleteSceneUseCase *sceneapp.DeleteSceneUseCase,
	listScenesUseCase *sceneapp.ListScenesUseCase,
	moveSceneUseCase *sceneapp.MoveSceneUseCase,
	addReferenceUC *sceneapp.AddSceneReferenceUseCase,
	removeReferenceUC *sceneapp.RemoveSceneReferenceUseCase,
	getReferencesUC *sceneapp.GetSceneReferencesUseCase,
	logger logger.Logger,
) *SceneHandler {
	return &SceneHandler{
		createSceneUseCase: createSceneUseCase,
		getSceneUseCase:    getSceneUseCase,
		updateSceneUseCase: updateSceneUseCase,
		deleteSceneUseCase: deleteSceneUseCase,
		listScenesUseCase:  listScenesUseCase,
		moveSceneUseCase:   moveSceneUseCase,
		addReferenceUC:     addReferenceUC,
		removeReferenceUC:  removeReferenceUC,
		getReferencesUC:    getReferencesUC,
		logger:             logger,
	}
}

// Create handles POST /api/v1/scenes
func (h *SceneHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		StoryID        string  `json:"story_id"`
		ChapterID      *string `json:"chapter_id,omitempty"`
		OrderNum       int     `json:"order_num"`
		POVCharacterID *string `json:"pov_character_id,omitempty"`
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
		chapterID = &parsedChapterID
	}

	var povCharacterID *uuid.UUID
	if req.POVCharacterID != nil && *req.POVCharacterID != "" {
		charID, err := uuid.Parse(*req.POVCharacterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "pov_character_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		povCharacterID = &charID
	}

	output, err := h.createSceneUseCase.Execute(r.Context(), sceneapp.CreateSceneInput{
		TenantID:       tenantID,
		StoryID:        storyID,
		ChapterID:      chapterID,
		OrderNum:       req.OrderNum,
		POVCharacterID: povCharacterID,
		TimeRef:        req.TimeRef,
		Goal:           req.Goal,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": output.Scene,
	})
}

// Get handles GET /api/v1/scenes/{id}
func (h *SceneHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getSceneUseCase.Execute(r.Context(), sceneapp.GetSceneInput{
		TenantID: tenantID,
		ID:       sceneID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": output.Scene,
	})
}

// Update handles PUT /api/v1/scenes/{id}
func (h *SceneHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		OrderNum       *int    `json:"order_num,omitempty"`
		POVCharacterID *string `json:"pov_character_id,omitempty"`
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

	input := sceneapp.UpdateSceneInput{
		TenantID: tenantID,
		ID:       sceneID,
		OrderNum: req.OrderNum,
		TimeRef:  req.TimeRef,
		Goal:     req.Goal,
	}

	if req.POVCharacterID != nil {
		if *req.POVCharacterID == "" {
			input.POVCharacterID = nil
		} else {
			charID, err := uuid.Parse(*req.POVCharacterID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "pov_character_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			input.POVCharacterID = &charID
		}
	}

	output, err := h.updateSceneUseCase.Execute(r.Context(), input)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": output.Scene,
	})
}

// List handles GET /api/v1/chapters/{id}/scenes
func (h *SceneHandler) List(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listScenesUseCase.Execute(r.Context(), sceneapp.ListScenesInput{
		TenantID:  tenantID,
		ChapterID: &chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": output.Scenes,
		"total":  output.Total,
	})
}

// Delete handles DELETE /api/v1/scenes/{id}
func (h *SceneHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteSceneUseCase.Execute(r.Context(), sceneapp.DeleteSceneInput{
		TenantID: tenantID,
		ID:       sceneID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReferences handles GET /api/v1/scenes/{id}/references
func (h *SceneHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUC.Execute(r.Context(), sceneapp.GetSceneReferencesInput{
		TenantID: tenantID,
		SceneID:  sceneID,
	})
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
		"references": output.References,
		"total":      len(output.References),
	})
}

// AddReference handles POST /api/v1/scenes/{id}/references
func (h *SceneHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	entityType := story.SceneReferenceEntityType(req.EntityType)
	if !isValidSceneReferenceEntityType(entityType) {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type, must be one of: character, location, artifact",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addReferenceUC.Execute(r.Context(), sceneapp.AddSceneReferenceInput{
		TenantID:   tenantID,
		SceneID:    sceneID,
		EntityType: entityType,
		EntityID:   entityID,
	}); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveReference handles DELETE /api/v1/scenes/{id}/references/{entity_type}/{entity_id}
func (h *SceneHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	entityTypeStr := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityType := story.SceneReferenceEntityType(entityTypeStr)
	if !isValidSceneReferenceEntityType(entityType) {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type, must be one of: character, location, artifact",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeReferenceUC.Execute(r.Context(), sceneapp.RemoveSceneReferenceInput{
		TenantID:   tenantID,
		SceneID:    sceneID,
		EntityType: entityType,
		EntityID:   entityID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidSceneReferenceEntityType(entityType story.SceneReferenceEntityType) bool {
	return entityType == story.SceneReferenceEntityTypeCharacter ||
		entityType == story.SceneReferenceEntityTypeLocation ||
		entityType == story.SceneReferenceEntityTypeArtifact
}

// ListByStory handles GET /api/v1/stories/{id}/scenes
func (h *SceneHandler) ListByStory(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listScenesUseCase.Execute(r.Context(), sceneapp.ListScenesInput{
		TenantID: tenantID,
		StoryID:  storyID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": output.Scenes,
		"total":  output.Total,
	})
}

// Move handles PUT /api/v1/scenes/{id}/move
func (h *SceneHandler) Move(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	id := r.PathValue("id")

	sceneID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
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
		newChapterID = &parsedChapterID
	}

	output, err := h.moveSceneUseCase.Execute(r.Context(), sceneapp.MoveSceneInput{
		TenantID:     tenantID,
		SceneID:      sceneID,
		NewChapterID: newChapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scene": output.Scene,
	})
}

