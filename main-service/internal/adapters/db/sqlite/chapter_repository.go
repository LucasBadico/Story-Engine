package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ChapterRepository = (*ChapterRepository)(nil)

// ChapterRepository implements the chapter repository interface for SQLite
type ChapterRepository struct {
	db *DB
}

// NewChapterRepository creates a new chapter repository
func NewChapterRepository(db *DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

// Create creates a new chapter
func (r *ChapterRepository) Create(ctx context.Context, c *story.Chapter) error {
	query := `
		INSERT INTO chapters (id, tenant_id, story_id, number, title, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var title sql.NullString
	if c.Title != "" {
		title = sql.NullString{String: c.Title, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		c.ID.String(),
		c.TenantID.String(),
		c.StoryID.String(),
		c.Number,
		title,
		string(c.Status),
		c.CreatedAt.Format(time.RFC3339),
		c.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a chapter by ID
func (r *ChapterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Chapter, error) {
	query := `
		SELECT id, tenant_id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE tenant_id = ? AND id = ?
	`
	var c story.Chapter
	var idStr, tenantIDStr, storyIDStr, createdAtStr, updatedAtStr string
	var title sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &storyIDStr, &c.Number, &title, &c.Status, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	c.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	c.TenantID = parsedTenantID

	parsedStoryID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return nil, err
	}
	c.StoryID = parsedStoryID

	// Parse timestamps
	c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable title
	if title.Valid {
		c.Title = title.String
	}

	return &c, nil
}

// ListByStory lists chapters for a story
func (r *ChapterRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error) {
	query := `
		SELECT id, tenant_id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE tenant_id = ? AND story_id = ?
		ORDER BY number ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), storyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChapters(rows)
}

// Update updates a chapter
func (r *ChapterRepository) Update(ctx context.Context, c *story.Chapter) error {
	query := `
		UPDATE chapters
		SET number = ?, title = ?, status = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var title sql.NullString
	if c.Title != "" {
		title = sql.NullString{String: c.Title, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		c.Number,
		title,
		string(c.Status),
		c.UpdatedAt.Format(time.RFC3339),
		c.TenantID.String(),
		c.ID.String(),
	)
	return err
}

// Delete deletes a chapter
func (r *ChapterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM chapters WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByStory deletes all chapters for a story
func (r *ChapterRepository) DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error {
	query := `DELETE FROM chapters WHERE tenant_id = ? AND story_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), storyID.String())
	return err
}

// ListByStoryOrdered lists chapters for a story ordered by number
func (r *ChapterRepository) ListByStoryOrdered(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error) {
	return r.ListByStory(ctx, tenantID, storyID)
}

func (r *ChapterRepository) scanChapters(rows *sql.Rows) ([]*story.Chapter, error) {
	chapters := make([]*story.Chapter, 0)
	for rows.Next() {
		var c story.Chapter
		var idStr, tenantIDStr, storyIDStr, createdAtStr, updatedAtStr string
		var title sql.NullString

		err := rows.Scan(&idStr, &tenantIDStr, &storyIDStr, &c.Number, &title, &c.Status, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		c.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		c.TenantID = parsedTenantID

		parsedStoryID, err := uuid.Parse(storyIDStr)
		if err != nil {
			return nil, err
		}
		c.StoryID = parsedStoryID

		// Parse timestamps
		c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable title
		if title.Valid {
			c.Title = title.String
		}

		chapters = append(chapters, &c)
	}

	return chapters, rows.Err()
}

