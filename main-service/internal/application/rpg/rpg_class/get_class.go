package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetRPGClassUseCase handles retrieving an RPG class
type GetRPGClassUseCase struct {
	classRepo repositories.RPGClassRepository
	logger    logger.Logger
}

// NewGetRPGClassUseCase creates a new GetRPGClassUseCase
func NewGetRPGClassUseCase(
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *GetRPGClassUseCase {
	return &GetRPGClassUseCase{
		classRepo: classRepo,
		logger:    logger,
	}
}

// GetRPGClassInput represents the input for getting an RPG class
type GetRPGClassInput struct {
	ID uuid.UUID
}

// GetRPGClassOutput represents the output of getting an RPG class
type GetRPGClassOutput struct {
	Class *rpg.RPGClass
}

// Execute retrieves an RPG class by ID
func (uc *GetRPGClassUseCase) Execute(ctx context.Context, input GetRPGClassInput) (*GetRPGClassOutput, error) {
	class, err := uc.classRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error("failed to get RPG class", "error", err, "class_id", input.ID)
		return nil, err
	}

	return &GetRPGClassOutput{
		Class: class,
	}, nil
}

