package beat

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// DeleteBeatUseCase handles beat deletion
type DeleteBeatUseCase struct {
	beatRepo repositories.BeatRepository
	logger   logger.Logger
}

// NewDeleteBeatUseCase creates a new DeleteBeatUseCase
func NewDeleteBeatUseCase(
	beatRepo repositories.BeatRepository,
	logger logger.Logger,
) *DeleteBeatUseCase {
	return &DeleteBeatUseCase{
		beatRepo: beatRepo,
		logger:   logger,
	}
}

// DeleteBeatInput represents the input for deleting a beat
type DeleteBeatInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
}

// Execute deletes a beat
func (uc *DeleteBeatUseCase) Execute(ctx context.Context, input DeleteBeatInput) error {
	// Check if beat exists
	_, err := uc.beatRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return err
	}

	if err := uc.beatRepo.Delete(ctx, input.TenantID, input.ID); err != nil {
		uc.logger.Error("failed to delete beat", "error", err, "beat_id", input.ID, "tenant_id", input.TenantID)
		return err
	}

	uc.logger.Info("beat deleted", "beat_id", input.ID, "tenant_id", input.TenantID)

	return nil
}

