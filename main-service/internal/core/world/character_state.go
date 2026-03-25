package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterStateCharacterIDRequired  = errors.New("character_id is required for character state")
	ErrCharacterStateChapterNumberRequired = errors.New("chapter_number is required for character state")
	ErrTenantIDRequired                    = errors.New("tenant_id is required")
)

// CharacterState represents the state of a character at a specific chapter (hot layer)
type CharacterState struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	CharacterID        uuid.UUID
	ChapterID          *uuid.UUID
	ChapterNumber      int
	Emotion            string
	Goal               string
	Secrets            []string
	Clothes            string
	PhysicalState      string
	TimelineDay        *int
	TimelineConfidence *string
	Metadata           map[string]interface{}
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewCharacterState creates a new character state
func NewCharacterState(tenantID, characterID uuid.UUID, chapterNumber int) (*CharacterState, error) {
	if tenantID == uuid.Nil {
		return nil, ErrTenantIDRequired
	}
	if characterID == uuid.Nil {
		return nil, ErrCharacterStateCharacterIDRequired
	}
	if chapterNumber < 0 {
		return nil, ErrCharacterStateChapterNumberRequired
	}

	now := time.Now()
	return &CharacterState{
		ID:            uuid.New(),
		TenantID:      tenantID,
		CharacterID:   characterID,
		ChapterNumber: chapterNumber,
		Secrets:       []string{},
		Metadata:      map[string]interface{}{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Validate validates the character state entity
func (cs *CharacterState) Validate() error {
	if cs.TenantID == uuid.Nil {
		return ErrTenantIDRequired
	}
	if cs.CharacterID == uuid.Nil {
		return ErrCharacterStateCharacterIDRequired
	}
	return nil
}

// SetEmotion sets the emotional state
func (cs *CharacterState) SetEmotion(emotion string) {
	cs.Emotion = emotion
	cs.UpdatedAt = time.Now()
}

// SetGoal sets the character's goal
func (cs *CharacterState) SetGoal(goal string) {
	cs.Goal = goal
	cs.UpdatedAt = time.Now()
}

// SetSecrets sets the character's secrets
func (cs *CharacterState) SetSecrets(secrets []string) {
	cs.Secrets = secrets
	cs.UpdatedAt = time.Now()
}

// SetClothes sets the character's clothes
func (cs *CharacterState) SetClothes(clothes string) {
	cs.Clothes = clothes
	cs.UpdatedAt = time.Now()
}

// SetPhysicalState sets the character's physical state
func (cs *CharacterState) SetPhysicalState(state string) {
	cs.PhysicalState = state
	cs.UpdatedAt = time.Now()
}

// SetTimeline sets the timeline day and confidence
func (cs *CharacterState) SetTimeline(day int, confidence string) {
	cs.TimelineDay = &day
	cs.TimelineConfidence = &confidence
	cs.UpdatedAt = time.Now()
}
