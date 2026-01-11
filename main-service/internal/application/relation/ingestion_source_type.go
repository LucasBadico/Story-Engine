package relation

import (
	"strings"

	corerelation "github.com/story-engine/main-service/internal/core/relation"
)

const (
	relationQueueType         = "relation"
	relationCitationQueueType = "relation_citation"
)

func relationIngestionSourceType(rel *corerelation.EntityRelation) string {
	if rel == nil {
		return relationQueueType
	}
	if isStoryEntityType(rel.SourceType) && isWorldEntityType(rel.TargetType) {
		return relationCitationQueueType
	}
	return relationQueueType
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
