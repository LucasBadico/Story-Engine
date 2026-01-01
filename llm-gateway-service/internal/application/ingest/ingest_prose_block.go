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

// IngestProseBlockInput is the input for ingesting a prose block
type IngestProseBlockInput struct {
	TenantID    uuid.UUID
	ProseBlockID uuid.UUID
}

// IngestProseBlockOutput is the output after ingesting a prose block
type IngestProseBlockOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestProseBlockUseCase handles prose block ingestion with enriched metadata
type IngestProseBlockUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	logger            logger.Logger
}

// NewIngestProseBlockUseCase creates a new IngestProseBlockUseCase
func NewIngestProseBlockUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger logger.Logger,
) *IngestProseBlockUseCase {
	return &IngestProseBlockUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:     documentRepo,
		chunkRepo:        chunkRepo,
		embedder:         embedder,
		logger:           logger,
	}
}

// Execute ingests a prose block by fetching its content, references, and generating enriched embeddings
func (uc *IngestProseBlockUseCase) Execute(ctx context.Context, input IngestProseBlockInput) (*IngestProseBlockOutput, error) {
	// Fetch prose block from main-service
	proseBlock, err := uc.mainServiceClient.GetProseBlock(ctx, input.ProseBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prose block: %w", err)
	}

	// Fetch references for the prose block
	references, err := uc.mainServiceClient.ListProseBlockReferences(ctx, input.ProseBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prose block references: %w", err)
	}

	// Fetch related entities and build metadata
	metadata, err := uc.buildMetadata(ctx, references)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata: %w", err)
	}

	// Build enriched content for embedding
	enrichedContent := uc.buildEnrichedContent(proseBlock, metadata)

	// Create or update document (one document per prose block)
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeProseBlock,
		input.ProseBlockID,
		fmt.Sprintf("Prose Block %d", proseBlock.OrderNum),
		enrichedContent,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeProseBlock, input.ProseBlockID)
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

	// Create chunk with metadata
	chunk, err := uc.createChunkWithMetadata(ctx, doc.ID, proseBlock, metadata, enrichedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk: %w", err)
	}

	// Save chunk
	if err := uc.chunkRepo.Create(ctx, chunk); err != nil {
		return nil, fmt.Errorf("failed to save chunk: %w", err)
	}

	uc.logger.Info("Prose block ingested successfully", "prose_block_id", input.ProseBlockID, "beat_type", metadata.BeatType)

	return &IngestProseBlockOutput{
		DocumentID: doc.ID,
		ChunkCount: 1,
	}, nil
}

// ProseBlockMetadata contains enriched metadata for a prose block
type ProseBlockMetadata struct {
	SceneID      *uuid.UUID
	BeatID       *uuid.UUID
	BeatType     *string
	BeatIntent   *string
	Characters   []string
	LocationID   *uuid.UUID
	LocationName *string
	Timeline     *string
	POVCharacter *string
}

// buildMetadata fetches related entities and builds metadata
func (uc *IngestProseBlockUseCase) buildMetadata(ctx context.Context, references []*grpcclient.ProseBlockReference) (*ProseBlockMetadata, error) {
	metadata := &ProseBlockMetadata{
		Characters: []string{},
	}

	for _, ref := range references {
		switch ref.EntityType {
		case "beat":
			beat, err := uc.mainServiceClient.GetBeat(ctx, ref.EntityID)
			if err != nil {
				uc.logger.Error("failed to fetch beat", "beat_id", ref.EntityID, "error", err)
				continue
			}
			metadata.BeatID = &beat.ID
			metadata.BeatType = &beat.Type
			if beat.Intent != "" {
				metadata.BeatIntent = &beat.Intent
			}

			// Fetch scene for beat
			scene, err := uc.mainServiceClient.GetScene(ctx, beat.SceneID)
			if err != nil {
				uc.logger.Error("failed to fetch scene", "scene_id", beat.SceneID, "error", err)
			} else {
				metadata.SceneID = &scene.ID
				if scene.TimeRef != "" {
					metadata.Timeline = &scene.TimeRef
				}
				if scene.POVCharacterID != nil {
					// For now, we'll store the UUID as string. In the future, we could fetch character name
					povStr := scene.POVCharacterID.String()
					metadata.POVCharacter = &povStr
				}
				if scene.LocationID != nil {
					metadata.LocationID = scene.LocationID
					// In the future, we could fetch location name from a location service
				}
			}

		case "scene":
			scene, err := uc.mainServiceClient.GetScene(ctx, ref.EntityID)
			if err != nil {
				uc.logger.Error("failed to fetch scene", "scene_id", ref.EntityID, "error", err)
				continue
			}
			if metadata.SceneID == nil {
				metadata.SceneID = &scene.ID
			}
			if scene.TimeRef != "" && metadata.Timeline == nil {
				metadata.Timeline = &scene.TimeRef
			}
			if scene.POVCharacterID != nil && metadata.POVCharacter == nil {
				povStr := scene.POVCharacterID.String()
				metadata.POVCharacter = &povStr
			}
			if scene.LocationID != nil && metadata.LocationID == nil {
				metadata.LocationID = scene.LocationID
			}

		case "character":
			// For now, we'll just store the character ID. In the future, fetch character name
			charID := ref.EntityID.String()
			metadata.Characters = append(metadata.Characters, charID)

		case "location":
			if metadata.LocationID == nil {
				metadata.LocationID = &ref.EntityID
			}
		}
	}

	return metadata, nil
}

// buildEnrichedContent builds content string enriched with metadata
func (uc *IngestProseBlockUseCase) buildEnrichedContent(proseBlock *grpcclient.ProseBlock, metadata *ProseBlockMetadata) string {
	var parts []string

	// Add prose block content
	parts = append(parts, proseBlock.Content)

	// Add beat context if available
	if metadata.BeatType != nil {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Beat Type: %s", *metadata.BeatType))
		if metadata.BeatIntent != nil && *metadata.BeatIntent != "" {
			parts = append(parts, fmt.Sprintf("Beat Intent: %s", *metadata.BeatIntent))
		}
	}

	// Add characters if available
	if len(metadata.Characters) > 0 {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Characters: %s", strings.Join(metadata.Characters, ", ")))
	}

	// Add location if available
	if metadata.LocationName != nil {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Location: %s", *metadata.LocationName))
	}

	// Add timeline if available
	if metadata.Timeline != nil {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Timeline: %s", *metadata.Timeline))
	}

	return strings.Join(parts, "\n")
}

// createChunkWithMetadata creates a chunk with enriched metadata
func (uc *IngestProseBlockUseCase) createChunkWithMetadata(ctx context.Context, documentID uuid.UUID, proseBlock *grpcclient.ProseBlock, metadata *ProseBlockMetadata, content string) (*memory.Chunk, error) {
	// Generate embedding
	embedding, err := uc.embedder.EmbedText(content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Estimate token count
	tokenCount := len(content) / 4

	chunk := memory.NewChunk(documentID, 0, content, embedding, tokenCount)

	// Set metadata
	chunk.SceneID = metadata.SceneID
	chunk.BeatID = metadata.BeatID
	chunk.BeatType = metadata.BeatType
	chunk.BeatIntent = metadata.BeatIntent
	chunk.Characters = metadata.Characters
	chunk.LocationID = metadata.LocationID
	chunk.LocationName = metadata.LocationName
	chunk.Timeline = metadata.Timeline
	chunk.POVCharacter = metadata.POVCharacter
	chunk.ProseKind = &proseBlock.Kind

	if err := chunk.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chunk: %w", err)
	}

	return chunk, nil
}

