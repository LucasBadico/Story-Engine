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

// IngestStoryInput is the input for ingesting a story
type IngestStoryInput struct {
	TenantID uuid.UUID
	StoryID  uuid.UUID
}

// IngestStoryOutput is the output after ingesting a story
type IngestStoryOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestStoryUseCase handles story ingestion
type IngestStoryUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestStoryUseCase creates a new IngestStoryUseCase
func NewIngestStoryUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestStoryUseCase {
	return &IngestStoryUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestStoryUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a story by fetching its content and generating embeddings
func (uc *IngestStoryUseCase) Execute(ctx context.Context, input IngestStoryInput) (*IngestStoryOutput, error) {
	// Fetch story from main-service
	story, err := uc.mainServiceClient.GetStory(ctx, input.StoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch story: %w", err)
	}

	// Build content from story metadata
	content := uc.buildStoryContent(story)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeStory,
		input.StoryID,
		story.Title,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeStory, input.StoryID)
	if err == nil && existingDoc != nil {
		// Update existing document
		doc.ID = existingDoc.ID
		doc.CreatedAt = existingDoc.CreatedAt
		if err := uc.documentRepo.Update(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
		// Delete old chunks
		if err := uc.chunkRepo.DeleteByDocument(ctx, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to delete old chunks: %w", err)
		}
	} else {
		// Create new document
		if err := uc.documentRepo.Create(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}
	}

	// Chunk content and generate embeddings
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeStory,
		input.StoryID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeStory),
		story.Title,
		summaryContents,
		chunks,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	// Save chunks
	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Story ingested successfully", "story_id", input.StoryID, "chunks", len(chunks))

	return &IngestStoryOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildStoryContent builds content string from story metadata
func (uc *IngestStoryUseCase) buildStoryContent(story *grpcclient.Story) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Title: %s", story.Title))
	parts = append(parts, fmt.Sprintf("Status: %s", story.Status))
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings
func (uc *IngestStoryUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, content string) ([]*memory.Chunk, error) {
	// Simple chunking: split by paragraphs (for now)
	paragraphs := strings.Split(content, "\n\n")
	chunks := make([]*memory.Chunk, 0, len(paragraphs))

	for i, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Generate embedding
		embedding, err := uc.embedder.EmbedText(para)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}

		// Estimate token count (rough: 1 token ≈ 4 characters)
		tokenCount := len(para) / 4

		chunk := memory.NewChunk(documentID, i, para, embedding, tokenCount)
		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
