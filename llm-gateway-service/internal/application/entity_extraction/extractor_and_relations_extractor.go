package entity_extraction

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type EntityAndRelationshipsExtractor struct {
	router    *Phase1EntityTypeRouterUseCase
	extractor *Phase2EntryUseCase
	matcher   *Phase3MatchUseCase
	payload   *PhaseTempPayloadUseCase
	logger    *logger.Logger
}

func NewEntityAndRelationshipsExtractor(
	router *Phase1EntityTypeRouterUseCase,
	extractor *Phase2EntryUseCase,
	matcher *Phase3MatchUseCase,
	payload *PhaseTempPayloadUseCase,
	logger *logger.Logger,
) *EntityAndRelationshipsExtractor {
	return &EntityAndRelationshipsExtractor{
		router:    router,
		extractor: extractor,
		matcher:   matcher,
		payload:   payload,
		logger:    logger,
	}
}

type EntityAndRelationshipsExtractorInput struct {
	TenantID              uuid.UUID
	WorldID               uuid.UUID
	Text                  string
	Context               string
	EntityTypes           []string
	MaxChunkChars         int
	OverlapChars          int
	MaxTypeCandidates     int
	MaxCandidatesPerChunk int
	MinSimilarity         float64
	MaxMatchCandidates    int
}

type EntityAndRelationshipsExtractorOutput struct {
	Payload PhaseTempPayload
}

func (u *EntityAndRelationshipsExtractor) Execute(ctx context.Context, input EntityAndRelationshipsExtractorInput) (EntityAndRelationshipsExtractorOutput, error) {
	if input.TenantID == uuid.Nil {
		return EntityAndRelationshipsExtractorOutput{}, errors.New("tenant id is required")
	}
	if input.WorldID == uuid.Nil {
		return EntityAndRelationshipsExtractorOutput{}, errors.New("world id is required")
	}
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return EntityAndRelationshipsExtractorOutput{}, errors.New("text is required")
	}
	if u.router == nil || u.extractor == nil || u.matcher == nil || u.payload == nil {
		return EntityAndRelationshipsExtractorOutput{}, errors.New("router, extractor, matcher, and payload are required")
	}

	entityTypes := input.EntityTypes
	if len(entityTypes) == 0 {
		entityTypes = []string{"character", "location", "artefact", "faction"}
	}

	split, err := SplitTextIntoParagraphChunks(Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: input.MaxChunkChars,
		OverlapChars:  input.OverlapChars,
	})
	if err != nil {
		return EntityAndRelationshipsExtractorOutput{}, err
	}

	routedChunks := make([]Phase2RoutedChunk, 0)
	for _, paragraph := range split.Paragraphs {
		for _, chunk := range paragraph.Chunks {
			chunkText := strings.TrimSpace(chunk.Text)
			if chunkText == "" {
				continue
			}

			types, err := u.routeChunkTypes(ctx, chunkText, input.Context, entityTypes, input.MaxTypeCandidates)
			if err != nil {
				u.logger.Error("phase1 routing failed", "error", err)
				continue
			}
			if len(types) == 0 {
				continue
			}

			routedChunks = append(routedChunks, Phase2RoutedChunk{
				ParagraphID: paragraph.ParagraphID,
				ChunkID:     chunk.ChunkID,
				StartOffset: chunk.StartOffset,
				EndOffset:   chunk.EndOffset,
				Text:        chunkText,
				Types:       types,
			})
		}
	}

	if len(routedChunks) == 0 {
		return EntityAndRelationshipsExtractorOutput{Payload: PhaseTempPayload{Entities: []PhaseTempEntity{}}}, nil
	}

	phase2Output, err := u.extractor.Execute(ctx, Phase2EntryInput{
		Context:               input.Context,
		MaxCandidatesPerChunk: input.MaxCandidatesPerChunk,
		Chunks:                routedChunks,
	})
	if err != nil {
		return EntityAndRelationshipsExtractorOutput{}, err
	}

	phase3Output, err := u.matcher.Execute(ctx, Phase3MatchInput{
		TenantID:      input.TenantID,
		WorldID:       &input.WorldID,
		Findings:      phase2Output.Findings,
		Context:       input.Context,
		MinSimilarity: input.MinSimilarity,
		MaxCandidates: input.MaxMatchCandidates,
	})
	if err != nil {
		return EntityAndRelationshipsExtractorOutput{}, err
	}

	return EntityAndRelationshipsExtractorOutput{
		Payload: u.payload.Execute(phase3Output),
	}, nil
}

func (u *EntityAndRelationshipsExtractor) routeChunkTypes(
	ctx context.Context,
	text string,
	context string,
	entityTypes []string,
	maxCandidates int,
) ([]string, error) {
	output, err := u.router.Execute(ctx, Phase1EntityTypeRouterInput{
		Text:          text,
		Context:       context,
		EntityTypes:   entityTypes,
		MaxCandidates: maxCandidates,
	})
	if err != nil {
		return nil, err
	}

	types := make([]string, 0, len(output.Candidates))
	seen := map[string]struct{}{}
	for _, candidate := range output.Candidates {
		entityType := strings.TrimSpace(candidate.Type)
		if entityType == "" {
			continue
		}
		if _, ok := seen[entityType]; ok {
			continue
		}
		seen[entityType] = struct{}{}
		types = append(types, entityType)
	}

	return types, nil
}
