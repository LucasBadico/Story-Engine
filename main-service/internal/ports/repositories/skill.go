package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
)

// SkillRepository defines the interface for skill persistence
type SkillRepository interface {
	Create(ctx context.Context, skill *rpg.Skill) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.Skill, error)
	ListBySystem(ctx context.Context, tenantID, rpgSystemID uuid.UUID) ([]*rpg.Skill, error)
	Update(ctx context.Context, skill *rpg.Skill) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// CharacterSkillRepository defines the interface for character skill persistence
type CharacterSkillRepository interface {
	Create(ctx context.Context, characterSkill *rpg.CharacterSkill) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.CharacterSkill, error)
	GetByCharacterAndSkill(ctx context.Context, tenantID, characterID, skillID uuid.UUID) (*rpg.CharacterSkill, error)
	ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterSkill, error)
	ListActiveByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*rpg.CharacterSkill, error)
	Update(ctx context.Context, characterSkill *rpg.CharacterSkill) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) error
}


