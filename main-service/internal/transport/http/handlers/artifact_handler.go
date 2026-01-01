package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
)

// ArtifactHandler handles HTTP requests for artifacts
type ArtifactHandler struct {
	createArtifactUseCase      *artifactapp.CreateArtifactUseCase
	getArtifactUseCase         *artifactapp.GetArtifactUseCase
	listArtifactsUseCase       *artifactapp.ListArtifactsUseCase
	updateArtifactUseCase      *artifactapp.UpdateArtifactUseCase
	deleteArtifactUseCase      *artifactapp.DeleteArtifactUseCase
	getReferencesUseCase       *artifactapp.GetArtifactReferencesUseCase
	addReferenceUseCase        *artifactapp.AddArtifactReferenceUseCase
	removeReferenceUseCase     *artifactapp.RemoveArtifactReferenceUseCase
	logger                     logger.Logger
}

// NewArtifactHandler creates a new ArtifactHandler
func NewArtifactHandler(
	createArtifactUseCase *artifactapp.CreateArtifactUseCase,
	getArtifactUseCase *artifactapp.GetArtifactUseCase,
	listArtifactsUseCase *artifactapp.ListArtifactsUseCase,
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase,
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase,
	getReferencesUseCase *artifactapp.GetArtifactReferencesUseCase,
	addReferenceUseCase *artifactapp.AddArtifactReferenceUseCase,
	removeReferenceUseCase *artifactapp.RemoveArtifactReferenceUseCase,
	logger logger.Logger,
) *ArtifactHandler {
	return &ArtifactHandler{
		createArtifactUseCase:  createArtifactUseCase,
		getArtifactUseCase:     getArtifactUseCase,
		listArtifactsUseCase:   listArtifactsUseCase,
		updateArtifactUseCase:  updateArtifactUseCase,
		deleteArtifactUseCase:  deleteArtifactUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/worlds/:world_id/artifacts
func (h *ArtifactHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		CharacterIDs []string `json:"character_ids"`
		LocationIDs  []string `json:"location_ids"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Rarity       string   `json:"rarity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	characterIDs := make([]uuid.UUID, 0, len(req.CharacterIDs))
	for _, cidStr := range req.CharacterIDs {
		cid, err := uuid.Parse(cidStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "character_ids",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		characterIDs = append(characterIDs, cid)
	}

	locationIDs := make([]uuid.UUID, 0, len(req.LocationIDs))
	for _, lidStr := range req.LocationIDs {
		lid, err := uuid.Parse(lidStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "location_ids",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		locationIDs = append(locationIDs, lid)
	}

	output, err := h.createArtifactUseCase.Execute(r.Context(), artifactapp.CreateArtifactInput{
		WorldID:      worldID,
		CharacterIDs: characterIDs,
		LocationIDs:  locationIDs,
		Name:         req.Name,
		Description:  req.Description,
		Rarity:       req.Rarity,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifact": output.Artifact,
	})
}

// Get handles GET /api/v1/artifacts/{id}
func (h *ArtifactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getArtifactUseCase.Execute(r.Context(), artifactapp.GetArtifactInput{
		ID: artifactID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifact": output.Artifact,
	})
}

// List handles GET /api/v1/worlds/:world_id/artifacts?limit=20&offset=0
func (h *ArtifactHandler) List(w http.ResponseWriter, r *http.Request) {
	worldIDStr := r.PathValue("world_id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "world_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
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

	output, err := h.listArtifactsUseCase.Execute(r.Context(), artifactapp.ListArtifactsInput{
		WorldID: worldID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifacts": output.Artifacts,
		"total":     output.Total,
	})
}

// Update handles PUT /api/v1/artifacts/{id}
func (h *ArtifactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name         *string   `json:"name"`
		Description  *string   `json:"description"`
		Rarity       *string   `json:"rarity"`
		CharacterIDs *[]string `json:"character_ids"`
		LocationIDs  *[]string `json:"location_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var characterIDs *[]uuid.UUID
	if req.CharacterIDs != nil {
		ids := make([]uuid.UUID, 0, len(*req.CharacterIDs))
		for _, cidStr := range *req.CharacterIDs {
			cid, err := uuid.Parse(cidStr)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "character_ids",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			ids = append(ids, cid)
		}
		characterIDs = &ids
	}

	var locationIDs *[]uuid.UUID
	if req.LocationIDs != nil {
		ids := make([]uuid.UUID, 0, len(*req.LocationIDs))
		for _, lidStr := range *req.LocationIDs {
			lid, err := uuid.Parse(lidStr)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "location_ids",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			ids = append(ids, lid)
		}
		locationIDs = &ids
	}

	output, err := h.updateArtifactUseCase.Execute(r.Context(), artifactapp.UpdateArtifactInput{
		ID:          artifactID,
		Name:        req.Name,
		Description: req.Description,
		Rarity:      req.Rarity,
		CharacterIDs: characterIDs,
		LocationIDs: locationIDs,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifact": output.Artifact,
	})
}

// Delete handles DELETE /api/v1/artifacts/{id}
func (h *ArtifactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteArtifactUseCase.Execute(r.Context(), artifactapp.DeleteArtifactInput{
		ID: artifactID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReferences handles GET /api/v1/artifacts/{id}/references
func (h *ArtifactHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUseCase.Execute(r.Context(), artifactapp.GetArtifactReferencesInput{
		ArtifactID: artifactID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": output.References,
	})
}

// AddReference handles POST /api/v1/artifacts/{id}/references
func (h *ArtifactHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
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

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityType := world.ArtifactReferenceEntityType(req.EntityType)
	if entityType != world.ArtifactReferenceEntityTypeCharacter && entityType != world.ArtifactReferenceEntityTypeLocation {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "must be 'character' or 'location'",
		}, http.StatusBadRequest)
		return
	}

	err = h.addReferenceUseCase.Execute(r.Context(), artifactapp.AddArtifactReferenceInput{
		ArtifactID: artifactID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveReference handles DELETE /api/v1/artifacts/{id}/references/{entity_type}/{entity_id}
func (h *ArtifactHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityTypeStr := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityType := world.ArtifactReferenceEntityType(entityTypeStr)
	if entityType != world.ArtifactReferenceEntityTypeCharacter && entityType != world.ArtifactReferenceEntityTypeLocation {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "must be 'character' or 'location'",
		}, http.StatusBadRequest)
		return
	}

	err = h.removeReferenceUseCase.Execute(r.Context(), artifactapp.RemoveArtifactReferenceInput{
		ArtifactID: artifactID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

