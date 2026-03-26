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

type IngestCharacterStateInput struct {
	TenantID         uuid.UUID
	CharacterStateID uuid.UUID
}

type IngestCharacterStateOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

type IngestCharacterStateUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	stateSnapshotRepo repositories.StateSnapshotRepository
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

func NewIngestCharacterStateUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	stateSnapshotRepo repositories.StateSnapshotRepository,
	logger *logger.Logger,
) *IngestCharacterStateUseCase {
	return &IngestCharacterStateUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		stateSnapshotRepo: stateSnapshotRepo,
		logger:            logger,
	}
}

func (uc *IngestCharacterStateUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

func (uc *IngestCharacterStateUseCase) Execute(ctx context.Context, input IngestCharacterStateInput) (*IngestCharacterStateOutput, error) {
	cs, err := uc.mainServiceClient.GetCharacterState(ctx, input.CharacterStateID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character state: %w", err)
	}

	character, err := uc.mainServiceClient.GetCharacter(ctx, cs.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}

	// Resolve StoryID via chapter
	var storyID uuid.UUID
	if cs.ChapterID != nil {
		chapter, err := uc.mainServiceClient.GetChapter(ctx, *cs.ChapterID)
		if err != nil {
			uc.logger.Warn("failed to fetch chapter for story ID", "chapter_id", cs.ChapterID, "error", err)
		} else {
			storyID = chapter.StoryID
		}
	}

	// Store in retrieval table
	stateData := map[string]interface{}{
		"emotion":        cs.Emotion,
		"goal":           cs.Goal,
		"secrets":        cs.Secrets,
		"clothes":        cs.Clothes,
		"physical_state": cs.PhysicalState,
	}
	snapshot := memory.NewStateSnapshot(input.TenantID, storyID, cs.CharacterID, "character", cs.ChapterNumber, stateData)
	snapshot.TimelineDay = cs.TimelineDay
	snapshot.TimelineConfidence = cs.TimelineConfidence
	if err := uc.stateSnapshotRepo.Create(ctx, snapshot); err != nil {
		uc.logger.Error("failed to store state snapshot", "error", err, "character_state_id", input.CharacterStateID)
	}

	// Build content for embedding
	content := uc.buildContent(cs, character)

	// Create or update document
	doc := memory.NewDocument(input.TenantID, memory.SourceTypeCharacterState, input.CharacterStateID, character.Name, content)
	if storyID != uuid.Nil {
		doc.StoryID = &storyID
	}
	chapterNum := cs.ChapterNumber
	doc.ChapterNumber = &chapterNum
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeCharacterState, input.CharacterStateID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, cs, character)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	chunks, err = runIngestPipeline(ctx, uc.logger, uc.embedder, uc.summaryGenerator,
		string(memory.SourceTypeCharacterState), character.Name, []string{content}, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Character state ingested", "character_state_id", input.CharacterStateID, "chunks", len(chunks))
	return &IngestCharacterStateOutput{DocumentID: doc.ID, ChunkCount: len(chunks)}, nil
}

func (uc *IngestCharacterStateUseCase) buildContent(cs *grpcclient.CharacterState, character *grpcclient.Character) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Character State: %s (Chapter %d)", character.Name, cs.ChapterNumber))

	if cs.Emotion != "" {
		parts = append(parts, fmt.Sprintf("Emotion: %s", cs.Emotion))
	}
	if cs.Goal != "" {
		parts = append(parts, fmt.Sprintf("Goal: %s", cs.Goal))
	}
	if len(cs.Secrets) > 0 {
		parts = append(parts, fmt.Sprintf("Secrets: %s", strings.Join(cs.Secrets, "; ")))
	}
	if cs.Clothes != "" {
		parts = append(parts, fmt.Sprintf("Clothes: %s", cs.Clothes))
	}
	if cs.PhysicalState != "" {
		parts = append(parts, fmt.Sprintf("Physical State: %s", cs.PhysicalState))
	}
	if cs.TimelineDay != nil {
		parts = append(parts, fmt.Sprintf("Timeline: Day %d", *cs.TimelineDay))
	}

	return strings.Join(parts, "\n")
}

func (uc *IngestCharacterStateUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, cs *grpcclient.CharacterState, character *grpcclient.Character) ([]*memory.Chunk, error) {
	content := uc.buildContent(cs, character)
	paragraphs := strings.Split(content, "\n\n")
	chunks := make([]*memory.Chunk, 0, len(paragraphs))

	for i, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		embedding, err := uc.embedder.EmbedText(para)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}
		tokenCount := len(para) / 4
		chunk := memory.NewChunk(documentID, i, para, embedding, tokenCount)
		chunk.WorldID = &character.WorldID
		chunk.EntityName = &character.Name
		chunk.ChapterNumber = &cs.ChapterNumber
		chunk.TimelineDay = cs.TimelineDay
		chunk.TimelineConfidence = cs.TimelineConfidence
		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}
