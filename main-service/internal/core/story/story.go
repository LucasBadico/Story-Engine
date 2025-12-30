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
	ID              uuid.UUID
	TenantID        uuid.UUID
	Title           string
	Status          StoryStatus
	VersionNumber   int
	RootStoryID     uuid.UUID  // All versions share the same root
	PreviousStoryID *uuid.UUID // NULL for the first version
	CreatedByUserID *uuid.UUID // nullable
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
