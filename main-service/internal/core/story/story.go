package story

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the status of a story
type StoryStatus string

const (
	StoryStatusDraft     StoryStatus = "draft"
	StoryStatusPublished StoryStatus = "published"
	StoryStatusArchived  StoryStatus = "archived"
)

// Story represents a story entity with versioning support
type Story struct {
	ID              uuid.UUID   `json:"id"`
	TenantID        uuid.UUID   `json:"tenant_id"`
	Title           string      `json:"title"`
	Status          StoryStatus `json:"status"`
	VersionNumber   int         `json:"version_number"`
	RootStoryID     uuid.UUID   `json:"root_story_id"`                // All versions share the same root
	PreviousStoryID *uuid.UUID  `json:"previous_story_id,omitempty"`  // NULL for the first version
	CreatedByUserID *uuid.UUID  `json:"created_by_user_id,omitempty"` // nullable
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// NewStory creates a new story as version 1 (root story)
func NewStory(tenantID uuid.UUID, title string, createdByUserID *uuid.UUID) (*Story, error) {
	if title == "" {
		return nil, ErrTitleRequired
	}

	now := time.Now()
	id := uuid.New()
	return &Story{
		ID:              id,
		TenantID:        tenantID,
		Title:           title,
		Status:          StoryStatusDraft,
		VersionNumber:   1,
		RootStoryID:     id, // First version is its own root
		PreviousStoryID: nil,
		CreatedByUserID: createdByUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Validate validates the story entity
func (s *Story) Validate() error {
	if s.Title == "" {
		return ErrTitleRequired
	}
	if s.Status != StoryStatusDraft && s.Status != StoryStatusPublished && s.Status != StoryStatusArchived {
		return ErrInvalidStatus
	}
	if s.VersionNumber < 1 {
		return ErrInvalidVersionNumber
	}
	return nil
}

// UpdateTitle updates the story title
func (s *Story) UpdateTitle(title string) error {
	if title == "" {
		return ErrTitleRequired
	}
	s.Title = title
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus updates the story status
func (s *Story) UpdateStatus(status StoryStatus) error {
	if status != StoryStatusDraft && status != StoryStatusPublished && status != StoryStatusArchived {
		return ErrInvalidStatus
	}
	s.Status = status
	s.UpdatedAt = time.Now()
	return nil
}

// IsRoot returns true if this story is the root version
func (s *Story) IsRoot() bool {
	return s.ID == s.RootStoryID
}

// IsFirstVersion returns true if this is the first version (no previous)
func (s *Story) IsFirstVersion() bool {
	return s.PreviousStoryID == nil
}
