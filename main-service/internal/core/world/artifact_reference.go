package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidArtifactReferenceEntityType = errors.New("invalid artifact reference entity type")
)

// ArtifactReferenceEntityType represents the type of entity that references an artifact
type ArtifactReferenceEntityType string

const (
	ArtifactReferenceEntityTypeCharacter ArtifactReferenceEntityType = "character"
	ArtifactReferenceEntityTypeLocation  ArtifactReferenceEntityType = "location"
)

// ArtifactReference represents a reference from an artifact to an entity
type ArtifactReference struct {
	ID          uuid.UUID                    `json:"id"`
	ArtifactID  uuid.UUID                   `json:"artifact_id"`
	EntityType  ArtifactReferenceEntityType  `json:"entity_type"`
	EntityID    uuid.UUID                    `json:"entity_id"`
	CreatedAt   time.Time                    `json:"created_at"`
}

// NewArtifactReference creates a new artifact reference
func NewArtifactReference(artifactID uuid.UUID, entityType ArtifactReferenceEntityType, entityID uuid.UUID) (*ArtifactReference, error) {
	if !isValidArtifactReferenceEntityType(entityType) {
		return nil, ErrInvalidArtifactReferenceEntityType
	}

	return &ArtifactReference{
		ID:         uuid.New(),
		ArtifactID: artifactID,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt: time.Now(),
	}, nil
}

// Validate validates the artifact reference entity
func (r *ArtifactReference) Validate() error {
	if !isValidArtifactReferenceEntityType(r.EntityType) {
		return ErrInvalidArtifactReferenceEntityType
	}
	return nil
}

func isValidArtifactReferenceEntityType(entityType ArtifactReferenceEntityType) bool {
	return entityType == ArtifactReferenceEntityTypeCharacter ||
		entityType == ArtifactReferenceEntityTypeLocation
}

