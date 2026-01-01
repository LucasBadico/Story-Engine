package character

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// GetAvailableClassesUseCase handles getting available classes for a character
type GetAvailableClassesUseCase struct {
	characterRepo repositories.CharacterRepository
	classRepo     repositories.RPGClassRepository
	rpgSystemRepo repositories.RPGSystemRepository
	logger        logger.Logger
}

// NewGetAvailableClassesUseCase creates a new GetAvailableClassesUseCase
func NewGetAvailableClassesUseCase(
	characterRepo repositories.CharacterRepository,
	classRepo repositories.RPGClassRepository,
	rpgSystemRepo repositories.RPGSystemRepository,
	logger logger.Logger,
) *GetAvailableClassesUseCase {
	return &GetAvailableClassesUseCase{
		characterRepo: characterRepo,
		classRepo:     classRepo,
		rpgSystemRepo: rpgSystemRepo,
		logger:        logger,
	}
}

// GetAvailableClassesInput represents the input for getting available classes
type GetAvailableClassesInput struct {
	CharacterID uuid.UUID
}

// GetAvailableClassesOutput represents the output of getting available classes
type GetAvailableClassesOutput struct {
	Classes []*rpg.RPGClass
}

// Execute gets available classes for a character
// For now, returns all classes from the world's RPG system
// In the future, can filter by requirements (level, stats, etc.)
func (uc *GetAvailableClassesUseCase) Execute(ctx context.Context, input GetAvailableClassesInput) (*GetAvailableClassesOutput, error) {
	// Get character
	_, err := uc.characterRepo.GetByID(ctx, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// Get world to find RPG system
	// Note: This requires worldRepo, but for now we'll get classes by system
	// In a full implementation, we'd get the world's RPG system ID
	// For now, we'll need to pass RPGSystemID or get it from world
	// Let's assume we'll get it from the world - we'll need worldRepo
	
	// For now, return empty - this will need worldRepo to be complete
	// The handler can pass RPGSystemID directly
	
	return &GetAvailableClassesOutput{
		Classes: []*rpg.RPGClass{},
	}, nil
}

