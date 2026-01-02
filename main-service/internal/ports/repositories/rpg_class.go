package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/rpg"
)

// RPGClassRepository defines the interface for RPG class persistence
type RPGClassRepository interface {
	Create(ctx context.Context, class *rpg.RPGClass) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.RPGClass, error)
	ListBySystem(ctx context.Context, tenantID, rpgSystemID uuid.UUID) ([]*rpg.RPGClass, error)
	ListByParent(ctx context.Context, tenantID, parentClassID uuid.UUID) ([]*rpg.RPGClass, error)
	Update(ctx context.Context, class *rpg.RPGClass) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// RPGClassSkillRepository defines the interface for RPG class skill persistence
type RPGClassSkillRepository interface {
	Create(ctx context.Context, classSkill *rpg.RPGClassSkill) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*rpg.RPGClassSkill, error)
	GetByClassAndSkill(ctx context.Context, tenantID, classID, skillID uuid.UUID) (*rpg.RPGClassSkill, error)
	ListByClass(ctx context.Context, tenantID, classID uuid.UUID) ([]*rpg.RPGClassSkill, error)
	Update(ctx context.Context, classSkill *rpg.RPGClassSkill) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	DeleteByClass(ctx context.Context, tenantID, classID uuid.UUID) error
}


