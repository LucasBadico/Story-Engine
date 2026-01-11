package relation

import (
	"testing"

	"github.com/google/uuid"
	corerelation "github.com/story-engine/main-service/internal/core/relation"
)

func TestRelationIngestionSourceType(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		targetType string
		wantType   string
	}{
		{
			name:       "story_to_world_is_citation",
			sourceType: "scene",
			targetType: "character",
			wantType:   relationCitationQueueType,
		},
		{
			name:       "world_to_world_is_relation",
			sourceType: "character",
			targetType: "faction",
			wantType:   relationQueueType,
		},
		{
			name:       "world_to_story_is_relation",
			sourceType: "character",
			targetType: "scene",
			wantType:   relationQueueType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := &corerelation.EntityRelation{
				ID:           uuid.New(),
				TenantID:     uuid.New(),
				WorldID:      uuid.New(),
				SourceType:   tt.sourceType,
				SourceID:     uuid.New(),
				TargetType:   tt.targetType,
				TargetID:     uuid.New(),
				RelationType: "mentions",
			}

			if got := relationIngestionSourceType(rel); got != tt.wantType {
				t.Fatalf("expected %q, got %q", tt.wantType, got)
			}
		})
	}
}
