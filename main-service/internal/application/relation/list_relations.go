package relation

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListRelationsBySourceUseCase handles listing relations by source
type ListRelationsBySourceUseCase struct {
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewListRelationsBySourceUseCase creates a new ListRelationsBySourceUseCase
func NewListRelationsBySourceUseCase(
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *ListRelationsBySourceUseCase {
	return &ListRelationsBySourceUseCase{
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// ListRelationsBySourceInput represents the input for listing relations by source
type ListRelationsBySourceInput struct {
	TenantID   uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Options    repositories.ListOptions
}

// ListRelationsBySourceOutput represents the output of listing relations by source
type ListRelationsBySourceOutput struct {
	Relations *repositories.ListResult
}

// Execute lists relations by source
func (uc *ListRelationsBySourceUseCase) Execute(ctx context.Context, input ListRelationsBySourceInput) (*ListRelationsBySourceOutput, error) {
	result, err := uc.relationRepo.ListBySource(ctx, input.TenantID, input.SourceType, input.SourceID, input.Options)
	if err != nil {
		uc.logger.Error("failed to list relations by source", "error", err, "source_type", input.SourceType, "source_id", input.SourceID)
		return nil, err
	}

	return &ListRelationsBySourceOutput{
		Relations: result,
	}, nil
}

// ListRelationsByTargetUseCase handles listing relations by target
type ListRelationsByTargetUseCase struct {
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewListRelationsByTargetUseCase creates a new ListRelationsByTargetUseCase
func NewListRelationsByTargetUseCase(
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *ListRelationsByTargetUseCase {
	return &ListRelationsByTargetUseCase{
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// ListRelationsByTargetInput represents the input for listing relations by target
type ListRelationsByTargetInput struct {
	TenantID   uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	Options    repositories.ListOptions
}

// ListRelationsByTargetOutput represents the output of listing relations by target
type ListRelationsByTargetOutput struct {
	Relations *repositories.ListResult
}

// Execute lists relations by target
func (uc *ListRelationsByTargetUseCase) Execute(ctx context.Context, input ListRelationsByTargetInput) (*ListRelationsByTargetOutput, error) {
	result, err := uc.relationRepo.ListByTarget(ctx, input.TenantID, input.TargetType, input.TargetID, input.Options)
	if err != nil {
		uc.logger.Error("failed to list relations by target", "error", err, "target_type", input.TargetType, "target_id", input.TargetID)
		return nil, err
	}

	return &ListRelationsByTargetOutput{
		Relations: result,
	}, nil
}

// ListRelationsByWorldUseCase handles listing relations by world
type ListRelationsByWorldUseCase struct {
	relationRepo repositories.EntityRelationRepository
	logger       logger.Logger
}

// NewListRelationsByWorldUseCase creates a new ListRelationsByWorldUseCase
func NewListRelationsByWorldUseCase(
	relationRepo repositories.EntityRelationRepository,
	logger logger.Logger,
) *ListRelationsByWorldUseCase {
	return &ListRelationsByWorldUseCase{
		relationRepo: relationRepo,
		logger:       logger,
	}
}

// ListRelationsByWorldInput represents the input for listing relations by world
type ListRelationsByWorldInput struct {
	TenantID uuid.UUID
	WorldID  uuid.UUID
	Options  repositories.ListOptions
}

// ListRelationsByWorldOutput represents the output of listing relations by world
type ListRelationsByWorldOutput struct {
	Relations *repositories.ListResult
}

// Execute lists relations by world
func (uc *ListRelationsByWorldUseCase) Execute(ctx context.Context, input ListRelationsByWorldInput) (*ListRelationsByWorldOutput, error) {
	result, err := uc.relationRepo.ListByWorld(ctx, input.TenantID, input.WorldID, input.Options)
	if err != nil {
		uc.logger.Error("failed to list relations by world", "error", err, "world_id", input.WorldID)
		return nil, err
	}

	return &ListRelationsByWorldOutput{
		Relations: result,
	}, nil
}

