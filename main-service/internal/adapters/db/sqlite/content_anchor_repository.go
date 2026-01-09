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

var _ repositories.ContentAnchorRepository = (*ContentAnchorRepository)(nil)

// ContentAnchorRepository implements content anchor persistence for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM content_blocks WHERE id = ?", anchor.ContentBlockID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       anchor.ContentBlockID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO content_anchors (id, tenant_id, content_block_id, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, query,
		anchor.ID.String(),
		tenantIDStr,
		anchor.ContentBlockID.String(),
		string(anchor.EntityType),
		anchor.EntityID.String(),
		anchor.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a content anchor by ID
func (r *ContentAnchorRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = ? AND id = ?
	`
	var anchor story.ContentAnchor
	var idStr, contentBlockIDStr, entityIDStr, createdAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &contentBlockIDStr, &anchor.EntityType, &entityIDStr, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_anchor",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	anchor.ID = parsedID

	parsedContentBlockID, err := uuid.Parse(contentBlockIDStr)
	if err != nil {
		return nil, err
	}
	anchor.ContentBlockID = parsedContentBlockID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	anchor.EntityID = parsedEntityID

	anchor.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &anchor, nil
}

// ListByContentBlock lists content anchors for a content block
func (r *ContentAnchorRepository) ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = ? AND content_block_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), contentBlockID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAnchors(rows)
}

// ListByEntity lists content anchors for an entity
func (r *ContentAnchorRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentAnchor, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_anchors
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), string(entityType), entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAnchors(rows)
}

// Delete deletes a content anchor
func (r *ContentAnchorRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_anchors WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByContentBlock deletes all anchors for a content block
func (r *ContentAnchorRepository) DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error {
	query := `DELETE FROM content_anchors WHERE tenant_id = ? AND content_block_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), contentBlockID.String())
	return err
}

func (r *ContentAnchorRepository) scanAnchors(rows *sql.Rows) ([]*story.ContentAnchor, error) {
	var anchors []*story.ContentAnchor
	for rows.Next() {
		var anchor story.ContentAnchor
		var idStr, contentBlockIDStr, entityIDStr, createdAtStr string

		err := rows.Scan(&idStr, &contentBlockIDStr, &anchor.EntityType, &entityIDStr, &createdAtStr)
		if err != nil {
			return nil, err
		}

		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		anchor.ID = parsedID

		parsedContentBlockID, err := uuid.Parse(contentBlockIDStr)
		if err != nil {
			return nil, err
		}
		anchor.ContentBlockID = parsedContentBlockID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		anchor.EntityID = parsedEntityID

		anchor.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		anchors = append(anchors, &anchor)
	}

	return anchors, rows.Err()
}


