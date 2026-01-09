package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/core/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// CharacterHandler handles HTTP requests for characters
type CharacterHandler struct {
	createCharacterUseCase *characterapp.CreateCharacterUseCase
	getCharacterUseCase    *characterapp.GetCharacterUseCase
	listCharactersUseCase  *characterapp.ListCharactersUseCase
	updateCharacterUseCase *characterapp.UpdateCharacterUseCase
	deleteCharacterUseCase *characterapp.DeleteCharacterUseCase
	addTraitUseCase        *characterapp.AddTraitToCharacterUseCase
	removeTraitUseCase     *characterapp.RemoveTraitFromCharacterUseCase
	updateTraitUseCase     *characterapp.UpdateCharacterTraitUseCase
	getTraitsUseCase       *characterapp.GetCharacterTraitsUseCase
	getEventsUseCase       *characterapp.GetCharacterEventsUseCase
	// Character relationship use cases - using entity_relations
	createRelationUseCase        *relationapp.CreateRelationUseCase
	getRelationUseCase           *relationapp.GetRelationUseCase
	listRelationsBySourceUseCase *relationapp.ListRelationsBySourceUseCase
	listRelationsByTargetUseCase *relationapp.ListRelationsByTargetUseCase
	updateRelationUseCase        *relationapp.UpdateRelationUseCase
	deleteRelationUseCase        *relationapp.DeleteRelationUseCase
	changeClassUseCase           *rpgcharacterapp.ChangeCharacterClassUseCase
	getAvailableClassesUseCase   *rpgcharacterapp.GetAvailableClassesUseCase
	logger                       logger.Logger
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
	createRelationUseCase *relationapp.CreateRelationUseCase,
	getRelationUseCase *relationapp.GetRelationUseCase,
	listRelationsBySourceUseCase *relationapp.ListRelationsBySourceUseCase,
	listRelationsByTargetUseCase *relationapp.ListRelationsByTargetUseCase,
	updateRelationUseCase *relationapp.UpdateRelationUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	changeClassUseCase *rpgcharacterapp.ChangeCharacterClassUseCase,
	getAvailableClassesUseCase *rpgcharacterapp.GetAvailableClassesUseCase,
	logger logger.Logger,
) *CharacterHandler {
	return &CharacterHandler{
		createCharacterUseCase:       createCharacterUseCase,
		getCharacterUseCase:          getCharacterUseCase,
		listCharactersUseCase:        listCharactersUseCase,
		updateCharacterUseCase:       updateCharacterUseCase,
		deleteCharacterUseCase:       deleteCharacterUseCase,
		addTraitUseCase:              addTraitUseCase,
		removeTraitUseCase:           removeTraitUseCase,
		updateTraitUseCase:           updateTraitUseCase,
		getTraitsUseCase:             getTraitsUseCase,
		getEventsUseCase:             getEventsUseCase,
		createRelationUseCase:        createRelationUseCase,
		getRelationUseCase:           getRelationUseCase,
		listRelationsBySourceUseCase: listRelationsBySourceUseCase,
		listRelationsByTargetUseCase: listRelationsByTargetUseCase,
		updateRelationUseCase:        updateRelationUseCase,
		deleteRelationUseCase:        deleteRelationUseCase,
		changeClassUseCase:           changeClassUseCase,
		getAvailableClassesUseCase:   getAvailableClassesUseCase,
		logger:                       logger,
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

// CharacterRelationshipDTO is a compatibility DTO for CharacterRelationship (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type CharacterRelationshipDTO struct {
	ID               string    `json:"id"`
	Character1ID     string    `json:"character1_id"`
	Character2ID     string    `json:"character2_id"`
	RelationshipType string    `json:"relationship_type"`
	Description      string    `json:"description"`
	Bidirectional    bool      `json:"bidirectional"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// entityRelationToCharacterRelationshipDTO converts EntityRelation to CharacterRelationshipDTO
func entityRelationToCharacterRelationshipDTO(rel *relation.EntityRelation, characterID uuid.UUID) *CharacterRelationshipDTO {
	// Determine character1_id and character2_id based on source/target
	var char1ID, char2ID uuid.UUID
	if rel.SourceType == "character" && rel.SourceID == characterID {
		char1ID = characterID
		if rel.TargetType == "character" {
			char2ID = rel.TargetID
		}
	} else if rel.TargetType == "character" && rel.TargetID == characterID {
		char1ID = characterID
		if rel.SourceType == "character" {
			char2ID = rel.SourceID
		}
	} else {
		// Fallback: use source/target as-is
		if rel.SourceType == "character" {
			char1ID = rel.SourceID
		}
		if rel.TargetType == "character" {
			char2ID = rel.TargetID
		}
	}

	// Extract description from attributes or summary
	description := rel.Summary
	if rel.Attributes != nil {
		if desc, ok := rel.Attributes["description"].(string); ok && desc != "" {
			description = desc
		}
	}

	// Check if bidirectional (has mirror)
	bidirectional := rel.MirrorID != nil

	return &CharacterRelationshipDTO{
		ID:               rel.ID.String(),
		Character1ID:     char1ID.String(),
		Character2ID:     char2ID.String(),
		RelationshipType: rel.RelationType,
		Description:      description,
		Bidirectional:    bidirectional,
		CreatedAt:        rel.CreatedAt,
		UpdatedAt:        rel.UpdatedAt,
	}
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

	// Get character to retrieve world_id
	charOutput, err := h.getCharacterUseCase.Execute(r.Context(), characterapp.GetCharacterInput{
		TenantID: tenantID,
		ID:       characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	_ = charOutput // Character retrieved but world_id not needed for listing

	// List relations where character is source
	sourceOutput, err := h.listRelationsBySourceUseCase.Execute(r.Context(), relationapp.ListRelationsBySourceInput{
		TenantID:   tenantID,
		SourceType: "character",
		SourceID:   characterID,
		Options: repositories.ListOptions{
			Limit:          100,
			ExcludeMirrors: true, // Only get primary relations
		},
	})
	if err != nil {
		h.logger.Error("failed to list relations by source", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// List relations where character is target (for bidirectional relationships)
	targetOutput, err := h.listRelationsByTargetUseCase.Execute(r.Context(), relationapp.ListRelationsByTargetInput{
		TenantID:   tenantID,
		TargetType: "character",
		TargetID:   characterID,
		Options: repositories.ListOptions{
			Limit:          100,
			ExcludeMirrors: true, // Only get primary relations
		},
	})
	if err != nil {
		h.logger.Error("failed to list relations by target", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Combine and convert to DTOs
	allRelations := make([]*relation.EntityRelation, 0)
	allRelations = append(allRelations, sourceOutput.Relations.Items...)
	// Only include target relations if they are bidirectional (have a mirror)
	for _, rel := range targetOutput.Relations.Items {
		if rel.MirrorID != nil {
			allRelations = append(allRelations, rel)
		}
	}

	// Filter to only character-to-character relations and convert
	relationships := make([]*CharacterRelationshipDTO, 0)
	for _, rel := range allRelations {
		if (rel.SourceType == "character" && rel.TargetType == "character") ||
			(rel.TargetType == "character" && rel.SourceType == "character") {
			dto := entityRelationToCharacterRelationshipDTO(rel, characterID)
			relationships = append(relationships, dto)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"relationships": relationships,
	})
}

// CreateRelationship handles POST /api/v1/characters/{id}/relationships
func (h *CharacterHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
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

	// Get character to retrieve world_id
	charOutput, err := h.getCharacterUseCase.Execute(r.Context(), characterapp.GetCharacterInput{
		TenantID: tenantID,
		ID:       characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	var req struct {
		OtherCharacterID string `json:"other_character_id"`
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

	otherCharacterID, err := uuid.Parse(req.OtherCharacterID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "other_character_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if req.RelationshipType == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "relationship_type",
			Message: "relationship_type is required",
		}, http.StatusBadRequest)
		return
	}

	if characterID == otherCharacterID {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "other_character_id",
			Message: "cannot create relationship with self",
		}, http.StatusBadRequest)
		return
	}

	// Create attributes map
	attributes := make(map[string]interface{})
	if req.Description != "" {
		attributes["description"] = req.Description
	}

	// Create relation
	output, err := h.createRelationUseCase.Execute(r.Context(), relationapp.CreateRelationInput{
		TenantID:     tenantID,
		WorldID:      charOutput.Character.WorldID,
		SourceType:   "character",
		SourceID:     characterID,
		TargetType:   "character",
		TargetID:     otherCharacterID,
		RelationType: req.RelationshipType,
		Attributes:   attributes,
		Summary:      req.Description,
		CreateMirror: req.Bidirectional,
	})
	if err != nil {
		h.logger.Error("failed to create relationship", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	dto := entityRelationToCharacterRelationshipDTO(output.Relation, characterID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// UpdateRelationship handles PUT /api/v1/character-relationships/{id}
func (h *CharacterHandler) UpdateRelationship(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	relationID, err := uuid.Parse(id)
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

	// Get existing relation
	relOutput, err := h.getRelationUseCase.Execute(r.Context(), relationapp.GetRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	rel := relOutput.Relation

	// Validate it's a character-to-character relation
	if rel.SourceType != "character" || rel.TargetType != "character" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "relation is not a character relationship",
		}, http.StatusBadRequest)
		return
	}

	// Prepare update input
	updateInput := relationapp.UpdateRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	}

	// Update attributes
	attributes := make(map[string]interface{})
	if rel.Attributes != nil {
		for k, v := range rel.Attributes {
			attributes[k] = v
		}
	}

	if req.Description != nil {
		attributes["description"] = *req.Description
		summary := *req.Description
		updateInput.Summary = &summary
	}

	if req.RelationshipType != nil {
		updateInput.RelationType = req.RelationshipType
	}

	if len(attributes) > 0 {
		updateInput.Attributes = &attributes
	}

	// Update relation
	updateOutput, err := h.updateRelationUseCase.Execute(r.Context(), updateInput)
	if err != nil {
		h.logger.Error("failed to update relationship", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Determine character_id for DTO conversion (use source)
	characterID := updateOutput.Relation.SourceID

	dto := entityRelationToCharacterRelationshipDTO(updateOutput.Relation, characterID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// DeleteRelationship handles DELETE /api/v1/character-relationships/{id}
func (h *CharacterHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	relationID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Verify relation exists and is a character relationship
	relOutput, err := h.getRelationUseCase.Execute(r.Context(), relationapp.GetRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	})
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	rel := relOutput.Relation
	if rel.SourceType != "character" || rel.TargetType != "character" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "relation is not a character relationship",
		}, http.StatusBadRequest)
		return
	}

	// Delete relation (this will also delete mirror if exists)
	err = h.deleteRelationUseCase.Execute(r.Context(), relationapp.DeleteRelationInput{
		TenantID: tenantID,
		ID:       relationID,
	})
	if err != nil {
		h.logger.Error("failed to delete relationship", "error", err)
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
