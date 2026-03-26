package httpapi

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/canon"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type CanonHandler struct {
	useCase *canon.RetrieveCanonUseCase
	logger  *logger.Logger
}

func NewCanonHandler(useCase *canon.RetrieveCanonUseCase, logger *logger.Logger) *CanonHandler {
	return &CanonHandler{useCase: useCase, logger: logger}
}

// Retrieve handles GET /api/v1/stories/{id}/canon
func (h *CanonHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract story ID from path: /api/v1/stories/{id}/canon
	storyIDStr := extractPathParam(r.URL.Path, "/api/v1/stories/", "/canon")
	if storyIDStr == "" {
		writeError(w, http.StatusBadRequest, "story_id is required in path")
		return
	}
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid story_id")
		return
	}

	// Parse query params
	query := r.URL.Query()

	asOfChapter := 0
	if v := query.Get("as_of_chapter"); v != "" {
		asOfChapter, err = strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid as_of_chapter")
			return
		}
	}

	var asOfDay *int
	if v := query.Get("as_of_day"); v != "" {
		day, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid as_of_day")
			return
		}
		asOfDay = &day
	}

	mode := canon.RetrievalMode(query.Get("mode"))
	if mode == "" {
		mode = canon.ModeRevealedNarrative
	}
	if mode != canon.ModeRevealedNarrative && mode != canon.ModeTimelineConsistent && mode != canon.ModeOmniscient {
		writeError(w, http.StatusBadRequest, "invalid mode: must be revealed_narrative, timeline_consistent, or omniscient")
		return
	}

	eventLimit := 100
	if v := query.Get("event_limit"); v != "" {
		eventLimit, err = strconv.Atoi(v)
		if err != nil || eventLimit <= 0 {
			writeError(w, http.StatusBadRequest, "invalid event_limit")
			return
		}
	}

	output, err := h.useCase.Execute(r.Context(), canon.RetrieveCanonInput{
		StoryID:     storyID,
		AsOfChapter: asOfChapter,
		AsOfDay:     asOfDay,
		Mode:        mode,
		EventLimit:  eventLimit,
	})
	if err != nil {
		h.logger.Error("canon retrieval failed", "story_id", storyID, "error", err)
		writeError(w, http.StatusInternalServerError, "canon retrieval failed")
		return
	}

	writeJSON(w, http.StatusOK, output)
}

// extractPathParam extracts a segment between prefix and suffix from a URL path
func extractPathParam(path, prefix, suffix string) string {
	if len(path) <= len(prefix)+len(suffix) {
		return ""
	}
	// Strip prefix
	rest := path[len(prefix):]
	// Find suffix
	for i := 0; i < len(rest); i++ {
		if len(rest[i:]) >= len(suffix) && rest[i:i+len(suffix)] == suffix {
			return rest[:i]
		}
	}
	return ""
}
