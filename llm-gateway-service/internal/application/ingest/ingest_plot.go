package ingest

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/llm/executor"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

//go:embed prompts/synthesize_plot_state.prompt
var synthesizePlotStatePrompt string

type IngestPlotInput struct {
	TenantID uuid.UUID
	PlotID   uuid.UUID
}

type IngestPlotOutput struct {
	PlotDocumentID uuid.UUID
	DevDocumentIDs []uuid.UUID
	ChunkCount     int
}

type IngestPlotUseCase struct {
	mainServiceClient grpcclient.MainServiceClient
	documentRepo      repositories.DocumentRepository
	chunkRepo         repositories.ChunkRepository
	embedder          embeddings.Embedder
	eventRepo         repositories.EventRepository
	summaryGenerator  SummaryGenerator
	executor          *executor.Executor
	llmProvider       string
	logger            *logger.Logger
}

func NewIngestPlotUseCase(
	mainServiceClient grpcclient.MainServiceClient,
	documentRepo repositories.DocumentRepository,
	chunkRepo repositories.ChunkRepository,
	embedder embeddings.Embedder,
	eventRepo repositories.EventRepository,
	logger *logger.Logger,
) *IngestPlotUseCase {
	return &IngestPlotUseCase{
		mainServiceClient: mainServiceClient,
		documentRepo:      documentRepo,
		chunkRepo:         chunkRepo,
		embedder:          embedder,
		eventRepo:         eventRepo,
		logger:            logger,
	}
}

func (uc *IngestPlotUseCase) SetSummaryGenerator(generator SummaryGenerator) {
	uc.summaryGenerator = generator
}

func (uc *IngestPlotUseCase) SetExecutor(exec *executor.Executor, provider string) {
	uc.executor = exec
	uc.llmProvider = provider
}

func (uc *IngestPlotUseCase) Execute(ctx context.Context, input IngestPlotInput) (*IngestPlotOutput, error) {
	plot, err := uc.mainServiceClient.GetPlot(ctx, input.PlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plot: %w", err)
	}

	// Fetch all developments
	devs, err := uc.mainServiceClient.ListPlotDevelopments(ctx, input.PlotID, 0)
	if err != nil {
		uc.logger.Warn("failed to fetch plot developments", "plot_id", input.PlotID, "error", err)
		devs = []*grpcclient.PlotDevelopment{}
	}

	output := &IngestPlotOutput{}
	totalChunks := 0

	// 1. Store each development as an event in the retrieval table
	for _, dev := range devs {
		description := fmt.Sprintf("[Plot: %s] %s: %s", plot.Title, dev.Action, dev.Description)
		event := memory.NewEvent(input.TenantID, plot.StoryID, dev.ChapterNumber, description)
		event.TimelineDay = dev.TimelineDay
		event.TimelineConfidence = dev.TimelineConfidence

		var involved []string
		for _, e := range plot.Entities {
			involved = append(involved, e.EntityID.String())
		}
		event.EntitiesInvolved = involved

		if err := uc.eventRepo.Create(ctx, event); err != nil {
			uc.logger.Error("failed to store plot development as event", "error", err, "plot_id", input.PlotID, "dev_id", dev.ID)
		}
	}

	// 2. Create versioned plot document (warm canon pattern)
	plotDocID, chunks, err := uc.createPlotDocument(ctx, input.TenantID, plot, devs)
	if err != nil {
		return nil, fmt.Errorf("failed to create plot document: %w", err)
	}
	output.PlotDocumentID = plotDocID
	totalChunks += chunks

	// 3. Create individual development documents
	for _, dev := range devs {
		devDocID, devChunks, err := uc.createDevelopmentDocument(ctx, input.TenantID, plot, dev)
		if err != nil {
			uc.logger.Error("failed to create development document", "error", err, "dev_id", dev.ID)
			continue
		}
		output.DevDocumentIDs = append(output.DevDocumentIDs, devDocID)
		totalChunks += devChunks
	}

	output.ChunkCount = totalChunks
	uc.logger.Info("Plot ingested",
		"plot_id", input.PlotID,
		"developments", len(devs),
		"total_chunks", totalChunks,
	)
	return output, nil
}

// createPlotDocument creates or versions a canonical plot document.
// On each ingestion: close old version (set valid_to_chapter), create new version with LLM-synthesized content.
func (uc *IngestPlotUseCase) createPlotDocument(ctx context.Context, tenantID uuid.UUID, plot *grpcclient.Plot, devs []*grpcclient.PlotDevelopment) (uuid.UUID, int, error) {
	title := fmt.Sprintf("Plot: %s", plot.Title)

	// Determine the latest chapter from developments
	latestChapter := 0
	for _, dev := range devs {
		if dev.ChapterNumber > latestChapter {
			latestChapter = dev.ChapterNumber
		}
	}

	// Generate content: LLM synthesis if available, mechanical fallback
	content, err := uc.synthesizePlotContent(ctx, plot, devs)
	if err != nil {
		uc.logger.Warn("LLM synthesis failed, using mechanical assembly", "error", err)
		content = uc.buildMechanicalContent(plot, devs)
	}

	// Close the existing version (if any)
	existingDoc, err := uc.documentRepo.GetBySource(ctx, tenantID, memory.SourceTypePlot, plot.ID)
	if err == nil && existingDoc != nil {
		// Close old version
		existingDoc.ValidToChapter = &latestChapter
		if err := uc.documentRepo.Update(ctx, existingDoc); err != nil {
			uc.logger.Error("failed to close old plot document version", "error", err)
		}
	}

	// Create new version
	doc := memory.NewDocument(tenantID, memory.SourceTypePlot, plot.ID, title, content)
	doc.IsCanonical = true
	doc.ValidFromChapter = latestChapter
	doc.StoryID = &plot.StoryID

	if err := doc.Validate(); err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid document: %w", err)
	}
	if err := uc.documentRepo.Create(ctx, doc); err != nil {
		return uuid.Nil, 0, fmt.Errorf("failed to create plot document: %w", err)
	}

	// Chunk, embed, and generate summaries
	chunks, err := uc.chunkAndEmbed(ctx, doc, content, title, string(memory.SourceTypePlot))
	if err != nil {
		return doc.ID, 0, fmt.Errorf("failed to chunk plot document: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return doc.ID, 0, fmt.Errorf("failed to save plot chunks: %w", err)
	}

	return doc.ID, len(chunks), nil
}

// createDevelopmentDocument creates a document for a single plot development.
func (uc *IngestPlotUseCase) createDevelopmentDocument(ctx context.Context, tenantID uuid.UUID, plot *grpcclient.Plot, dev *grpcclient.PlotDevelopment) (uuid.UUID, int, error) {
	title := fmt.Sprintf("Plot Development: %s — %s (Ch %d)", plot.Title, dev.Action, dev.ChapterNumber)

	content := uc.buildDevelopmentContent(plot, dev)

	doc := memory.NewDocument(tenantID, memory.SourceTypePlotDevelopment, dev.ID, title, content)
	chapterNum := dev.ChapterNumber
	doc.ChapterNumber = &chapterNum
	doc.StoryID = &plot.StoryID

	if err := doc.Validate(); err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid development document: %w", err)
	}

	// Upsert: check if already exists
	existingDoc, err := uc.documentRepo.GetBySource(ctx, tenantID, memory.SourceTypePlotDevelopment, dev.ID)
	if err == nil && existingDoc != nil {
		doc.ID = existingDoc.ID
		doc.CreatedAt = existingDoc.CreatedAt
		if err := uc.documentRepo.Update(ctx, doc); err != nil {
			return uuid.Nil, 0, fmt.Errorf("failed to update development document: %w", err)
		}
		if err := uc.chunkRepo.DeleteByDocument(ctx, doc.ID); err != nil {
			return uuid.Nil, 0, fmt.Errorf("failed to delete old development chunks: %w", err)
		}
	} else {
		if err := uc.documentRepo.Create(ctx, doc); err != nil {
			return uuid.Nil, 0, fmt.Errorf("failed to create development document: %w", err)
		}
	}

	chunks, err := uc.chunkAndEmbed(ctx, doc, content, title, string(memory.SourceTypePlotDevelopment))
	if err != nil {
		return doc.ID, 0, fmt.Errorf("failed to chunk development document: %w", err)
	}

	if err := uc.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return doc.ID, 0, fmt.Errorf("failed to save development chunks: %w", err)
	}

	return doc.ID, len(chunks), nil
}

// synthesizePlotContent uses the LLM to generate a rich narrative summary of the plot state.
func (uc *IngestPlotUseCase) synthesizePlotContent(ctx context.Context, plot *grpcclient.Plot, devs []*grpcclient.PlotDevelopment) (string, error) {
	if uc.executor == nil || uc.llmProvider == "" {
		return "", fmt.Errorf("LLM executor not configured")
	}

	prompt := synthesizePlotStatePrompt
	prompt = strings.ReplaceAll(prompt, "{{plot_title}}", plot.Title)
	prompt = strings.ReplaceAll(prompt, "{{plot_type}}", plot.Type)
	prompt = strings.ReplaceAll(prompt, "{{plot_status}}", plot.Status)
	prompt = strings.ReplaceAll(prompt, "{{plot_description}}", plot.Description)
	prompt = strings.ReplaceAll(prompt, "{{plot_entities}}", formatPlotEntities(plot.Entities))
	prompt = strings.ReplaceAll(prompt, "{{plot_developments}}", formatPlotDevelopments(devs))

	result, err := uc.executor.Submit(ctx, prompt, uc.llmProvider)
	if err != nil {
		return "", fmt.Errorf("LLM synthesis failed: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return "", fmt.Errorf("LLM returned empty synthesis")
	}

	return result, nil
}

// buildMechanicalContent is the fallback when LLM is unavailable.
func (uc *IngestPlotUseCase) buildMechanicalContent(plot *grpcclient.Plot, devs []*grpcclient.PlotDevelopment) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Plot: %s (%s, %s)", plot.Title, plot.Type, plot.Status))
	if plot.Description != "" {
		parts = append(parts, plot.Description)
	}

	if len(plot.Entities) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Entities involved:")
		for _, e := range plot.Entities {
			parts = append(parts, fmt.Sprintf("- %s: %s (%s)", e.Role, e.EntityID.String(), e.EntityType))
		}
	}

	if len(devs) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Developments:")
		for _, d := range devs {
			header := fmt.Sprintf("Chapter %d — %s", d.ChapterNumber, d.Action)
			if d.TimelineDay != nil {
				header += fmt.Sprintf(" (Day %d)", *d.TimelineDay)
			}
			parts = append(parts, header)
			if d.Description != "" {
				parts = append(parts, d.Description)
			}
			if d.Notes != "" {
				parts = append(parts, fmt.Sprintf("Notes: %s", d.Notes))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// buildDevelopmentContent builds the content for a single development document.
func (uc *IngestPlotUseCase) buildDevelopmentContent(plot *grpcclient.Plot, dev *grpcclient.PlotDevelopment) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Plot: %s — %s at Chapter %d", plot.Title, dev.Action, dev.ChapterNumber))

	if dev.TimelineDay != nil {
		confidence := "unknown"
		if dev.TimelineConfidence != nil {
			confidence = *dev.TimelineConfidence
		}
		parts = append(parts, fmt.Sprintf("Timeline: Day %d (confidence: %s)", *dev.TimelineDay, confidence))
	}

	if dev.Description != "" {
		parts = append(parts, "")
		parts = append(parts, dev.Description)
	}

	if dev.Notes != "" {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Notes: %s", dev.Notes))
	}

	if dev.TimelineEventID != nil {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Linked timeline event: %s", dev.TimelineEventID.String()))
	}

	if len(plot.Entities) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Plot entities:")
		for _, e := range plot.Entities {
			parts = append(parts, fmt.Sprintf("- %s: %s (%s)", e.Role, e.EntityID.String(), e.EntityType))
		}
	}

	return strings.Join(parts, "\n")
}

// chunkAndEmbed splits content into paragraphs, embeds them, and runs the summary pipeline.
func (uc *IngestPlotUseCase) chunkAndEmbed(ctx context.Context, doc *memory.Document, content, title, sourceType string) ([]*memory.Chunk, error) {
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
		chunk := memory.NewChunk(doc.ID, i, para, embedding, len(para)/4)
		if err := chunk.Validate(); err != nil {
			return nil, fmt.Errorf("invalid chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	// Run summary pipeline
	chunks, err := runIngestPipeline(ctx, uc.logger, uc.embedder, uc.summaryGenerator,
		sourceType, title, []string{content}, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to run ingest pipeline: %w", err)
	}

	return chunks, nil
}

func formatPlotEntities(entities []grpcclient.PlotEntity) string {
	if len(entities) == 0 {
		return "None specified"
	}
	var lines []string
	for _, e := range entities {
		lines = append(lines, fmt.Sprintf("- %s: %s (%s)", e.Role, e.EntityID.String(), e.EntityType))
	}
	return strings.Join(lines, "\n")
}

func formatPlotDevelopments(devs []*grpcclient.PlotDevelopment) string {
	if len(devs) == 0 {
		return "No developments yet"
	}
	var lines []string
	for _, d := range devs {
		header := fmt.Sprintf("### Chapter %d — %s", d.ChapterNumber, d.Action)
		if d.TimelineDay != nil {
			header += fmt.Sprintf(" (Day %d)", *d.TimelineDay)
		}
		lines = append(lines, header)
		if d.Description != "" {
			lines = append(lines, d.Description)
		}
		if d.Notes != "" {
			lines = append(lines, fmt.Sprintf("Notes: %s", d.Notes))
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
