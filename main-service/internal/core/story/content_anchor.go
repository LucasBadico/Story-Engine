package story

import (
	"time"

	"github.com/google/uuid"
)

// EntityType represents the type of entity anchored within a content block
type EntityType string

const (
	EntityTypeScene            EntityType = "scene"
	EntityTypeBeat             EntityType = "beat"
	EntityTypeChapter          EntityType = "chapter"
	EntityTypeCharacter        EntityType = "character"
	EntityTypeLocation         EntityType = "location"
	EntityTypeArtifact         EntityType = "artifact"
	EntityTypeEvent            EntityType = "event"
	EntityTypeWorld            EntityType = "world"
	EntityTypeRPGSystem        EntityType = "rpg_system"
	EntityTypeRPGSkill         EntityType = "rpg_skill"
	EntityTypeRPGClass         EntityType = "rpg_class"
	EntityTypeInventoryItem    EntityType = "inventory_item"
	EntityTypeFaction          EntityType = "faction"
	EntityTypeLore             EntityType = "lore"
	EntityTypeFactionReference EntityType = "faction_reference"
	EntityTypeLoreReference    EntityType = "lore_reference"
)

// ContentAnchor represents an entity mention anchored within a content block
type ContentAnchor struct {
	ID             uuid.UUID  `json:"id"`
	ContentBlockID uuid.UUID  `json:"content_block_id"`
	EntityType     EntityType `json:"entity_type"`
	EntityID       uuid.UUID  `json:"entity_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NewContentAnchor creates a new content anchor
func NewContentAnchor(contentBlockID uuid.UUID, entityType EntityType, entityID uuid.UUID) (*ContentAnchor, error) {
	if !isValidEntityType(entityType) {
		return nil, ErrInvalidEntityType
	}

	return &ContentAnchor{
		ID:             uuid.New(),
		ContentBlockID: contentBlockID,
		EntityType:     entityType,
		EntityID:       entityID,
		CreatedAt:      time.Now(),
	}, nil
}

// Validate validates the content anchor entity
func (a *ContentAnchor) Validate() error {
	if !isValidEntityType(a.EntityType) {
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
		entityType == EntityTypeInventoryItem ||
		entityType == EntityTypeFaction ||
		entityType == EntityTypeLore ||
		entityType == EntityTypeFactionReference ||
		entityType == EntityTypeLoreReference
}


