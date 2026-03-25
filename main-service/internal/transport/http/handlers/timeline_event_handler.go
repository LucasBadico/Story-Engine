package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	timelineeventapp "github.com/story-engine/main-service/internal/application/story/timeline_event"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// TimelineEventHandler handles HTTP requests for timeline events
type TimelineEventHandler struct {
	createTimelineEventUseCase         *timelineeventapp.CreateTimelineEventUseCase
	getTimelineEventUseCase            *timelineeventapp.GetTimelineEventUseCase
	listTimelineEventsByStoryUseCase   *timelineeventapp.ListTimelineEventsByStoryUseCase
	listTimelineEventsByChapterUseCase *timelineeventapp.ListTimelineEventsByChapterUseCase
	deleteTimelineEventUseCase         *timelineeventapp.DeleteTimelineEventUseCase
	logger                             logger.Logger
}

// NewTimelineEventHandler creates a new TimelineEventHandler
func NewTimelineEventHandler(
	createTimelineEventUseCase *timelineeventapp.CreateTimelineEventUseCase,
	getTimelineEventUseCase *timelineeventapp.GetTimelineEventUseCase,
	listTimelineEventsByStoryUseCase *timelineeventapp.ListTimelineEventsByStoryUseCase,
	listTimelineEventsByChapterUseCase *timelineeventapp.ListTimelineEventsByChapterUseCase,
	deleteTimelineEventUseCase *timelineeventapp.DeleteTimelineEventUseCase,
	logger logger.Logger,
) *TimelineEventHandler {
	return &TimelineEventHandler{
		createTimelineEventUseCase:         createTimelineEventUseCase,
		getTimelineEventUseCase:            getTimelineEventUseCase,
		listTimelineEventsByStoryUseCase:   listTimelineEventsByStoryUseCase,
		listTimelineEventsByChapterUseCase: listTimelineEventsByChapterUseCase,
		deleteTimelineEventUseCase:         deleteTimelineEventUseCase,
		logger:                             logger,
	}
}

// Create handles POST /api/v1/stories/{story_id}/timeline-events
func (h *TimelineEventHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		ChapterID          string  `json:"chapter_id"`
		ChapterNumber      int     `json:"chapter_number"`
		TimelineDay        *int    `json:"timeline_day"`
		TimelineConfidence *string `json:"timeline_confidence"`
		Description        string  `json:"description"`
		Entities           []struct {
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

	chapterID, err := uuid.Parse(req.ChapterID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "chapter_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var entities []timelineeventapp.EntityLinkInput
	for _, e := range req.Entities {
		eid, err := uuid.Parse(e.EntityID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "entities.entity_id",
				Message: "invalid UUID format: " + e.EntityID,
			}, http.StatusBadRequest)
			return
		}
		entities = append(entities, timelineeventapp.EntityLinkInput{
			Role:       e.Role,
			EntityType: e.EntityType,
			EntityID:   eid,
		})
	}

	output, err := h.createTimelineEventUseCase.Execute(r.Context(), timelineeventapp.CreateTimelineEventInput{
		TenantID:           tenantID,
		StoryID:            storyID,
		ChapterID:          chapterID,
		ChapterNumber:      req.ChapterNumber,
		TimelineDay:        req.TimelineDay,
		TimelineConfidence: req.TimelineConfidence,
		Description:        req.Description,
		Entities:           entities,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline_event": output.TimelineEvent,
	})
}

// Get handles GET /api/v1/timeline-events/{id}
func (h *TimelineEventHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	teID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getTimelineEventUseCase.Execute(r.Context(), timelineeventapp.GetTimelineEventInput{
		TenantID: tenantID,
		ID:       teID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline_event": output.TimelineEvent,
	})
}

// ListByStory handles GET /api/v1/stories/{story_id}/timeline-events?as_of_chapter=N&limit=50
func (h *TimelineEventHandler) ListByStory(w http.ResponseWriter, r *http.Request) {
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

	asOfChapterStr := r.URL.Query().Get("as_of_chapter")
	asOfChapter := 0
	if asOfChapterStr != "" {
		if n, err := strconv.Atoi(asOfChapterStr); err == nil && n > 0 {
			asOfChapter = n
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	output, err := h.listTimelineEventsByStoryUseCase.Execute(r.Context(), timelineeventapp.ListTimelineEventsByStoryInput{
		TenantID:    tenantID,
		StoryID:     storyID,
		AsOfChapter: asOfChapter,
		Limit:       limit,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline_events": output.TimelineEvents,
	})
}

// ListByChapter handles GET /api/v1/chapters/{chapter_id}/timeline-events
func (h *TimelineEventHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	chapterIDStr := r.PathValue("chapter_id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "chapter_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listTimelineEventsByChapterUseCase.Execute(r.Context(), timelineeventapp.ListTimelineEventsByChapterInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline_events": output.TimelineEvents,
	})
}

// Delete handles DELETE /api/v1/timeline-events/{id}
func (h *TimelineEventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	teID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteTimelineEventUseCase.Execute(r.Context(), timelineeventapp.DeleteTimelineEventInput{
		TenantID: tenantID,
		ID:       teID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
