package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ArtifactRPGStatsHandler handles HTTP requests for artifact RPG stats
type ArtifactRPGStatsHandler struct {
	createStatsUseCase     *artifactstatsapp.CreateArtifactStatsUseCase
	getActiveStatsUseCase  *artifactstatsapp.GetActiveArtifactStatsUseCase
	listHistoryUseCase     *artifactstatsapp.ListArtifactStatsHistoryUseCase
	activateVersionUseCase *artifactstatsapp.ActivateArtifactStatsVersionUseCase
	logger                 logger.Logger
}

// NewArtifactRPGStatsHandler creates a new ArtifactRPGStatsHandler
func NewArtifactRPGStatsHandler(
	createStatsUseCase *artifactstatsapp.CreateArtifactStatsUseCase,
	getActiveStatsUseCase *artifactstatsapp.GetActiveArtifactStatsUseCase,
	listHistoryUseCase *artifactstatsapp.ListArtifactStatsHistoryUseCase,
	activateVersionUseCase *artifactstatsapp.ActivateArtifactStatsVersionUseCase,
	logger logger.Logger,
) *ArtifactRPGStatsHandler {
	return &ArtifactRPGStatsHandler{
		createStatsUseCase:     createStatsUseCase,
		getActiveStatsUseCase:  getActiveStatsUseCase,
		listHistoryUseCase:     listHistoryUseCase,
		activateVersionUseCase: activateVersionUseCase,
		logger:                 logger,
	}
}

// GetActive handles GET /api/v1/artifacts/{id}/rpg-stats
func (h *ArtifactRPGStatsHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getActiveStatsUseCase.Execute(r.Context(), artifactstatsapp.GetActiveArtifactStatsInput{
		ArtifactID: artifactID,
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

// Create handles POST /api/v1/artifacts/{id}/rpg-stats
func (h *ArtifactRPGStatsHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
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
		EventID            *string         `json:"event_id,omitempty"`
		Stats              json.RawMessage `json:"stats"`
		Reason             *string          `json:"reason,omitempty"`
		Timeline           *string          `json:"timeline,omitempty"`
		DeactivatePrevious bool            `json:"deactivate_previous"`
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

	output, err := h.createStatsUseCase.Execute(r.Context(), artifactstatsapp.CreateArtifactStatsInput{
		ArtifactID:         artifactID,
		EventID:            eventID,
		Stats:              req.Stats,
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

// ListHistory handles GET /api/v1/artifacts/{id}/rpg-stats/history
func (h *ArtifactRPGStatsHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	artifactID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listHistoryUseCase.Execute(r.Context(), artifactstatsapp.ListArtifactStatsHistoryInput{
		ArtifactID: artifactID,
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

// ActivateVersion handles PUT /api/v1/artifacts/{id}/rpg-stats/{stats_id}/activate
func (h *ArtifactRPGStatsHandler) ActivateVersion(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.activateVersionUseCase.Execute(r.Context(), artifactstatsapp.ActivateArtifactStatsVersionInput{
		StatsID: statsID,
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


