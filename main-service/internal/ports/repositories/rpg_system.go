package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
)

// RPGSystemRepository defines the interface for RPG system persistence
type RPGSystemRepository interface {
	Create(ctx context.Context, system *rpg.RPGSystem) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.RPGSystem, error)
	List(ctx context.Context, tenantID *uuid.UUID) ([]*rpg.RPGSystem, error) // nil = builtin only
	Update(ctx context.Context, system *rpg.RPGSystem) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// CharacterRPGStatsRepository defines the interface for character RPG stats persistence
type CharacterRPGStatsRepository interface {
	Create(ctx context.Context, stats *rpg.CharacterRPGStats) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.CharacterRPGStats, error)
	GetActiveByCharacter(ctx context.Context, characterID uuid.UUID) (*rpg.CharacterRPGStats, error)
	ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*rpg.CharacterRPGStats, error)
	ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*rpg.CharacterRPGStats, error)
	Update(ctx context.Context, stats *rpg.CharacterRPGStats) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCharacter(ctx context.Context, characterID uuid.UUID) error
	DeactivateAllByCharacter(ctx context.Context, characterID uuid.UUID) error
	GetNextVersion(ctx context.Context, characterID uuid.UUID) (int, error)
}

// ArtifactRPGStatsRepository defines the interface for artifact RPG stats persistence
type ArtifactRPGStatsRepository interface {
	Create(ctx context.Context, stats *rpg.ArtifactRPGStats) error
	GetByID(ctx context.Context, id uuid.UUID) (*rpg.ArtifactRPGStats, error)
	GetActiveByArtifact(ctx context.Context, artifactID uuid.UUID) (*rpg.ArtifactRPGStats, error)
	ListByArtifact(ctx context.Context, artifactID uuid.UUID) ([]*rpg.ArtifactRPGStats, error)
	ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*rpg.ArtifactRPGStats, error)
	Update(ctx context.Context, stats *rpg.ArtifactRPGStats) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByArtifact(ctx context.Context, artifactID uuid.UUID) error
	DeactivateAllByArtifact(ctx context.Context, artifactID uuid.UUID) error
	GetNextVersion(ctx context.Context, artifactID uuid.UUID) (int, error)
}


