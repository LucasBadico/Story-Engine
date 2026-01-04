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

var _ repositories.SceneRepository = (*SceneRepository)(nil)

// SceneRepository implements the scene repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var chapterID sql.NullString
	if s.ChapterID != nil {
		chapterID = sql.NullString{String: s.ChapterID.String(), Valid: true}
	}

	var povCharacterID sql.NullString
	if s.POVCharacterID != nil {
		povCharacterID = sql.NullString{String: s.POVCharacterID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		s.ID.String(),
		s.TenantID.String(),
		s.StoryID.String(),
		chapterID,
		s.OrderNum,
		povCharacterID,
		s.TimeRef,
		s.Goal,
		s.CreatedAt.Format(time.RFC3339),
		s.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a scene by ID
func (r *SceneRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Scene, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, order_num, pov_character_id, time_ref, goal, created_at, updated_at
		FROM scenes
		WHERE tenant_id = ? AND id = ?
	`
	var s story.Scene
	var idStr, tenantIDStr, storyIDStr, createdAtStr, updatedAtStr string
	var chapterID, povCharacterID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &storyIDStr, &chapterID, &s.OrderNum, &povCharacterID,
		&s.TimeRef, &s.Goal, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "scene",
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
	s.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	s.TenantID = parsedTenantID

	parsedStoryID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return nil, err
	}
	s.StoryID = parsedStoryID

	// Parse timestamps
	s.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	s.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUIDs
	if chapterID.Valid {
		if parsedChapterID, err := uuid.Parse(chapterID.String); err == nil {
			s.ChapterID = &parsedChapterID
		}
	}
	if povCharacterID.Valid {
		if parsedPOVCharacterID, err := uuid.Parse(povCharacterID.String); err == nil {
			s.POVCharacterID = &parsedPOVCharacterID
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
		WHERE tenant_id = ? AND chapter_id = ?
		ORDER BY order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), chapterID.String())
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
		WHERE tenant_id = ? AND story_id = ?
		ORDER BY chapter_id, order_num ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), storyID.String())
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
		SET chapter_id = ?, order_num = ?, pov_character_id = ?, time_ref = ?, goal = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var chapterID sql.NullString
	if s.ChapterID != nil {
		chapterID = sql.NullString{String: s.ChapterID.String(), Valid: true}
	}

	var povCharacterID sql.NullString
	if s.POVCharacterID != nil {
		povCharacterID = sql.NullString{String: s.POVCharacterID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		chapterID,
		s.OrderNum,
		povCharacterID,
		s.TimeRef,
		s.Goal,
		s.UpdatedAt.Format(time.RFC3339),
		s.TenantID.String(),
		s.ID.String(),
	)
	return err
}

// Delete deletes a scene
func (r *SceneRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByChapter deletes all scenes for a chapter
func (r *SceneRepository) DeleteByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = ? AND chapter_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), chapterID.String())
	return err
}

// DeleteByStory deletes all scenes for a story
func (r *SceneRepository) DeleteByStory(ctx context.Context, tenantID, storyID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE tenant_id = ? AND story_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), storyID.String())
	return err
}

func (r *SceneRepository) scanScenes(rows *sql.Rows) ([]*story.Scene, error) {
	var scenes []*story.Scene
	for rows.Next() {
		var s story.Scene
		var idStr, tenantIDStr, storyIDStr, createdAtStr, updatedAtStr string
		var chapterID, povCharacterID sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &storyIDStr, &chapterID, &s.OrderNum, &povCharacterID,
			&s.TimeRef, &s.Goal, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		s.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		s.TenantID = parsedTenantID

		parsedStoryID, err := uuid.Parse(storyIDStr)
		if err != nil {
			return nil, err
		}
		s.StoryID = parsedStoryID

		// Parse timestamps
		s.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		s.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable UUIDs
		if chapterID.Valid {
			if parsedChapterID, err := uuid.Parse(chapterID.String); err == nil {
				s.ChapterID = &parsedChapterID
			}
		}
		if povCharacterID.Valid {
			if parsedPOVCharacterID, err := uuid.Parse(povCharacterID.String); err == nil {
				s.POVCharacterID = &parsedPOVCharacterID
			}
		}

		scenes = append(scenes, &s)
	}

	return scenes, rows.Err()
}

