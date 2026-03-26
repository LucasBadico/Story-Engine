package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRelationshipStateChapterIDRequired    = errors.New("chapter_id is required for relationship state")
	ErrRelationshipStateChapterNumberRequired = errors.New("chapter_number is required for relationship state")
	ErrRelationshipStateTensionRequired       = errors.New("tension_description is required for relationship state")
)

// RelationshipStateEntity represents an entity linked to a relationship state
type RelationshipStateEntity struct {
	ID                  uuid.UUID
	RelationshipStateID uuid.UUID
	Role                string
	EntityType          string
	EntityID            uuid.UUID
	CreatedAt           time.Time
}

// RelationshipState represents the state of a relationship between entities at a specific chapter
type RelationshipState struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	ChapterID          uuid.UUID
	ChapterNumber      int
	TensionDescription string
	TimelineDay        *int
	TimelineConfidence *string
	Metadata           map[string]interface{}
	Entities           []RelationshipStateEntity
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewRelationshipState creates a new relationship state
func NewRelationshipState(tenantID, chapterID uuid.UUID, chapterNumber int, tensionDescription string) (*RelationshipState, error) {
	if tenantID == uuid.Nil {
		return nil, ErrTenantIDRequired
	}
	if chapterID == uuid.Nil {
		return nil, ErrRelationshipStateChapterIDRequired
	}
	if chapterNumber < 0 {
		return nil, ErrRelationshipStateChapterNumberRequired
	}
	if tensionDescription == "" {
		return nil, ErrRelationshipStateTensionRequired
	}

	now := time.Now()
	return &RelationshipState{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		ChapterID:          chapterID,
		ChapterNumber:      chapterNumber,
		TensionDescription: tensionDescription,
		Metadata:           map[string]interface{}{},
		Entities:           []RelationshipStateEntity{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// Validate validates the relationship state entity
func (rs *RelationshipState) Validate() error {
	if rs.TenantID == uuid.Nil {
		return ErrTenantIDRequired
	}
	if rs.TensionDescription == "" {
		return ErrRelationshipStateTensionRequired
	}
	return nil
}

// AddEntity adds an entity link to this relationship state
func (rs *RelationshipState) AddEntity(role, entityType string, entityID uuid.UUID) {
	rs.Entities = append(rs.Entities, RelationshipStateEntity{
		ID:                  uuid.New(),
		RelationshipStateID: rs.ID,
		Role:                role,
		EntityType:          entityType,
		EntityID:            entityID,
		CreatedAt:           time.Now(),
	})
}

// SetTensionDescription updates the tension description
func (rs *RelationshipState) SetTensionDescription(tension string) {
	rs.TensionDescription = tension
	rs.UpdatedAt = time.Now()
}

// SetTimeline sets the timeline day and confidence
func (rs *RelationshipState) SetTimeline(day int, confidence string) {
	rs.TimelineDay = &day
	rs.TimelineConfidence = &confidence
	rs.UpdatedAt = time.Now()
}
