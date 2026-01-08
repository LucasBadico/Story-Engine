package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
)

type SummaryGenerator interface {
	Execute(ctx context.Context, input GenerateSummaryInput) (GenerateSummaryOutput, error)
}

type IngestPipelineContext struct {
	DocumentID       uuid.UUID
	EntityType       string
	EntityName       string
	Contents         []string
	Context          string
	Chunks           []*memory.Chunk
	TemplateChunk    *memory.Chunk
	NextChunkIndex   int
	Embedder         embeddings.Embedder
	SummaryGenerator SummaryGenerator
	Logger           *logger.Logger
}

type IngestPipelineStep interface {
	Name() string
	Execute(ctx context.Context, input *IngestPipelineContext) error
}

type IngestPipeline struct {
	steps  []IngestPipelineStep
	logger *logger.Logger
}

func NewIngestPipeline(logger *logger.Logger, steps ...IngestPipelineStep) *IngestPipeline {
	return &IngestPipeline{
		steps:  steps,
		logger: logger,
	}
}

func (p *IngestPipeline) Execute(ctx context.Context, input *IngestPipelineContext) error {
	for _, step := range p.steps {
		if err := step.Execute(ctx, input); err != nil {
			return fmt.Errorf("ingest step %s failed: %w", step.Name(), err)
		}
	}
	return nil
}

func runIngestPipeline(
	ctx context.Context,
	log *logger.Logger,
	embedder embeddings.Embedder,
	summaryGenerator SummaryGenerator,
	entityType string,
	entityName string,
	contents []string,
	chunks []*memory.Chunk,
) ([]*memory.Chunk, error) {
	if log == nil || embedder == nil {
		return chunks, nil
	}

	var templateChunk *memory.Chunk
	if len(chunks) > 0 {
		templateChunk = chunks[0]
	}

	pipeline := NewIngestPipeline(log, SummaryIngestStep{})
	pipeCtx := &IngestPipelineContext{
		DocumentID:       uuid.Nil,
		EntityType:       entityType,
		EntityName:       entityName,
		Contents:         contents,
		Context:          "",
		Chunks:           chunks,
		TemplateChunk:    templateChunk,
		NextChunkIndex:   len(chunks),
		Embedder:         embedder,
		SummaryGenerator: summaryGenerator,
		Logger:           log,
	}

	if len(chunks) > 0 {
		pipeCtx.DocumentID = chunks[0].DocumentID
	}

	if err := pipeline.Execute(ctx, pipeCtx); err != nil {
		return nil, err
	}

	return pipeCtx.Chunks, nil
}

type SummaryIngestStep struct {
	MaxItems int
}

func (s SummaryIngestStep) Name() string {
	return "summary"
}

func (s SummaryIngestStep) Execute(ctx context.Context, input *IngestPipelineContext) error {
	if input == nil || input.SummaryGenerator == nil || input.Embedder == nil {
		return nil
	}
	if input.TemplateChunk == nil || input.EntityName == "" || len(input.Contents) == 0 {
		return nil
	}

	maxItems := s.MaxItems
	if maxItems <= 0 {
		maxItems = 3
	}

	output, err := input.SummaryGenerator.Execute(ctx, GenerateSummaryInput{
		EntityType: input.EntityType,
		Name:       input.EntityName,
		Contents:   input.Contents,
		Context:    input.Context,
		MaxItems:   maxItems,
	})
	if err != nil {
		return err
	}

	if len(output.Summaries) == 0 {
		return nil
	}

	for _, summary := range output.Summaries {
		embedding, err := input.Embedder.EmbedText(summary)
		if err != nil {
			return fmt.Errorf("failed to generate summary embedding: %w", err)
		}

		tokenCount := len(summary) / 4
		chunk := memory.NewChunk(input.DocumentID, input.NextChunkIndex, summary, embedding, tokenCount)
		copyChunkMetadata(chunk, input.TemplateChunk)
		setSummaryMetadata(chunk, summary)

		if err := chunk.Validate(); err != nil {
			return fmt.Errorf("invalid summary chunk: %w", err)
		}

		input.Chunks = append(input.Chunks, chunk)
		input.NextChunkIndex++
	}

	return nil
}

func setSummaryMetadata(chunk *memory.Chunk, summary string) {
	if chunk == nil {
		return
	}
	chunkType := "summary"
	chunk.ChunkType = &chunkType
	if value := strings.TrimSpace(summary); value != "" {
		chunk.EmbedText = &value
	}
}

func copyChunkMetadata(dst *memory.Chunk, src *memory.Chunk) {
	if dst == nil || src == nil {
		return
	}

	if src.SceneID != nil {
		val := *src.SceneID
		dst.SceneID = &val
	}
	if src.BeatID != nil {
		val := *src.BeatID
		dst.BeatID = &val
	}
	if src.BeatType != nil {
		val := *src.BeatType
		dst.BeatType = &val
	}
	if src.BeatIntent != nil {
		val := *src.BeatIntent
		dst.BeatIntent = &val
	}
	if src.LocationID != nil {
		val := *src.LocationID
		dst.LocationID = &val
	}
	if src.LocationName != nil {
		val := *src.LocationName
		dst.LocationName = &val
	}
	if src.Timeline != nil {
		val := *src.Timeline
		dst.Timeline = &val
	}
	if src.POVCharacter != nil {
		val := *src.POVCharacter
		dst.POVCharacter = &val
	}
	if src.ContentType != nil {
		val := *src.ContentType
		dst.ContentType = &val
	}
	if src.ContentKind != nil {
		val := *src.ContentKind
		dst.ContentKind = &val
	}
	if src.WorldID != nil {
		val := *src.WorldID
		dst.WorldID = &val
	}
	if src.WorldName != nil {
		val := *src.WorldName
		dst.WorldName = &val
	}
	if src.WorldGenre != nil {
		val := *src.WorldGenre
		dst.WorldGenre = &val
	}
	if src.EntityName != nil {
		val := *src.EntityName
		dst.EntityName = &val
	}
	if src.EventTimeline != nil {
		val := *src.EventTimeline
		dst.EventTimeline = &val
	}
	if src.Importance != nil {
		val := *src.Importance
		dst.Importance = &val
	}

	if len(src.Characters) > 0 {
		dst.Characters = append([]string(nil), src.Characters...)
	}
	if len(src.RelatedCharacters) > 0 {
		dst.RelatedCharacters = append([]string(nil), src.RelatedCharacters...)
	}
	if len(src.RelatedLocations) > 0 {
		dst.RelatedLocations = append([]string(nil), src.RelatedLocations...)
	}
	if len(src.RelatedArtifacts) > 0 {
		dst.RelatedArtifacts = append([]string(nil), src.RelatedArtifacts...)
	}
	if len(src.RelatedEvents) > 0 {
		dst.RelatedEvents = append([]string(nil), src.RelatedEvents...)
	}
}
