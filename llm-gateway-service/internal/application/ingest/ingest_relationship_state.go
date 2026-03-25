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

type IngestRelationshipStateInput struct {
	TenantID            uuid.UUID
	RelationshipStateID uuid.UUID
}

type IngestRelationshipStateOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

type IngestRelationshipStateUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	stateSnapshotRepo repositories.StateSnapshotRepository
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

func NewIngestRelationshipStateUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	stateSnapshotRepo repositories.StateSnapshotRepository,
	logger *logger.Logger,
) *IngestRelationshipStateUseCase {
	return &IngestRelationshipStateUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		stateSnapshotRepo: stateSnapshotRepo,
		logger:            logger,
	}
}

func (uc *IngestRelationshipStateUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

func (uc *IngestRelationshipStateUseCase) Execute(ctx context.Context, input IngestRelationshipStateInput) (*IngestRelationshipStateOutput, error) {
	rs, err := uc.mainServiceClient.GetRelationshipState(ctx, input.RelationshipStateID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationship state: %w", err)
	}

	// Resolve StoryID via chapter
	var storyID uuid.UUID
	chapter, err := uc.mainServiceClient.GetChapter(ctx, rs.ChapterID)
	if err != nil {
		uc.logger.Warn("failed to fetch chapter for story ID", "chapter_id", rs.ChapterID, "error", err)
	} else {
		storyID = chapter.StoryID
	}

	// Resolve entity names for content building
	entityNames := uc.resolveEntityNames(ctx, rs.Entities)

	// Store in retrieval table
	entityDescriptions := make([]interface{}, len(rs.Entities))
	for i, e := range rs.Entities {
		name := entityNames[e.EntityID]
		entityDescriptions[i] = map[string]interface{}{
			"role": e.Role, "type": e.EntityType, "name": name, "id": e.EntityID.String(),
		}
	}
	stateData := map[string]interface{}{
		"tension_description": rs.TensionDescription,
		"entities":            entityDescriptions,
	}
	snapshot := memory.NewStateSnapshot(input.TenantID, storyID, input.RelationshipStateID, "relationship", rs.ChapterNumber, stateData)
	snapshot.TimelineDay = rs.TimelineDay
	snapshot.TimelineConfidence = rs.TimelineConfidence
	if err := uc.stateSnapshotRepo.Create(ctx, snapshot); err != nil {
		uc.logger.Error("failed to store state snapshot", "error", err, "relationship_state_id", input.RelationshipStateID)
	}

	// Build content
	title := uc.buildTitle(rs, entityNames)
	content := uc.buildContent(rs, entityNames)

	doc := memory.NewDocument(input.TenantID, memory.SourceTypeRelationshipState, input.RelationshipStateID, title, content)
	if storyID != uuid.Nil {
		doc.StoryID = &storyID
	}
	chapterNum := rs.ChapterNumber
	doc.ChapterNumber = &chapterNum
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeRelationshipState, input.RelationshipStateID)
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
	chunk.ChapterNumber = &rs.ChapterNumber
	chunk.TimelineDay = rs.TimelineDay
	chunk.TimelineConfidence = rs.TimelineConfidence
	// Set related characters from entities
	for _, e := range rs.Entities {
		if e.EntityType == "character" {
			if name, ok := entityNames[e.EntityID]; ok {
				chunk.RelatedCharacters = append(chunk.RelatedCharacters, name)
			}
		}
	}
	if err := chunk.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chunk: %w", err)
	}
	chunks := []*memory.Chunk{chunk}

	chunks, err = runIngestPipeline(ctx, uc.logger, uc.embedder, uc.summaryGenerator,
		string(memory.SourceTypeRelationshipState), title, []string{content}, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Relationship state ingested", "relationship_state_id", input.RelationshipStateID, "chunks", len(chunks))
	return &IngestRelationshipStateOutput{DocumentID: doc.ID, ChunkCount: len(chunks)}, nil
}

func (uc *IngestRelationshipStateUseCase) resolveEntityNames(ctx context.Context, entities []grpcclient.RelationshipStateEntity) map[uuid.UUID]string {
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
		default:
			names[e.EntityID] = e.EntityID.String()
		}
	}
	return names
}

func (uc *IngestRelationshipStateUseCase) buildTitle(rs *grpcclient.RelationshipState, names map[uuid.UUID]string) string {
	var entityStrs []string
	for _, e := range rs.Entities {
		name := names[e.EntityID]
		if name == "" {
			name = e.EntityID.String()
		}
		entityStrs = append(entityStrs, name)
	}
	return fmt.Sprintf("Relationship: %s", strings.Join(entityStrs, " ↔ "))
}

func (uc *IngestRelationshipStateUseCase) buildContent(rs *grpcclient.RelationshipState, names map[uuid.UUID]string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Relationship State (Chapter %d)", rs.ChapterNumber))
	parts = append(parts, fmt.Sprintf("Tension: %s", rs.TensionDescription))
	if rs.TimelineDay != nil {
		parts = append(parts, fmt.Sprintf("Timeline: Day %d", *rs.TimelineDay))
	}
	parts = append(parts, "Entities:")
	for _, e := range rs.Entities {
		name := names[e.EntityID]
		parts = append(parts, fmt.Sprintf("- %s (%s, %s)", name, e.Role, e.EntityType))
	}
	return strings.Join(parts, "\n")
}
