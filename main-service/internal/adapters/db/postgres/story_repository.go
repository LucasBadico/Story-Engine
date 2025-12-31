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

var _ repositories.StoryRepository = (*StoryRepository)(nil)

// StoryRepository implements the story repository interface
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
		INSERT INTO stories (id, tenant_id, title, status, version_number, root_story_id, previous_story_id, created_by_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.TenantID, s.Title, string(s.Status), s.VersionNumber,
		s.RootStoryID, s.PreviousStoryID, s.CreatedByUserID, s.CreatedAt, s.UpdatedAt)
	return err
}

// GetByID retrieves a story by ID
func (r *StoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE id = $1
	`
	var s story.Story
	var previousStoryID sql.NullString
	var createdByUserID sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.TenantID, &s.Title, &s.Status, &s.VersionNumber,
		&s.RootStoryID, &previousStoryID, &createdByUserID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "story",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if previousStoryID.Valid {
		if id, err := uuid.Parse(previousStoryID.String); err == nil {
			s.PreviousStoryID = &id
		}
	}
	if createdByUserID.Valid {
		if id, err := uuid.Parse(createdByUserID.String); err == nil {
			s.CreatedByUserID = &id
		}
	}

	return &s, nil
}

// ListByTenant lists stories for a tenant
func (r *StoryRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStories(rows)
}

// ListVersionsByRoot lists all versions of a story by root story ID
func (r *StoryRepository) ListVersionsByRoot(ctx context.Context, rootStoryID uuid.UUID) ([]*story.Story, error) {
	query := `
		SELECT id, tenant_id, title, status, version_number, root_story_id, previous_story_id, created_by_user_id, created_at, updated_at
		FROM stories
		WHERE root_story_id = $1
		ORDER BY version_number ASC
	`
	rows, err := r.db.Query(ctx, query, rootStoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStories(rows)
}

// GetVersionGraph retrieves all versions for building a version graph
func (r *StoryRepository) GetVersionGraph(ctx context.Context, rootStoryID uuid.UUID) ([]*story.Story, error) {
	return r.ListVersionsByRoot(ctx, rootStoryID)
}

// Update updates a story
func (r *StoryRepository) Update(ctx context.Context, s *story.Story) error {
	query := `
		UPDATE stories
		SET title = $2, status = $3, version_number = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.Title, string(s.Status), s.VersionNumber, s.UpdatedAt)
	return err
}

// Delete deletes a story
func (r *StoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM stories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByTenant counts stories for a tenant
func (r *StoryRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM stories WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *StoryRepository) scanStories(rows pgx.Rows) ([]*story.Story, error) {
	stories := make([]*story.Story, 0) // Initialize as empty slice, not nil
	for rows.Next() {
		var s story.Story
		var previousStoryID sql.NullString
		var createdByUserID sql.NullString

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.Title, &s.Status, &s.VersionNumber,
			&s.RootStoryID, &previousStoryID, &createdByUserID, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if previousStoryID.Valid {
			if id, err := uuid.Parse(previousStoryID.String); err == nil {
				s.PreviousStoryID = &id
			}
		}
		if createdByUserID.Valid {
			if id, err := uuid.Parse(createdByUserID.String); err == nil {
				s.CreatedByUserID = &id
			}
		}

		stories = append(stories, &s)
	}

	return stories, rows.Err()
}

