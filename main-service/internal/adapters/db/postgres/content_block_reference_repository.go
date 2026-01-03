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

var _ repositories.ContentBlockReferenceRepository = (*ContentBlockReferenceRepository)(nil)

// ContentBlockReferenceRepository implements the content block reference repository interface
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
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM content_blocks WHERE id = $1", ref.ContentBlockID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "content_block",
				ID:       ref.ContentBlockID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO content_block_references (id, tenant_id, content_block_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID, tenantID, ref.ContentBlockID, string(ref.EntityType), ref.EntityID, ref.CreatedAt)
	return err
}

// GetByID retrieves a content block reference by ID
func (r *ContentBlockReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = $1 AND id = $2
	`
	var ref story.ContentBlockReference
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ref.ID, &ref.ContentBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "content_block_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	return &ref, nil
}

// ListByContentBlock lists content block references for a content block
func (r *ContentBlockReferenceRepository) ListByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) ([]*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = $1 AND content_block_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, contentBlockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*story.ContentBlockReference
	for rows.Next() {
		var ref story.ContentBlockReference
		err := rows.Scan(&ref.ID, &ref.ContentBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// ListByEntity lists content block references for an entity
func (r *ContentBlockReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ContentBlockReference, error) {
	query := `
		SELECT id, content_block_id, entity_type, entity_id, created_at
		FROM content_block_references
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*story.ContentBlockReference
	for rows.Next() {
		var ref story.ContentBlockReference
		err := rows.Scan(&ref.ID, &ref.ContentBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// Delete deletes a content block reference
func (r *ContentBlockReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM content_block_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByContentBlock deletes all content block references for a content block
func (r *ContentBlockReferenceRepository) DeleteByContentBlock(ctx context.Context, tenantID, contentBlockID uuid.UUID) error {
	query := `DELETE FROM content_block_references WHERE tenant_id = $1 AND content_block_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, contentBlockID)
	return err
}
