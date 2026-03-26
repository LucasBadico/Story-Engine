package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
)

// CharacterStateRepository defines the interface for character state persistence
type CharacterStateRepository interface {
	Create(ctx context.Context, cs *world.CharacterState) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.CharacterState, error)
	GetLatestByCharacter(ctx context.Context, tenantID, characterID uuid.UUID, asOfChapter int) (*world.CharacterState, error)
	ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*world.CharacterState, error)
	ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.CharacterState, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}
