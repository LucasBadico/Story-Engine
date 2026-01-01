package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.SceneRepository = (*SceneRepository)(nil)

// SceneRepository implements the scene repository interface
type SceneRepository struct {
	db *DB
}

// NewSceneRepository creates a new scene repository
func NewSceneRepository(db *DB) *SceneRepository {
	return &SceneRepository{db: db}
}

// Create creates a new scene
func (r *SceneRepository) Create(ctx context.Context, s *story.Scene) error {
	query := `
		INSERT INTO scenes (id, story_id, chapter_id, order_num, pov_character_id, location_id, time_ref, goal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.StoryID, s.ChapterID, s.OrderNum, s.POVCharacterID, s.LocationID,
		s.TimeRef, s.Goal, s.CreatedAt, s.UpdatedAt)
	return err
}

// GetByID retrieves a scene by ID
func (r *SceneRepository) GetByID(ctx context.Context, id uuid.UUID) (*story.Scene, error) {
	query := `
		SELECT id, story_id, chapter_id, order_num, pov_character_id, location_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE id = $1
	`
	var s story.Scene
	var chapterID sql.NullString
	var povCharacterID sql.NullString
	var locationID sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.StoryID, &chapterID, &s.OrderNum, &povCharacterID, &locationID,
		&s.TimeRef, &s.Goal, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("scene not found")
		}
		return nil, err
	}

	if chapterID.Valid {
		if id, err := uuid.Parse(chapterID.String); err == nil {
			s.ChapterID = &id
		}
	}
	if povCharacterID.Valid {
		if id, err := uuid.Parse(povCharacterID.String); err == nil {
			s.POVCharacterID = &id
		}
	}
	if locationID.Valid {
		if id, err := uuid.Parse(locationID.String); err == nil {
			s.LocationID = &id
		}
	}

	return &s, nil
}

// ListByChapter lists scenes for a chapter
func (r *SceneRepository) ListByChapter(ctx context.Context, chapterID uuid.UUID) ([]*story.Scene, error) {
	return r.ListByChapterOrdered(ctx, chapterID)
}

// ListByChapterOrdered lists scenes for a chapter ordered by order_num
func (r *SceneRepository) ListByChapterOrdered(ctx context.Context, chapterID uuid.UUID) ([]*story.Scene, error) {
	query := `
		SELECT id, story_id, chapter_id, order_num, pov_character_id, location_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE chapter_id = $1
		ORDER BY order_num ASC
	`
	rows, err := r.db.Query(ctx, query, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanScenes(rows)
}

// ListByStory lists scenes for a story
func (r *SceneRepository) ListByStory(ctx context.Context, storyID uuid.UUID) ([]*story.Scene, error) {
	query := `
		SELECT id, story_id, chapter_id, order_num, pov_character_id, location_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE story_id = $1
		ORDER BY chapter_id, order_num ASC
	`
	rows, err := r.db.Query(ctx, query, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanScenes(rows)
}

// Update updates a scene
func (r *SceneRepository) Update(ctx context.Context, s *story.Scene) error {
	query := `
		UPDATE scenes
		SET chapter_id = $2, order_num = $3, pov_character_id = $4, location_id = $5, time_ref = $6, goal = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.ChapterID, s.OrderNum, s.POVCharacterID, s.LocationID, s.TimeRef, s.Goal, s.UpdatedAt)
	return err
}

// Delete deletes a scene
func (r *SceneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM scenes WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByChapter deletes all scenes for a chapter
func (r *SceneRepository) DeleteByChapter(ctx context.Context, chapterID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE chapter_id = $1`
	_, err := r.db.Exec(ctx, query, chapterID)
	return err
}

// DeleteByStory deletes all scenes for a story
func (r *SceneRepository) DeleteByStory(ctx context.Context, storyID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE story_id = $1`
	_, err := r.db.Exec(ctx, query, storyID)
	return err
}

func (r *SceneRepository) scanScenes(rows pgx.Rows) ([]*story.Scene, error) {
	var scenes []*story.Scene
	for rows.Next() {
		var s story.Scene
		var chapterID sql.NullString
		var povCharacterID sql.NullString
		var locationID sql.NullString

		err := rows.Scan(
			&s.ID, &s.StoryID, &chapterID, &s.OrderNum, &povCharacterID, &locationID,
			&s.TimeRef, &s.Goal, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if chapterID.Valid {
			if id, err := uuid.Parse(chapterID.String); err == nil {
				s.ChapterID = &id
			}
		}
		if povCharacterID.Valid {
			if id, err := uuid.Parse(povCharacterID.String); err == nil {
				s.POVCharacterID = &id
			}
		}
		if locationID.Valid {
			if id, err := uuid.Parse(locationID.String); err == nil {
				s.LocationID = &id
			}
		}

		scenes = append(scenes, &s)
	}

	return scenes, rows.Err()
}

