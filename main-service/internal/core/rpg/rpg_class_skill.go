package rpg

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClassSkillRequired = errors.New("class and skill IDs are required")
	ErrInvalidUnlockLevel = errors.New("unlock_level must be at least 1")
)

// RPGClassSkill represents a skill that belongs to a class
type RPGClassSkill struct {
	ID           uuid.UUID `json:"id"`
	ClassID      uuid.UUID `json:"class_id"`
	SkillID      uuid.UUID `json:"skill_id"`
	UnlockLevel  int       `json:"unlock_level"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewRPGClassSkill creates a new RPG class skill
func NewRPGClassSkill(classID, skillID uuid.UUID, unlockLevel int) (*RPGClassSkill, error) {
	if classID == uuid.Nil {
		return nil, ErrClassSkillRequired
	}
	if skillID == uuid.Nil {
		return nil, ErrClassSkillRequired
	}
	if unlockLevel < 1 {
		return nil, ErrInvalidUnlockLevel
	}

	return &RPGClassSkill{
		ID:          uuid.New(),
		ClassID:     classID,
		SkillID:     skillID,
		UnlockLevel: unlockLevel,
		CreatedAt:   time.Now(),
	}, nil
}

// Validate validates the RPG class skill entity
func (cs *RPGClassSkill) Validate() error {
	if cs.ClassID == uuid.Nil {
		return ErrClassSkillRequired
	}
	if cs.SkillID == uuid.Nil {
		return ErrClassSkillRequired
	}
	if cs.UnlockLevel < 1 {
		return ErrInvalidUnlockLevel
	}
	return nil
}

// SetUnlockLevel sets the unlock level
func (cs *RPGClassSkill) SetUnlockLevel(level int) error {
	if level < 1 {
		return ErrInvalidUnlockLevel
	}
	cs.UnlockLevel = level
	return nil
}


