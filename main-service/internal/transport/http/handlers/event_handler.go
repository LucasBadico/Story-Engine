package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	rpgeventapp "github.com/story-engine/main-service/internal/application/rpg/event"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// EventHandler handles HTTP requests for events
type EventHandler struct {
	createEventUseCase      *eventapp.CreateEventUseCase
	getEventUseCase         *eventapp.GetEventUseCase
	listEventsUseCase       *eventapp.ListEventsUseCase
	updateEventUseCase      *eventapp.UpdateEventUseCase
	deleteEventUseCase      *eventapp.DeleteEventUseCase
	addReferenceUseCase     *eventapp.AddReferenceUseCase
	removeReferenceUseCase  *eventapp.RemoveReferenceUseCase
	getReferencesUseCase    *eventapp.GetReferencesUseCase
	updateReferenceUseCase  *eventapp.UpdateReferenceUseCase
	getChildrenUseCase      *eventapp.GetChildrenUseCase
	getAncestorsUseCase     *eventapp.GetAncestorsUseCase
	getDescendantsUseCase   *eventapp.GetDescendantsUseCase
	moveEventUseCase        *eventapp.MoveEventUseCase
	setEpochUseCase         *eventapp.SetEpochUseCase
	getEpochUseCase         *eventapp.GetEpochUseCase
	getTimelineUseCase      *eventapp.GetTimelineUseCase
	getStatChangesUseCase   *rpgeventapp.GetEventStatChangesUseCase
	logger                  logger.Logger
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(
	createEventUseCase *eventapp.CreateEventUseCase,
	getEventUseCase *eventapp.GetEventUseCase,
	listEventsUseCase *eventapp.ListEventsUseCase,
	updateEventUseCase *eventapp.UpdateEventUseCase,
	deleteEventUseCase *eventapp.DeleteEventUseCase,
	addReferenceUseCase *eventapp.AddReferenceUseCase,
	removeReferenceUseCase *eventapp.RemoveReferenceUseCase,
	getReferencesUseCase *eventapp.GetReferencesUseCase,
	updateReferenceUseCase *eventapp.UpdateReferenceUseCase,
	getChildrenUseCase *eventapp.GetChildrenUseCase,
	getAncestorsUseCase *eventapp.GetAncestorsUseCase,
	getDescendantsUseCase *eventapp.GetDescendantsUseCase,
	moveEventUseCase *eventapp.MoveEventUseCase,
	setEpochUseCase *eventapp.SetEpochUseCase,
	getEpochUseCase *eventapp.GetEpochUseCase,
	getTimelineUseCase *eventapp.GetTimelineUseCase,
	getStatChangesUseCase *rpgeventapp.GetEventStatChangesUseCase,
	logger logger.Logger,
) *EventHandler {
	return &EventHandler{
		createEventUseCase:     createEventUseCase,
		getEventUseCase:        getEventUseCase,
		listEventsUseCase:      listEventsUseCase,
		updateEventUseCase:     updateEventUseCase,
		deleteEventUseCase:     deleteEventUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		getChildrenUseCase:    getChildrenUseCase,
		getAncestorsUseCase:   getAncestorsUseCase,
		getDescendantsUseCase:  getDescendantsUseCase,
		moveEventUseCase:       moveEventUseCase,
		setEpochUseCase:        setEpochUseCase,
		getEpochUseCase:        getEpochUseCase,
		getTimelineUseCase:     getTimelineUseCase,
		getStatChangesUseCase:  getStatChangesUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/worlds/:world_id/events
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name            string   `json:"name"`
		Type            *string  `json:"type,omitempty"`
		Description     *string  `json:"description,omitempty"`
		Timeline        *string  `json:"timeline,omitempty"`
		Importance      int      `json:"importance"`
		ParentID        *string  `json:"parent_id,omitempty"`
		TimelinePosition float64 `json:"timeline_position"`
		IsEpoch         bool     `json:"is_epoch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsedParentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentID = &parsedParentID
	}

	output, err := h.createEventUseCase.Execute(r.Context(), eventapp.CreateEventInput{
		TenantID:        tenantID,
		WorldID:         worldID,
		Name:            req.Name,
		Type:            req.Type,
		Description:     req.Description,
		Timeline:        req.Timeline,
		Importance:      req.Importance,
		ParentID:        parentID,
		TimelinePosition: req.TimelinePosition,
		IsEpoch:         req.IsEpoch,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": output.Event,
	})
}

// Get handles GET /api/v1/events/{id}
func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getEventUseCase.Execute(r.Context(), eventapp.GetEventInput{
		TenantID: tenantID,
		ID:       eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": output.Event,
	})
}

// List handles GET /api/v1/worlds/:world_id/events
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listEventsUseCase.Execute(r.Context(), eventapp.ListEventsInput{
		TenantID: tenantID,
		WorldID:  worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": output.Events,
		"total":   len(output.Events),
	})
}

// Update handles PUT /api/v1/events/{id}
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name            *string  `json:"name,omitempty"`
		Type            *string  `json:"type,omitempty"`
		Description     *string  `json:"description,omitempty"`
		Timeline        *string  `json:"timeline,omitempty"`
		Importance      *int     `json:"importance,omitempty"`
		TimelinePosition *float64 `json:"timeline_position,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateEventUseCase.Execute(r.Context(), eventapp.UpdateEventInput{
		TenantID:        tenantID,
		ID:              eventID,
		Name:            req.Name,
		Type:            req.Type,
		Description:     req.Description,
		Timeline:        req.Timeline,
		Importance:      req.Importance,
		TimelinePosition: req.TimelinePosition,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": output.Event,
	})
}

// Delete handles DELETE /api/v1/events/{id}
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteEventUseCase.Execute(r.Context(), eventapp.DeleteEventInput{
		TenantID: tenantID,
		ID:       eventID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddReference handles POST /api/v1/events/{id}/references
func (h *EventHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType       string  `json:"entity_type"`
		EntityID         string  `json:"entity_id"`
		RelationshipType *string `json:"relationship_type,omitempty"`
		Notes            string  `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
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

	if err := h.addReferenceUseCase.Execute(r.Context(), eventapp.AddReferenceInput{
		TenantID:         tenantID,
		EventID:          eventID,
		EntityType:       req.EntityType,
		EntityID:         entityID,
		RelationshipType: req.RelationshipType,
		Notes:            req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveReference handles DELETE /api/v1/events/{id}/references/{entity_type}/{entity_id}
func (h *EventHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
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

	if err := h.removeReferenceUseCase.Execute(r.Context(), eventapp.RemoveReferenceInput{
		TenantID:   tenantID,
		EventID:    eventID,
		EntityType: entityType,
		EntityID:   entityID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReferences handles GET /api/v1/events/{id}/references
func (h *EventHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUseCase.Execute(r.Context(), eventapp.GetReferencesInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": output.References,
		"total":       len(output.References),
	})
}

// UpdateReference handles PUT /api/v1/event-references/{id}
func (h *EventHandler) UpdateReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	referenceID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		RelationshipType *string `json:"relationship_type,omitempty"`
		Notes            *string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if err := h.updateReferenceUseCase.Execute(r.Context(), eventapp.UpdateReferenceInput{
		TenantID:         tenantID,
		ID:               referenceID,
		RelationshipType: req.RelationshipType,
		Notes:            req.Notes,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetChildren handles GET /api/v1/events/{id}/children
func (h *EventHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	parentID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getChildrenUseCase.Execute(r.Context(), eventapp.GetChildrenInput{
		TenantID: tenantID,
		ParentID: parentID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": output.Events,
		"total":   len(output.Events),
	})
}

// GetAncestors handles GET /api/v1/events/{id}/ancestors
func (h *EventHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getAncestorsUseCase.Execute(r.Context(), eventapp.GetAncestorsInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": output.Events,
		"total":   len(output.Events),
	})
}

// GetDescendants handles GET /api/v1/events/{id}/descendants
func (h *EventHandler) GetDescendants(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getDescendantsUseCase.Execute(r.Context(), eventapp.GetDescendantsInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": output.Events,
		"total":   len(output.Events),
	})
}

// MoveEvent handles PUT /api/v1/events/{id}/move
func (h *EventHandler) MoveEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ParentID *string `json:"parent_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsedParentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentID = &parsedParentID
	}

	if err := h.moveEventUseCase.Execute(r.Context(), eventapp.MoveEventInput{
		TenantID: tenantID,
		EventID:  eventID,
		ParentID: parentID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetEpoch handles PUT /api/v1/events/{id}/epoch
func (h *EventHandler) SetEpoch(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.setEpochUseCase.Execute(r.Context(), eventapp.SetEpochInput{
		TenantID: tenantID,
		EventID:  eventID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEpoch handles GET /api/v1/worlds/{world_id}/epoch
func (h *EventHandler) GetEpoch(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getEpochUseCase.Execute(r.Context(), eventapp.GetEpochInput{
		TenantID: tenantID,
		WorldID:  worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": output.Event,
	})
}

// GetTimeline handles GET /api/v1/worlds/{world_id}/timeline
func (h *EventHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Parse query parameters for filtering
	var fromPos, toPos *float64
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		var f float64
		if _, err := fmt.Sscanf(fromStr, "%f", &f); err == nil {
			fromPos = &f
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		var t float64
		if _, err := fmt.Sscanf(toStr, "%f", &t); err == nil {
			toPos = &t
		}
	}

	output, err := h.getTimelineUseCase.Execute(r.Context(), eventapp.GetTimelineInput{
		TenantID: tenantID,
		WorldID:  worldID,
		FromPos:  fromPos,
		ToPos:    toPos,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": output.Events,
		"total":   len(output.Events),
	})
}

// GetStatChanges handles GET /api/v1/events/{id}/stat-changes
func (h *EventHandler) GetStatChanges(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getStatChangesUseCase.Execute(r.Context(), rpgeventapp.GetEventStatChangesInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_stats": output.CharacterStats,
		"artifact_stats":  output.ArtifactStats,
		"total":           len(output.CharacterStats) + len(output.ArtifactStats),
	})
}

