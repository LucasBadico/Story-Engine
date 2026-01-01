package story

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSceneReferenceEntityType = errors.New("invalid scene reference entity type")
)

// SceneReferenceEntityType represents the type of entity that references a scene
type SceneReferenceEntityType string

const (
	SceneReferenceEntityTypeCharacter SceneReferenceEntityType = "character"
	SceneReferenceEntityTypeLocation  SceneReferenceEntityType = "location"
	SceneReferenceEntityTypeArtifact  SceneReferenceEntityType = "artifact"
)

// SceneReference represents a reference from a scene to an entity
type SceneReference struct {
	ID         uuid.UUID                   `json:"id"`
	SceneID    uuid.UUID                   `json:"scene_id"`
	EntityType SceneReferenceEntityType     `json:"entity_type"`
	EntityID   uuid.UUID                   `json:"entity_id"`
	CreatedAt  time.Time                   `json:"created_at"`
}

// NewSceneReference creates a new scene reference
func NewSceneReference(sceneID uuid.UUID, entityType SceneReferenceEntityType, entityID uuid.UUID) (*SceneReference, error) {
	if !isValidSceneReferenceEntityType(entityType) {
		return nil, ErrInvalidSceneReferenceEntityType
	}

	return &SceneReference{
		ID:         uuid.New(),
		SceneID:    sceneID,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt:  time.Now(),
	}, nil
}

// Validate validates the scene reference entity
func (r *SceneReference) Validate() error {
	if !isValidSceneReferenceEntityType(r.EntityType) {
		return ErrInvalidSceneReferenceEntityType
	}
	return nil
}

func isValidSceneReferenceEntityType(entityType SceneReferenceEntityType) bool {
	return entityType == SceneReferenceEntityTypeCharacter ||
		entityType == SceneReferenceEntityTypeLocation ||
		entityType == SceneReferenceEntityTypeArtifact
}

