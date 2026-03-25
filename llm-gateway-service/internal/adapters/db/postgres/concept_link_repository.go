package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

var _ repositories.ConceptLinkRepository = (*ConceptLinkRepository)(nil)

// ConceptLinkRepository implements the concept link repository interface
type ConceptLinkRepository struct {
	db *DB
}

// NewConceptLinkRepository creates a new concept link repository
func NewConceptLinkRepository(db *DB) *ConceptLinkRepository {
	return &ConceptLinkRepository{db: db}
}

// Create creates a new concept link
func (r *ConceptLinkRepository) Create(ctx context.Context, link *memory.ConceptLink) error {
	metadataJSON, err := json.Marshal(link.Metadata)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO concept_links (id, tenant_id, source_concept_id, target_concept_id, relation_type, summary, metadata, valid_from_chapter, valid_to_chapter, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		link.ID, link.TenantID, link.SourceConceptID, link.TargetConceptID, link.RelationType, link.Summary, metadataJSON, link.ValidFromChapter, link.ValidToChapter, link.CreatedAt)
	return err
}

// GetByID retrieves a concept link by ID
func (r *ConceptLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.ConceptLink, error) {
	query := `
		SELECT id, tenant_id, source_concept_id, target_concept_id, relation_type, summary, metadata, valid_from_chapter, valid_to_chapter, created_at
		FROM concept_links
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

// FindByPair finds the current (non-closed) link between two concepts
func (r *ConceptLinkRepository) FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) (*memory.ConceptLink, error) {
	query := `
		SELECT id, tenant_id, source_concept_id, target_concept_id, relation_type, summary, metadata, valid_from_chapter, valid_to_chapter, created_at
		FROM concept_links
		WHERE source_concept_id = $1 AND target_concept_id = $2 AND valid_to_chapter IS NULL
	`
	return r.scanOne(ctx, query, sourceID, targetID)
}

// List lists concept links for a story with optional chapter filter
func (r *ConceptLinkRepository) List(ctx context.Context, storyID uuid.UUID, asOfChapter *int) ([]*memory.ConceptLink, error) {
	query := `
		SELECT cl.id, cl.tenant_id, cl.source_concept_id, cl.target_concept_id, cl.relation_type, cl.summary, cl.metadata, cl.valid_from_chapter, cl.valid_to_chapter, cl.created_at
		FROM concept_links cl
		INNER JOIN concepts c ON cl.source_concept_id = c.id
		WHERE c.story_id = $1
	`
	args := []interface{}{storyID}
	argIndex := 2

	if asOfChapter != nil {
		query += fmt.Sprintf(" AND cl.valid_from_chapter <= $%d AND (cl.valid_to_chapter IS NULL OR cl.valid_to_chapter > $%d)", argIndex, argIndex)
		args = append(args, *asOfChapter)
		argIndex++
	}

	query += " ORDER BY cl.created_at"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByConceptID lists all links where the concept is source or target
func (r *ConceptLinkRepository) ListByConceptID(ctx context.Context, conceptID uuid.UUID) ([]*memory.ConceptLink, error) {
	query := `
		SELECT id, tenant_id, source_concept_id, target_concept_id, relation_type, summary, metadata, valid_from_chapter, valid_to_chapter, created_at
		FROM concept_links
		WHERE source_concept_id = $1 OR target_concept_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, conceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// CloseVersion sets the valid_to_chapter on a concept link
func (r *ConceptLinkRepository) CloseVersion(ctx context.Context, id uuid.UUID, validToChapter int) error {
	query := `UPDATE concept_links SET valid_to_chapter = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, validToChapter)
	return err
}

// Delete deletes a concept link by ID
func (r *ConceptLinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM concept_links WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// scanOne scans a single concept link row
func (r *ConceptLinkRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*memory.ConceptLink, error) {
	var link memory.ConceptLink
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&link.ID, &link.TenantID, &link.SourceConceptID, &link.TargetConceptID, &link.RelationType, &link.Summary, &metadataJSON, &link.ValidFromChapter, &link.ValidToChapter, &link.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	link.Metadata = map[string]interface{}{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
			return nil, err
		}
	}
	return &link, nil
}

// scanRows scans multiple concept link rows
func (r *ConceptLinkRepository) scanRows(rows pgx.Rows) ([]*memory.ConceptLink, error) {
	var links []*memory.ConceptLink
	for rows.Next() {
		var link memory.ConceptLink
		var metadataJSON []byte
		if err := rows.Scan(
			&link.ID, &link.TenantID, &link.SourceConceptID, &link.TargetConceptID, &link.RelationType, &link.Summary, &metadataJSON, &link.ValidFromChapter, &link.ValidToChapter, &link.CreatedAt); err != nil {
			return nil, err
		}
		link.Metadata = map[string]interface{}{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &link.Metadata); err != nil {
				return nil, err
			}
		}
		links = append(links, &link)
	}
	return links, rows.Err()
}
