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

// IngestChapterInput is the input for ingesting a chapter
type IngestChapterInput struct {
	TenantID  uuid.UUID
	ChapterID uuid.UUID
}

// IngestChapterOutput is the output after ingesting a chapter
type IngestChapterOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestChapterUseCase handles chapter ingestion
type IngestChapterUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestChapterUseCase creates a new IngestChapterUseCase
func NewIngestChapterUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestChapterUseCase {
	return &IngestChapterUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestChapterUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a chapter by fetching its content and generating embeddings
func (uc *IngestChapterUseCase) Execute(ctx context.Context, input IngestChapterInput) (*IngestChapterOutput, error) {
	// Fetch chapter from main-service
	chapter, err := uc.mainServiceClient.GetChapter(ctx, input.ChapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chapter: %w", err)
	}

	// Fetch content blocks for the chapter
	contentBlocks, err := uc.mainServiceClient.ListContentBlocksByChapter(ctx, input.ChapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content blocks: %w", err)
	}

	// Build content from chapter and content blocks
	content := uc.buildChapterContent(chapter, contentBlocks)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeChapter,
		input.ChapterID,
		chapter.Title,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeChapter, input.ChapterID)
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
	entityName := strings.TrimSpace(chapter.Title)
	if entityName == "" {
		entityName = fmt.Sprintf("Chapter %d", chapter.Number)
	}
	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeChapter,
		input.ChapterID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeChapter),
		entityName,
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

	uc.logger.Info("Chapter ingested successfully", "chapter_id", input.ChapterID, "chunks", len(chunks))

	return &IngestChapterOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildChapterContent builds content string from chapter and content blocks
func (uc *IngestChapterUseCase) buildChapterContent(chapter *grpcclient.Chapter, contentBlocks []*grpcclient.ContentBlock) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Chapter %d: %s", chapter.Number, chapter.Title))
	parts = append(parts, fmt.Sprintf("Status: %s", chapter.Status))
	parts = append(parts, "")

	// Add content blocks content (only text type)
	for _, cb := range contentBlocks {
		if cb.Type == "text" {
			parts = append(parts, cb.Content)
			parts = append(parts, "")
		}
	}

	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings (same as IngestStoryUseCase)
func (uc *IngestChapterUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, content string) ([]*memory.Chunk, error) {
	// Simple chunking: split by paragraphs
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

		// Estimate token count
		tokenCount := len(para) / 4

		chunk := memory.NewChunk(documentID, i, para, embedding, tokenCount)
		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
