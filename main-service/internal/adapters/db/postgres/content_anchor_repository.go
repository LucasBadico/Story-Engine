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

var _ repositories.ContentAnchorRepository = (*ContentAnchorRepository)(nil)

// ContentAnchorRepository implements persistence for content anchors in Postgres
type ContentAnchorRepository struct {
	db *DB
}

// NewContentAnchorRepository creates a new content anchor repository
func NewContentAnchorRepository(db *DB) *ContentAnchorRepository {
	return &ContentAnchorRepository{db: db}
}

// Create creates a new content anchor
func (r *ContentAnchorRepository) Create(ctx context.Context, anchor *story.ContentAnchor) error {
	// Get tenant_id from content_block
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM content_blocks WHERE id = $1", anchor.ContentBlockID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       anchor.ContentBlockID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO content_anchors (id, tenant_id, content_block_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		anchor.ID, tenantID, anchor.ContentBlockID, string(anchor.EntityType), anchor.EntityID, anchor.CreatedAt)
	return err
}

// GetByID retrieves a content anchor by ID
func (r *ContentAnchorRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = $1 AND id = $2
	`
	var anchor story.ContentAnchor
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&anchor.ID, &anchor.ContentBlockID, &anchor.EntityType, &anchor.EntityID, &anchor.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_anchor",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	return &anchor, nil
}

// ListByContentBlock lists content anchors for a content block
func (r *ContentAnchorRepository) ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = $1 AND content_block_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, contentBlockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []*story.ContentAnchor
	for rows.Next() {
		var anchor story.ContentAnchor
		err := rows.Scan(&anchor.ID, &anchor.ContentBlockID, &anchor.EntityType, &anchor.EntityID, &anchor.CreatedAt)
		if err != nil {
			return nil, err
		}
		anchors = append(anchors, &anchor)
	}

	return anchors, rows.Err()
}

// ListByEntity lists content anchors for an entity
func (r *ContentAnchorRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []*story.ContentAnchor
	for rows.Next() {
		var anchor story.ContentAnchor
		err := rows.Scan(&anchor.ID, &anchor.ContentBlockID, &anchor.EntityType, &anchor.EntityID, &anchor.CreatedAt)
		if err != nil {
			return nil, err
		}
		anchors = append(anchors, &anchor)
	}

	return anchors, rows.Err()
}

// Delete deletes a content anchor
func (r *ContentAnchorRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_anchors WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByContentBlock deletes all content anchors for a content block
func (r *ContentAnchorRepository) DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error {
	query := `DELETE FROM content_anchors WHERE tenant_id = $1 AND content_block_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, contentBlockID)
	return err
}
