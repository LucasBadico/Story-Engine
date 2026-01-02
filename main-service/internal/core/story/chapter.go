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
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenant_id"`
	StoryID   uuid.UUID     `json:"story_id"`
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	Status    ChapterStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewChapter creates a new chapter
func NewChapter(tenantID, storyID uuid.UUID, number int, title string) (*Chapter, error) {
	if number < 1 {
		return nil, ErrInvalidChapterNumber
	}

	now := time.Now()
	return &Chapter{
		ID:        uuid.New(),
		TenantID:  tenantID,
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
