package entity_extraction

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/search"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type Phase7RelationMatchUseCase struct {
	searcher *search.SearchMemoryUseCase
	logger   *logger.Logger
}

func NewPhase7RelationMatchUseCase(searcher *search.SearchMemoryUseCase, logger *logger.Logger) *Phase7RelationMatchUseCase {
	return &Phase7RelationMatchUseCase{
		searcher: searcher,
		logger:   logger,
	}
}

type Phase7RelationMatchInput struct {
	TenantID      uuid.UUID
	Relations     []Phase6NormalizedRelation
	SourceTypes   []memory.SourceType
	MaxMatches    int
	MinSimilarity float64
}

type Phase7RelationMatchOutput struct {
	Matches []Phase7RelationMatch
}

type Phase7RelationMatch struct {
	RelationIndex int
	RelationKey   string
	Matches       []Phase7RelationMatchCandidate
}

type Phase7RelationMatchCandidate struct {
	ChunkID    uuid.UUID         `json:"chunk_id"`
	DocumentID uuid.UUID         `json:"document_id"`
	SourceType memory.SourceType `json:"source_type"`
	SourceID   uuid.UUID         `json:"source_id"`
	Content    string            `json:"content"`
	Similarity float64           `json:"similarity"`
}

func (uc *Phase7RelationMatchUseCase) Execute(ctx context.Context, input Phase7RelationMatchInput) (Phase7RelationMatchOutput, error) {
	if input.TenantID == uuid.Nil {
		return Phase7RelationMatchOutput{}, fmt.Errorf("tenant id is required")
	}
	if uc.searcher == nil {
		return Phase7RelationMatchOutput{}, fmt.Errorf("search use case is required")
	}
	if len(input.Relations) == 0 {
		return Phase7RelationMatchOutput{Matches: []Phase7RelationMatch{}}, nil
	}

	limit := input.MaxMatches
	if limit <= 0 {
		limit = 5
	}
	minSimilarity := input.MinSimilarity
	if minSimilarity < 0 {
		minSimilarity = 0
	}

	output := Phase7RelationMatchOutput{
		Matches: make([]Phase7RelationMatch, 0, len(input.Relations)),
	}

	for idx, relation := range input.Relations {
		query := buildRelationMatchQuery(relation)
		if query == "" {
			continue
		}

		searchOutput, err := uc.searcher.Execute(ctx, search.SearchMemoryInput{
			TenantID:    input.TenantID,
			Query:       query,
			Limit:       limit,
			SourceTypes: input.SourceTypes,
		})
		if err != nil {
			if uc.logger != nil {
				uc.logger.Error("relation match search failed", "relation_index", idx, "error", err)
			}
			continue
		}

		candidates := make([]Phase7RelationMatchCandidate, 0, len(searchOutput.Chunks))
		for _, chunk := range searchOutput.Chunks {
			if chunk == nil {
				continue
			}
			if minSimilarity > 0 && chunk.Similarity < minSimilarity {
				continue
			}
			candidates = append(candidates, Phase7RelationMatchCandidate{
				ChunkID:    chunk.ChunkID,
				DocumentID: chunk.DocumentID,
				SourceType: chunk.SourceType,
				SourceID:   chunk.SourceID,
				Content:    chunk.Content,
				Similarity: chunk.Similarity,
			})
		}

		output.Matches = append(output.Matches, Phase7RelationMatch{
			RelationIndex: idx,
			RelationKey:   phase6RelationKey(relation),
			Matches:       candidates,
		})
	}

	return output, nil
}

func buildRelationMatchQuery(relation Phase6NormalizedRelation) string {
	summary := strings.TrimSpace(relation.Summary)
	if summary == "" {
		sourceName := displayName(relation.Source)
		targetName := displayName(relation.Target)
		relationType := strings.TrimSpace(relation.RelationType)
		if relationType == "" {
			relationType = "related_to"
		}
		summary = strings.TrimSpace(fmt.Sprintf("%s %s %s", sourceName, relationType, targetName))
	}

	quote := strings.TrimSpace(relation.Evidence.Quote)
	if quote != "" {
		return strings.TrimSpace(fmt.Sprintf("%s\nEvidence: %s", summary, quote))
	}

	return summary
}
