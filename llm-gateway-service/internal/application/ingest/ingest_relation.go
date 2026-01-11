package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

// IngestRelationInput is the input for ingesting a relation.
type IngestRelationInput struct {
	TenantID   uuid.UUID
	RelationID uuid.UUID
}

// IngestRelationOutput is the output after ingesting a relation.
type IngestRelationOutput struct {
	DocumentID uuid.UUID
	ChunkCount int
}

// IngestRelationUseCase handles relation ingestion.
type IngestRelationUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	summaryGenerator  SummaryGenerator
	logger            *logger.Logger
}

// NewIngestRelationUseCase creates a new IngestRelationUseCase.
func NewIngestRelationUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	logger *logger.Logger,
) *IngestRelationUseCase {
	return &IngestRelationUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		summaryGenerator:  nil,
		logger:            logger,
	}
}

func (uc *IngestRelationUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

// Execute ingests a relation by fetching its data and generating embeddings.
func (uc *IngestRelationUseCase) Execute(ctx context.Context, input IngestRelationInput) (*IngestRelationOutput, error) {
	relation, err := uc.mainServiceClient.GetRelation(ctx, input.RelationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relation: %w", err)
	}

	source := uc.fetchEntityDetails(ctx, relation.SourceType, relation.SourceID)
	target := uc.fetchEntityDetails(ctx, relation.TargetType, relation.TargetID)
	relationKind := classifyRelationKind(relation.SourceType, relation.TargetType)

	title := fmt.Sprintf("%s: %s -> %s", capitalizeWord(relationKind), source.Name, target.Name)
	content := uc.buildRelationContent(relation, relationKind, source, target)

	doc := memory.NewDocument(
		input.TenantID,
		memory.SourceTypeRelation,
		input.RelationID,
		title,
		content,
	)
	applyRelationMetadata(doc, relation, relationKind, source, target)

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	existingDoc, err := uc.documentRepo.GetBySource(ctx, input.TenantID, memory.SourceTypeRelation, input.RelationID)
	if err == nil && existingDoc != nil {
		doc.ID = existingDoc.ID
		doc.CreatedAt = existingDoc.CreatedAt
		if err := uc.documentRepo.Update(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}
		if err := uc.chunkRepo.DeleteByDocument(ctx, doc.ID); err != nil {
			return nil, fmt.Errorf("failed to delete old chunks: %w", err)
		}
	} else {
		if err := uc.documentRepo.Create(ctx, doc); err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}
	}

	chunks, err := uc.chunkAndEmbed(ctx, doc.ID, relation, title, content)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk and embed: %w", err)
	}

	summaryContents := buildRelationSummaryContents(content, relation, source, target)
	sourceData, targetData, relationData := buildRelationSummaryDetails(relation, source, target)
	chunks, err = runIngestPipelineWithDetails(
		ctx,
		uc.logger,
		uc.embedder,
		uc.summaryGenerator,
		string(memory.SourceTypeRelation),
		title,
		summaryContents,
		sourceData,
		targetData,
		relationData,
		chunks,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to save chunks: %w", err)
	}

	uc.logger.Info("Relation ingested successfully", "relation_id", input.RelationID, "chunks", len(chunks))

	return &IngestRelationOutput{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

type relationEntityDetails struct {
	EntityType  string
	Name        string
	Subtype     string
	Description string
	WorldID     *uuid.UUID
}

func (uc *IngestRelationUseCase) fetchEntityDetails(ctx context.Context, entityType string, entityID uuid.UUID) relationEntityDetails {
	details := relationEntityDetails{
		EntityType: strings.TrimSpace(entityType),
		Name:       fmt.Sprintf("%s %s", entityType, entityID.String()),
	}

	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "story":
		if story, err := uc.mainServiceClient.GetStory(ctx, entityID); err == nil && story != nil {
			details.Name = story.Title
		}
	case "chapter":
		if chapter, err := uc.mainServiceClient.GetChapter(ctx, entityID); err == nil && chapter != nil {
			if strings.TrimSpace(chapter.Title) != "" {
				details.Name = chapter.Title
			} else {
				details.Name = fmt.Sprintf("Chapter %d", chapter.Number)
			}
		}
	case "scene":
		if scene, err := uc.mainServiceClient.GetScene(ctx, entityID); err == nil && scene != nil {
			if strings.TrimSpace(scene.Goal) != "" {
				details.Name = scene.Goal
			} else {
				details.Name = fmt.Sprintf("Scene %d", scene.OrderNum)
			}
			if strings.TrimSpace(scene.TimeRef) != "" {
				details.Description = scene.TimeRef
			}
		}
	case "beat":
		if beat, err := uc.mainServiceClient.GetBeat(ctx, entityID); err == nil && beat != nil {
			if strings.TrimSpace(beat.Type) != "" {
				details.Name = beat.Type
			} else {
				details.Name = fmt.Sprintf("Beat %d", beat.OrderNum)
			}
			if strings.TrimSpace(beat.Intent) != "" {
				details.Description = beat.Intent
			} else if strings.TrimSpace(beat.Outcome) != "" {
				details.Description = beat.Outcome
			}
		}
	case "content_block":
		if block, err := uc.mainServiceClient.GetContentBlock(ctx, entityID); err == nil && block != nil {
			details.Name = fmt.Sprintf("Content Block (%s)", block.Type)
			details.Description = shortText(block.Content, 160)
		}
	case "world":
		if world, err := uc.mainServiceClient.GetWorld(ctx, entityID); err == nil && world != nil {
			details.Name = world.Name
			if strings.TrimSpace(world.Genre) != "" {
				details.Subtype = world.Genre
			}
			details.WorldID = &world.ID
		}
	case "character":
		if character, err := uc.mainServiceClient.GetCharacter(ctx, entityID); err == nil && character != nil {
			details.Name = character.Name
			details.Description = character.Description
			details.WorldID = &character.WorldID
		}
	case "location":
		if location, err := uc.mainServiceClient.GetLocation(ctx, entityID); err == nil && location != nil {
			details.Name = location.Name
			if strings.TrimSpace(location.Type) != "" {
				details.Subtype = location.Type
			}
			details.Description = location.Description
			details.WorldID = &location.WorldID
		}
	case "event":
		if event, err := uc.mainServiceClient.GetEvent(ctx, entityID); err == nil && event != nil {
			details.Name = event.Name
			if event.Type != nil && strings.TrimSpace(*event.Type) != "" {
				details.Subtype = strings.TrimSpace(*event.Type)
			}
			if event.Description != nil {
				details.Description = strings.TrimSpace(*event.Description)
			}
			details.WorldID = &event.WorldID
		}
	case "artifact":
		if artifact, err := uc.mainServiceClient.GetArtifact(ctx, entityID); err == nil && artifact != nil {
			details.Name = artifact.Name
			if strings.TrimSpace(artifact.Rarity) != "" {
				details.Subtype = artifact.Rarity
			}
			details.Description = artifact.Description
			details.WorldID = &artifact.WorldID
		}
	case "faction":
		if faction, err := uc.mainServiceClient.GetFaction(ctx, entityID); err == nil && faction != nil {
			details.Name = faction.Name
			if faction.Type != nil && strings.TrimSpace(*faction.Type) != "" {
				details.Subtype = strings.TrimSpace(*faction.Type)
			}
			details.Description = faction.Description
			details.WorldID = &faction.WorldID
		}
	case "lore":
		if lore, err := uc.mainServiceClient.GetLore(ctx, entityID); err == nil && lore != nil {
			details.Name = lore.Name
			if lore.Category != nil && strings.TrimSpace(*lore.Category) != "" {
				details.Subtype = strings.TrimSpace(*lore.Category)
			}
			details.Description = lore.Description
			details.WorldID = &lore.WorldID
		}
	}

	if details.Name == "" {
		details.Name = fmt.Sprintf("%s %s", entityType, entityID.String())
	}

	return details
}

func (uc *IngestRelationUseCase) buildRelationContent(relation *grpcclient.EntityRelation, kind string, source relationEntityDetails, target relationEntityDetails) string {
	parts := []string{
		fmt.Sprintf("%s: %s -> %s", capitalizeWord(kind), source.Name, target.Name),
		fmt.Sprintf("Type: %s", relation.RelationType),
	}

	if relation.Summary != "" {
		parts = append(parts, fmt.Sprintf("Summary: %s", relation.Summary))
	}

	sourceLines := []string{
		fmt.Sprintf("Source: %s - %s", source.EntityType, source.Name),
	}
	if source.Subtype != "" {
		sourceLines = append(sourceLines, fmt.Sprintf("Subtype: %s", source.Subtype))
	}
	if source.Description != "" {
		sourceLines = append(sourceLines, fmt.Sprintf("Details: %s", source.Description))
	}
	parts = append(parts, strings.Join(sourceLines, "\n"))

	targetLines := []string{
		fmt.Sprintf("Target: %s - %s", target.EntityType, target.Name),
	}
	if target.Subtype != "" {
		targetLines = append(targetLines, fmt.Sprintf("Subtype: %s", target.Subtype))
	}
	if target.Description != "" {
		targetLines = append(targetLines, fmt.Sprintf("Details: %s", target.Description))
	}
	parts = append(parts, strings.Join(targetLines, "\n"))

	if relation.ContextType != nil && strings.TrimSpace(*relation.ContextType) != "" {
		contextLine := fmt.Sprintf("Context: %s", strings.TrimSpace(*relation.ContextType))
		if relation.ContextID != nil && strings.TrimSpace(*relation.ContextID) != "" {
			contextLine = fmt.Sprintf("%s %s", contextLine, strings.TrimSpace(*relation.ContextID))
		}
		parts = append(parts, contextLine)
	}

	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func (uc *IngestRelationUseCase) chunkAndEmbed(ctx context.Context, documentID uuid.UUID, relation *grpcclient.EntityRelation, title string, content string) ([]*memory.Chunk, error) {
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
		chunk.WorldID = &relation.WorldID
		if title != "" {
			chunk.EntityName = &title
		}

		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func applyRelationMetadata(doc *memory.Document, relation *grpcclient.EntityRelation, kind string, source relationEntityDetails, target relationEntityDetails) {
	if doc == nil || relation == nil {
		return
	}

	doc.Metadata["relation_type"] = relation.RelationType
	doc.Metadata["relation_kind"] = kind
	doc.Metadata["source_type"] = relation.SourceType
	doc.Metadata["source_id"] = relation.SourceID.String()
	doc.Metadata["source_name"] = source.Name
	doc.Metadata["target_type"] = relation.TargetType
	doc.Metadata["target_id"] = relation.TargetID.String()
	doc.Metadata["target_name"] = target.Name
	if source.Subtype != "" {
		doc.Metadata["source_subtype"] = source.Subtype
	}
	if target.Subtype != "" {
		doc.Metadata["target_subtype"] = target.Subtype
	}
	if source.WorldID != nil && *source.WorldID != uuid.Nil {
		doc.Metadata["source_world_id"] = source.WorldID.String()
	}
	if target.WorldID != nil && *target.WorldID != uuid.Nil {
		doc.Metadata["target_world_id"] = target.WorldID.String()
	}
	if relation.ContextType != nil && strings.TrimSpace(*relation.ContextType) != "" {
		doc.Metadata["context_type"] = strings.TrimSpace(*relation.ContextType)
	}
	if relation.ContextID != nil && strings.TrimSpace(*relation.ContextID) != "" {
		doc.Metadata["context_id"] = strings.TrimSpace(*relation.ContextID)
	}
	if relation.Summary != "" {
		doc.Metadata["relation_summary"] = relation.Summary
	}
}

func buildRelationSummaryContents(content string, relation *grpcclient.EntityRelation, source relationEntityDetails, target relationEntityDetails) []string {
	contents := make([]string, 0, 5)
	appendUnique := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		for _, existing := range contents {
			if existing == trimmed {
				return
			}
		}
		contents = append(contents, trimmed)
	}

	appendUnique(content)
	appendUnique(source.Description)
	appendUnique(target.Description)
	if relation != nil {
		appendUnique(relation.Summary)
	}

	return contents
}

func buildRelationSummaryDetails(relation *grpcclient.EntityRelation, source relationEntityDetails, target relationEntityDetails) (string, string, string) {
	sourceData, _ := json.Marshal(map[string]interface{}{
		"type":        source.EntityType,
		"id":          relation.SourceID.String(),
		"name":        source.Name,
		"subtype":     source.Subtype,
		"description": source.Description,
		"world_id":    formatUUIDPtr(source.WorldID),
	})

	targetData, _ := json.Marshal(map[string]interface{}{
		"type":        target.EntityType,
		"id":          relation.TargetID.String(),
		"name":        target.Name,
		"subtype":     target.Subtype,
		"description": target.Description,
		"world_id":    formatUUIDPtr(target.WorldID),
	})

	relationData := map[string]interface{}{
		"id":            relation.ID.String(),
		"type":          relation.RelationType,
		"summary":       relation.Summary,
		"context_type":  formatStringPtr(relation.ContextType),
		"context_id":    formatStringPtr(relation.ContextID),
		"source_type":   relation.SourceType,
		"source_id":     relation.SourceID.String(),
		"target_type":   relation.TargetType,
		"target_id":     relation.TargetID.String(),
		"relation_kind": classifyRelationKind(relation.SourceType, relation.TargetType),
	}
	relationDataJSON, _ := json.Marshal(relationData)

	return string(sourceData), string(targetData), string(relationDataJSON)
}

func formatStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func formatUUIDPtr(value *uuid.UUID) string {
	if value == nil || *value == uuid.Nil {
		return ""
	}
	return value.String()
}

func classifyRelationKind(sourceType string, targetType string) string {
	if isStoryEntityType(sourceType) && isWorldEntityType(targetType) {
		return "citation"
	}
	return "relationship"
}

func isStoryEntityType(entityType string) bool {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "story", "chapter", "scene", "beat", "content_block":
		return true
	default:
		return false
	}
}

func isWorldEntityType(entityType string) bool {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "world", "character", "location", "event", "artifact", "faction", "lore":
		return true
	default:
		return false
	}
}

func shortText(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || limit <= 0 {
		return ""
	}
	if len(trimmed) <= limit {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:limit]) + "..."
}

func capitalizeWord(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return strings.ToUpper(trimmed[:1]) + trimmed[1:]
}
