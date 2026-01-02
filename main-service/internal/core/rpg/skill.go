package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSkillNameRequired = errors.New("skill name is required")
	ErrSkillSystemRequired = errors.New("RPG system ID is required")
	ErrInvalidSkillJSON = errors.New("invalid skill JSON")
)

// SkillCategory represents the category of a skill
type SkillCategory string

const (
	SkillCategoryCombat  SkillCategory = "combat"
	SkillCategoryMagic   SkillCategory = "magic"
	SkillCategoryPassive SkillCategory = "passive"
	SkillCategoryUtility SkillCategory = "utility"
)

// SkillType represents the type of a skill
type SkillType string

const (
	SkillTypeActive  SkillType = "active"
	SkillTypePassive SkillType = "passive"
	SkillTypeSpell   SkillType = "spell"
	SkillTypeAbility SkillType = "ability"
)

// Skill represents an RPG skill definition
type Skill struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	RPGSystemID    uuid.UUID       `json:"rpg_system_id"`
	Name           string          `json:"name"`
	Category       *SkillCategory  `json:"category,omitempty"`
	Type           *SkillType      `json:"type,omitempty"`
	Description    *string         `json:"description,omitempty"`
	Prerequisites  *json.RawMessage `json:"prerequisites,omitempty"` // JSONB
	MaxRank        int             `json:"max_rank"`
	EffectsSchema  *json.RawMessage `json:"effects_schema,omitempty"` // JSONB
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// NewSkill creates a new skill
func NewSkill(tenantID, rpgSystemID uuid.UUID, name string) (*Skill, error) {
	if name == "" {
		return nil, ErrSkillNameRequired
	}
	if rpgSystemID == uuid.Nil {
		return nil, ErrSkillSystemRequired
	}

	maxRank := 10
	now := time.Now()
	return &Skill{
		ID:          uuid.New(),
		TenantID:    tenantID,
		RPGSystemID: rpgSystemID,
		Name:        name,
		MaxRank:     maxRank,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate validates the skill entity
func (s *Skill) Validate() error {
	if s.Name == "" {
		return ErrSkillNameRequired
	}
	if s.RPGSystemID == uuid.Nil {
		return ErrSkillSystemRequired
	}
	if s.MaxRank < 1 {
		return errors.New("max_rank must be at least 1")
	}

	// Validate JSON fields
	if s.Prerequisites != nil && len(*s.Prerequisites) > 0 {
		var prereq map[string]interface{}
		if err := json.Unmarshal(*s.Prerequisites, &prereq); err != nil {
			return ErrInvalidSkillJSON
		}
	}

	if s.EffectsSchema != nil && len(*s.EffectsSchema) > 0 {
		var effects map[string]interface{}
		if err := json.Unmarshal(*s.EffectsSchema, &effects); err != nil {
			return ErrInvalidSkillJSON
		}
	}

	return nil
}

// UpdateName updates the skill name
func (s *Skill) UpdateName(name string) error {
	if name == "" {
		return ErrSkillNameRequired
	}
	s.Name = name
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateCategory updates the skill category
func (s *Skill) UpdateCategory(category *SkillCategory) {
	s.Category = category
	s.UpdatedAt = time.Now()
}

// UpdateType updates the skill type
func (s *Skill) UpdateType(skillType *SkillType) {
	s.Type = skillType
	s.UpdatedAt = time.Now()
}

// UpdateDescription updates the description
func (s *Skill) UpdateDescription(description *string) {
	s.Description = description
	s.UpdatedAt = time.Now()
}

// UpdatePrerequisites updates the prerequisites
func (s *Skill) UpdatePrerequisites(prerequisites *json.RawMessage) error {
	if prerequisites != nil && len(*prerequisites) > 0 {
		var prereq map[string]interface{}
		if err := json.Unmarshal(*prerequisites, &prereq); err != nil {
			return ErrInvalidSkillJSON
		}
	}
	s.Prerequisites = prerequisites
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateMaxRank updates the max rank
func (s *Skill) UpdateMaxRank(maxRank int) error {
	if maxRank < 1 {
		return errors.New("max_rank must be at least 1")
	}
	s.MaxRank = maxRank
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateEffectsSchema updates the effects schema
func (s *Skill) UpdateEffectsSchema(effectsSchema *json.RawMessage) error {
	if effectsSchema != nil && len(*effectsSchema) > 0 {
		var effects map[string]interface{}
		if err := json.Unmarshal(*effectsSchema, &effects); err != nil {
			return ErrInvalidSkillJSON
		}
	}
	s.EffectsSchema = effectsSchema
	s.UpdatedAt = time.Now()
	return nil
}


