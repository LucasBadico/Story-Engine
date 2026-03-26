package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	plotapp "github.com/story-engine/main-service/internal/application/story/plot"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// PlotHandler handles HTTP requests for plots
type PlotHandler struct {
	createPlotUseCase            *plotapp.CreatePlotUseCase
	getPlotUseCase               *plotapp.GetPlotUseCase
	listPlotsUseCase             *plotapp.ListPlotsUseCase
	updatePlotUseCase            *plotapp.UpdatePlotUseCase
	deletePlotUseCase            *plotapp.DeletePlotUseCase
	addPlotEntityUseCase         *plotapp.AddPlotEntityUseCase
	createPlotDevelopmentUseCase *plotapp.CreatePlotDevelopmentUseCase
	listPlotDevelopmentsUseCase  *plotapp.ListPlotDevelopmentsUseCase
	logger                       logger.Logger
}

// NewPlotHandler creates a new PlotHandler
func NewPlotHandler(
	createPlotUseCase *plotapp.CreatePlotUseCase,
	getPlotUseCase *plotapp.GetPlotUseCase,
	listPlotsUseCase *plotapp.ListPlotsUseCase,
	updatePlotUseCase *plotapp.UpdatePlotUseCase,
	deletePlotUseCase *plotapp.DeletePlotUseCase,
	addPlotEntityUseCase *plotapp.AddPlotEntityUseCase,
	createPlotDevelopmentUseCase *plotapp.CreatePlotDevelopmentUseCase,
	listPlotDevelopmentsUseCase *plotapp.ListPlotDevelopmentsUseCase,
	logger logger.Logger,
) *PlotHandler {
	return &PlotHandler{
		createPlotUseCase:            createPlotUseCase,
		getPlotUseCase:               getPlotUseCase,
		listPlotsUseCase:             listPlotsUseCase,
		updatePlotUseCase:            updatePlotUseCase,
		deletePlotUseCase:            deletePlotUseCase,
		addPlotEntityUseCase:         addPlotEntityUseCase,
		createPlotDevelopmentUseCase: createPlotDevelopmentUseCase,
		listPlotDevelopmentsUseCase:  listPlotDevelopmentsUseCase,
		logger:                       logger,
	}
}

// Create handles POST /api/v1/stories/{story_id}/plots
func (h *PlotHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	storyIDStr := r.PathValue("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "story_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Entities    []struct {
			Role       string `json:"role"`
			EntityType string `json:"entity_type"`
			EntityID   string `json:"entity_id"`
		} `json:"entities"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var entities []plotapp.EntityLinkInput
	for _, e := range req.Entities {
		eid, err := uuid.Parse(e.EntityID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "entities.entity_id",
				Message: "invalid UUID format: " + e.EntityID,
			}, http.StatusBadRequest)
			return
		}
		entities = append(entities, plotapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   eid,
		})
	}

	output, err := h.createPlotUseCase.Execute(r.Context(), plotapp.CreatePlotInput{
		TenantID:    tenantID,
		StoryID:     storyID,
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		Entities:    entities,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot": output.Plot,
	})
}

// Get handles GET /api/v1/plots/{id}
func (h *PlotHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getPlotUseCase.Execute(r.Context(), plotapp.GetPlotInput{
		TenantID: tenantID,
		ID:       plotID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot": output.Plot,
	})
}

// ListByStory handles GET /api/v1/stories/{story_id}/plots?status=open
func (h *PlotHandler) ListByStory(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	storyIDStr := r.PathValue("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "story_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")

	output, err := h.listPlotsUseCase.Execute(r.Context(), plotapp.ListPlotsInput{
		TenantID: tenantID,
		StoryID:  storyID,
		Status:   status,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plots": output.Plots,
	})
}

// Update handles PUT /api/v1/plots/{id}
func (h *PlotHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Type        *string `json:"type"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.updatePlotUseCase.Execute(r.Context(), plotapp.UpdatePlotInput{
		TenantID:    tenantID,
		ID:          plotID,
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot": output.Plot,
	})
}

// Delete handles DELETE /api/v1/plots/{id}
func (h *PlotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deletePlotUseCase.Execute(r.Context(), plotapp.DeletePlotInput{
		TenantID: tenantID,
		ID:       plotID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEntity handles POST /api/v1/plots/{id}/entities
func (h *PlotHandler) AddEntity(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Role       string `json:"role"`
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.addPlotEntityUseCase.Execute(r.Context(), plotapp.AddPlotEntityInput{
		TenantID:   tenantID,
		PlotID:     plotID,
		Role:       req.Role,
		EntityType: req.EntityType,
		EntityID:   entityID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot_entity": output.PlotEntity,
	})
}

// CreateDevelopment handles POST /api/v1/plots/{id}/developments
func (h *PlotHandler) CreateDevelopment(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ChapterNumber      int     `json:"chapter_number"`
		Action             string  `json:"action"`
		Description        string  `json:"description"`
		Notes              string  `json:"notes"`
		TimelineDay        *int    `json:"timeline_day"`
		TimelineConfidence *string `json:"timeline_confidence"`
		TimelineEventID    *string `json:"timeline_event_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var timelineEventID *uuid.UUID
	if req.TimelineEventID != nil && *req.TimelineEventID != "" {
		eid, err := uuid.Parse(*req.TimelineEventID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "timeline_event_id",
				Message: "invalid UUID format: " + *req.TimelineEventID,
			}, http.StatusBadRequest)
			return
		}
		timelineEventID = &eid
	}

	output, err := h.createPlotDevelopmentUseCase.Execute(r.Context(), plotapp.CreatePlotDevelopmentInput{
		TenantID:           tenantID,
		PlotID:             plotID,
		ChapterNumber:      req.ChapterNumber,
		Action:             req.Action,
		Description:        req.Description,
		Notes:              req.Notes,
		TimelineDay:        req.TimelineDay,
		TimelineConfidence: req.TimelineConfidence,
		TimelineEventID:    timelineEventID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot_development": output.PlotDevelopment,
	})
}

// ListDevelopments handles GET /api/v1/plots/{id}/developments?as_of_chapter=N
func (h *PlotHandler) ListDevelopments(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	plotID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var asOfChapter *int
	asOfChapterStr := r.URL.Query().Get("as_of_chapter")
	if asOfChapterStr != "" {
		if n, err := strconv.Atoi(asOfChapterStr); err == nil && n > 0 {
			asOfChapter = &n
		}
	}

	output, err := h.listPlotDevelopmentsUseCase.Execute(r.Context(), plotapp.ListPlotDevelopmentsInput{
		TenantID:    tenantID,
		PlotID:      plotID,
		AsOfChapter: asOfChapter,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plot_developments": output.PlotDevelopments,
	})
}
