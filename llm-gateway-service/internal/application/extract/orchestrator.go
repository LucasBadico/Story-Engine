package extract

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/entities"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/events"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/relations"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type ExtractOrchestrator struct {
	router            *entities.Phase1EntityTypeRouterUseCase
	extractor         *entities.Phase2EntryUseCase
	matcher           *entities.Phase3MatchUseCase
	payload           *entities.Phase4EntitiesPayloadUseCase
	relationDiscovery *relations.Phase5RelationDiscoveryUseCase
	relationNormalize *relations.Phase6RelationNormalizeUseCase
	relationMatcher   *relations.Phase7RelationMatchUseCase
	logger            *logger.Logger
}

func NewExtractOrchestrator(
	router *entities.Phase1EntityTypeRouterUseCase,
	extractor *entities.Phase2EntryUseCase,
	matcher *entities.Phase3MatchUseCase,
	payload *entities.Phase4EntitiesPayloadUseCase,
	relationDiscovery *relations.Phase5RelationDiscoveryUseCase,
	relationNormalize *relations.Phase6RelationNormalizeUseCase,
	relationMatcher *relations.Phase7RelationMatchUseCase,
	logger *logger.Logger,
) *ExtractOrchestrator {
	return &ExtractOrchestrator{
		router:            router,
		extractor:         extractor,
		matcher:           matcher,
		payload:           payload,
		relationDiscovery: relationDiscovery,
		relationNormalize: relationNormalize,
		relationMatcher:   relationMatcher,
		logger:            logger,
	}
}

type ExtractRequest struct {
	TenantID              uuid.UUID
	WorldID               uuid.UUID
	RequestID             string
	Text                  string
	Context               string
	EntityTypes           []string
	MaxChunkChars         int
	OverlapChars          int
	MaxTypeCandidates     int
	MaxCandidatesPerChunk int
	MinSimilarity         float64
	MaxMatchCandidates    int
	MaxRelationMatches    int
	RelationMatchMinSim   float64
	SuggestedRelations    map[string]relations.Phase5PerEntityRelationMap
	RelationTypes         map[string]relations.Phase6RelationTypeDefinition
	RelationTypeSemantics map[string]string
	EventLogger           events.ExtractionEventLogger
	IncludeRelations      bool
}

type ExtractResult struct {
	Payload ExtractPayload
}

func (u *ExtractOrchestrator) Execute(ctx context.Context, input ExtractRequest) (ExtractResult, error) {
	if input.TenantID == uuid.Nil {
		return ExtractResult{}, errors.New("tenant id is required")
	}
	if input.WorldID == uuid.Nil {
		return ExtractResult{}, errors.New("world id is required")
	}
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return ExtractResult{}, errors.New("text is required")
	}
	if u.router == nil || u.extractor == nil || u.matcher == nil || u.payload == nil {
		return ExtractResult{}, errors.New("router, extractor, matcher, and payload are required")
	}

	eventLogger := events.NormalizeEventLogger(input.EventLogger)
	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
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

	split, err := entities.SplitTextIntoParagraphChunks(entities.Phase0TextSplitInput{
		Text:          text,
		MaxChunkChars: input.MaxChunkChars,
		OverlapChars:  input.OverlapChars,
	})
	if err != nil {
		return ExtractResult{}, err
	}

	routedChunks := make([]entities.Phase2RoutedChunk, 0)
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

			events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
				Type:    "router.chunk",
				Phase:   "entities.detect",
				Message: "router types identified",
				Data: map[string]interface{}{
					"paragraph_id": paragraph.ParagraphID,
					"chunk_id":     chunk.ChunkID,
					"types":        types,
				},
			})

			routedChunks = append(routedChunks, entities.Phase2RoutedChunk{
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
		events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
			Type:    "pipeline.done",
			Message: "no routed chunks to extract",
			Data: map[string]interface{}{
				"entities": 0,
			},
		})
		return ExtractResult{Payload: ExtractPayload{Entities: []entities.Phase4Entity{}, Relations: []relations.Phase8RelationResult{}}}, nil
	}

	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "phase.start",
		Phase:   "entities.candidates",
		Message: "extractor started",
		Data: map[string]interface{}{
			"chunks": len(routedChunks),
		},
	})
	phase2Output, err := u.extractor.Execute(ctx, entities.Phase2EntryInput{
		Context:               input.Context,
		MaxCandidatesPerChunk: input.MaxCandidatesPerChunk,
		Chunks:                routedChunks,
		EventLogger:           eventLogger,
	})
	if err != nil {
		return ExtractResult{}, err
	}
	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "phase.done",
		Phase:   "entities.candidates",
		Message: "extractor finished",
		Data: map[string]interface{}{
			"findings": len(phase2Output.Findings),
		},
	})

	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "phase.start",
		Phase:   "entities.resolve",
		Message: "matcher started",
		Data: map[string]interface{}{
			"findings": len(phase2Output.Findings),
		},
	})
	phase3Output, err := u.matcher.Execute(ctx, entities.Phase3MatchInput{
		TenantID:      input.TenantID,
		WorldID:       &input.WorldID,
		Findings:      phase2Output.Findings,
		Context:       input.Context,
		MinSimilarity: input.MinSimilarity,
		MaxCandidates: input.MaxMatchCandidates,
		EventLogger:   eventLogger,
	})
	if err != nil {
		return ExtractResult{}, err
	}
	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "phase.done",
		Phase:   "entities.resolve",
		Message: "matcher finished",
		Data: map[string]interface{}{
			"results": len(phase3Output.Results),
		},
	})

	relationsResult := []relations.Phase8RelationResult{}
	basePayload := u.payload.Execute(phase3Output)

	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "result_entities",
		Phase:   "entities.result",
		Message: "entity extraction completed",
		Data: map[string]interface{}{
			"entities": basePayload.Entities,
		},
	})

	if input.IncludeRelations && u.relationDiscovery != nil && u.relationNormalize != nil && u.relationMatcher != nil {
		if len(input.SuggestedRelations) == 0 || len(input.RelationTypes) == 0 {
			u.logger.Warn("relation maps missing; skipping relation extraction",
				"suggested", len(input.SuggestedRelations),
				"types", len(input.RelationTypes))
			events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
				Type:    "relation.error",
				Phase:   "relations.result",
				Message: "relation extraction skipped (maps missing)",
				Data: map[string]interface{}{
					"suggested": len(input.SuggestedRelations),
					"types":     len(input.RelationTypes),
				},
			})
		} else {
			relationExtractor := relations.NewExtractor(
				u.relationDiscovery,
				u.relationNormalize,
				u.relationMatcher,
				u.logger,
			)
			relationsResult, err = relationExtractor.Execute(ctx, relations.ExtractInput{
				TenantID:              input.TenantID,
				WorldID:               input.WorldID,
				RequestID:             input.RequestID,
				Text:                  input.Text,
				Context:               input.Context,
				MaxRelationMatches:    input.MaxRelationMatches,
				RelationMatchMinSim:   input.RelationMatchMinSim,
				SuggestedRelations:    input.SuggestedRelations,
				RelationTypes:         input.RelationTypes,
				RelationTypeSemantics: input.RelationTypeSemantics,
				EventLogger:           input.EventLogger,
			}, phase2Output, phase3Output)
			if err != nil {
				u.logger.Error("relation extraction failed", "error", err)
				events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
					Type:    "relation.error",
					Phase:   "relations.result",
					Message: "relation extraction failed",
					Data: map[string]interface{}{
						"error": err.Error(),
					},
				})
			}
		}
	}

	if len(relationsResult) > 0 {
		events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
			Type:    "result_relations",
			Phase:   "relations.result",
			Message: "relation extraction completed",
			Data: map[string]interface{}{
				"relations": relationsResult,
			},
		})
		events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
			Type:    "relation.success",
			Phase:   "relations.result",
			Message: "relation extraction succeeded",
			Data: map[string]interface{}{
				"count": len(relationsResult),
			},
		})
	} else if input.IncludeRelations {
		events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
			Type:    "relation.success",
			Phase:   "relations.result",
			Message: "relation extraction completed with no relations",
			Data: map[string]interface{}{
				"count": 0,
			},
		})
	}

	events.EmitEvent(ctx, eventLogger, events.ExtractionEvent{
		Type:    "phase.start",
		Phase:   "entities.result",
		Message: "payload formatting started",
	})
	return ExtractResult{
		Payload: mergeRelationsPayload(basePayload, relationsResult),
	}, nil
}

func mergeRelationsPayload(payload entities.Phase4EntitiesPayload, relations []relations.Phase8RelationResult) ExtractPayload {
	return ExtractPayload{
		Entities:  payload.Entities,
		Relations: relations,
	}
}

func uuidString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func (u *ExtractOrchestrator) routeChunkTypes(
	ctx context.Context,
	text string,
	context string,
	entityTypes []string,
	maxCandidates int,
) ([]string, error) {
	output, err := u.router.Execute(ctx, entities.Phase1EntityTypeRouterInput{
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
