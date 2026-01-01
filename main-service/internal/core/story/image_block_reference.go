package story

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidImageBlockReferenceEntityType = errors.New("invalid image block reference entity type")
)

// ImageBlockReferenceEntityType represents the type of entity that references an image block
type ImageBlockReferenceEntityType string

const (
	ImageBlockReferenceEntityTypeScene     ImageBlockReferenceEntityType = "scene"
	ImageBlockReferenceEntityTypeBeat      ImageBlockReferenceEntityType = "beat"
	ImageBlockReferenceEntityTypeChapter   ImageBlockReferenceEntityType = "chapter"
	ImageBlockReferenceEntityTypeCharacter ImageBlockReferenceEntityType = "character"
	ImageBlockReferenceEntityTypeLocation  ImageBlockReferenceEntityType = "location"
	ImageBlockReferenceEntityTypeArtifact  ImageBlockReferenceEntityType = "artifact"
	ImageBlockReferenceEntityTypeEvent     ImageBlockReferenceEntityType = "event"
	ImageBlockReferenceEntityTypeWorld     ImageBlockReferenceEntityType = "world"
)

// ImageBlockReference represents a reference from an image block to an entity
type ImageBlockReference struct {
	ID           uuid.UUID                       `json:"id"`
	ImageBlockID uuid.UUID                       `json:"image_block_id"`
	EntityType   ImageBlockReferenceEntityType   `json:"entity_type"`
	EntityID     uuid.UUID                       `json:"entity_id"`
	CreatedAt    time.Time                       `json:"created_at"`
}

// NewImageBlockReference creates a new image block reference
func NewImageBlockReference(imageBlockID uuid.UUID, entityType ImageBlockReferenceEntityType, entityID uuid.UUID) (*ImageBlockReference, error) {
	if !isValidImageBlockReferenceEntityType(entityType) {
		return nil, ErrInvalidImageBlockReferenceEntityType
	}

	return &ImageBlockReference{
		ID:           uuid.New(),
		ImageBlockID: imageBlockID,
		EntityType:   entityType,
		EntityID:     entityID,
		CreatedAt:    time.Now(),
	}, nil
}

// Validate validates the image block reference entity
func (r *ImageBlockReference) Validate() error {
	if !isValidImageBlockReferenceEntityType(r.EntityType) {
		return ErrInvalidImageBlockReferenceEntityType
	}
	return nil
}

func isValidImageBlockReferenceEntityType(entityType ImageBlockReferenceEntityType) bool {
	return entityType == ImageBlockReferenceEntityTypeScene ||
		entityType == ImageBlockReferenceEntityTypeBeat ||
		entityType == ImageBlockReferenceEntityTypeChapter ||
		entityType == ImageBlockReferenceEntityTypeCharacter ||
		entityType == ImageBlockReferenceEntityTypeLocation ||
		entityType == ImageBlockReferenceEntityTypeArtifact ||
		entityType == ImageBlockReferenceEntityTypeEvent ||
		entityType == ImageBlockReferenceEntityTypeWorld
}

