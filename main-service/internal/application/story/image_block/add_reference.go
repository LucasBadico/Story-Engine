package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddImageBlockReferenceUseCase handles adding a reference to an image block
type AddImageBlockReferenceUseCase struct {
	imageBlockRepo         repositories.ImageBlockRepository
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository
	logger                 logger.Logger
}

// NewAddImageBlockReferenceUseCase creates a new AddImageBlockReferenceUseCase
func NewAddImageBlockReferenceUseCase(
	imageBlockRepo repositories.ImageBlockRepository,
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository,
	logger logger.Logger,
) *AddImageBlockReferenceUseCase {
	return &AddImageBlockReferenceUseCase{
		imageBlockRepo:         imageBlockRepo,
		imageBlockReferenceRepo: imageBlockReferenceRepo,
		logger:                 logger,
	}
}

// AddImageBlockReferenceInput represents the input for adding a reference
type AddImageBlockReferenceInput struct {
	ImageBlockID uuid.UUID
	EntityType   story.ImageBlockReferenceEntityType
	EntityID     uuid.UUID
}

// Execute adds a reference to an image block
func (uc *AddImageBlockReferenceUseCase) Execute(ctx context.Context, input AddImageBlockReferenceInput) error {
	// Validate image block exists
	_, err := uc.imageBlockRepo.GetByID(ctx, input.ImageBlockID)
	if err != nil {
		return err
	}

	// Prevent duplicate references
	existingRefs, err := uc.imageBlockReferenceRepo.ListByImageBlock(ctx, input.ImageBlockID)
	if err == nil {
		for _, ref := range existingRefs {
			if ref.EntityType == input.EntityType && ref.EntityID == input.EntityID {
				return &platformerrors.ValidationError{
					Field:   "entity_id",
					Message: "reference already exists",
				}
			}
		}
	}

	ref, err := story.NewImageBlockReference(input.ImageBlockID, input.EntityType, input.EntityID)
	if err != nil {
		return err
	}

	if err := uc.imageBlockReferenceRepo.Create(ctx, ref); err != nil {
		uc.logger.Error("failed to add image block reference", "error", err, "image_block_id", input.ImageBlockID)
		return err
	}

	uc.logger.Info("image block reference added", "image_block_id", input.ImageBlockID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}


