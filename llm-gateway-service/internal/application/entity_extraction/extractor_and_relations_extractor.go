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
	EventLogger           ExtractionEventLogger
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

	eventLogger := normalizeEventLogger(input.EventLogger)
	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "pipeline.start",
		Message: "entity extraction started",
		Data: map[string]interface{}{
			"tenant_id": uuidString(input.TenantID),
			"world_id":  uuidString(input.WorldID),
			"text_len":  len(text),
		},
	})

	entityTypes := input.EntityTypes
	if len(entityTypes) == 0 {
		entityTypes = []string{"character", "location", "artefact", "faction", "event"}
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

			emitEvent(ctx, eventLogger, ExtractionEvent{
				Type:    "router.chunk",
				Phase:   "router",
				Message: "router types identified",
				Data: map[string]interface{}{
					"paragraph_id": paragraph.ParagraphID,
					"chunk_id":     chunk.ChunkID,
					"types":        types,
				},
			})

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
		emitEvent(ctx, eventLogger, ExtractionEvent{
			Type:    "pipeline.done",
			Message: "no routed chunks to extract",
			Data: map[string]interface{}{
				"entities": 0,
			},
		})
		return EntityAndRelationshipsExtractorOutput{Payload: PhaseTempPayload{Entities: []PhaseTempEntity{}}}, nil
	}

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "extractor",
		Message: "extractor started",
		Data: map[string]interface{}{
			"chunks": len(routedChunks),
		},
	})
	phase2Output, err := u.extractor.Execute(ctx, Phase2EntryInput{
		Context:               input.Context,
		MaxCandidatesPerChunk: input.MaxCandidatesPerChunk,
		Chunks:                routedChunks,
		EventLogger:           eventLogger,
	})
	if err != nil {
		return EntityAndRelationshipsExtractorOutput{}, err
	}
	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.done",
		Phase:   "extractor",
		Message: "extractor finished",
		Data: map[string]interface{}{
			"findings": len(phase2Output.Findings),
		},
	})

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "matcher",
		Message: "matcher started",
		Data: map[string]interface{}{
			"findings": len(phase2Output.Findings),
		},
	})
	phase3Output, err := u.matcher.Execute(ctx, Phase3MatchInput{
		TenantID:      input.TenantID,
		WorldID:       &input.WorldID,
		Findings:      phase2Output.Findings,
		Context:       input.Context,
		MinSimilarity: input.MinSimilarity,
		MaxCandidates: input.MaxMatchCandidates,
		EventLogger:   eventLogger,
	})
	if err != nil {
		return EntityAndRelationshipsExtractorOutput{}, err
	}
	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.done",
		Phase:   "matcher",
		Message: "matcher finished",
		Data: map[string]interface{}{
			"results": len(phase3Output.Results),
		},
	})

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "payload",
		Message: "payload formatting started",
	})
	return EntityAndRelationshipsExtractorOutput{
		Payload: u.payload.Execute(phase3Output),
	}, nil
}

func uuidString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
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
