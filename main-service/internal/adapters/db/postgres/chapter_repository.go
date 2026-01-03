package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ChapterRepository = (*ChapterRepository)(nil)

// ChapterRepository implements the chapter repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		c.ID, c.TenantID, c.StoryID, c.Number, c.Title, string(c.Status), c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID retrieves a chapter by ID
func (r *ChapterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Chapter, error) {
	query := `
		SELECT id, tenant_id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE tenant_id = $1 AND id = $2
	`
	var c story.Chapter
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.StoryID, &c.Number, &c.Title, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "chapter",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	return &c, nil
}

// ListByStory lists chapters for a story
func (r *ChapterRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error) {
	return r.ListByStoryOrdered(ctx, tenantID, storyID)
}

// ListByStoryOrdered lists chapters for a story ordered by number
func (r *ChapterRepository) ListByStoryOrdered(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Chapter, error) {
	query := `
		SELECT id, tenant_id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE tenant_id = $1 AND story_id = $2
		ORDER BY number ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []*story.Chapter
	for rows.Next() {
		var c story.Chapter
		err := rows.Scan(&c.ID, &c.TenantID, &c.StoryID, &c.Number, &c.Title, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, &c)
	}

	return chapters, rows.Err()
}

// Update updates a chapter
func (r *ChapterRepository) Update(ctx context.Context, c *story.Chapter) error {
	query := `
		UPDATE chapters
		SET number = $2, title = $3, status = $4, updated_at = $5
		WHERE tenant_id = $6 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.Number, c.Title, string(c.Status), c.UpdatedAt, c.TenantID)
	return err
}

// Delete deletes a chapter
func (r *ChapterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM chapters WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByStory deletes all chapters for a story
func (r *ChapterRepository) DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error {
	query := `DELETE FROM chapters WHERE tenant_id = $1 AND story_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, storyID)
	return err
}

