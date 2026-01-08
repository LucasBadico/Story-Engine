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

// IngestCharacterInput is the input for ingesting a character
type IngestCharacterInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// IngestCharacterOutput is the output after ingesting a character
type IngestCharacterOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestCharacterUseCase handles character ingestion with traits
type IngestCharacterUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestCharacterUseCase creates a new IngestCharacterUseCase
func NewIngestCharacterUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestCharacterUseCase {
	return &IngestCharacterUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestCharacterUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a character by fetching its content, traits, and generating embeddings
func (uc *IngestCharacterUseCase) Execute(ctx context.Context, input IngestCharacterInput) (*IngestCharacterOutput, error) {
	// Fetch character from main-service
	character, err := uc.mainServiceClient.GetCharacter(ctx, input.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}

	// Fetch character traits
	traits, err := uc.mainServiceClient.GetCharacterTraits(ctx, input.CharacterID)
	if err != nil {
		uc.logger.Warn("failed to fetch character traits", "character_id", input.CharacterID, "error", err)
		traits = []*grpcclient.CharacterTrait{}
	}

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, character.WorldID)
	if err != nil {
		uc.logger.Warn("failed to fetch world", "world_id", character.WorldID, "error", err)
	}

	// Build content from character and traits
	content := uc.buildCharacterContent(character, traits, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeCharacter,
		input.CharacterID,
		character.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeCharacter, input.CharacterID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, character, world, traits, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeCharacter,
		input.CharacterID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeCharacter),
		character.Name,
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

	uc.logger.Info("Character ingested successfully", "character_id", input.CharacterID, "chunks", len(chunks))

	return &IngestCharacterOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildCharacterContent builds content string from character and traits
func (uc *IngestCharacterUseCase) buildCharacterContent(character *grpcclient.Character, traits []*grpcclient.CharacterTrait, world *grpcclient.World) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Character: %s", character.Name))
	if character.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", character.Description))
	}
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}

	if len(traits) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Traits:")
		for _, trait := range traits {
			traitLine := fmt.Sprintf("- %s: %s", trait.TraitName, trait.Value)
			if trait.Notes != "" {
				traitLine += fmt.Sprintf(" (%s)", trait.Notes)
			}
			parts = append(parts, traitLine)
		}
	}

	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with character metadata
func (uc *IngestCharacterUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, character *grpcclient.Character, world *grpcclient.World, traits []*grpcclient.CharacterTrait, content string) ([]*memory.Chunk, error) {
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
		chunk.WorldID = &character.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &character.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
