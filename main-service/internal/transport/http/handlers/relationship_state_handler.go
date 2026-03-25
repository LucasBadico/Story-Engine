package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	relationshipstateapp "github.com/story-engine/main-service/internal/application/world/relationship_state"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// RelationshipStateHandler handles HTTP requests for relationship states
type RelationshipStateHandler struct {
	createRelationshipStateUseCase    *relationshipstateapp.CreateRelationshipStateUseCase
	getRelationshipStateUseCase       *relationshipstateapp.GetRelationshipStateUseCase
	getLatestRelationshipStateUseCase *relationshipstateapp.GetLatestRelationshipStateUseCase
	listRelationshipStatesUseCase     *relationshipstateapp.ListRelationshipStatesUseCase
	deleteRelationshipStateUseCase    *relationshipstateapp.DeleteRelationshipStateUseCase
	logger                            logger.Logger
}

// NewRelationshipStateHandler creates a new RelationshipStateHandler
func NewRelationshipStateHandler(
	createRelationshipStateUseCase *relationshipstateapp.CreateRelationshipStateUseCase,
	getRelationshipStateUseCase *relationshipstateapp.GetRelationshipStateUseCase,
	getLatestRelationshipStateUseCase *relationshipstateapp.GetLatestRelationshipStateUseCase,
	listRelationshipStatesUseCase *relationshipstateapp.ListRelationshipStatesUseCase,
	deleteRelationshipStateUseCase *relationshipstateapp.DeleteRelationshipStateUseCase,
	logger logger.Logger,
) *RelationshipStateHandler {
	return &RelationshipStateHandler{
		createRelationshipStateUseCase:    createRelationshipStateUseCase,
		getRelationshipStateUseCase:       getRelationshipStateUseCase,
		getLatestRelationshipStateUseCase: getLatestRelationshipStateUseCase,
		listRelationshipStatesUseCase:     listRelationshipStatesUseCase,
		deleteRelationshipStateUseCase:    deleteRelationshipStateUseCase,
		logger:                            logger,
	}
}

// Create handles POST /api/v1/relationship-states
func (h *RelationshipStateHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		ChapterID          string  `json:"chapter_id"`
		ChapterNumber      int     `json:"chapter_number"`
		TensionDescription string  `json:"tension_description"`
		TimelineDay        *int    `json:"timeline_day"`
		TimelineConfidence *string `json:"timeline_confidence"`
		Entities           []struct {
			Role       string `json:"role"`
			EntityType string `json:"entity_type"`
			EntityID   string `json:"entity_id"`
		} `json:"entities"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	chapterID, err := uuid.Parse(req.ChapterID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "chapter_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var entities []relationshipstateapp.EntityLinkInput
	for _, e := range req.Entities {
		eid, err := uuid.Parse(e.EntityID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "entities.entity_id",
				Message: "invalid UUID format: " + e.EntityID,
			}, http.StatusBadRequest)
			return
		}
		entities = append(entities, relationshipstateapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   eid,
		})
	}

	output, err := h.createRelationshipStateUseCase.Execute(r.Context(), relationshipstateapp.CreateRelationshipStateInput{
		TenantID:           tenantID,
		ChapterID:          chapterID,
		ChapterNumber:      req.ChapterNumber,
		TensionDescription: req.TensionDescription,
		TimelineDay:        req.TimelineDay,
		TimelineConfidence: req.TimelineConfidence,
		Entities:           entities,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship_state": output.RelationshipState,
	})
}

// Get handles GET /api/v1/relationship-states/{id}
func (h *RelationshipStateHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	rsID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getRelationshipStateUseCase.Execute(r.Context(), relationshipstateapp.GetRelationshipStateInput{
		TenantID: tenantID,
		ID:       rsID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship_state": output.RelationshipState,
	})
}

// GetLatest handles GET /api/v1/relationship-states/latest?source_entity_type=X&source_entity_id=X&target_entity_type=Y&target_entity_id=Y&as_of_chapter=N
func (h *RelationshipStateHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sourceEntityType := r.URL.Query().Get("source_entity_type")
	sourceEntityIDStr := r.URL.Query().Get("source_entity_id")
	sourceEntityID, err := uuid.Parse(sourceEntityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "source_entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	targetEntityType := r.URL.Query().Get("target_entity_type")
	targetEntityIDStr := r.URL.Query().Get("target_entity_id")
	targetEntityID, err := uuid.Parse(targetEntityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "target_entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	asOfChapterStr := r.URL.Query().Get("as_of_chapter")
	asOfChapter := 0
	if asOfChapterStr != "" {
		if n, err := strconv.Atoi(asOfChapterStr); err == nil && n > 0 {
			asOfChapter = n
		}
	}

	output, err := h.getLatestRelationshipStateUseCase.Execute(r.Context(), relationshipstateapp.GetLatestRelationshipStateInput{
		TenantID:         tenantID,
		SourceEntityType: sourceEntityType,
		SourceEntityID:   sourceEntityID,
		TargetEntityType: targetEntityType,
		TargetEntityID:   targetEntityID,
		AsOfChapter:      asOfChapter,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship_state": output.RelationshipState,
	})
}

// ListByChapter handles GET /api/v1/chapters/{chapter_id}/relationship-states
func (h *RelationshipStateHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	chapterIDStr := r.PathValue("chapter_id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "chapter_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listRelationshipStatesUseCase.Execute(r.Context(), relationshipstateapp.ListRelationshipStatesInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship_states": output.RelationshipStates,
	})
}

// Delete handles DELETE /api/v1/relationship-states/{id}
func (h *RelationshipStateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	rsID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteRelationshipStateUseCase.Execute(r.Context(), relationshipstateapp.DeleteRelationshipStateInput{
		TenantID: tenantID,
		ID:       rsID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
