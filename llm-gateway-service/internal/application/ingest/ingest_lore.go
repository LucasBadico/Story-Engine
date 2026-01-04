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

// IngestLoreInput is the input for ingesting a lore
type IngestLoreInput struct {
	TenantID uuid.UUID
	LoreID   uuid.UUID
}

// IngestLoreOutput is the output after ingesting a lore
type IngestLoreOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestLoreUseCase handles lore ingestion
type IngestLoreUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	logger            *logger.Logger
}

// NewIngestLoreUseCase creates a new IngestLoreUseCase
func NewIngestLoreUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestLoreUseCase {
	return &IngestLoreUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:     documentRepo,
		chunkRepo:        chunkRepo,
		embedder:         embedder,
		logger:           logger,
	}
}

// Execute ingests a lore by fetching its content and generating embeddings
func (uc *IngestLoreUseCase) Execute(ctx context.Context, input IngestLoreInput) (*IngestLoreOutput, error) {
	// Fetch lore from main-service
	lore, err := uc.mainServiceClient.GetLore(ctx, input.LoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lore: %w", err)
	}

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, lore.WorldID)
	if err != nil {
		uc.logger.Error("failed to fetch world", "world_id", lore.WorldID, "error", err)
	}

	// Build content from lore
	content := uc.buildLoreContent(lore, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeLore,
		input.LoreID,
		lore.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeLore, input.LoreID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, lore, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	// Save chunks
	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Lore ingested successfully", "lore_id", input.LoreID, "chunks", len(chunks))

	return &IngestLoreOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildLoreContent builds content string from lore
func (uc *IngestLoreUseCase) buildLoreContent(lore *grpcclient.Lore, world *grpcclient.World) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Lore: %s", lore.Name))
	if lore.Category != nil && *lore.Category != "" {
		parts = append(parts, fmt.Sprintf("Category: %s", *lore.Category))
	}
	if lore.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", lore.Description))
	}
	if lore.Rules != "" {
		parts = append(parts, fmt.Sprintf("Rules: %s", lore.Rules))
	}
	if lore.Limitations != "" {
		parts = append(parts, fmt.Sprintf("Limitations: %s", lore.Limitations))
	}
	if lore.Requirements != "" {
		parts = append(parts, fmt.Sprintf("Requirements: %s", lore.Requirements))
	}
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with lore metadata
func (uc *IngestLoreUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, lore *grpcclient.Lore, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
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
		chunk.WorldID = &lore.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &lore.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

