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

// IngestEventInput is the input for ingesting an event
type IngestEventInput struct {
	TenantID uuid.UUID
	EventID  uuid.UUID
}

// IngestEventOutput is the output after ingesting an event
type IngestEventOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestEventUseCase handles event ingestion with relationships
type IngestEventUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestEventUseCase creates a new IngestEventUseCase
func NewIngestEventUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestEventUseCase {
	return &IngestEventUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestEventUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests an event by fetching its content, relationships, and generating embeddings
func (uc *IngestEventUseCase) Execute(ctx context.Context, input IngestEventInput) (*IngestEventOutput, error) {
	// Fetch event from main-service
	event, err := uc.mainServiceClient.GetEvent(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	// Fetch related entities
	eventCharacters, err := uc.mainServiceClient.GetEventCharacters(ctx, input.EventID)
	if err != nil {
		uc.logger.Warn("failed to fetch event characters", "event_id", input.EventID, "error", err)
		eventCharacters = []*grpcclient.EventCharacter{}
	}

	eventLocations, err := uc.mainServiceClient.GetEventLocations(ctx, input.EventID)
	if err != nil {
		uc.logger.Warn("failed to fetch event locations", "event_id", input.EventID, "error", err)
		eventLocations = []*grpcclient.EventLocation{}
	}

	eventArtifacts, err := uc.mainServiceClient.GetEventArtifacts(ctx, input.EventID)
	if err != nil {
		uc.logger.Warn("failed to fetch event artifacts", "event_id", input.EventID, "error", err)
		eventArtifacts = []*grpcclient.EventArtifact{}
	}

	// Fetch world to get world metadata
	world, err := uc.mainServiceClient.GetWorld(ctx, event.WorldID)
	if err != nil {
		uc.logger.Warn("failed to fetch world", "world_id", event.WorldID, "error", err)
	}

	// Build content from event and relationships
	content := uc.buildEventContent(event, eventCharacters, eventLocations, eventArtifacts, world)

	// Create or update document
	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeEvent,
		input.EventID,
		event.Name,
		content,
	)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// Check if document already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeEvent, input.EventID)
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
	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, event, eventCharacters, eventLocations, eventArtifacts, world, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}
	summaryContents := collectSummaryContents(
		ctx,
		uc.mainServiceClient,
		memory.SourceTypeEvent,
		input.EventID,
		content,
		uc.logger,
	)
	chunks, err = runIngestPipeline(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeEvent),
		event.Name,
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

	uc.logger.Info("Event ingested successfully", "event_id", input.EventID, "chunks", len(chunks))

	return &IngestEventOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

// buildEventContent builds content string from event and relationships
func (uc *IngestEventUseCase) buildEventContent(event *grpcclient.Event, eventCharacters []*grpcclient.EventCharacter, eventLocations []*grpcclient.EventLocation, eventArtifacts []*grpcclient.EventArtifact, world *grpcclient.World) string {
	var parts []string
	header := fmt.Sprintf("Event: %s", event.Name)
	if event.Description != nil && *event.Description != "" {
		header = fmt.Sprintf("Event: %s - %s", event.Name, *event.Description)
	}
	parts = append(parts, header)
	if event.Type != nil && *event.Type != "" {
		parts = append(parts, fmt.Sprintf("Type: %s", *event.Type))
	}
	if event.Timeline != nil && *event.Timeline != "" {
		parts = append(parts, fmt.Sprintf("Timeline: %s", *event.Timeline))
	}
	parts = append(parts, fmt.Sprintf("Importance: %d/10", event.Importance))
	if world != nil {
		parts = append(parts, fmt.Sprintf("World: %s", world.Name))
	}

	if len(eventCharacters) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Characters:")
		for _, ec := range eventCharacters {
			char, err := uc.mainServiceClient.GetCharacter(context.Background(), ec.CharacterID)
			if err == nil {
				role := ""
				if ec.Role != nil {
					role = fmt.Sprintf(" (%s)", *ec.Role)
				}
				parts = append(parts, fmt.Sprintf("- %s%s", char.Name, role))
			}
		}
	}

	if len(eventLocations) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Locations:")
		for _, el := range eventLocations {
			loc, err := uc.mainServiceClient.GetLocation(context.Background(), el.LocationID)
			if err == nil {
				sig := ""
				if el.Significance != nil {
					sig = fmt.Sprintf(" (%s)", *el.Significance)
				}
				parts = append(parts, fmt.Sprintf("- %s%s", loc.Name, sig))
			}
		}
	}

	if len(eventArtifacts) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Artifacts:")
		for _, ea := range eventArtifacts {
			art, err := uc.mainServiceClient.GetArtifact(context.Background(), ea.ArtifactID)
			if err == nil {
				role := ""
				if ea.Role != nil {
					role = fmt.Sprintf(" (%s)", *ea.Role)
				}
				parts = append(parts, fmt.Sprintf("- %s%s", art.Name, role))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// chunkAndEmbed chunks content and generates embeddings with event metadata
func (uc *IngestEventUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, event *grpcclient.Event, eventCharacters []*grpcclient.EventCharacter, eventLocations []*grpcclient.EventLocation, eventArtifacts []*grpcclient.EventArtifact, world *grpcclient.World, content string) ([]*memory.Chunk, error) {
	paragraphs := strings.Split(content, "\n\n")
	chunks := make([]*memory.Chunk, 0, len(paragraphs))

	// Collect related entity names
	relatedChars := make([]string, 0, len(eventCharacters))
	for _, ec := range eventCharacters {
		char, err := uc.mainServiceClient.GetCharacter(ctx, ec.CharacterID)
		if err == nil {
			relatedChars = append(relatedChars, char.Name)
		}
	}

	relatedLocs := make([]string, 0, len(eventLocations))
	for _, el := range eventLocations {
		loc, err := uc.mainServiceClient.GetLocation(ctx, el.LocationID)
		if err == nil {
			relatedLocs = append(relatedLocs, loc.Name)
		}
	}

	relatedArts := make([]string, 0, len(eventArtifacts))
	for _, ea := range eventArtifacts {
		art, err := uc.mainServiceClient.GetArtifact(ctx, ea.ArtifactID)
		if err == nil {
			relatedArts = append(relatedArts, art.Name)
		}
	}

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
		chunk.WorldID = &event.WorldID
		if world != nil {
			chunk.WorldName = &world.Name
			if world.Genre != "" {
				chunk.WorldGenre = &world.Genre
			}
		}
		chunk.EntityName = &event.Name
		if event.Timeline != nil {
			chunk.EventTimeline = event.Timeline
		}
		importance := event.Importance
		chunk.Importance = &importance
		chunk.RelatedCharacters = relatedChars
		chunk.RelatedLocations = relatedLocs
		chunk.RelatedArtifacts = relatedArts

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
