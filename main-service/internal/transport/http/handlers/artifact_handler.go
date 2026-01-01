package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
)

// ArtifactHandler handles HTTP requests for artifacts
type ArtifactHandler struct {
	createArtifactUseCase *artifactapp.CreateArtifactUseCase
	getArtifactUseCase    *artifactapp.GetArtifactUseCase
	listArtifactsUseCase  *artifactapp.ListArtifactsUseCase
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase
	logger                logger.Logger
}

// NewArtifactHandler creates a new ArtifactHandler
func NewArtifactHandler(
	createArtifactUseCase *artifactapp.CreateArtifactUseCase,
	getArtifactUseCase *artifactapp.GetArtifactUseCase,
	listArtifactsUseCase *artifactapp.ListArtifactsUseCase,
	updateArtifactUseCase *artifactapp.UpdateArtifactUseCase,
	deleteArtifactUseCase *artifactapp.DeleteArtifactUseCase,
	logger logger.Logger,
) *ArtifactHandler {
	return &ArtifactHandler{
		createArtifactUseCase: createArtifactUseCase,
		getArtifactUseCase:    getArtifactUseCase,
		listArtifactsUseCase:  listArtifactsUseCase,
		updateArtifactUseCase: updateArtifactUseCase,
		deleteArtifactUseCase: deleteArtifactUseCase,
		logger:                logger,
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
		CharacterID *string `json:"character_id"`
		LocationID  *string `json:"location_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Rarity      string  `json:"rarity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var characterID *uuid.UUID
	if req.CharacterID != nil && *req.CharacterID != "" {
		cid, err := uuid.Parse(*req.CharacterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "character_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		characterID = &cid
	}

	var locationID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		lid, err := uuid.Parse(*req.LocationID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "location_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		locationID = &lid
	}

	output, err := h.createArtifactUseCase.Execute(r.Context(), artifactapp.CreateArtifactInput{
		WorldID:     worldID,
		CharacterID: characterID,
		LocationID:  locationID,
		Name:        req.Name,
		Description: req.Description,
		Rarity:      req.Rarity,
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
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Rarity      *string `json:"rarity"`
		CharacterID *string `json:"character_id"`
		LocationID  *string `json:"location_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var characterID *uuid.UUID
	if req.CharacterID != nil {
		if *req.CharacterID == "" {
			characterID = nil
		} else {
			cid, err := uuid.Parse(*req.CharacterID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "character_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			characterID = &cid
		}
	}

	var locationID *uuid.UUID
	if req.LocationID != nil {
		if *req.LocationID == "" {
			locationID = nil
		} else {
			lid, err := uuid.Parse(*req.LocationID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "location_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			locationID = &lid
		}
	}

	output, err := h.updateArtifactUseCase.Execute(r.Context(), artifactapp.UpdateArtifactInput{
		ID:          artifactID,
		Name:        req.Name,
		Description: req.Description,
		Rarity:      req.Rarity,
		CharacterID: characterID,
		LocationID:  locationID,
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

