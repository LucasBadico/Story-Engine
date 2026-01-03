package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/application/ingest"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/queue"
)

// DebouncedWorker processes ingestion queue items with debounce
type DebouncedWorker struct {
	queue              queue.IngestionQueue
	ingestStory        *ingest.IngestStoryUseCase
	ingestChapter      *ingest.IngestChapterUseCase
	ingestContentBlock *ingest.IngestContentBlockUseCase
	ingestWorld        *ingest.IngestWorldUseCase
	ingestCharacter    *ingest.IngestCharacterUseCase
	ingestLocation     *ingest.IngestLocationUseCase
	ingestEvent        *ingest.IngestEventUseCase
	ingestArtifact     *ingest.IngestArtifactUseCase
	logger             *logger.Logger
	debounceInterval   time.Duration
	pollInterval       time.Duration
	batchSize          int
}

// NewDebouncedWorker creates a new debounced worker
func NewDebouncedWorker(
	queue queue.IngestionQueue,
	ingestStory *ingest.IngestStoryUseCase,
	ingestChapter *ingest.IngestChapterUseCase,
	ingestContentBlock *ingest.IngestContentBlockUseCase,
	ingestWorld *ingest.IngestWorldUseCase,
	ingestCharacter *ingest.IngestCharacterUseCase,
	ingestLocation *ingest.IngestLocationUseCase,
	ingestEvent *ingest.IngestEventUseCase,
	ingestArtifact *ingest.IngestArtifactUseCase,
	logger *logger.Logger,
	cfg *config.Config,
) *DebouncedWorker {
	return &DebouncedWorker{
		queue:              queue,
		ingestStory:        ingestStory,
		ingestChapter:      ingestChapter,
		ingestContentBlock: ingestContentBlock,
		ingestWorld:        ingestWorld,
		ingestCharacter:    ingestCharacter,
		ingestLocation:     ingestLocation,
		ingestEvent:        ingestEvent,
		ingestArtifact:     ingestArtifact,
		logger:             logger,
		debounceInterval:   time.Duration(cfg.Worker.DebounceMinutes) * time.Minute,
		pollInterval:       time.Duration(cfg.Worker.PollSeconds) * time.Second,
		batchSize:          cfg.Worker.BatchSize,
	}
}

// Run starts the worker loop
func (w *DebouncedWorker) Run(ctx context.Context) {
	w.logger.Info("Starting debounced worker",
		"debounce_interval", w.debounceInterval,
		"poll_interval", w.pollInterval,
		"batch_size", w.batchSize)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Process immediately on start
	w.processStableItems(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker shutting down")
			return
		case <-ticker.C:
			w.processStableItems(ctx)
		}
	}
}

// processStableItems finds and processes items that haven't been updated recently
func (w *DebouncedWorker) processStableItems(ctx context.Context) {
	stableAt := time.Now().Add(-w.debounceInterval)

	w.logger.Debug("Processing stable items", "stable_at", stableAt)

	// Get list of tenants that have items in queue
	tenantIDs, err := w.queue.ListTenantsWithItems(ctx)
	if err != nil {
		w.logger.Error("Failed to list tenants with items", "error", err)
		return
	}

	if len(tenantIDs) == 0 {
		return
	}

	w.logger.Debug("Found tenants with items", "count", len(tenantIDs))

	// Process each tenant's queue
	for _, tenantID := range tenantIDs {
		if err := w.processTenantQueue(ctx, tenantID); err != nil {
			w.logger.Error("Failed to process tenant queue",
				"tenant_id", tenantID,
				"error", err)
			// Continue with other tenants
			continue
		}
	}
}

// processTenantQueue processes queue for a specific tenant
func (w *DebouncedWorker) processTenantQueue(ctx context.Context, tenantID uuid.UUID) error {
	stableAt := time.Now().Add(-w.debounceInterval)

	items, err := w.queue.PopStable(ctx, tenantID, stableAt, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to pop stable items: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	w.logger.Info("Processing stable items", "tenant_id", tenantID, "count", len(items))

	for _, item := range items {
		if err := w.processItem(ctx, item); err != nil {
			w.logger.Error("Failed to process item",
				"tenant_id", item.TenantID,
				"source_type", item.SourceType,
				"source_id", item.SourceID,
				"error", err)
			// Continue processing other items
			continue
		}
	}

	return nil
}

// processItem processes a single queue item
func (w *DebouncedWorker) processItem(ctx context.Context, item *queue.QueueItem) error {
	switch memory.SourceType(item.SourceType) {
	case memory.SourceTypeStory:
		_, err := w.ingestStory.Execute(ctx, ingest.IngestStoryInput{
			TenantID: item.TenantID,
			StoryID:  item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest story: %w", err)
		}
		w.logger.Info("Successfully ingested story", "story_id", item.SourceID)

	case memory.SourceTypeChapter:
		_, err := w.ingestChapter.Execute(ctx, ingest.IngestChapterInput{
			TenantID:  item.TenantID,
			ChapterID: item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest chapter: %w", err)
		}
		w.logger.Info("Successfully ingested chapter", "chapter_id", item.SourceID)

	case memory.SourceTypeContentBlock:
		_, err := w.ingestContentBlock.Execute(ctx, ingest.IngestContentBlockInput{
			TenantID:       item.TenantID,
			ContentBlockID: item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest content block: %w", err)
		}
		w.logger.Info("Successfully ingested content block", "content_block_id", item.SourceID)

	case memory.SourceTypeWorld:
		_, err := w.ingestWorld.Execute(ctx, ingest.IngestWorldInput{
			TenantID: item.TenantID,
			WorldID:  item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest world: %w", err)
		}
		w.logger.Info("Successfully ingested world", "world_id", item.SourceID)

	case memory.SourceTypeCharacter:
		_, err := w.ingestCharacter.Execute(ctx, ingest.IngestCharacterInput{
			TenantID:    item.TenantID,
			CharacterID: item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest character: %w", err)
		}
		w.logger.Info("Successfully ingested character", "character_id", item.SourceID)

	case memory.SourceTypeLocation:
		_, err := w.ingestLocation.Execute(ctx, ingest.IngestLocationInput{
			TenantID:   item.TenantID,
			LocationID: item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest location: %w", err)
		}
		w.logger.Info("Successfully ingested location", "location_id", item.SourceID)

	case memory.SourceTypeEvent:
		_, err := w.ingestEvent.Execute(ctx, ingest.IngestEventInput{
			TenantID: item.TenantID,
			EventID:  item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest event: %w", err)
		}
		w.logger.Info("Successfully ingested event", "event_id", item.SourceID)

	case memory.SourceTypeArtifact:
		_, err := w.ingestArtifact.Execute(ctx, ingest.IngestArtifactInput{
			TenantID:   item.TenantID,
			ArtifactID: item.SourceID,
		})
		if err != nil {
			return fmt.Errorf("failed to ingest artifact: %w", err)
		}
		w.logger.Info("Successfully ingested artifact", "artifact_id", item.SourceID)

	default:
		return fmt.Errorf("unsupported source type: %s", item.SourceType)
	}

	return nil
}

