package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// ArtifactRepository defines the interface for artifact persistence
type ArtifactRepository interface {
	Create(ctx context.Context, a *world.Artifact) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Artifact, error)
	ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Artifact, error)
	Update(ctx context.Context, a *world.Artifact) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error)
}

// ArtifactReferenceRepository defines the interface for artifact reference persistence
type ArtifactReferenceRepository interface {
	Create(ctx context.Context, ref *world.ArtifactReference) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.ArtifactReference, error)
	ListByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) ([]*world.ArtifactReference, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) ([]*world.ArtifactReference, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) error
	DeleteByArtifactAndEntity(ctx context.Context, tenantID, artifactID uuid.UUID, entityType world.ArtifactReferenceEntityType, entityID uuid.UUID) error
}

