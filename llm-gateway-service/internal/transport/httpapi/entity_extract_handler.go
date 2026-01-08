package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	Entities []entityExtractEntity `json:"entities"`
}

type entityExtractEntity struct {
	Type       string                             `json:"type"`
	Name       string                             `json:"name"`
	Summary    string                             `json:"summary,omitempty"`
	Found      bool                               `json:"found"`
	Match      *entity_extraction.PhaseTempMatch  `json:"match,omitempty"`
	Candidates []entity_extraction.PhaseTempMatch `json:"candidates,omitempty"`
}

func (h *EntityExtractHandler) Extract(w http.ResponseWriter, r *http.Request) {
	tenantID, err := extractTenantID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req entityExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	if req.WorldID == "" {
		writeError(w, http.StatusBadRequest, "world_id is required")
		return
	}
	worldID, err := uuid.Parse(req.WorldID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "world_id must be a valid UUID")
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
		Entities: make([]entityExtractEntity, 0, len(output.Payload.Entities)),
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
