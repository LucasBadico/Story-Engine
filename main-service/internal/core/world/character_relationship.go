package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterRelationshipTypeRequired = errors.New("character relationship type is required")
	ErrCharactersMustBeDifferent         = errors.New("character1_id and character2_id must be different")
)

// CharacterRelationship represents a relationship between two characters
type CharacterRelationship struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	Character1ID    uuid.UUID `json:"character1_id"`
	Character2ID    uuid.UUID `json:"character2_id"`
	RelationshipType string    `json:"relationship_type"` // "ally", "enemy", "family", "lover", "rival", "mentor", "student", etc
	Description     string    `json:"description"`
	Bidirectional   bool      `json:"bidirectional"` // if true, applies in both directions
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewCharacterRelationship creates a new character relationship
func NewCharacterRelationship(tenantID, character1ID, character2ID uuid.UUID, relationshipType string) (*CharacterRelationship, error) {
	if relationshipType == "" {
		return nil, ErrCharacterRelationshipTypeRequired
	}
	if character1ID == character2ID {
		return nil, ErrCharactersMustBeDifferent
	}

	now := time.Now()
	return &CharacterRelationship{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Character1ID:    character1ID,
		Character2ID:    character2ID,
		RelationshipType: relationshipType,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Validate validates the character relationship
func (cr *CharacterRelationship) Validate() error {
	if cr.RelationshipType == "" {
		return ErrCharacterRelationshipTypeRequired
	}
	if cr.Character1ID == cr.Character2ID {
		return ErrCharactersMustBeDifferent
	}
	return nil
}

// UpdateRelationshipType updates the relationship type
func (cr *CharacterRelationship) UpdateRelationshipType(relationshipType string) error {
	if relationshipType == "" {
		return ErrCharacterRelationshipTypeRequired
	}
	cr.RelationshipType = relationshipType
	cr.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the relationship description
func (cr *CharacterRelationship) UpdateDescription(description string) {
	cr.Description = description
	cr.UpdatedAt = time.Now()
}

// UpdateBidirectional updates the bidirectional flag
func (cr *CharacterRelationship) UpdateBidirectional(bidirectional bool) {
	cr.Bidirectional = bidirectional
	cr.UpdatedAt = time.Now()
}

