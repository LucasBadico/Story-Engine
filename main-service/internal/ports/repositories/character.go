package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// CharacterRepository defines the interface for character persistence
type CharacterRepository interface {
	Create(ctx context.Context, c *world.Character) error
	GetByID(ctx context.Context, id uuid.UUID) (*world.Character, error)
	ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Character, error)
	Update(ctx context.Context, c *world.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error)
}

// CharacterTraitRepository defines the interface for character-trait persistence
type CharacterTraitRepository interface {
	Create(ctx context.Context, ct *world.CharacterTrait) error
	GetByCharacter(ctx context.Context, characterID uuid.UUID) ([]*world.CharacterTrait, error)
	GetByID(ctx context.Context, id uuid.UUID) (*world.CharacterTrait, error)
	GetByCharacterAndTrait(ctx context.Context, characterID, traitID uuid.UUID) (*world.CharacterTrait, error)
	Update(ctx context.Context, ct *world.CharacterTrait) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCharacter(ctx context.Context, characterID uuid.UUID) error
}


