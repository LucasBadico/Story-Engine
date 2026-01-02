package story

import (
	"time"

	"github.com/google/uuid"
)

// EntityType represents the type of entity that references a prose block
type EntityType string

const (
	EntityTypeScene         EntityType = "scene"
	EntityTypeBeat          EntityType = "beat"
	EntityTypeChapter       EntityType = "chapter"
	EntityTypeCharacter     EntityType = "character"
	EntityTypeLocation      EntityType = "location"
	EntityTypeArtifact      EntityType = "artifact"
	EntityTypeEvent         EntityType = "event"
	EntityTypeWorld         EntityType = "world"
	EntityTypeRPGSystem     EntityType = "rpg_system"
	EntityTypeRPGSkill      EntityType = "rpg_skill"
	EntityTypeRPGClass      EntityType = "rpg_class"
	EntityTypeInventoryItem EntityType = "inventory_item"
)

// ProseBlockReference represents a reference from an entity to a prose block
type ProseBlockReference struct {
	ID          uuid.UUID  `json:"id"`
	ProseBlockID uuid.UUID `json:"prose_block_id"`
	EntityType  EntityType `json:"entity_type"`
	EntityID    uuid.UUID  `json:"entity_id"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewProseBlockReference creates a new prose block reference
func NewProseBlockReference(proseBlockID uuid.UUID, entityType EntityType, entityID uuid.UUID) (*ProseBlockReference, error) {
	if !isValidEntityType(entityType) {
		return nil, ErrInvalidEntityType
	}

	return &ProseBlockReference{
		ID:          uuid.New(),
		ProseBlockID: proseBlockID,
		EntityType:  entityType,
		EntityID:    entityID,
		CreatedAt:   time.Now(),
	}, nil
}

// Validate validates the prose block reference entity
func (r *ProseBlockReference) Validate() error {
	if !isValidEntityType(r.EntityType) {
		return ErrInvalidEntityType
	}
	return nil
}

func isValidEntityType(entityType EntityType) bool {
	return entityType == EntityTypeScene ||
		entityType == EntityTypeBeat ||
		entityType == EntityTypeChapter ||
		entityType == EntityTypeCharacter ||
		entityType == EntityTypeLocation ||
		entityType == EntityTypeArtifact ||
		entityType == EntityTypeEvent ||
		entityType == EntityTypeWorld ||
		entityType == EntityTypeRPGSystem ||
		entityType == EntityTypeRPGSkill ||
		entityType == EntityTypeRPGClass ||
		entityType == EntityTypeInventoryItem
}

