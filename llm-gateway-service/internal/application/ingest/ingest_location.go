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

// IngestLocationInput is the input for ingesting a location
type IngestLocationInput struct {
	TenantID   uuid.UUID
	LocationID uuid.UUID
}

// IngestLocationOutput is the output after ingesting a location
type IngestLocationOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestLocationUseCase handles location ingestion with hierarchy
type IngestLocationUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestLocationUseCase creates a new IngestLocationUseCase
func NewIngestLocationUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestLocationUseCase {
	return &IngestLocationUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestLocationUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a location by fetching its content, hierarchy, and generating embeddings
func (uc *IngestLocationUseCase) Execute(ctx context.Context, input IngestLocationInput) (*IngestLocationOutput, error) {
	// Fetch location from main-service
	location, err := uc.mainServiceClient.GetLocation(ctx, input.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location: %w", err)
	}

	// Build hierarchy path
	hierarchyPath := uc.buildHierarchyPath(ctx, location)

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, location.WorldID)
	if err != nil {
		uc.logger.Warn("failed to fetch world", "world_id", location.WorldID, "error", err)
	}

	// Build content from location and hierarchy
	content := uc.buildLocationContent(location, hierarchyPath, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeLocation,
		input.LocationID,
		location.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeLocation, input.LocationID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, location, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}
	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeLocation,
		input.LocationID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeLocation),
		location.Name,
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

	uc.logger.Info("Location ingested successfully", "location_id", input.LocationID, "chunks", len(chunks))

	return &IngestLocationOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildHierarchyPath builds the hierarchy path for a location
func (uc *IngestLocationUseCase) buildHierarchyPath(ctx context.Context, location *grpcclient.Location) []string {
	var path []string
	current := location

	// Build path by traversing up the hierarchy
	for current != nil {
		path = append([]string{current.Name}, path...)
		if current.ParentID == nil {
			break
		}
		parent, err := uc.mainServiceClient.GetLocation(ctx, *current.ParentID)
		if err != nil {
			uc.logger.Warn("failed to fetch parent location", "parent_id", *current.ParentID, "error", err)
			break
		}
		current = parent
	}

	return path
}

// buildLocationContent builds content string from location and hierarchy
func (uc *IngestLocationUseCase) buildLocationContent(location *grpcclient.Location, hierarchyPath []string, world *grpcclient.World) string {
	var parts []string
	baseName := location.Name
	if len(hierarchyPath) > 1 {
		baseName = strings.Join(hierarchyPath, " > ")
	}
	header := fmt.Sprintf("Location: %s", baseName)
	if location.Description != "" {
		header = fmt.Sprintf("Location: %s - %s", baseName, location.Description)
	}
	parts = append(parts, header)
	if location.Type != "" {
		parts = append(parts, fmt.Sprintf("Type: %s", location.Type))
	}
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}
	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with location metadata
func (uc *IngestLocationUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, location *grpcclient.Location, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
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
		chunk.WorldID = &location.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &location.Name

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
