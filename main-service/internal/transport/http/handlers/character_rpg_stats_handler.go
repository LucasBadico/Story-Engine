package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// CharacterRPGStatsHandler handles HTTP requests for character RPG stats
type CharacterRPGStatsHandler struct {
	createStatsUseCase      *characterstatsapp.CreateCharacterStatsUseCase
	getActiveStatsUseCase   *characterstatsapp.GetActiveCharacterStatsUseCase
	listHistoryUseCase      *characterstatsapp.ListCharacterStatsHistoryUseCase
	activateVersionUseCase  *characterstatsapp.ActivateCharacterStatsVersionUseCase
	deleteAllStatsUseCase   *characterstatsapp.DeleteAllCharacterStatsUseCase
	logger                   logger.Logger
}

// NewCharacterRPGStatsHandler creates a new CharacterRPGStatsHandler
func NewCharacterRPGStatsHandler(
	createStatsUseCase *characterstatsapp.CreateCharacterStatsUseCase,
	getActiveStatsUseCase *characterstatsapp.GetActiveCharacterStatsUseCase,
	listHistoryUseCase *characterstatsapp.ListCharacterStatsHistoryUseCase,
	activateVersionUseCase *characterstatsapp.ActivateCharacterStatsVersionUseCase,
	deleteAllStatsUseCase *characterstatsapp.DeleteAllCharacterStatsUseCase,
	logger logger.Logger,
) *CharacterRPGStatsHandler {
	return &CharacterRPGStatsHandler{
		createStatsUseCase:     createStatsUseCase,
		getActiveStatsUseCase:  getActiveStatsUseCase,
		listHistoryUseCase:     listHistoryUseCase,
		activateVersionUseCase: activateVersionUseCase,
		deleteAllStatsUseCase:  deleteAllStatsUseCase,
		logger:                 logger,
	}
}

// GetActive handles GET /api/v1/characters/{id}/rpg-stats
func (h *CharacterRPGStatsHandler) GetActive(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.getActiveStatsUseCase.Execute(r.Context(), characterstatsapp.GetActiveCharacterStatsInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": output.Stats,
	})
}

// Create handles POST /api/v1/characters/{id}/rpg-stats
func (h *CharacterRPGStatsHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		EventID             *string          `json:"event_id,omitempty"`
		BaseStats           json.RawMessage  `json:"base_stats"`
		DerivedStats        *json.RawMessage `json:"derived_stats,omitempty"`
		Progression         *json.RawMessage `json:"progression,omitempty"`
		Reason              *string          `json:"reason,omitempty"`
		Timeline            *string          `json:"timeline,omitempty"`
		DeactivatePrevious  bool             `json:"deactivate_previous"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var eventID *uuid.UUID
	if req.EventID != nil && *req.EventID != "" {
		parsedEventID, err := uuid.Parse(*req.EventID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "event_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		eventID = &parsedEventID
	}

	output, err := h.createStatsUseCase.Execute(r.Context(), characterstatsapp.CreateCharacterStatsInput{
		TenantID:          tenantID,
		CharacterID:       characterID,
		EventID:            eventID,
		BaseStats:          req.BaseStats,
		DerivedStats:       req.DerivedStats,
		Progression:        req.Progression,
		Reason:             req.Reason,
		Timeline:           req.Timeline,
		DeactivatePrevious: req.DeactivatePrevious,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": output.Stats,
	})
}

// ListHistory handles GET /api/v1/characters/{id}/rpg-stats/history
func (h *CharacterRPGStatsHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listHistoryUseCase.Execute(r.Context(), characterstatsapp.ListCharacterStatsHistoryInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": output.Stats,
		"total": len(output.Stats),
	})
}

// ActivateVersion handles PUT /api/v1/characters/{id}/rpg-stats/{stats_id}/activate
func (h *CharacterRPGStatsHandler) ActivateVersion(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	statsIDStr := r.PathValue("stats_id")
	statsID, err := uuid.Parse(statsIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "stats_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.activateVersionUseCase.Execute(r.Context(), characterstatsapp.ActivateCharacterStatsVersionInput{
		TenantID: tenantID,
		StatsID:  statsID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": output.Stats,
	})
}

// DeleteAll handles DELETE /api/v1/characters/{id}/rpg-stats
func (h *CharacterRPGStatsHandler) DeleteAll(w http.ResponseWriter, r *http.Request) {
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

	if err := h.deleteAllStatsUseCase.Execute(r.Context(), characterstatsapp.DeleteAllCharacterStatsInput{
		TenantID:    tenantID,
		CharacterID: characterID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


