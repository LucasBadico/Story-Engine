package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// SkillHandler handles HTTP requests for RPG skills
type SkillHandler struct {
	createSkillUseCase *skillapp.CreateSkillUseCase
	getSkillUseCase    *skillapp.GetSkillUseCase
	listSkillsUseCase  *skillapp.ListSkillsUseCase
	updateSkillUseCase *skillapp.UpdateSkillUseCase
	deleteSkillUseCase *skillapp.DeleteSkillUseCase
	logger             logger.Logger
}

// NewSkillHandler creates a new SkillHandler
func NewSkillHandler(
	createSkillUseCase *skillapp.CreateSkillUseCase,
	getSkillUseCase *skillapp.GetSkillUseCase,
	listSkillsUseCase *skillapp.ListSkillsUseCase,
	updateSkillUseCase *skillapp.UpdateSkillUseCase,
	deleteSkillUseCase *skillapp.DeleteSkillUseCase,
	logger logger.Logger,
) *SkillHandler {
	return &SkillHandler{
		createSkillUseCase: createSkillUseCase,
		getSkillUseCase:    getSkillUseCase,
		listSkillsUseCase:  listSkillsUseCase,
		updateSkillUseCase: updateSkillUseCase,
		deleteSkillUseCase: deleteSkillUseCase,
		logger:             logger,
	}
}

// Create handles POST /api/v1/rpg-systems/{id}/skills
func (h *SkillHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name          string           `json:"name"`
		Category      *string          `json:"category,omitempty"`
		Type          *string          `json:"type,omitempty"`
		Description   *string          `json:"description,omitempty"`
		Prerequisites *json.RawMessage `json:"prerequisites,omitempty"`
		MaxRank       *int             `json:"max_rank,omitempty"`
		EffectsSchema *json.RawMessage `json:"effects_schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var category *rpg.SkillCategory
	if req.Category != nil {
		cat := rpg.SkillCategory(*req.Category)
		category = &cat
	}
	var skillType *rpg.SkillType
	if req.Type != nil {
		st := rpg.SkillType(*req.Type)
		skillType = &st
	}

	output, err := h.createSkillUseCase.Execute(r.Context(), skillapp.CreateSkillInput{
		TenantID:      tenantID,
		RPGSystemID:   rpgSystemID,
		Name:          req.Name,
		Category:      category,
		Type:          skillType,
		Description:   req.Description,
		Prerequisites: req.Prerequisites,
		MaxRank:       req.MaxRank,
		EffectsSchema: req.EffectsSchema,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skill": output.Skill,
	})
}

// Get handles GET /api/v1/rpg-skills/{id}
func (h *SkillHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skillID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	output, err := h.getSkillUseCase.Execute(r.Context(), skillapp.GetSkillInput{
		TenantID: tenantID,
		ID:       skillID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skill": output.Skill,
	})
}

// List handles GET /api/v1/rpg-systems/{id}/skills
func (h *SkillHandler) List(w http.ResponseWriter, r *http.Request) {
	rpgSystemIDStr := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(rpgSystemIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	output, err := h.listSkillsUseCase.Execute(r.Context(), skillapp.ListSkillsInput{
		TenantID:    tenantID,
		RPGSystemID: rpgSystemID,
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

// Update handles PUT /api/v1/rpg-skills/{id}
func (h *SkillHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skillID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name          *string          `json:"name,omitempty"`
		Category      *string          `json:"category,omitempty"`
		Type          *string          `json:"type,omitempty"`
		Description   *string          `json:"description,omitempty"`
		Prerequisites *json.RawMessage `json:"prerequisites,omitempty"`
		MaxRank       *int             `json:"max_rank,omitempty"`
		EffectsSchema *json.RawMessage `json:"effects_schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var category *rpg.SkillCategory
	if req.Category != nil {
		cat := rpg.SkillCategory(*req.Category)
		category = &cat
	}
	var skillType *rpg.SkillType
	if req.Type != nil {
		st := rpg.SkillType(*req.Type)
		skillType = &st
	}

	tenantID := middleware.GetTenantID(r.Context())
	output, err := h.updateSkillUseCase.Execute(r.Context(), skillapp.UpdateSkillInput{
		TenantID:      tenantID,
		ID:            skillID,
		Name:          req.Name,
		Category:      category,
		Type:          skillType,
		Description:   req.Description,
		Prerequisites: req.Prerequisites,
		MaxRank:       req.MaxRank,
		EffectsSchema: req.EffectsSchema,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skill": output.Skill,
	})
}

// Delete handles DELETE /api/v1/rpg-skills/{id}
func (h *SkillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skillID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	if err := h.deleteSkillUseCase.Execute(r.Context(), skillapp.DeleteSkillInput{
		TenantID: tenantID,
		ID:       skillID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


