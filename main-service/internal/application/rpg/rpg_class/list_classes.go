package rpg_class

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ListRPGClassesUseCase handles listing RPG classes
type ListRPGClassesUseCase struct {
	classRepo repositories.RPGClassRepository
	logger    logger.Logger
}

// NewListRPGClassesUseCase creates a new ListRPGClassesUseCase
func NewListRPGClassesUseCase(
	classRepo repositories.RPGClassRepository,
	logger logger.Logger,
) *ListRPGClassesUseCase {
	return &ListRPGClassesUseCase{
		classRepo: classRepo,
		logger:    logger,
	}
}

// ListRPGClassesInput represents the input for listing RPG classes
type ListRPGClassesInput struct {
	RPGSystemID  uuid.UUID
	ParentClassID *uuid.UUID // optional: filter by parent
}

// ListRPGClassesOutput represents the output of listing RPG classes
type ListRPGClassesOutput struct {
	Classes []*rpg.RPGClass
}

// Execute lists RPG classes
func (uc *ListRPGClassesUseCase) Execute(ctx context.Context, input ListRPGClassesInput) (*ListRPGClassesOutput, error) {
	var classes []*rpg.RPGClass
	var err error

	if input.ParentClassID != nil {
		classes, err = uc.classRepo.ListByParent(ctx, *input.ParentClassID)
	} else {
		classes, err = uc.classRepo.ListBySystem(ctx, input.RPGSystemID)
	}

	if err != nil {
		uc.logger.Error("failed to list RPG classes", "error", err, "rpg_system_id", input.RPGSystemID)
		return nil, err
	}

	return &ListRPGClassesOutput{
		Classes: classes,
	}, nil
}


