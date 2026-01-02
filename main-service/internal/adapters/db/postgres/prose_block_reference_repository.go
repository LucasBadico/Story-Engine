package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ProseBlockReferenceRepository = (*ProseBlockReferenceRepository)(nil)

// ProseBlockReferenceRepository implements the prose block reference repository interface
type ProseBlockReferenceRepository struct {
	db *DB
}

// NewProseBlockReferenceRepository creates a new prose block reference repository
func NewProseBlockReferenceRepository(db *DB) *ProseBlockReferenceRepository {
	return &ProseBlockReferenceRepository{db: db}
}

// Create creates a new prose block reference
func (r *ProseBlockReferenceRepository) Create(ctx context.Context, ref *story.ProseBlockReference) error {
	// Get tenant_id from prose_block
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM prose_blocks WHERE id = $1", ref.ProseBlockID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO prose_block_references (id, tenant_id, prose_block_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		ref.ID, tenantID, ref.ProseBlockID, string(ref.EntityType), ref.EntityID, ref.CreatedAt)
	return err
}

// GetByID retrieves a prose block reference by ID
func (r *ProseBlockReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.ProseBlockReference, error) {
	query := `
		SELECT id, prose_block_id, entity_type, entity_id, created_at
		FROM prose_block_references
		WHERE tenant_id = $1 AND id = $2
	`
	var ref story.ProseBlockReference
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ref.ID, &ref.ProseBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("prose block reference not found")
		}
		return nil, err
	}
	return &ref, nil
}

// ListByProseBlock lists prose block references for a prose block
func (r *ProseBlockReferenceRepository) ListByProseBlock(ctx context.Context, tenantID, proseBlockID uuid.UUID) ([]*story.ProseBlockReference, error) {
	query := `
		SELECT id, prose_block_id, entity_type, entity_id, created_at
		FROM prose_block_references
		WHERE tenant_id = $1 AND prose_block_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, proseBlockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*story.ProseBlockReference
	for rows.Next() {
		var ref story.ProseBlockReference
		err := rows.Scan(&ref.ID, &ref.ProseBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// ListByEntity lists prose block references for an entity
func (r *ProseBlockReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType story.EntityType, entityID uuid.UUID) ([]*story.ProseBlockReference, error) {
	query := `
		SELECT id, prose_block_id, entity_type, entity_id, created_at
		FROM prose_block_references
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*story.ProseBlockReference
	for rows.Next() {
		var ref story.ProseBlockReference
		err := rows.Scan(&ref.ID, &ref.ProseBlockID, &ref.EntityType, &ref.EntityID, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// Delete deletes a prose block reference
func (r *ProseBlockReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM prose_block_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByProseBlock deletes all prose block references for a prose block
func (r *ProseBlockReferenceRepository) DeleteByProseBlock(ctx context.Context, tenantID, proseBlockID uuid.UUID) error {
	query := `DELETE FROM prose_block_references WHERE tenant_id = $1 AND prose_block_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, proseBlockID)
	return err
}


