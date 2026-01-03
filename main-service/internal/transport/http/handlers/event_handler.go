package handlers

import (
	"encoding/json"
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
	createEventUseCase         *eventapp.CreateEventUseCase
	getEventUseCase            *eventapp.GetEventUseCase
	listEventsUseCase          *eventapp.ListEventsUseCase
	updateEventUseCase         *eventapp.UpdateEventUseCase
	deleteEventUseCase         *eventapp.DeleteEventUseCase
	addCharacterUseCase        *eventapp.AddCharacterToEventUseCase
	removeCharacterUseCase     *eventapp.RemoveCharacterFromEventUseCase
	getCharactersUseCase       *eventapp.GetEventCharactersUseCase
	addLocationUseCase         *eventapp.AddLocationToEventUseCase
	removeLocationUseCase      *eventapp.RemoveLocationFromEventUseCase
	getLocationsUseCase        *eventapp.GetEventLocationsUseCase
	addArtifactUseCase         *eventapp.AddArtifactToEventUseCase
	removeArtifactUseCase      *eventapp.RemoveArtifactFromEventUseCase
	getArtifactsUseCase        *eventapp.GetEventArtifactsUseCase
	getStatChangesUseCase      *rpgeventapp.GetEventStatChangesUseCase
	logger                     logger.Logger
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(
	createEventUseCase *eventapp.CreateEventUseCase,
	getEventUseCase *eventapp.GetEventUseCase,
	listEventsUseCase *eventapp.ListEventsUseCase,
	updateEventUseCase *eventapp.UpdateEventUseCase,
	deleteEventUseCase *eventapp.DeleteEventUseCase,
	addCharacterUseCase *eventapp.AddCharacterToEventUseCase,
	removeCharacterUseCase *eventapp.RemoveCharacterFromEventUseCase,
	getCharactersUseCase *eventapp.GetEventCharactersUseCase,
	addLocationUseCase *eventapp.AddLocationToEventUseCase,
	removeLocationUseCase *eventapp.RemoveLocationFromEventUseCase,
	getLocationsUseCase *eventapp.GetEventLocationsUseCase,
	addArtifactUseCase *eventapp.AddArtifactToEventUseCase,
	removeArtifactUseCase *eventapp.RemoveArtifactFromEventUseCase,
	getArtifactsUseCase *eventapp.GetEventArtifactsUseCase,
	getStatChangesUseCase *rpgeventapp.GetEventStatChangesUseCase,
	logger logger.Logger,
) *EventHandler {
	return &EventHandler{
		createEventUseCase:    createEventUseCase,
		getEventUseCase:       getEventUseCase,
		listEventsUseCase:     listEventsUseCase,
		updateEventUseCase:    updateEventUseCase,
		deleteEventUseCase:    deleteEventUseCase,
		addCharacterUseCase:   addCharacterUseCase,
		removeCharacterUseCase: removeCharacterUseCase,
		getCharactersUseCase:  getCharactersUseCase,
		addLocationUseCase:    addLocationUseCase,
		removeLocationUseCase: removeLocationUseCase,
		getLocationsUseCase:   getLocationsUseCase,
		addArtifactUseCase:    addArtifactUseCase,
		removeArtifactUseCase: removeArtifactUseCase,
		getArtifactsUseCase:   getArtifactsUseCase,
		getStatChangesUseCase: getStatChangesUseCase,
		logger:                logger,
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
		Name        string  `json:"name"`
		Type        *string `json:"type,omitempty"`
		Description *string `json:"description,omitempty"`
		Timeline    *string `json:"timeline,omitempty"`
		Importance  int     `json:"importance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createEventUseCase.Execute(r.Context(), eventapp.CreateEventInput{
		TenantID:    tenantID,
		WorldID:     worldID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Timeline:    req.Timeline,
		Importance:  req.Importance,
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
		Name        *string `json:"name,omitempty"`
		Type        *string `json:"type,omitempty"`
		Description *string `json:"description,omitempty"`
		Timeline    *string `json:"timeline,omitempty"`
		Importance  *int    `json:"importance,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateEventUseCase.Execute(r.Context(), eventapp.UpdateEventInput{
		TenantID:    tenantID,
		ID:          eventID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Timeline:    req.Timeline,
		Importance:  req.Importance,
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

// AddCharacter handles POST /api/v1/events/{id}/characters
func (h *EventHandler) AddCharacter(w http.ResponseWriter, r *http.Request) {
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
		CharacterID string  `json:"character_id"`
		Role        *string `json:"role,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	characterID, err := uuid.Parse(req.CharacterID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "character_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addCharacterUseCase.Execute(r.Context(), eventapp.AddCharacterToEventInput{
		TenantID:    tenantID,
		EventID:     eventID,
		CharacterID: characterID,
		Role:        req.Role,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveCharacter handles DELETE /api/v1/events/{id}/characters/{character_id}
func (h *EventHandler) RemoveCharacter(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterIDStr := r.PathValue("character_id")

	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "character_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeCharacterUseCase.Execute(r.Context(), eventapp.RemoveCharacterFromEventInput{
		TenantID:    tenantID,
		EventID:     eventID,
		CharacterID: characterID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCharacters handles GET /api/v1/events/{id}/characters
func (h *EventHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.getCharactersUseCase.Execute(r.Context(), eventapp.GetEventCharactersInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"characters": output.Characters,
		"total":       len(output.Characters),
	})
}

// AddLocation handles POST /api/v1/events/{id}/locations
func (h *EventHandler) AddLocation(w http.ResponseWriter, r *http.Request) {
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
		LocationID   string  `json:"location_id"`
		Significance *string `json:"significance,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "location_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addLocationUseCase.Execute(r.Context(), eventapp.AddLocationToEventInput{
		TenantID:     tenantID,
		EventID:      eventID,
		LocationID:   locationID,
		Significance: req.Significance,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveLocation handles DELETE /api/v1/events/{id}/locations/{location_id}
func (h *EventHandler) RemoveLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	locationIDStr := r.PathValue("location_id")

	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "location_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeLocationUseCase.Execute(r.Context(), eventapp.RemoveLocationFromEventInput{
		TenantID:   tenantID,
		EventID:    eventID,
		LocationID: locationID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLocations handles GET /api/v1/events/{id}/locations
func (h *EventHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.getLocationsUseCase.Execute(r.Context(), eventapp.GetEventLocationsInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locations": output.Locations,
		"total":      len(output.Locations),
	})
}

// AddArtifact handles POST /api/v1/events/{id}/artifacts
func (h *EventHandler) AddArtifact(w http.ResponseWriter, r *http.Request) {
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
		ArtifactID string  `json:"artifact_id"`
		Role       *string `json:"role,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	artifactID, err := uuid.Parse(req.ArtifactID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "artifact_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addArtifactUseCase.Execute(r.Context(), eventapp.AddArtifactToEventInput{
		TenantID:   tenantID,
		EventID:    eventID,
		ArtifactID: artifactID,
		Role:       req.Role,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveArtifact handles DELETE /api/v1/events/{id}/artifacts/{artifact_id}
func (h *EventHandler) RemoveArtifact(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	artifactIDStr := r.PathValue("artifact_id")

	eventID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	artifactID, err := uuid.Parse(artifactIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "artifact_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeArtifactUseCase.Execute(r.Context(), eventapp.RemoveArtifactFromEventInput{
		TenantID:   tenantID,
		EventID:    eventID,
		ArtifactID: artifactID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetArtifacts handles GET /api/v1/events/{id}/artifacts
func (h *EventHandler) GetArtifacts(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.getArtifactsUseCase.Execute(r.Context(), eventapp.GetEventArtifactsInput{
		TenantID: tenantID,
		EventID:  eventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifacts": output.Artifacts,
		"total":      len(output.Artifacts),
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

