package image_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveImageBlockReferenceUseCase handles removing a reference from an image block
type RemoveImageBlockReferenceUseCase struct {
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository
	logger                  logger.Logger
}

// NewRemoveImageBlockReferenceUseCase creates a new RemoveImageBlockReferenceUseCase
func NewRemoveImageBlockReferenceUseCase(
	imageBlockReferenceRepo repositories.ImageBlockReferenceRepository,
	logger logger.Logger,
) *RemoveImageBlockReferenceUseCase {
	return &RemoveImageBlockReferenceUseCase{
		imageBlockReferenceRepo: imageBlockReferenceRepo,
		logger:                  logger,
	}
}

// RemoveImageBlockReferenceInput represents the input for removing a reference
type RemoveImageBlockReferenceInput struct {
	ImageBlockID uuid.UUID
	EntityType   story.ImageBlockReferenceEntityType
	EntityID     uuid.UUID
}

// Execute removes a reference from an image block
func (uc *RemoveImageBlockReferenceUseCase) Execute(ctx context.Context, input RemoveImageBlockReferenceInput) error {
	if err := uc.imageBlockReferenceRepo.DeleteByImageBlockAndEntity(ctx, input.ImageBlockID, input.EntityType, input.EntityID); err != nil {
		uc.logger.Error("failed to remove image block reference", "error", err, "image_block_id", input.ImageBlockID)
		return err
	}

	uc.logger.Info("image block reference removed", "image_block_id", input.ImageBlockID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}


