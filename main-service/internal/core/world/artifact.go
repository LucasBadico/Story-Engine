package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrArtifactNameRequired = errors.New("artifact name is required")
)

// Artifact represents an artifact entity
type Artifact struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	WorldID     uuid.UUID  `json:"world_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Rarity      string     `json:"rarity"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewArtifact creates a new artifact
func NewArtifact(tenantID, worldID uuid.UUID, name string) (*Artifact, error) {
	if name == "" {
		return nil, ErrArtifactNameRequired
	}

	now := time.Now()
	return &Artifact{
		ID:        uuid.New(),
		TenantID:  tenantID,
		WorldID:   worldID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the artifact entity
func (a *Artifact) Validate() error {
	if a.Name == "" {
		return ErrArtifactNameRequired
	}
	return nil
}

// UpdateName updates the artifact name
func (a *Artifact) UpdateName(name string) error {
	if name == "" {
		return ErrArtifactNameRequired
	}
	a.Name = name
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the artifact description
func (a *Artifact) UpdateDescription(description string) {
	a.Description = description
	a.UpdatedAt = time.Now()
}

// UpdateRarity updates the artifact rarity
func (a *Artifact) UpdateRarity(rarity string) {
	a.Rarity = rarity
	a.UpdatedAt = time.Now()
}


