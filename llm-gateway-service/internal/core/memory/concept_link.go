package memory

import (
	"time"

	"github.com/google/uuid"
)

// ConceptLink represents a relationship between two concepts
type ConceptLink struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	SourceConceptID  uuid.UUID
	TargetConceptID  uuid.UUID
	RelationType     string
	Summary          string
	Metadata         map[string]interface{}
	ValidFromChapter int
	ValidToChapter   *int
	CreatedAt        time.Time
}

// NewConceptLink creates a new concept link
func NewConceptLink(tenantID, sourceID, targetID uuid.UUID, relationType string) *ConceptLink {
	return &ConceptLink{
		ID:               uuid.New(),
		TenantID:         tenantID,
		SourceConceptID:  sourceID,
		TargetConceptID:  targetID,
		RelationType:     relationType,
		Metadata:         map[string]interface{}{},
		ValidFromChapter: 0,
		CreatedAt:        time.Now(),
	}
}
