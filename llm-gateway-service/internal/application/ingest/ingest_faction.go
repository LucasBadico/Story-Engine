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

// IngestFactionInput is the input for ingesting a faction
type IngestFactionInput struct {
	TenantID  uuid.UUID
	FactionID uuid.UUID
}

// IngestFactionOutput is the output after ingesting a faction
type IngestFactionOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestFactionUseCase handles faction ingestion
type IngestFactionUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	logger            *logger.Logger
}

// NewIngestFactionUseCase creates a new IngestFactionUseCase
func NewIngestFactionUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestFactionUseCase {
	return &IngestFactionUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:     documentRepo,
		chunkRepo:        chunkRepo,
		embedder:         embedder,
		logger:           logger,
	}
}

// Execute ingests a faction by fetching its content and generating embeddings
func (uc *IngestFactionUseCase) Execute(ctx context.Context, input IngestFactionInput) (*IngestFactionOutput, error) {
	// Fetch faction from main-service
	faction, err := uc.mainServiceClient.GetFaction(ctx, input.FactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch faction: %w", err)
	}

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, faction.WorldID)
	if err != nil {
		uc.logger.Error("failed to fetch world", "world_id", faction.WorldID, "error", err)
	}

	// Build content from faction
	content := uc.buildFactionContent(faction, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeFaction,
		input.FactionID,
		faction.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeFaction, input.FactionID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, faction, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	// Save chunks
	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Faction ingested successfully", "faction_id", input.FactionID, "chunks", len(chunks))

	return &IngestFactionOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildFactionContent builds content string from faction
func (uc *IngestFactionUseCase) buildFactionContent(faction *grpcclient.Faction, world *grpcclient.World) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Faction: %s", faction.Name))
	if faction.Type != nil && *faction.Type != "" {
		parts = append(parts, fmt.Sprintf("Type: %s", *faction.Type))
	}
	if faction.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", faction.Description))
	}
	if faction.Beliefs != "" {
		parts = append(parts, fmt.Sprintf("Beliefs: %s", faction.Beliefs))
	}
	if faction.Structure != "" {
		parts = append(parts, fmt.Sprintf("Structure: %s", faction.Structure))
	}
	if faction.Symbols != "" {
		parts = append(parts, fmt.Sprintf("Symbols: %s", faction.Symbols))
	}
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with faction metadata
func (uc *IngestFactionUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, faction *grpcclient.Faction, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
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
		chunk.WorldID = &faction.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &faction.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

