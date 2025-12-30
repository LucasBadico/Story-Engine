package story

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the status of a chapter
type ChapterStatus string

const (
	ChapterStatusDraft     ChapterStatus = "draft"
	ChapterStatusPublished ChapterStatus = "published"
	ChapterStatusArchived  ChapterStatus = "archived"
)

// Chapter represents a chapter entity
type Chapter struct {
	ID        uuid.UUID
	StoryID   uuid.UUID
	Number    int
	Title     string
	Status    ChapterStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewChapter creates a new chapter
func NewChapter(storyID uuid.UUID, number int, title string) (*Chapter, error) {
	if number < 1 {
		return nil, ErrInvalidChapterNumber
	}

	now := time.Now()
	return &Chapter{
		ID:        uuid.New(),
		StoryID:   storyID,
		Number:    number,
		Title:     title,
		Status:    ChapterStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the chapter entity
func (c *Chapter) Validate() error {
	if c.Number < 1 {
		return ErrInvalidChapterNumber
	}
	if c.Status != ChapterStatusDraft && c.Status != ChapterStatusPublished && c.Status != ChapterStatusArchived {
		return ErrInvalidStatus
	}
	return nil
}

// UpdateTitle updates the chapter title
func (c *Chapter) UpdateTitle(title string) {
	c.Title = title
	c.UpdatedAt = time.Now()
}

// UpdateStatus updates the chapter status
func (c *Chapter) UpdateStatus(status ChapterStatus) error {
	if status != ChapterStatusDraft && status != ChapterStatusPublished && status != ChapterStatusArchived {
		return ErrInvalidStatus
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}
