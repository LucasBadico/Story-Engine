package rpg

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterSkillRequired = errors.New("character and skill IDs are required")
	ErrInvalidRank = errors.New("rank must be at least 1")
)

// CharacterSkill represents a character's learned skill
type CharacterSkill struct {
	ID          uuid.UUID `json:"id"`
	CharacterID uuid.UUID `json:"character_id"`
	SkillID     uuid.UUID `json:"skill_id"`
	Rank        int       `json:"rank"`
	XPInSkill   int       `json:"xp_in_skill"`
	IsActive    bool      `json:"is_active"`
	AcquiredAt  time.Time `json:"acquired_at"`
}

// NewCharacterSkill creates a new character skill
func NewCharacterSkill(characterID, skillID uuid.UUID) (*CharacterSkill, error) {
	if characterID == uuid.Nil {
		return nil, ErrCharacterSkillRequired
	}
	if skillID == uuid.Nil {
		return nil, ErrCharacterSkillRequired
	}

	return &CharacterSkill{
		ID:          uuid.New(),
		CharacterID: characterID,
		SkillID:     skillID,
		Rank:        1,
		XPInSkill:   0,
		IsActive:    true,
		AcquiredAt:  time.Now(),
	}, nil
}

// Validate validates the character skill entity
func (cs *CharacterSkill) Validate() error {
	if cs.CharacterID == uuid.Nil {
		return ErrCharacterSkillRequired
	}
	if cs.SkillID == uuid.Nil {
		return ErrCharacterSkillRequired
	}
	if cs.Rank < 1 {
		return ErrInvalidRank
	}
	if cs.XPInSkill < 0 {
		return errors.New("xp_in_skill cannot be negative")
	}
	return nil
}

// SetRank sets the skill rank
func (cs *CharacterSkill) SetRank(rank int) error {
	if rank < 1 {
		return ErrInvalidRank
	}
	cs.Rank = rank
	return nil
}

// AddXP adds XP to the skill
func (cs *CharacterSkill) AddXP(xp int) {
	if xp > 0 {
		cs.XPInSkill += xp
	}
}

// SetActive sets whether the skill is active/equipped
func (cs *CharacterSkill) SetActive(isActive bool) {
	cs.IsActive = isActive
}

