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

var _ repositories.StoryRepository = (*StoryRepository)(nil)

// StoryRepository implements the story repository interface for SQLite
type StoryRepository struct {
	db *DB
}

// NewStoryRepository creates a new story repository
func NewStoryRepository(db *DB) *StoryRepository {
	return &StoryRepository{db: db}
}

// Create creates a new story
func (r *StoryRepository) Create(ctx context.Context, s *story.Story) error {
	query := `
		INSERT INTO stories (id, tenant_id, title, status, version_number, root_story_id, previous_story_id, world_id, created_by_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var previousStoryID sql.NullString
	if s.PreviousStoryID != nil {
		previousStoryID = sql.NullString{String: s.PreviousStoryID.String(), Valid: true}
	}

	var worldID sql.NullString
	if s.WorldID != nil {
		worldID = sql.NullString{String: s.WorldID.String(), Valid: true}
	}

	var createdByUserID sql.NullString
	if s.CreatedByUserID != nil {
		createdByUserID = sql.NullString{String: s.CreatedByUserID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		s.ID.String(),
		s.TenantID.String(),
		s.Title,
		string(s.Status),
		s.VersionNumber,
		s.RootStoryID.String(),
		previousStoryID,
		worldID,
		createdByUserID,
		s.CreatedAt.Format(time.RFC3339),
		s.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a story by ID
func (r *StoryRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, world_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE tenant_id = ? AND id = ?
	`
	var s story.Story
	var idStr, tenantIDStr, rootStoryIDStr, createdAtStr, updatedAtStr string
	var previousStoryID, worldID, createdByUserID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &s.Title, &s.Status, &s.VersionNumber,
		&rootStoryIDStr, &previousStoryID, &worldID, &createdByUserID, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "story",
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

	parsedRootStoryID, err := uuid.Parse(rootStoryIDStr)
	if err != nil {
		return nil, err
	}
	s.RootStoryID = parsedRootStoryID

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
	if previousStoryID.Valid {
		if parsedID, err := uuid.Parse(previousStoryID.String); err == nil {
			s.PreviousStoryID = &parsedID
		}
	}
	if worldID.Valid {
		if parsedID, err := uuid.Parse(worldID.String); err == nil {
			s.WorldID = &parsedID
		}
	}
	if createdByUserID.Valid {
		if parsedID, err := uuid.Parse(createdByUserID.String); err == nil {
			s.CreatedByUserID = &parsedID
		}
	}

	return &s, nil
}

// ListByTenant lists stories for a tenant
func (r *StoryRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, world_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStories(rows)
}

// ListVersionsByRoot lists all versions of a story by root story ID
func (r *StoryRepository) ListVersionsByRoot(ctx context.Context, tenantID, rootStoryID uuid.UUID) ([]*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, world_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE tenant_id = ? AND root_story_id = ?
		ORDER BY version_number ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), rootStoryID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStories(rows)
}

// GetVersionGraph retrieves all versions for building a version graph
func (r *StoryRepository) GetVersionGraph(ctx context.Context, tenantID, rootStoryID uuid.UUID) ([]*story.Story, error) {
	return r.ListVersionsByRoot(ctx, tenantID, rootStoryID)
}

// Update updates a story
func (r *StoryRepository) Update(ctx context.Context, s *story.Story) error {
	query := `
		UPDATE stories
		SET title = ?, status = ?, version_number = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err := r.db.Exec(ctx, query,
		s.Title,
		string(s.Status),
		s.VersionNumber,
		s.UpdatedAt.Format(time.RFC3339),
		s.TenantID.String(),
		s.ID.String(),
	)
	return err
}

// Delete deletes a story
func (r *StoryRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM stories WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByTenant counts stories for a tenant
func (r *StoryRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM stories WHERE tenant_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&count)
	return count, err
}

func (r *StoryRepository) scanStories(rows *sql.Rows) ([]*story.Story, error) {
	stories := make([]*story.Story, 0)
	for rows.Next() {
		var s story.Story
		var idStr, tenantIDStr, rootStoryIDStr, createdAtStr, updatedAtStr string
		var previousStoryID, worldID, createdByUserID sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &s.Title, &s.Status, &s.VersionNumber,
			&rootStoryIDStr, &previousStoryID, &worldID, &createdByUserID, &createdAtStr, &updatedAtStr)
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

		parsedRootStoryID, err := uuid.Parse(rootStoryIDStr)
		if err != nil {
			return nil, err
		}
		s.RootStoryID = parsedRootStoryID

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
		if previousStoryID.Valid {
			if parsedID, err := uuid.Parse(previousStoryID.String); err == nil {
				s.PreviousStoryID = &parsedID
			}
		}
		if worldID.Valid {
			if parsedID, err := uuid.Parse(worldID.String); err == nil {
				s.WorldID = &parsedID
			}
		}
		if createdByUserID.Valid {
			if parsedID, err := uuid.Parse(createdByUserID.String); err == nil {
				s.CreatedByUserID = &parsedID
			}
		}

		stories = append(stories, &s)
	}

	return stories, rows.Err()
}

