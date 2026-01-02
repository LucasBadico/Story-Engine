package prose_block

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateProseBlockReferenceUseCase handles prose block reference creation
type CreateProseBlockReferenceUseCase struct {
	refRepo        repositories.ProseBlockReferenceRepository
	proseBlockRepo repositories.ProseBlockRepository
	logger         logger.Logger
}

// NewCreateProseBlockReferenceUseCase creates a new CreateProseBlockReferenceUseCase
func NewCreateProseBlockReferenceUseCase(
	refRepo repositories.ProseBlockReferenceRepository,
	proseBlockRepo repositories.ProseBlockRepository,
	logger logger.Logger,
) *CreateProseBlockReferenceUseCase {
	return &CreateProseBlockReferenceUseCase{
		refRepo:        refRepo,
		proseBlockRepo: proseBlockRepo,
		logger:         logger,
	}
}

// CreateProseBlockReferenceInput represents the input for creating a reference
type CreateProseBlockReferenceInput struct {
	TenantID     uuid.UUID
	ProseBlockID uuid.UUID
	EntityType   story.EntityType
	EntityID     uuid.UUID
}

// CreateProseBlockReferenceOutput represents the output of creating a reference
type CreateProseBlockReferenceOutput struct {
	Reference *story.ProseBlockReference
}

// Execute creates a new prose block reference
func (uc *CreateProseBlockReferenceUseCase) Execute(ctx context.Context, input CreateProseBlockReferenceInput) (*CreateProseBlockReferenceOutput, error) {
	// Validate prose block exists
	_, err := uc.proseBlockRepo.GetByID(ctx, input.TenantID, input.ProseBlockID)
	if err != nil {
		return nil, err
	}

	if input.EntityType == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "entity_type is required",
		}
	}

	ref, err := story.NewProseBlockReference(input.ProseBlockID, input.EntityType, input.EntityID)
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
		uc.logger.Error("failed to create prose block reference", "error", err, "prose_block_id", input.ProseBlockID, "tenant_id", input.TenantID)
		return nil, err
	}

	uc.logger.Info("prose block reference created", "reference_id", ref.ID, "prose_block_id", input.ProseBlockID, "tenant_id", input.TenantID)

	return &CreateProseBlockReferenceOutput{
		Reference: ref,
	}, nil
}

