package ingest

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
)

func collectSummaryContents(
	ctx context.Context,
	client grpcclient.MainServiceClient,
	entityType memory.SourceType,
	entityID uuid.UUID,
	baseContent string,
	log *logger.Logger,
) []string {
	contents := make([]string, 0, 4)
	seen := make(map[string]struct{})

	appendUnique := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		contents = append(contents, trimmed)
	}

	appendUnique(baseContent)

	if client == nil || entityID == uuid.Nil {
		return contents
	}
	if entityType == memory.SourceTypeContentBlock {
		return contents
	}

	entityTypeValue := strings.TrimSpace(string(entityType))
	if entityTypeValue == "" {
		return contents
	}

	blocks, err := client.ListContentBlocksByEntity(ctx, entityTypeValue, entityID)
	if err != nil {
		if log != nil {
			log.Warn("failed to fetch content blocks for summary", "entity_type", entityTypeValue, "entity_id", entityID, "error", err)
		}
		return contents
	}

	for _, block := range blocks {
		if block == nil {
			continue
		}
		appendUnique(block.Content)
	}

	return contents
}
