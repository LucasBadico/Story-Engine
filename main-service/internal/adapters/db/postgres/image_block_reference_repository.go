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

var _ repositories.ImageBlockReferenceRepository = (*ImageBlockReferenceRepository)(nil)

// ImageBlockReferenceRepository implements the image block reference repository interface
type ImageBlockReferenceRepository struct {
	db *DB
}

// NewImageBlockReferenceRepository creates a new image block reference repository
func NewImageBlockReferenceRepository(db *DB) *ImageBlockReferenceRepository {
	return &ImageBlockReferenceRepository{db: db}
}

// Create creates a new image block reference
func (r *ImageBlockReferenceRepository) Create(ctx context.Context, ref *story.ImageBlockReference) error {
	// Get tenant_id from image_block
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM image_blocks WHERE id = $1", ref.ImageBlockID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO image_block_references (id, tenant_id, image_block_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID, tenantID, ref.ImageBlockID, string(ref.EntityType), ref.EntityID, ref.CreatedAt)
	return err
}

// GetByID retrieves an image block reference by ID
func (r *ImageBlockReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ImageBlockReference, error) {
	query := `
		SELECT id, image_block_id, entity_type, entity_id, created_at
		FROM image_block_references
		WHERE tenant_id = $1 AND id = $2
	`
	var ref story.ImageBlockReference
	var entityTypeStr string

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ref.ID, &ref.ImageBlockID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "image_block_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	ref.EntityType = story.ImageBlockReferenceEntityType(entityTypeStr)
	return &ref, nil
}

// ListByImageBlock lists image block references for an image block
func (r *ImageBlockReferenceRepository) ListByImageBlock(ctx context.Context, tenantID, imageBlockID uuid.UUID) ([]*story.ImageBlockReference, error) {
	query := `
		SELECT id, image_block_id, entity_type, entity_id, created_at
		FROM image_block_references
		WHERE tenant_id = $1 AND image_block_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, imageBlockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanImageBlockReferences(rows)
}

// ListByEntity lists image block references for an entity
func (r *ImageBlockReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.ImageBlockReferenceEntityType, entityID uuid.UUID) ([]*story.ImageBlockReference, error) {
	query := `
		SELECT id, image_block_id, entity_type, entity_id, created_at
		FROM image_block_references
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanImageBlockReferences(rows)
}

// Delete deletes an image block reference
func (r *ImageBlockReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM image_block_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByImageBlock deletes all image block references for an image block
func (r *ImageBlockReferenceRepository) DeleteByImageBlock(ctx context.Context, tenantID, imageBlockID uuid.UUID) error {
	query := `DELETE FROM image_block_references WHERE tenant_id = $1 AND image_block_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, imageBlockID)
	return err
}

// DeleteByImageBlockAndEntity deletes a specific image block reference
func (r *ImageBlockReferenceRepository) DeleteByImageBlockAndEntity(ctx context.Context, tenantID, imageBlockID uuid.UUID, entityType story.ImageBlockReferenceEntityType, entityID uuid.UUID) error {
	query := `DELETE FROM image_block_references WHERE tenant_id = $1 AND image_block_id = $2 AND entity_type = $3 AND entity_id = $4`
	_, err := r.db.Exec(ctx, query, tenantID, imageBlockID, string(entityType), entityID)
	return err
}

func (r *ImageBlockReferenceRepository) scanImageBlockReferences(rows pgx.Rows) ([]*story.ImageBlockReference, error) {
	references := make([]*story.ImageBlockReference, 0)
	for rows.Next() {
		var ref story.ImageBlockReference
		var entityTypeStr string

		err := rows.Scan(
			&ref.ID, &ref.ImageBlockID, &entityTypeStr, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}

		ref.EntityType = story.ImageBlockReferenceEntityType(entityTypeStr)
		references = append(references, &ref)
	}

	return references, rows.Err()
}


