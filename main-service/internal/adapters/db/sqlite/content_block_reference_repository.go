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

var _ repositories.ContentBlockReferenceRepository = (*ContentBlockReferenceRepository)(nil)

// ContentBlockReferenceRepository implements the content block reference repository interface for SQLite
type ContentBlockReferenceRepository struct {
	db *DB
}

// NewContentBlockReferenceRepository creates a new content block reference repository
func NewContentBlockReferenceRepository(db *DB) *ContentBlockReferenceRepository {
	return &ContentBlockReferenceRepository{db: db}
}

// Create creates a new content block reference
func (r *ContentBlockReferenceRepository) Create(ctx context.Context, ref *story.ContentBlockReference) error {
	// Get tenant_id from content_block
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM content_blocks WHERE id = ?", ref.ContentBlockID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       ref.ContentBlockID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO content_block_references (id, tenant_id, content_block_id, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID.String(),
		tenantIDStr,
		ref.ContentBlockID.String(),
		string(ref.EntityType),
		ref.EntityID.String(),
		ref.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a content block reference by ID
func (r *ContentBlockReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = ? AND id = ?
	`
	var ref story.ContentBlockReference
	var idStr, contentBlockIDStr, entityIDStr, createdAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &contentBlockIDStr, &ref.EntityType, &entityIDStr, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block_reference",
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
	ref.ID = parsedID

	parsedContentBlockID, err := uuid.Parse(contentBlockIDStr)
	if err != nil {
		return nil, err
	}
	ref.ContentBlockID = parsedContentBlockID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	ref.EntityID = parsedEntityID

	// Parse timestamp
	ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// ListByContentBlock lists content block references for a content block
func (r *ContentBlockReferenceRepository) ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = ? AND content_block_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), contentBlockID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReferences(rows)
}

// ListByEntity lists content block references for an entity
func (r *ContentBlockReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), string(entityType), entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReferences(rows)
}

// Delete deletes a content block reference
func (r *ContentBlockReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_block_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByContentBlock deletes all content block references for a content block
func (r *ContentBlockReferenceRepository) DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error {
	query := `DELETE FROM content_block_references WHERE tenant_id = ? AND content_block_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), contentBlockID.String())
	return err
}

func (r *ContentBlockReferenceRepository) scanReferences(rows *sql.Rows) ([]*story.ContentBlockReference, error) {
	var references []*story.ContentBlockReference
	for rows.Next() {
		var ref story.ContentBlockReference
		var idStr, contentBlockIDStr, entityIDStr, createdAtStr string

		err := rows.Scan(&idStr, &contentBlockIDStr, &ref.EntityType, &entityIDStr, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ref.ID = parsedID

		parsedContentBlockID, err := uuid.Parse(contentBlockIDStr)
		if err != nil {
			return nil, err
		}
		ref.ContentBlockID = parsedContentBlockID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		ref.EntityID = parsedEntityID

		// Parse timestamp
		ref.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &ref)
	}

	return references, rows.Err()
}

