package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateContentBlockReferenceUseCase handles content block reference creation
type CreateContentBlockReferenceUseCase struct {
	refRepo         repositories.ContentBlockReferenceRepository
	contentBlockRepo repositories.ContentBlockRepository
	logger          logger.Logger
}

// NewCreateContentBlockReferenceUseCase creates a new CreateContentBlockReferenceUseCase
func NewCreateContentBlockReferenceUseCase(
	refRepo repositories.ContentBlockReferenceRepository,
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *CreateContentBlockReferenceUseCase {
	return &CreateContentBlockReferenceUseCase{
		refRepo:         refRepo,
		contentBlockRepo: contentBlockRepo,
		logger:          logger,
	}
}

// CreateContentBlockReferenceInput represents the input for creating a reference
type CreateContentBlockReferenceInput struct {
	TenantID      uuid.UUID
	ContentBlockID uuid.UUID
	EntityType    story.EntityType
	EntityID      uuid.UUID
}

// CreateContentBlockReferenceOutput represents the output of creating a reference
type CreateContentBlockReferenceOutput struct {
	Reference *story.ContentBlockReference
}

// Execute creates a new content block reference
func (uc *CreateContentBlockReferenceUseCase) Execute(ctx context.Context, input CreateContentBlockReferenceInput) (*CreateContentBlockReferenceOutput, error) {
	// Validate content block exists
	_, err := uc.contentBlockRepo.GetByID(ctx, input.TenantID, input.ContentBlockID)
	if err != nil {
		return nil, err
	}

	if input.EntityType == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}
	}

	ref, err := story.NewContentBlockReference(input.ContentBlockID, input.EntityType, input.EntityID)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "reference",
			Message: err.Error(),
		}
	}

	if err := ref.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "reference",
			Message: err.Error(),
		}
	}

	if err := uc.refRepo.Create(ctx, ref); err != nil {
		uc.logger.Error("failed to create content block reference", "error", err, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("content block reference created", "reference_id", ref.ID, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)

	return &CreateContentBlockReferenceOutput{
		Reference: ref,
	}, nil
}

