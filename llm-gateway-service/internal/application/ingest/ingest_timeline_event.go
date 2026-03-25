package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

type IngestTimelineEventInput struct {
	TenantID        uuid.UUID
	TimelineEventID uuid.UUID
}

type IngestTimelineEventOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

type IngestTimelineEventUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	eventRepo         repositories.EventRepository
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

func NewIngestTimelineEventUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	eventRepo repositories.EventRepository,
	logger *logger.Logger,
) *IngestTimelineEventUseCase {
	return &IngestTimelineEventUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		eventRepo:         eventRepo,
		logger:            logger,
	}
}

func (uc *IngestTimelineEventUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

func (uc *IngestTimelineEventUseCase) Execute(ctx context.Context, input IngestTimelineEventInput) (*IngestTimelineEventOutput, error) {
	te, err := uc.mainServiceClient.GetTimelineEvent(ctx, input.TimelineEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timeline event: %w", err)
	}

	// Resolve entity names
	entityNames := uc.resolveEntityNames(ctx, te.Entities)

	// Store in retrieval table
	entitiesInvolved := make([]string, 0, len(te.Entities))
	for _, e := range te.Entities {
		name := entityNames[e.EntityID]
		if name != "" {
			entitiesInvolved = append(entitiesInvolved, name)
		} else {
			entitiesInvolved = append(entitiesInvolved, e.EntityID.String())
		}
	}
	event := memory.NewEvent(input.TenantID, te.StoryID, te.ChapterNumber, te.Description)
	event.TimelineDay = te.TimelineDay
	event.TimelineConfidence = te.TimelineConfidence
	event.EntitiesInvolved = entitiesInvolved
	if err := uc.eventRepo.Create(ctx, event); err != nil {
		uc.logger.Error("failed to store event", "error", err, "timeline_event_id", input.TimelineEventID)
	}

	// Build content
	content := uc.buildContent(te, entityNames)
	title := fmt.Sprintf("Event: Chapter %d", te.ChapterNumber)
	if te.TimelineDay != nil {
		title = fmt.Sprintf("Event: Chapter %d, Day %d", te.ChapterNumber, *te.TimelineDay)
	}

	doc := memory.NewDocument(input.TenantID, memory.SourceTypeTimelineEvent, input.TimelineEventID, title, content)
	doc.StoryID = &te.StoryID
	chapterNum := te.ChapterNumber
	doc.ChapterNumber = &chapterNum
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeTimelineEvent, input.TimelineEventID)
	if err == nil && existingDoc != nil {
		doc.ID = existingDoc.ID
		doc.CreatedAt = existingDoc.CreatedAt
		if err := uc.documentRepo.Update(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
		if err := uc.chunkRepo.DeleteByDocument(ctx, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to delete old chunks: %w", err)
		}
	} else {
		if err := uc.documentRepo.Create(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}
	}

	// Chunk and embed
	embedding, err := uc.embedder.EmbedText(content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	chunk := memory.NewChunk(doc.ID, 0, content, embedding, len(content)/4)
	chunk.ChapterNumber = &te.ChapterNumber
	chunk.TimelineDay = te.TimelineDay
	chunk.TimelineConfidence = te.TimelineConfidence
	if err := chunk.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chunk: %w", err)
	}
	chunks := []*memory.Chunk{chunk}

	chunks, err = runIngestPipeline(ctx, uc.logger, uc.embedder, uc.summaryGenerator,
		string(memory.SourceTypeTimelineEvent), title, []string{content}, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Timeline event ingested", "timeline_event_id", input.TimelineEventID, "chunks", len(chunks))
	return &IngestTimelineEventOutput{DocumentID: doc.ID, ChunkCount: len(chunks)}, nil
}

func (uc *IngestTimelineEventUseCase) resolveEntityNames(ctx context.Context, entities []grpcclient.TimelineEventEntity) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string)
	for _, e := range entities {
		switch e.EntityType {
		case "character":
			if ch, err := uc.mainServiceClient.GetCharacter(ctx, e.EntityID); err == nil {
				names[e.EntityID] = ch.Name
			}
		case "location":
			if loc, err := uc.mainServiceClient.GetLocation(ctx, e.EntityID); err == nil {
				names[e.EntityID] = loc.Name
			}
		case "artifact":
			if art, err := uc.mainServiceClient.GetArtifact(ctx, e.EntityID); err == nil {
				names[e.EntityID] = art.Name
			}
		}
	}
	return names
}

func (uc *IngestTimelineEventUseCase) buildContent(te *grpcclient.TimelineEvent, names map[uuid.UUID]string) string {
	var parts []string
	header := fmt.Sprintf("Timeline Event (Chapter %d)", te.ChapterNumber)
	if te.TimelineDay != nil {
		header = fmt.Sprintf("Timeline Event (Chapter %d, Day %d)", te.ChapterNumber, *te.TimelineDay)
	}
	parts = append(parts, header)
	parts = append(parts, te.Description)
	if len(te.Entities) > 0 {
		parts = append(parts, "Entities involved:")
		for _, e := range te.Entities {
			name := names[e.EntityID]
			if name == "" {
				name = e.EntityID.String()
			}
			parts = append(parts, fmt.Sprintf("- %s (%s, %s)", name, e.Role, e.EntityType))
		}
	}
	return strings.Join(parts, "\n")
}
