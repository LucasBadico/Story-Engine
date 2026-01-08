package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

type SearchHandler struct {
	searchUseCase *search.SearchMemoryUseCase
	logger        *logger.Logger
}

func NewSearchHandler(searchUseCase *search.SearchMemoryUseCase, logger *logger.Logger) *SearchHandler {
	return &SearchHandler{
		searchUseCase: searchUseCase,
		logger:        logger,
	}
}

type searchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

type searchChunkResponse struct {
	ChunkID      uuid.UUID `json:"chunk_id"`
	DocumentID   uuid.UUID `json:"document_id"`
	SourceType   string    `json:"source_type"`
	SourceID     uuid.UUID `json:"source_id"`
	Content      string    `json:"content"`
	Score        float64   `json:"score"`
	BeatType     *string   `json:"beat_type,omitempty"`
	BeatIntent   *string   `json:"beat_intent,omitempty"`
	Characters   []string  `json:"characters,omitempty"`
	LocationName *string   `json:"location_name,omitempty"`
	Timeline     *string   `json:"timeline,omitempty"`
	POVCharacter *string   `json:"pov_character,omitempty"`
	ContentKind  *string   `json:"content_kind,omitempty"`
}

type searchResponse struct {
	Chunks     []searchChunkResponse `json:"chunks"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractTenantID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	var cursor *repositories.SearchCursor
	if req.Cursor != "" {
		decoded, err := decodeCursor(req.Cursor)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		cursor = decoded
	}

	output, err := h.searchUseCase.Execute(r.Context(), search.SearchMemoryInput{
		TenantID: tenantID,
		Query:    req.Query,
		Limit:    req.Limit,
		Cursor:   cursor,
	})
	if err != nil {
		h.logger.Error(fmt.Sprintf("search failed tenant_id=%s error=%v", tenantID, err))
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	resp := searchResponse{
		Chunks: make([]searchChunkResponse, 0, len(output.Chunks)),
	}

	for _, chunk := range output.Chunks {
		resp.Chunks = append(resp.Chunks, searchChunkResponse{
			ChunkID:      chunk.ChunkID,
			DocumentID:   chunk.DocumentID,
			SourceType:   string(chunk.SourceType),
			SourceID:     chunk.SourceID,
			Content:      chunk.Content,
			Score:        chunk.Similarity,
			BeatType:     chunk.BeatType,
			BeatIntent:   chunk.BeatIntent,
			Characters:   chunk.Characters,
			LocationName: chunk.LocationName,
			Timeline:     chunk.Timeline,
			POVCharacter: chunk.POVCharacter,
			ContentKind:  chunk.ContentKind,
		})
	}

	if output.NextCursor != nil {
		resp.NextCursor = encodeCursor(output.NextCursor)
	}

	writeJSON(w, http.StatusOK, resp)
}

func decodeCursor(encoded string) (*repositories.SearchCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	var cursor repositories.SearchCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, errors.New("invalid cursor")
	}

	if cursor.ChunkID == uuid.Nil {
		return nil, errors.New("invalid cursor")
	}

	return &cursor, nil
}

func encodeCursor(cursor *repositories.SearchCursor) string {
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func extractTenantID(r *http.Request) (uuid.UUID, error) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return uuid.Nil, errors.New("X-Tenant-ID is required")
	}
	return uuid.Parse(tenantIDStr)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
