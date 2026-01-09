package content_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateContentAnchorUseCase handles content anchor creation
type CreateContentAnchorUseCase struct {
	anchorRepo       repositories.ContentAnchorRepository
	contentBlockRepo repositories.ContentBlockRepository
	logger           logger.Logger
}

// NewCreateContentAnchorUseCase creates a new CreateContentAnchorUseCase
func NewCreateContentAnchorUseCase(
	anchorRepo repositories.ContentAnchorRepository,
	contentBlockRepo repositories.ContentBlockRepository,
	logger logger.Logger,
) *CreateContentAnchorUseCase {
	return &CreateContentAnchorUseCase{
		anchorRepo:       anchorRepo,
		contentBlockRepo: contentBlockRepo,
		logger:           logger,
	}
}

// CreateContentAnchorInput represents the input for creating an anchor
type CreateContentAnchorInput struct {
	TenantID       uuid.UUID
	ContentBlockID uuid.UUID
	EntityType     story.EntityType
	EntityID       uuid.UUID
}

// CreateContentAnchorOutput represents the output of creating an anchor
type CreateContentAnchorOutput struct {
	Anchor *story.ContentAnchor
}

// Execute creates a new content anchor
func (uc *CreateContentAnchorUseCase) Execute(ctx context.Context, input CreateContentAnchorInput) (*CreateContentAnchorOutput, error) {
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

	anchor, err := story.NewContentAnchor(input.ContentBlockID, input.EntityType, input.EntityID)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "anchor",
			Message: err.Error(),
		}
	}

	if err := anchor.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "anchor",
			Message: err.Error(),
		}
	}

	if err := uc.anchorRepo.Create(ctx, anchor); err != nil {
		uc.logger.Error("failed to create content anchor", "error", err, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("content anchor created", "anchor_id", anchor.ID, "content_block_id", input.ContentBlockID, "tenant_id", input.TenantID)

	return &CreateContentAnchorOutput{
		Anchor: anchor,
	}, nil
}


