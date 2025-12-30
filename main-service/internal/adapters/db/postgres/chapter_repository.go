package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
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
		INSERT INTO chapters (id, story_id, number, title, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		c.ID, c.StoryID, c.Number, c.Title, string(c.Status), c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID retrieves a chapter by ID
func (r *ChapterRepository) GetByID(ctx context.Context, id uuid.UUID) (*story.Chapter, error) {
	query := `
		SELECT id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE id = $1
	`
	var c story.Chapter
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.StoryID, &c.Number, &c.Title, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("chapter not found")
		}
		return nil, err
	}
	return &c, nil
}

// ListByStory lists chapters for a story
func (r *ChapterRepository) ListByStory(ctx context.Context, storyID uuid.UUID) ([]*story.Chapter, error) {
	return r.ListByStoryOrdered(ctx, storyID)
}

// ListByStoryOrdered lists chapters for a story ordered by number
func (r *ChapterRepository) ListByStoryOrdered(ctx context.Context, storyID uuid.UUID) ([]*story.Chapter, error) {
	query := `
		SELECT id, story_id, number, title, status, created_at, updated_at
		FROM chapters
		WHERE story_id = $1
		ORDER BY number ASC
	`
	rows, err := r.db.Query(ctx, query, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []*story.Chapter
	for rows.Next() {
		var c story.Chapter
		err := rows.Scan(&c.ID, &c.StoryID, &c.Number, &c.Title, &c.Status, &c.CreatedAt, &c.UpdatedAt)
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
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.Number, c.Title, string(c.Status), c.UpdatedAt)
	return err
}

// Delete deletes a chapter
func (r *ChapterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chapters WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByStory deletes all chapters for a story
func (r *ChapterRepository) DeleteByStory(ctx context.Context, storyID uuid.UUID) error {
	query := `DELETE FROM chapters WHERE story_id = $1`
	_, err := r.db.Exec(ctx, query, storyID)
	return err
}

