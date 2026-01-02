package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// WorldHandler handles HTTP requests for worlds
type WorldHandler struct {
	createWorldUseCase *worldapp.CreateWorldUseCase
	getWorldUseCase    *worldapp.GetWorldUseCase
	listWorldsUseCase  *worldapp.ListWorldsUseCase
	updateWorldUseCase *worldapp.UpdateWorldUseCase
	deleteWorldUseCase *worldapp.DeleteWorldUseCase
	logger             logger.Logger
}

// NewWorldHandler creates a new WorldHandler
func NewWorldHandler(
	createWorldUseCase *worldapp.CreateWorldUseCase,
	getWorldUseCase *worldapp.GetWorldUseCase,
	listWorldsUseCase *worldapp.ListWorldsUseCase,
	updateWorldUseCase *worldapp.UpdateWorldUseCase,
	deleteWorldUseCase *worldapp.DeleteWorldUseCase,
	logger logger.Logger,
) *WorldHandler {
	return &WorldHandler{
		createWorldUseCase: createWorldUseCase,
		getWorldUseCase:    getWorldUseCase,
		listWorldsUseCase:  listWorldsUseCase,
		updateWorldUseCase: updateWorldUseCase,
		deleteWorldUseCase: deleteWorldUseCase,
		logger:             logger,
	}
}

// Create handles POST /api/v1/worlds
func (h *WorldHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Genre       string `json:"genre"`
		IsImplicit  *bool  `json:"is_implicit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	isImplicit := false
	if req.IsImplicit != nil {
		isImplicit = *req.IsImplicit
	}

	output, err := h.createWorldUseCase.Execute(r.Context(), worldapp.CreateWorldInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Genre:       req.Genre,
		IsImplicit:  isImplicit,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"world": output.World,
	})
}

// Get handles GET /api/v1/worlds/{id}
func (h *WorldHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	worldID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getWorldUseCase.Execute(r.Context(), worldapp.GetWorldInput{
		ID: worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"world": output.World,
	})
}

// List handles GET /api/v1/worlds?limit=20&offset=0
func (h *WorldHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

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

	output, err := h.listWorldsUseCase.Execute(r.Context(), worldapp.ListWorldsInput{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"worlds": output.Worlds,
		"total":  output.Total,
	})
}

// Update handles PUT /api/v1/worlds/{id}
func (h *WorldHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	worldID, err := uuid.Parse(id)
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
		Genre       *string `json:"genre"`
		IsImplicit  *bool   `json:"is_implicit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updateWorldUseCase.Execute(r.Context(), worldapp.UpdateWorldInput{
		ID: worldID,
		Name:        req.Name,
		Description: req.Description,
		Genre:       req.Genre,
		IsImplicit:  req.IsImplicit,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"world": output.World,
	})
}

// Delete handles DELETE /api/v1/worlds/{id}
func (h *WorldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	worldID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteWorldUseCase.Execute(r.Context(), worldapp.DeleteWorldInput{
		ID: worldID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


