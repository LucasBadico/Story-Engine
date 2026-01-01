package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrArtifactStatsRequired = errors.New("artifact stats are required")
	ErrInvalidArtifactStatsJSON = errors.New("invalid artifact stats JSON")
)

// ArtifactRPGStats represents RPG stats for an artifact with versioning
type ArtifactRPGStats struct {
	ID          uuid.UUID       `json:"id"`
	ArtifactID  uuid.UUID       `json:"artifact_id"`
	EventID     *uuid.UUID      `json:"event_id,omitempty"`
	Stats       json.RawMessage `json:"stats"` // JSONB
	IsActive    bool            `json:"is_active"`
	Version     int             `json:"version"`
	Reason      *string         `json:"reason,omitempty"`
	Timeline    *string         `json:"timeline,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// NewArtifactRPGStats creates a new artifact RPG stats entry
func NewArtifactRPGStats(artifactID uuid.UUID, stats json.RawMessage) (*ArtifactRPGStats, error) {
	if len(stats) == 0 {
		return nil, ErrArtifactStatsRequired
	}

	// Validate JSON
	var statsMap map[string]interface{}
	if err := json.Unmarshal(stats, &statsMap); err != nil {
		return nil, ErrInvalidArtifactStatsJSON
	}

	return &ArtifactRPGStats{
		ID:         uuid.New(),
		ArtifactID: artifactID,
		Stats:      stats,
		IsActive:   true,
		Version:    1,
		CreatedAt:  time.Now(),
	}, nil
}

// Validate validates the artifact RPG stats
func (a *ArtifactRPGStats) Validate() error {
	if len(a.Stats) == 0 {
		return ErrArtifactStatsRequired
	}

	// Validate JSON
	var stats map[string]interface{}
	if err := json.Unmarshal(a.Stats, &stats); err != nil {
		return ErrInvalidArtifactStatsJSON
	}

	return nil
}

// SetEventID sets the event that caused this stats version
func (a *ArtifactRPGStats) SetEventID(eventID *uuid.UUID) {
	a.EventID = eventID
}

// SetReason sets the reason for this stats version
func (a *ArtifactRPGStats) SetReason(reason *string) {
	a.Reason = reason
}

// SetTimeline sets the timeline for this stats version
func (a *ArtifactRPGStats) SetTimeline(timeline *string) {
	a.Timeline = timeline
}

// SetVersion sets the version number
func (a *ArtifactRPGStats) SetVersion(version int) {
	a.Version = version
}

// SetActive sets whether this version is active
func (a *ArtifactRPGStats) SetActive(isActive bool) {
	a.IsActive = isActive
}

