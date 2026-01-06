package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	characterrelationshipapp "github.com/story-engine/main-service/internal/application/world/character_relationship"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// CharacterHandler handles HTTP requests for characters
type CharacterHandler struct {
	createCharacterUseCase          *characterapp.CreateCharacterUseCase
	getCharacterUseCase             *characterapp.GetCharacterUseCase
	listCharactersUseCase           *characterapp.ListCharactersUseCase
	updateCharacterUseCase          *characterapp.UpdateCharacterUseCase
	deleteCharacterUseCase          *characterapp.DeleteCharacterUseCase
	addTraitUseCase                 *characterapp.AddTraitToCharacterUseCase
	removeTraitUseCase              *characterapp.RemoveTraitFromCharacterUseCase
	updateTraitUseCase              *characterapp.UpdateCharacterTraitUseCase
	getTraitsUseCase                *characterapp.GetCharacterTraitsUseCase
	getEventsUseCase                *characterapp.GetCharacterEventsUseCase
	createRelationshipUseCase       *characterrelationshipapp.CreateCharacterRelationshipUseCase
	getRelationshipUseCase          *characterrelationshipapp.GetCharacterRelationshipUseCase
	listRelationshipsUseCase        *characterrelationshipapp.ListCharacterRelationshipsUseCase
	updateRelationshipUseCase       *characterrelationshipapp.UpdateCharacterRelationshipUseCase
	deleteRelationshipUseCase       *characterrelationshipapp.DeleteCharacterRelationshipUseCase
	changeClassUseCase              *rpgcharacterapp.ChangeCharacterClassUseCase
	getAvailableClassesUseCase      *rpgcharacterapp.GetAvailableClassesUseCase
	logger                          logger.Logger
}

// NewCharacterHandler creates a new CharacterHandler
func NewCharacterHandler(
	createCharacterUseCase *characterapp.CreateCharacterUseCase,
	getCharacterUseCase *characterapp.GetCharacterUseCase,
	listCharactersUseCase *characterapp.ListCharactersUseCase,
	updateCharacterUseCase *characterapp.UpdateCharacterUseCase,
	deleteCharacterUseCase *characterapp.DeleteCharacterUseCase,
	addTraitUseCase *characterapp.AddTraitToCharacterUseCase,
	removeTraitUseCase *characterapp.RemoveTraitFromCharacterUseCase,
	updateTraitUseCase *characterapp.UpdateCharacterTraitUseCase,
	getTraitsUseCase *characterapp.GetCharacterTraitsUseCase,
	getEventsUseCase *characterapp.GetCharacterEventsUseCase,
	createRelationshipUseCase *characterrelationshipapp.CreateCharacterRelationshipUseCase,
	getRelationshipUseCase *characterrelationshipapp.GetCharacterRelationshipUseCase,
	listRelationshipsUseCase *characterrelationshipapp.ListCharacterRelationshipsUseCase,
	updateRelationshipUseCase *characterrelationshipapp.UpdateCharacterRelationshipUseCase,
	deleteRelationshipUseCase *characterrelationshipapp.DeleteCharacterRelationshipUseCase,
	changeClassUseCase *rpgcharacterapp.ChangeCharacterClassUseCase,
	getAvailableClassesUseCase *rpgcharacterapp.GetAvailableClassesUseCase,
	logger logger.Logger,
) *CharacterHandler {
	return &CharacterHandler{
		createCharacterUseCase:     createCharacterUseCase,
		getCharacterUseCase:        getCharacterUseCase,
		listCharactersUseCase:      listCharactersUseCase,
		updateCharacterUseCase:     updateCharacterUseCase,
		deleteCharacterUseCase:     deleteCharacterUseCase,
		addTraitUseCase:            addTraitUseCase,
		removeTraitUseCase:         removeTraitUseCase,
		updateTraitUseCase:         updateTraitUseCase,
		getTraitsUseCase:           getTraitsUseCase,
		getEventsUseCase:           getEventsUseCase,
		createRelationshipUseCase:  createRelationshipUseCase,
		getRelationshipUseCase:     getRelationshipUseCase,
		listRelationshipsUseCase:   listRelationshipsUseCase,
		updateRelationshipUseCase:  updateRelationshipUseCase,
		deleteRelationshipUseCase:  deleteRelationshipUseCase,
		changeClassUseCase:         changeClassUseCase,
		getAvailableClassesUseCase: getAvailableClassesUseCase,
		logger:                     logger,
	}
}

// Create handles POST /api/v1/worlds/:world_id/characters
func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		ArchetypeID *string `json:"archetype_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var archetypeID *uuid.UUID
	if req.ArchetypeID != nil && *req.ArchetypeID != "" {
		aid, err := uuid.Parse(*req.ArchetypeID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "archetype_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		archetypeID = &aid
	}

	output, err := h.createCharacterUseCase.Execute(r.Context(), characterapp.CreateCharacterInput{
		TenantID:    tenantID,
		WorldID:     worldID,
		ArchetypeID: archetypeID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": output.Character,
	})
}

// Get handles GET /api/v1/characters/{id}
func (h *CharacterHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getCharacterUseCase.Execute(r.Context(), characterapp.GetCharacterInput{
		TenantID: tenantID,
		ID:       characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": output.Character,
	})
}

// List handles GET /api/v1/worlds/:world_id/characters?limit=20&offset=0
func (h *CharacterHandler) List(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listCharactersUseCase.Execute(r.Context(), characterapp.ListCharactersInput{
		TenantID: tenantID,
		WorldID:  worldID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"characters": output.Characters,
		"total":      output.Total,
	})
}

// Update handles PUT /api/v1/characters/{id}
func (h *CharacterHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
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
		ArchetypeID *string `json:"archetype_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var archetypeID *uuid.UUID
	if req.ArchetypeID != nil {
		if *req.ArchetypeID == "" {
			archetypeID = nil
		} else {
			aid, err := uuid.Parse(*req.ArchetypeID)
			if err != nil {
				WriteError(w, &platformerrors.ValidationError{
					Field:   "archetype_id",
					Message: "invalid UUID format",
				}, http.StatusBadRequest)
				return
			}
			archetypeID = &aid
		}
	}

	output, err := h.updateCharacterUseCase.Execute(r.Context(), characterapp.UpdateCharacterInput{
		TenantID:    tenantID,
		ID:          characterID,
		Name:        req.Name,
		Description: req.Description,
		ArchetypeID: archetypeID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": output.Character,
	})
}

// Delete handles DELETE /api/v1/characters/{id}
func (h *CharacterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteCharacterUseCase.Execute(r.Context(), characterapp.DeleteCharacterInput{
		TenantID: tenantID,
		ID:       characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTraits handles GET /api/v1/characters/{id}/traits
func (h *CharacterHandler) GetTraits(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getTraitsUseCase.Execute(r.Context(), characterapp.GetCharacterTraitsInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"traits": output.Traits,
	})
}

// AddTrait handles POST /api/v1/characters/{id}/traits
func (h *CharacterHandler) AddTrait(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		TraitID string `json:"trait_id"`
		Value   string `json:"value"`
		Notes   string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	traitID, err := uuid.Parse(req.TraitID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.addTraitUseCase.Execute(r.Context(), characterapp.AddTraitToCharacterInput{
		TenantID:    tenantID,
		CharacterID: characterID,
		TraitID:     traitID,
		Value:       req.Value,
		Notes:       req.Notes,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveTrait handles DELETE /api/v1/characters/{id}/traits/{trait_id}
func (h *CharacterHandler) RemoveTrait(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	traitIDStr := r.PathValue("trait_id")
	traitID, err := uuid.Parse(traitIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.removeTraitUseCase.Execute(r.Context(), characterapp.RemoveTraitFromCharacterInput{
		TenantID:    tenantID,
		CharacterID: characterID,
		TraitID:     traitID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateTrait handles PUT /api/v1/characters/{id}/traits/{trait_id}
func (h *CharacterHandler) UpdateTrait(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	traitIDStr := r.PathValue("trait_id")
	traitID, err := uuid.Parse(traitIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "trait_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Value *string `json:"value"`
		Notes *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateTraitUseCase.Execute(r.Context(), characterapp.UpdateCharacterTraitInput{
		TenantID:    tenantID,
		CharacterID: characterID,
		TraitID:     traitID,
		Value:       req.Value,
		Notes:       req.Notes,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_trait": output.CharacterTrait,
	})
}

// ChangeClass handles PUT /api/v1/characters/{id}/class
func (h *CharacterHandler) ChangeClass(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ClassID    *string `json:"class_id,omitempty"`
		ClassLevel *int    `json:"class_level,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var classID *uuid.UUID
	if req.ClassID != nil && *req.ClassID != "" {
		parsedID, err := uuid.Parse(*req.ClassID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "class_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		classID = &parsedID
	}

	output, err := h.changeClassUseCase.Execute(r.Context(), rpgcharacterapp.ChangeCharacterClassInput{
		TenantID:    tenantID,
		CharacterID: characterID,
		ClassID:     classID,
		ClassLevel:  req.ClassLevel,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": output.Character,
	})
}

// GetAvailableClasses handles GET /api/v1/characters/{id}/available-classes
func (h *CharacterHandler) GetAvailableClasses(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getAvailableClassesUseCase.Execute(r.Context(), rpgcharacterapp.GetAvailableClassesInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"classes": output.Classes,
		"total":   len(output.Classes),
	})
}

// GetEvents handles GET /api/v1/characters/{id}/events
func (h *CharacterHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getEventsUseCase.Execute(r.Context(), characterapp.GetCharacterEventsInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event_references": output.EventReferences,
	})
}

// ListRelationships handles GET /api/v1/characters/{id}/relationships
func (h *CharacterHandler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listRelationshipsUseCase.Execute(r.Context(), characterrelationshipapp.ListCharacterRelationshipsInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationships": output.Relationships,
	})
}

// CreateRelationship handles POST /api/v1/characters/{id}/relationships
func (h *CharacterHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	character1ID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Character2ID     string `json:"character2_id"`
		RelationshipType string `json:"relationship_type"`
		Description      string `json:"description"`
		Bidirectional    bool   `json:"bidirectional"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	character2ID, err := uuid.Parse(req.Character2ID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "character2_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createRelationshipUseCase.Execute(r.Context(), characterrelationshipapp.CreateCharacterRelationshipInput{
		TenantID:         tenantID,
		Character1ID:     character1ID,
		Character2ID:     character2ID,
		RelationshipType: req.RelationshipType,
		Description:      req.Description,
		Bidirectional:    req.Bidirectional,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship": output.Relationship,
	})
}

// UpdateRelationship handles PUT /api/v1/character-relationships/{id}
func (h *CharacterHandler) UpdateRelationship(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	relationshipID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		RelationshipType *string `json:"relationship_type"`
		Description      *string `json:"description"`
		Bidirectional    *bool   `json:"bidirectional"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateRelationshipUseCase.Execute(r.Context(), characterrelationshipapp.UpdateCharacterRelationshipInput{
		TenantID:         tenantID,
		ID:               relationshipID,
		RelationshipType: req.RelationshipType,
		Description:      req.Description,
		Bidirectional:    req.Bidirectional,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationship": output.Relationship,
	})
}

// DeleteRelationship handles DELETE /api/v1/character-relationships/{id}
func (h *CharacterHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	relationshipID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteRelationshipUseCase.Execute(r.Context(), characterrelationshipapp.DeleteCharacterRelationshipInput{
		TenantID: tenantID,
		ID:       relationshipID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

