package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// RPGSystemHandler handles HTTP requests for RPG systems
type RPGSystemHandler struct {
	createRPGSystemUseCase *rpgsystemapp.CreateRPGSystemUseCase
	getRPGSystemUseCase    *rpgsystemapp.GetRPGSystemUseCase
	listRPGSystemsUseCase  *rpgsystemapp.ListRPGSystemsUseCase
	updateRPGSystemUseCase *rpgsystemapp.UpdateRPGSystemUseCase
	deleteRPGSystemUseCase *rpgsystemapp.DeleteRPGSystemUseCase
	logger                  logger.Logger
}

// NewRPGSystemHandler creates a new RPGSystemHandler
func NewRPGSystemHandler(
	createRPGSystemUseCase *rpgsystemapp.CreateRPGSystemUseCase,
	getRPGSystemUseCase *rpgsystemapp.GetRPGSystemUseCase,
	listRPGSystemsUseCase *rpgsystemapp.ListRPGSystemsUseCase,
	updateRPGSystemUseCase *rpgsystemapp.UpdateRPGSystemUseCase,
	deleteRPGSystemUseCase *rpgsystemapp.DeleteRPGSystemUseCase,
	logger logger.Logger,
) *RPGSystemHandler {
	return &RPGSystemHandler{
		createRPGSystemUseCase: createRPGSystemUseCase,
		getRPGSystemUseCase:    getRPGSystemUseCase,
		listRPGSystemsUseCase:  listRPGSystemsUseCase,
		updateRPGSystemUseCase: updateRPGSystemUseCase,
		deleteRPGSystemUseCase: deleteRPGSystemUseCase,
		logger:                 logger,
	}
}

// Create handles POST /api/v1/rpg-systems
func (h *RPGSystemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string           `json:"name"`
		Description       *string           `json:"description,omitempty"`
		BaseStatsSchema   json.RawMessage  `json:"base_stats_schema"`
		DerivedStatsSchema *json.RawMessage `json:"derived_stats_schema,omitempty"`
		ProgressionSchema *json.RawMessage `json:"progression_schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	output, err := h.createRPGSystemUseCase.Execute(r.Context(), rpgsystemapp.CreateRPGSystemInput{
		TenantID:          &tenantID,
		Name:              req.Name,
		Description:       req.Description,
		BaseStatsSchema:   req.BaseStatsSchema,
		DerivedStatsSchema: req.DerivedStatsSchema,
		ProgressionSchema: req.ProgressionSchema,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rpg_system": output.RPGSystem,
	})
}

// Get handles GET /api/v1/rpg-systems/{id}
func (h *RPGSystemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getRPGSystemUseCase.Execute(r.Context(), rpgsystemapp.GetRPGSystemInput{
		ID: rpgSystemID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rpg_system": output.RPGSystem,
	})
}

// List handles GET /api/v1/rpg-systems
func (h *RPGSystemHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	output, err := h.listRPGSystemsUseCase.Execute(r.Context(), rpgsystemapp.ListRPGSystemsInput{
		TenantID: &tenantID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rpg_systems": output.RPGSystems,
		"total":       len(output.RPGSystems),
	})
}

// Update handles PUT /api/v1/rpg-systems/{id}
func (h *RPGSystemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name              *string          `json:"name,omitempty"`
		Description       *string           `json:"description,omitempty"`
		BaseStatsSchema   *json.RawMessage  `json:"base_stats_schema,omitempty"`
		DerivedStatsSchema *json.RawMessage `json:"derived_stats_schema,omitempty"`
		ProgressionSchema *json.RawMessage `json:"progression_schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateRPGSystemUseCase.Execute(r.Context(), rpgsystemapp.UpdateRPGSystemInput{
		ID:                rpgSystemID,
		Name:              req.Name,
		Description:       req.Description,
		BaseStatsSchema:   req.BaseStatsSchema,
		DerivedStatsSchema: req.DerivedStatsSchema,
		ProgressionSchema: req.ProgressionSchema,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rpg_system": output.RPGSystem,
	})
}

// Delete handles DELETE /api/v1/rpg-systems/{id}
func (h *RPGSystemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rpgSystemID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteRPGSystemUseCase.Execute(r.Context(), rpgsystemapp.DeleteRPGSystemInput{
		ID: rpgSystemID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


