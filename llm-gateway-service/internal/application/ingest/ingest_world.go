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

// IngestWorldInput is the input for ingesting a world
type IngestWorldInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
}

// IngestWorldOutput is the output after ingesting a world
type IngestWorldOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestWorldUseCase handles world ingestion
type IngestWorldUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestWorldUseCase creates a new IngestWorldUseCase
func NewIngestWorldUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestWorldUseCase {
	return &IngestWorldUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestWorldUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a world by fetching its content and generating embeddings
func (uc *IngestWorldUseCase) Execute(ctx context.Context, input IngestWorldInput) (*IngestWorldOutput, error) {
	// Fetch world from main-service
	world, err := uc.mainServiceClient.GetWorld(ctx, input.WorldID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch world: %w", err)
	}

	// Build content from world metadata
	content := uc.buildWorldContent(world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeWorld,
		input.WorldID,
		world.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeWorld, input.WorldID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}
	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeWorld,
		input.WorldID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeWorld),
		world.Name,
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

	uc.logger.Info("World ingested successfully", "world_id", input.WorldID, "chunks", len(chunks))

	return &IngestWorldOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildWorldContent builds content string from world metadata
func (uc *IngestWorldUseCase) buildWorldContent(world *grpcclient.World) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	if world.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", world.Description))
	}
	if world.Genre != "" {
		parts = append(parts, fmt.Sprintf("Genre: %s", world.Genre))
	}
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with world metadata
func (uc *IngestWorldUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
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

		// Set world metadata
		chunk.WorldID = &world.ID
		chunk.WorldName = &world.Name
		if world.Genre != "" {
			chunk.WorldGenre = &world.Genre
		}
		chunk.EntityName = &world.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
