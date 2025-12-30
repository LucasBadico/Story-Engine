package story

import (
	"time"

	"github.com/google/uuid"
)

// BeatType represents the type of a beat
type BeatType string

const (
	BeatTypeSetup      BeatType = "setup"
	BeatTypeTurn       BeatType = "turn"
	BeatTypeReveal     BeatType = "reveal"
	BeatTypeConflict   BeatType = "conflict"
	BeatTypeClimax     BeatType = "climax"
	BeatTypeResolution BeatType = "resolution"
	BeatTypeHook       BeatType = "hook"
	BeatTypeTransition BeatType = "transition"
)

// Beat represents a beat entity
type Beat struct {
	ID        uuid.UUID
	SceneID   uuid.UUID
	OrderNum  int
	Type      BeatType
	Intent    string
	Outcome   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewBeat creates a new beat
func NewBeat(sceneID uuid.UUID, orderNum int, beatType BeatType) (*Beat, error) {
	if orderNum < 1 {
		return nil, ErrInvalidOrderNumber
	}
	if !isValidBeatType(beatType) {
		return nil, ErrInvalidBeatType
	}

	now := time.Now()
	return &Beat{
		ID:        uuid.New(),
		SceneID:   sceneID,
		OrderNum:  orderNum,
		Type:      beatType,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the beat entity
func (b *Beat) Validate() error {
	if b.OrderNum < 1 {
		return ErrInvalidOrderNumber
	}
	if !isValidBeatType(b.Type) {
		return ErrInvalidBeatType
	}
	return nil
}

// UpdateIntent updates the beat intent
func (b *Beat) UpdateIntent(intent string) {
	b.Intent = intent
	b.UpdatedAt = time.Now()
}

// UpdateOutcome updates the beat outcome
func (b *Beat) UpdateOutcome(outcome string) {
	b.Outcome = outcome
	b.UpdatedAt = time.Now()
}

func isValidBeatType(bt BeatType) bool {
	return bt == BeatTypeSetup ||
		bt == BeatTypeTurn ||
		bt == BeatTypeReveal ||
		bt == BeatTypeConflict ||
		bt == BeatTypeClimax ||
		bt == BeatTypeResolution ||
		bt == BeatTypeHook ||
		bt == BeatTypeTransition
}

