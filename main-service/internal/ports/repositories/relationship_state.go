package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// RelationshipStateRepository defines the interface for relationship state persistence
type RelationshipStateRepository interface {
	Create(ctx context.Context, rs *world.RelationshipState) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.RelationshipState, error)
	GetLatestByEntityPair(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, targetType string, targetID uuid.UUID, asOfChapter int) (*world.RelationshipState, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*world.RelationshipState, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// Entity links
	AddEntity(ctx context.Context, entity *world.RelationshipStateEntity) error
	ListEntities(ctx context.Context, relationshipStateID uuid.UUID) ([]*world.RelationshipStateEntity, error)
}
