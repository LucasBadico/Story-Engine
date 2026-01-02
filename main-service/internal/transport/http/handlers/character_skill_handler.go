package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// CharacterSkillHandler handles HTTP requests for character skills
type CharacterSkillHandler struct {
	learnSkillUseCase      *characterskillapp.LearnSkillUseCase
	listSkillsUseCase       *characterskillapp.ListCharacterSkillsUseCase
	updateSkillUseCase      *characterskillapp.UpdateCharacterSkillUseCase
	deleteSkillUseCase      *characterskillapp.DeleteCharacterSkillUseCase
	logger                  logger.Logger
}

// NewCharacterSkillHandler creates a new CharacterSkillHandler
func NewCharacterSkillHandler(
	learnSkillUseCase *characterskillapp.LearnSkillUseCase,
	listSkillsUseCase *characterskillapp.ListCharacterSkillsUseCase,
	updateSkillUseCase *characterskillapp.UpdateCharacterSkillUseCase,
	deleteSkillUseCase *characterskillapp.DeleteCharacterSkillUseCase,
	logger logger.Logger,
) *CharacterSkillHandler {
	return &CharacterSkillHandler{
		learnSkillUseCase: learnSkillUseCase,
		listSkillsUseCase: listSkillsUseCase,
		updateSkillUseCase: updateSkillUseCase,
		deleteSkillUseCase: deleteSkillUseCase,
		logger:            logger,
	}
}

// Learn handles POST /api/v1/characters/{id}/skills
func (h *CharacterSkillHandler) Learn(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	characterIDStr := r.PathValue("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		SkillID uuid.UUID `json:"skill_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.learnSkillUseCase.Execute(r.Context(), characterskillapp.LearnSkillInput{
		CharacterID: characterID,
		SkillID:     req.SkillID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_skill": output.CharacterSkill,
	})
}

// List handles GET /api/v1/characters/{id}/skills
func (h *CharacterSkillHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	characterIDStr := r.PathValue("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	activeOnly := r.URL.Query().Get("active_only") == "true"

	output, err := h.listSkillsUseCase.Execute(r.Context(), characterskillapp.ListCharacterSkillsInput{
		CharacterID: characterID,
		ActiveOnly:  activeOnly,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": output.Skills,
		"total":   len(output.Skills),
	})
}

// Update handles PUT /api/v1/characters/{id}/skills/{skill_id}
func (h *CharacterSkillHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	_, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	_, err = uuid.Parse(r.PathValue("skill_id"))
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "skill_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Rank     *int  `json:"rank,omitempty"`
		AddXP    *int  `json:"add_xp,omitempty"`
		IsActive *bool `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	// Get character skill ID first
	// Note: In a real implementation, we'd need a method to get by character+skill
	// For now, we'll need to update the use case to accept character_id + skill_id
	// or add a GetByCharacterAndSkill method to the handler
	// For simplicity, let's assume the use case will handle this internally
	// We'll need to modify UpdateCharacterSkillUseCase to accept character_id + skill_id instead of ID

	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "use character_skill_id endpoint instead",
	})
}

// Delete handles DELETE /api/v1/characters/{id}/skills/{skill_id}
func (h *CharacterSkillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	_, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	_, err = uuid.Parse(r.PathValue("skill_id"))
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "skill_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	// Similar to Update, we'd need to get the character_skill ID first
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "use character_skill_id endpoint instead",
	})
}

// UpdateByID handles PUT /api/v1/character-skills/{id}
func (h *CharacterSkillHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterSkillID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Rank     *int  `json:"rank,omitempty"`
		AddXP    *int  `json:"add_xp,omitempty"`
		IsActive *bool `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateSkillUseCase.Execute(r.Context(), characterskillapp.UpdateCharacterSkillInput{
		ID:       characterSkillID,
		Rank:     req.Rank,
		AddXP:    req.AddXP,
		IsActive: req.IsActive,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_skill": output.CharacterSkill,
	})
}

// DeleteByID handles DELETE /api/v1/character-skills/{id}
func (h *CharacterSkillHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	characterSkillID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteSkillUseCase.Execute(r.Context(), characterskillapp.DeleteCharacterSkillInput{
		ID: characterSkillID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

