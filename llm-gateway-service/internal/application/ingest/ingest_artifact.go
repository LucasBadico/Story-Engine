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

// IngestArtifactInput is the input for ingesting an artifact
type IngestArtifactInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
}

// IngestArtifactOutput is the output after ingesting an artifact
type IngestArtifactOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestArtifactUseCase handles artifact ingestion
type IngestArtifactUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestArtifactUseCase creates a new IngestArtifactUseCase
func NewIngestArtifactUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestArtifactUseCase {
	return &IngestArtifactUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestArtifactUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests an artifact by fetching its content and generating embeddings
func (uc *IngestArtifactUseCase) Execute(ctx context.Context, input IngestArtifactInput) (*IngestArtifactOutput, error) {
	// Fetch artifact from main-service
	artifact, err := uc.mainServiceClient.GetArtifact(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artifact: %w", err)
	}

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, artifact.WorldID)
	if err != nil {
		uc.logger.Warn("failed to fetch world", "world_id", artifact.WorldID, "error", err)
	}

	// Build content from artifact
	content := uc.buildArtifactContent(artifact, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeArtifact,
		input.ArtifactID,
		artifact.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeArtifact, input.ArtifactID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, artifact, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}
	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeArtifact,
		input.ArtifactID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeArtifact),
		artifact.Name,
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

	uc.logger.Info("Artifact ingested successfully", "artifact_id", input.ArtifactID, "chunks", len(chunks))

	return &IngestArtifactOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildArtifactContent builds content string from artifact
func (uc *IngestArtifactUseCase) buildArtifactContent(artifact *grpcclient.Artifact, world *grpcclient.World) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Artifact: %s", artifact.Name))
	if artifact.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", artifact.Description))
	}
	if artifact.Rarity != "" {
		parts = append(parts, fmt.Sprintf("Rarity: %s", artifact.Rarity))
	}
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with artifact metadata
func (uc *IngestArtifactUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, artifact *grpcclient.Artifact, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
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

		// Set world metadata
		chunk.WorldID = &artifact.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &artifact.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
