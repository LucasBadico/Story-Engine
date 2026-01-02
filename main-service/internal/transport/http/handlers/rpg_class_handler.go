package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// RPGClassHandler handles HTTP requests for RPG classes
type RPGClassHandler struct {
	createClassUseCase     *rpgclassapp.CreateRPGClassUseCase
	getClassUseCase        *rpgclassapp.GetRPGClassUseCase
	listClassesUseCase     *rpgclassapp.ListRPGClassesUseCase
	updateClassUseCase     *rpgclassapp.UpdateRPGClassUseCase
	deleteClassUseCase     *rpgclassapp.DeleteRPGClassUseCase
	addSkillUseCase        *rpgclassapp.AddSkillToClassUseCase
	listClassSkillsUseCase *rpgclassapp.ListClassSkillsUseCase
	logger                 logger.Logger
}

// NewRPGClassHandler creates a new RPGClassHandler
func NewRPGClassHandler(
	createClassUseCase *rpgclassapp.CreateRPGClassUseCase,
	getClassUseCase *rpgclassapp.GetRPGClassUseCase,
	listClassesUseCase *rpgclassapp.ListRPGClassesUseCase,
	updateClassUseCase *rpgclassapp.UpdateRPGClassUseCase,
	deleteClassUseCase *rpgclassapp.DeleteRPGClassUseCase,
	addSkillUseCase *rpgclassapp.AddSkillToClassUseCase,
	listClassSkillsUseCase *rpgclassapp.ListClassSkillsUseCase,
	logger logger.Logger,
) *RPGClassHandler {
	return &RPGClassHandler{
		createClassUseCase:     createClassUseCase,
		getClassUseCase:        getClassUseCase,
		listClassesUseCase:     listClassesUseCase,
		updateClassUseCase:     updateClassUseCase,
		deleteClassUseCase:     deleteClassUseCase,
		addSkillUseCase:        addSkillUseCase,
		listClassSkillsUseCase: listClassSkillsUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/rpg-systems/{id}/classes
func (h *RPGClassHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		ParentClassID *string          `json:"parent_class_id,omitempty"`
		Name          string           `json:"name"`
		Tier          *int             `json:"tier,omitempty"`
		Description   *string          `json:"description,omitempty"`
		Requirements  *json.RawMessage `json:"requirements,omitempty"`
		StatBonuses   *json.RawMessage `json:"stat_bonuses,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentClassID *uuid.UUID
	if req.ParentClassID != nil && *req.ParentClassID != "" {
		parsedID, err := uuid.Parse(*req.ParentClassID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_class_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentClassID = &parsedID
	}

	output, err := h.createClassUseCase.Execute(r.Context(), rpgclassapp.CreateRPGClassInput{
		RPGSystemID:   rpgSystemID,
		ParentClassID: parentClassID,
		Name:          req.Name,
		Tier:          req.Tier,
		Description:   req.Description,
		Requirements:  req.Requirements,
		StatBonuses:   req.StatBonuses,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"class": output.Class,
	})
}

// Get handles GET /api/v1/rpg-classes/{id}
func (h *RPGClassHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	classID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getClassUseCase.Execute(r.Context(), rpgclassapp.GetRPGClassInput{
		ID: classID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"class": output.Class,
	})
}

// List handles GET /api/v1/rpg-systems/{id}/classes
func (h *RPGClassHandler) List(w http.ResponseWriter, r *http.Request) {
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

	parentClassIDStr := r.URL.Query().Get("parent_class_id")
	var parentClassID *uuid.UUID
	if parentClassIDStr != "" {
		parsedID, err := uuid.Parse(parentClassIDStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_class_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentClassID = &parsedID
	}

	output, err := h.listClassesUseCase.Execute(r.Context(), rpgclassapp.ListRPGClassesInput{
		RPGSystemID:   rpgSystemID,
		ParentClassID: parentClassID,
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

// Update handles PUT /api/v1/rpg-classes/{id}
func (h *RPGClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	classID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ParentClassID *string          `json:"parent_class_id,omitempty"`
		Name          *string          `json:"name,omitempty"`
		Tier          *int             `json:"tier,omitempty"`
		Description   *string          `json:"description,omitempty"`
		Requirements  *json.RawMessage `json:"requirements,omitempty"`
		StatBonuses   *json.RawMessage `json:"stat_bonuses,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var parentClassID *uuid.UUID
	if req.ParentClassID != nil && *req.ParentClassID != "" {
		parsedID, err := uuid.Parse(*req.ParentClassID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "parent_class_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		parentClassID = &parsedID
	}

	output, err := h.updateClassUseCase.Execute(r.Context(), rpgclassapp.UpdateRPGClassInput{
		ID:            classID,
		ParentClassID: parentClassID,
		Name:          req.Name,
		Tier:          req.Tier,
		Description:   req.Description,
		Requirements:  req.Requirements,
		StatBonuses:   req.StatBonuses,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"class": output.Class,
	})
}

// Delete handles DELETE /api/v1/rpg-classes/{id}
func (h *RPGClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	classID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteClassUseCase.Execute(r.Context(), rpgclassapp.DeleteRPGClassInput{
		ID: classID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListSkills handles GET /api/v1/rpg-classes/{id}/skills
func (h *RPGClassHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	classID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listClassSkillsUseCase.Execute(r.Context(), rpgclassapp.ListClassSkillsInput{
		ClassID: classID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": output.Skills,
		"total":  len(output.Skills),
	})
}

// AddSkill handles POST /api/v1/rpg-classes/{id}/skills
func (h *RPGClassHandler) AddSkill(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	classID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		SkillID     uuid.UUID `json:"skill_id"`
		UnlockLevel int       `json:"unlock_level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.addSkillUseCase.Execute(r.Context(), rpgclassapp.AddSkillToClassInput{
		ClassID:     classID,
		SkillID:     req.SkillID,
		UnlockLevel: req.UnlockLevel,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"class_skill": output.ClassSkill,
	})
}
