package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	characterstateapp "github.com/story-engine/main-service/internal/application/world/character_state"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// CharacterStateHandler handles HTTP requests for character states
type CharacterStateHandler struct {
	createCharacterStateUseCase    *characterstateapp.CreateCharacterStateUseCase
	getCharacterStateUseCase       *characterstateapp.GetCharacterStateUseCase
	getLatestCharacterStateUseCase *characterstateapp.GetLatestCharacterStateUseCase
	listCharacterStatesUseCase     *characterstateapp.ListCharacterStatesUseCase
	deleteCharacterStateUseCase    *characterstateapp.DeleteCharacterStateUseCase
	logger                         logger.Logger
}

// NewCharacterStateHandler creates a new CharacterStateHandler
func NewCharacterStateHandler(
	createCharacterStateUseCase *characterstateapp.CreateCharacterStateUseCase,
	getCharacterStateUseCase *characterstateapp.GetCharacterStateUseCase,
	getLatestCharacterStateUseCase *characterstateapp.GetLatestCharacterStateUseCase,
	listCharacterStatesUseCase *characterstateapp.ListCharacterStatesUseCase,
	deleteCharacterStateUseCase *characterstateapp.DeleteCharacterStateUseCase,
	logger logger.Logger,
) *CharacterStateHandler {
	return &CharacterStateHandler{
		createCharacterStateUseCase:    createCharacterStateUseCase,
		getCharacterStateUseCase:       getCharacterStateUseCase,
		getLatestCharacterStateUseCase: getLatestCharacterStateUseCase,
		listCharacterStatesUseCase:     listCharacterStatesUseCase,
		deleteCharacterStateUseCase:    deleteCharacterStateUseCase,
		logger:                         logger,
	}
}

// Create handles POST /api/v1/characters/{character_id}/states
func (h *CharacterStateHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	characterIDStr := r.PathValue("character_id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "character_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ChapterID          *string  `json:"chapter_id"`
		ChapterNumber      int      `json:"chapter_number"`
		Emotion            string   `json:"emotion"`
		Goal               string   `json:"goal"`
		Secrets            []string `json:"secrets"`
		Clothes            string   `json:"clothes"`
		PhysicalState      string   `json:"physical_state"`
		TimelineDay        *int     `json:"timeline_day"`
		TimelineConfidence *string  `json:"timeline_confidence"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var chapterID *uuid.UUID
	if req.ChapterID != nil && *req.ChapterID != "" {
		cid, err := uuid.Parse(*req.ChapterID)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "chapter_id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		chapterID = &cid
	}

	output, err := h.createCharacterStateUseCase.Execute(r.Context(), characterstateapp.CreateCharacterStateInput{
		TenantID:           tenantID,
		CharacterID:        characterID,
		ChapterID:          chapterID,
		ChapterNumber:      req.ChapterNumber,
		Emotion:            req.Emotion,
		Goal:               req.Goal,
		Secrets:            req.Secrets,
		Clothes:            req.Clothes,
		PhysicalState:      req.PhysicalState,
		TimelineDay:        req.TimelineDay,
		TimelineConfidence: req.TimelineConfidence,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_state": output.CharacterState,
	})
}

// Get handles GET /api/v1/character-states/{id}
func (h *CharacterStateHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	csID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getCharacterStateUseCase.Execute(r.Context(), characterstateapp.GetCharacterStateInput{
		TenantID: tenantID,
		ID:       csID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_state": output.CharacterState,
	})
}

// GetLatest handles GET /api/v1/characters/{character_id}/states/latest?as_of_chapter=N
func (h *CharacterStateHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	characterIDStr := r.PathValue("character_id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "character_id",
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

	output, err := h.getLatestCharacterStateUseCase.Execute(r.Context(), characterstateapp.GetLatestCharacterStateInput{
		TenantID:    tenantID,
		CharacterID: characterID,
		AsOfChapter: asOfChapter,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_state": output.CharacterState,
	})
}

// ListByChapter handles GET /api/v1/chapters/{chapter_id}/character-states
func (h *CharacterStateHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
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

	output, err := h.listCharacterStatesUseCase.Execute(r.Context(), characterstateapp.ListCharacterStatesInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_states": output.CharacterStates,
	})
}

// Delete handles DELETE /api/v1/character-states/{id}
func (h *CharacterStateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	csID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	err = h.deleteCharacterStateUseCase.Execute(r.Context(), characterstateapp.DeleteCharacterStateInput{
		TenantID: tenantID,
		ID:       csID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
