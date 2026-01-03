package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
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
		INSERT INTO scenes (id, tenant_id, story_id, chapter_id, order_num, pov_character_id, time_ref, goal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.TenantID, s.StoryID, s.ChapterID, s.OrderNum, s.POVCharacterID,
		s.TimeRef, s.Goal, s.CreatedAt, s.UpdatedAt)
	return err
}

// GetByID retrieves a scene by ID
func (r *SceneRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Scene, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, order_num, pov_character_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE tenant_id = $1 AND id = $2
	`
	var s story.Scene
	var chapterID sql.NullString
	var povCharacterID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&s.ID, &s.TenantID, &s.StoryID, &chapterID, &s.OrderNum, &povCharacterID,
		&s.TimeRef, &s.Goal, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "scene",
				ID:       id.String(),
			}
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

	return &s, nil
}

// ListByChapter lists scenes for a chapter
func (r *SceneRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.Scene, error) {
	return r.ListByChapterOrdered(ctx, tenantID, chapterID)
}

// ListByChapterOrdered lists scenes for a chapter ordered by order_num
func (r *SceneRepository) ListByChapterOrdered(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.Scene, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, order_num, pov_character_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanScenes(rows)
}

// ListByStory lists scenes for a story
func (r *SceneRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Scene, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, order_num, pov_character_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE tenant_id = $1 AND story_id = $2
		ORDER BY chapter_id, order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID)
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
		SET chapter_id = $2, order_num = $3, pov_character_id = $4, time_ref = $5, goal = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.ChapterID, s.OrderNum, s.POVCharacterID, s.TimeRef, s.Goal, s.UpdatedAt, s.TenantID)
	return err
}

// Delete deletes a scene
func (r *SceneRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByChapter deletes all scenes for a chapter
func (r *SceneRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = $1 AND chapter_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, chapterID)
	return err
}

// DeleteByStory deletes all scenes for a story
func (r *SceneRepository) DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = $1 AND story_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, storyID)
	return err
}

func (r *SceneRepository) scanScenes(rows pgx.Rows) ([]*story.Scene, error) {
	var scenes []*story.Scene
	for rows.Next() {
		var s story.Scene
		var chapterID sql.NullString
		var povCharacterID sql.NullString

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.StoryID, &chapterID, &s.OrderNum, &povCharacterID,
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

		scenes = append(scenes, &s)
	}

	return scenes, rows.Err()
}

