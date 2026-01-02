package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterStatsRequired = errors.New("character stats are required")
	ErrInvalidStatsJSON       = errors.New("invalid stats JSON")
)

// CharacterRPGStats represents RPG stats for a character with versioning
type CharacterRPGStats struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	CharacterID   uuid.UUID       `json:"character_id"`
	EventID       *uuid.UUID      `json:"event_id,omitempty"`
	BaseStats     json.RawMessage `json:"base_stats"`     // JSONB
	DerivedStats  *json.RawMessage `json:"derived_stats,omitempty"` // JSONB
	Progression   *json.RawMessage `json:"progression,omitempty"` // JSONB
	IsActive      bool            `json:"is_active"`
	Version       int             `json:"version"`
	Reason        *string         `json:"reason,omitempty"`
	Timeline      *string         `json:"timeline,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// NewCharacterRPGStats creates a new character RPG stats entry
func NewCharacterRPGStats(tenantID, characterID uuid.UUID, baseStats json.RawMessage) (*CharacterRPGStats, error) {
	if len(baseStats) == 0 {
		return nil, ErrCharacterStatsRequired
	}

	// Validate JSON
	var stats map[string]interface{}
	if err := json.Unmarshal(baseStats, &stats); err != nil {
		return nil, ErrInvalidStatsJSON
	}

	return &CharacterRPGStats{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CharacterID: characterID,
		BaseStats:   baseStats,
		IsActive:    true,
		Version:     1,
		CreatedAt:   time.Now(),
	}, nil
}

// Validate validates the character RPG stats
func (c *CharacterRPGStats) Validate() error {
	if len(c.BaseStats) == 0 {
		return ErrCharacterStatsRequired
	}

	// Validate JSON
	var stats map[string]interface{}
	if err := json.Unmarshal(c.BaseStats, &stats); err != nil {
		return ErrInvalidStatsJSON
	}

	if c.DerivedStats != nil && len(*c.DerivedStats) > 0 {
		var derived map[string]interface{}
		if err := json.Unmarshal(*c.DerivedStats, &derived); err != nil {
			return ErrInvalidStatsJSON
		}
	}

	if c.Progression != nil && len(*c.Progression) > 0 {
		var progression map[string]interface{}
		if err := json.Unmarshal(*c.Progression, &progression); err != nil {
			return ErrInvalidStatsJSON
		}
	}

	return nil
}

// SetEventID sets the event that caused this stats version
func (c *CharacterRPGStats) SetEventID(eventID *uuid.UUID) {
	c.EventID = eventID
}

// SetReason sets the reason for this stats version
func (c *CharacterRPGStats) SetReason(reason *string) {
	c.Reason = reason
}

// SetTimeline sets the timeline for this stats version
func (c *CharacterRPGStats) SetTimeline(timeline *string) {
	c.Timeline = timeline
}

// SetVersion sets the version number
func (c *CharacterRPGStats) SetVersion(version int) {
	c.Version = version
}

// SetActive sets whether this version is active
func (c *CharacterRPGStats) SetActive(isActive bool) {
	c.IsActive = isActive
}


