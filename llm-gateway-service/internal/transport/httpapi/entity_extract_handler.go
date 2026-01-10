package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type EntityExtractHandler struct {
	useCase *entity_extraction.EntityAndRelationshipsExtractor
	logger  *logger.Logger
}

func NewEntityExtractHandler(useCase *entity_extraction.EntityAndRelationshipsExtractor, logger *logger.Logger) *EntityExtractHandler {
	return &EntityExtractHandler{
		useCase: useCase,
		logger:  logger,
	}
}

type entityExtractRequest struct {
	Text                  string   `json:"text"`
	WorldID               string   `json:"world_id"`
	Context               string   `json:"context,omitempty"`
	EntityTypes           []string `json:"entity_types,omitempty"`
	MaxChunkChars         int      `json:"max_chunk_chars,omitempty"`
	OverlapChars          int      `json:"overlap_chars,omitempty"`
	MaxTypeCandidates     int      `json:"max_type_candidates,omitempty"`
	MaxCandidatesPerChunk int      `json:"max_candidates_per_chunk,omitempty"`
	MinSimilarity         float64  `json:"min_similarity,omitempty"`
	MaxMatchCandidates    int      `json:"max_match_candidates,omitempty"`
}

type entityExtractResponse struct {
	Entities  []entityExtractEntity              `json:"entities"`
	Relations []entity_extraction.Phase4Relation `json:"relations"`
}

type entityExtractEntity struct {
	Type       string                          `json:"type"`
	Name       string                          `json:"name"`
	Summary    string                          `json:"summary,omitempty"`
	Found      bool                            `json:"found"`
	Match      *entity_extraction.Phase4Match  `json:"match,omitempty"`
	Candidates []entity_extraction.Phase4Match `json:"candidates,omitempty"`
}

func (h *EntityExtractHandler) Extract(w http.ResponseWriter, r *http.Request) {
	tenantID, worldID, req, err := decodeEntityExtractRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info(formatLogBlock("entity extract request", []string{
		fmt.Sprintf("tenant_id: %s", tenantID),
		fmt.Sprintf("world_id: %s", worldID),
		fmt.Sprintf("text_len: %d", len(req.Text)),
		fmt.Sprintf("context_len: %d", len(req.Context)),
		fmt.Sprintf("entity_types: %v", req.EntityTypes),
		fmt.Sprintf("max_chunk_chars: %d", req.MaxChunkChars),
		fmt.Sprintf("overlap_chars: %d", req.OverlapChars),
		fmt.Sprintf("max_type_candidates: %d", req.MaxTypeCandidates),
		fmt.Sprintf("max_candidates_per_chunk: %d", req.MaxCandidatesPerChunk),
		fmt.Sprintf("min_similarity: %g", req.MinSimilarity),
		fmt.Sprintf("max_match_candidates: %d", req.MaxMatchCandidates),
	}))

	output, err := h.useCase.Execute(r.Context(), entity_extraction.EntityAndRelationshipsExtractorInput{
		TenantID:              tenantID,
		WorldID:               worldID,
		Text:                  req.Text,
		Context:               req.Context,
		EntityTypes:           req.EntityTypes,
		MaxChunkChars:         req.MaxChunkChars,
		OverlapChars:          req.OverlapChars,
		MaxTypeCandidates:     req.MaxTypeCandidates,
		MaxCandidatesPerChunk: req.MaxCandidatesPerChunk,
		MinSimilarity:         req.MinSimilarity,
		MaxMatchCandidates:    req.MaxMatchCandidates,
	})
	if err != nil {
		h.logger.Error(fmt.Sprintf("entity extract failed tenant_id=%s error=%v", tenantID, err))
		writeError(w, http.StatusInternalServerError, "entity extraction failed")
		return
	}

	resp := entityExtractResponse{
		Entities:  make([]entityExtractEntity, 0, len(output.Payload.Entities)),
		Relations: []entity_extraction.Phase4Relation{},
	}

	foundCount := 0
	for _, entity := range output.Payload.Entities {
		if entity.Found {
			foundCount++
		}
		resp.Entities = append(resp.Entities, entityExtractEntity{
			Type:       entity.Type,
			Name:       entity.Name,
			Summary:    entity.Summary,
			Found:      entity.Found,
			Match:      entity.Match,
			Candidates: entity.Candidates,
		})
	}

	entitySummaries := make([]string, 0, len(resp.Entities))
	for _, entity := range resp.Entities {
		entitySummaries = append(entitySummaries, fmt.Sprintf("%s:%s:%t", entity.Type, entity.Name, entity.Found))
	}

	h.logger.Info(formatLogBlock("entity extract response", []string{
		fmt.Sprintf("tenant_id: %s", tenantID),
		fmt.Sprintf("world_id: %s", worldID),
		fmt.Sprintf("entities_count: %d", len(resp.Entities)),
		fmt.Sprintf("found_count: %d", foundCount),
		fmt.Sprintf("entities: %v", entitySummaries),
	}))

	writeJSON(w, http.StatusOK, resp)
}

func (h *EntityExtractHandler) ExtractStream(w http.ResponseWriter, r *http.Request) {
	tenantID, worldID, req, err := decodeEntityExtractRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	eventLogger := &sseEventLogger{
		writer:  w,
		flusher: flusher,
	}

	emitEvent(r.Context(), eventLogger, entity_extraction.ExtractionEvent{
		Type:    "request.received",
		Message: "request received",
		Data: map[string]interface{}{
			"tenant_id": tenantID.String(),
			"world_id":  worldID.String(),
			"text_len":  len(req.Text),
		},
		Timestamp: time.Now().UTC(),
	})

	output, err := h.useCase.Execute(r.Context(), entity_extraction.EntityAndRelationshipsExtractorInput{
		TenantID:              tenantID,
		WorldID:               worldID,
		Text:                  req.Text,
		Context:               req.Context,
		EntityTypes:           req.EntityTypes,
		MaxChunkChars:         req.MaxChunkChars,
		OverlapChars:          req.OverlapChars,
		MaxTypeCandidates:     req.MaxTypeCandidates,
		MaxCandidatesPerChunk: req.MaxCandidatesPerChunk,
		MinSimilarity:         req.MinSimilarity,
		MaxMatchCandidates:    req.MaxMatchCandidates,
		EventLogger:           eventLogger,
	})
	if err != nil {
		emitEvent(r.Context(), eventLogger, entity_extraction.ExtractionEvent{
			Type:    "error",
			Message: "entity extraction failed",
			Data: map[string]interface{}{
				"error": err.Error(),
			},
			Timestamp: time.Now().UTC(),
		})
		return
	}

	emitEvent(r.Context(), eventLogger, entity_extraction.ExtractionEvent{
		Type:    "result_entities",
		Message: "entity extraction completed",
		Data: map[string]interface{}{
			"entities": output.Payload.Entities,
		},
		Timestamp: time.Now().UTC(),
	})
}

func formatLogBlock(title string, lines []string) string {
	var builder strings.Builder
	builder.WriteString("====\n")
	builder.WriteString(title)
	builder.WriteString("\n====\n")
	for i, line := range lines {
		builder.WriteString(line)
		if i < len(lines)-1 {
			builder.WriteString("\n---\n")
		}
	}
	builder.WriteString("\n====")
	return builder.String()
}

func decodeEntityExtractRequest(r *http.Request) (uuid.UUID, uuid.UUID, entityExtractRequest, error) {
	tenantID, err := extractTenantID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, entityExtractRequest{}, err
	}

	var req entityExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return uuid.Nil, uuid.Nil, entityExtractRequest{}, fmt.Errorf("invalid JSON body")
	}

	if req.Text == "" {
		return uuid.Nil, uuid.Nil, entityExtractRequest{}, fmt.Errorf("text is required")
	}
	if req.WorldID == "" {
		return uuid.Nil, uuid.Nil, entityExtractRequest{}, fmt.Errorf("world_id is required")
	}
	worldID, err := uuid.Parse(req.WorldID)
	if err != nil {
		return uuid.Nil, uuid.Nil, entityExtractRequest{}, fmt.Errorf("world_id must be a valid UUID")
	}

	return tenantID, worldID, req, nil
}

type sseEventLogger struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
}

func (l *sseEventLogger) Emit(ctx context.Context, event entity_extraction.ExtractionEvent) {
	if ctx.Err() != nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.writer, "event: %s\n", event.Type)
	fmt.Fprintf(l.writer, "data: %s\n\n", payload)
	l.flusher.Flush()
}

func emitEvent(ctx context.Context, logger entity_extraction.ExtractionEventLogger, event entity_extraction.ExtractionEvent) {
	if logger == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	logger.Emit(ctx, event)
}
