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

// IngestContentBlockInput is the input for ingesting a content block
type IngestContentBlockInput struct {
	TenantID       uuid.UUID
	ContentBlockID uuid.UUID
}

// IngestContentBlockOutput is the output after ingesting a content block
type IngestContentBlockOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestContentBlockUseCase handles content block ingestion with enriched metadata
type IngestContentBlockUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	logger            *logger.Logger
}

// NewIngestContentBlockUseCase creates a new IngestContentBlockUseCase
func NewIngestContentBlockUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestContentBlockUseCase {
	return &IngestContentBlockUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:     documentRepo,
		chunkRepo:        chunkRepo,
		embedder:         embedder,
		logger:           logger,
	}
}

// Execute ingests a content block by fetching its content, references, and generating enriched embeddings
// Only processes content blocks of type "text" for embeddings
func (uc *IngestContentBlockUseCase) Execute(ctx context.Context, input IngestContentBlockInput) (*IngestContentBlockOutput, error) {
	// Fetch content block from main-service
	contentBlock, err := uc.mainServiceClient.GetContentBlock(ctx, input.ContentBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content block: %w", err)
	}

	// Process text and image content blocks for embeddings
	if contentBlock.Type != "text" && contentBlock.Type != "image" {
		uc.logger.Debug("Skipping unsupported content block type", "content_block_id", input.ContentBlockID, "type", contentBlock.Type)
		return &IngestContentBlockOutput{
			DocumentID: uuid.Nil,
			ChunkCount: 0,
		}, nil
	}

	// Fetch references for the content block
	references, err := uc.mainServiceClient.ListContentBlockReferences(ctx, input.ContentBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content block references: %w", err)
	}

	// Fetch related entities and build metadata
	metadata, err := uc.buildMetadata(ctx, references)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata: %w", err)
	}

	// Build enriched content for embedding
	enrichedContent := uc.buildEnrichedContent(contentBlock, metadata)

	// Create or update document (one document per content block)
	title := fmt.Sprintf("Content Block")
	if contentBlock.OrderNum != nil {
		title = fmt.Sprintf("Content Block %d", *contentBlock.OrderNum)
	}
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeContentBlock,
		input.ContentBlockID,
		title,
		enrichedContent,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeContentBlock, input.ContentBlockID)
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
	chunk, err := uc.createChunkWithMetadata(ctx, doc.ID, contentBlock, metadata, enrichedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk: %w", err)
	}

	// Save chunk
	if err := uc.chunkRepo.Create(ctx, chunk); err != nil {
		return nil, fmt.Errorf("failed to save chunk: %w", err)
	}

	uc.logger.Info("Content block ingested successfully", "content_block_id", input.ContentBlockID, "beat_type", metadata.BeatType)

	return &IngestContentBlockOutput{
		DocumentID: doc.ID,
		ChunkCount: 1,
	}, nil
}

// ContentBlockMetadata contains enriched metadata for a content block
type ContentBlockMetadata struct {
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
func (uc *IngestContentBlockUseCase) buildMetadata(ctx context.Context, references []*grpcclient.ContentBlockReference) (*ContentBlockMetadata, error) {
	metadata := &ContentBlockMetadata{
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
					// Fetch character name
					char, err := uc.mainServiceClient.GetCharacter(ctx, *scene.POVCharacterID)
					if err == nil {
						metadata.POVCharacter = &char.Name
					} else {
						uc.logger.Error("failed to fetch POV character", "character_id", *scene.POVCharacterID, "error", err)
						povStr := scene.POVCharacterID.String()
						metadata.POVCharacter = &povStr
					}
				}
				if scene.LocationID != nil {
					metadata.LocationID = scene.LocationID
					// Fetch location name
					loc, err := uc.mainServiceClient.GetLocation(ctx, *scene.LocationID)
					if err == nil {
						metadata.LocationName = &loc.Name
			} else {
				uc.logger.Error("failed to fetch location", "location_id", *scene.LocationID, "error", err)
			}
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
				// Fetch character name
				char, err := uc.mainServiceClient.GetCharacter(ctx, *scene.POVCharacterID)
				if err == nil {
					metadata.POVCharacter = &char.Name
				} else {
					uc.logger.Error("failed to fetch POV character", "character_id", *scene.POVCharacterID, "error", err)
					povStr := scene.POVCharacterID.String()
					metadata.POVCharacter = &povStr
				}
			}
			if scene.LocationID != nil && metadata.LocationID == nil {
				metadata.LocationID = scene.LocationID
				// Fetch location name
				loc, err := uc.mainServiceClient.GetLocation(ctx, *scene.LocationID)
				if err == nil {
					metadata.LocationName = &loc.Name
				} else {
					uc.logger.Error("failed to fetch location", "location_id", *scene.LocationID, "error", err)
				}
			}

		case "character":
			// Fetch character name
			char, err := uc.mainServiceClient.GetCharacter(ctx, ref.EntityID)
			if err == nil {
				metadata.Characters = append(metadata.Characters, char.Name)
			} else {
				uc.logger.Error("failed to fetch character", "character_id", ref.EntityID, "error", err)
				charID := ref.EntityID.String()
				metadata.Characters = append(metadata.Characters, charID)
			}

		case "location":
			if metadata.LocationID == nil {
				metadata.LocationID = &ref.EntityID
			}
			// Fetch location name
			loc, err := uc.mainServiceClient.GetLocation(ctx, ref.EntityID)
			if err == nil {
				metadata.LocationName = &loc.Name
			} else {
				uc.logger.Error("failed to fetch location", "location_id", ref.EntityID, "error", err)
			}
		}
	}

	return metadata, nil
}

// buildEnrichedContent builds content string enriched with metadata
func (uc *IngestContentBlockUseCase) buildEnrichedContent(contentBlock *grpcclient.ContentBlock, metadata *ContentBlockMetadata) string {
	var parts []string

	// Add content block content
	if contentBlock.Type == "image" {
		// For images, use URL as content with descriptive prefix
		parts = append(parts, fmt.Sprintf("Image: %s", contentBlock.Content))
		// Add metadata description if available
		if contentBlock.Metadata != nil {
			if desc, ok := contentBlock.Metadata["description"].(string); ok && desc != "" {
				parts = append(parts, fmt.Sprintf("Description: %s", desc))
			}
			if alt, ok := contentBlock.Metadata["alt"].(string); ok && alt != "" {
				parts = append(parts, fmt.Sprintf("Alt text: %s", alt))
			}
		}
	} else {
		parts = append(parts, contentBlock.Content)
	}

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
func (uc *IngestContentBlockUseCase) createChunkWithMetadata(ctx context.Context, documentID uuid.UUID, contentBlock *grpcclient.ContentBlock, metadata *ContentBlockMetadata, content string) (*memory.Chunk, error) {
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
	
	// Set ContentBlock metadata
	chunk.ContentType = &contentBlock.Type
	chunk.ContentKind = &contentBlock.Kind

	if err := chunk.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chunk: %w", err)
	}

	return chunk, nil
}

