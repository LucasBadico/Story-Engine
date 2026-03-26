package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

var _ repositories.ConceptRepository = (*ConceptRepository)(nil)

// ConceptRepository implements the concept repository interface
type ConceptRepository struct {
	db *DB
}

// NewConceptRepository creates a new concept repository
func NewConceptRepository(db *DB) *ConceptRepository {
	return &ConceptRepository{db: db}
}

// Create creates a new concept
func (r *ConceptRepository) Create(ctx context.Context, concept *memory.Concept) error {
	metadataJSON, err := json.Marshal(concept.Metadata)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO concepts (id, tenant_id, story_id, name, type, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.Exec(ctx, query,
		concept.ID, concept.TenantID, concept.StoryID, concept.Name, string(concept.Type), metadataJSON, concept.CreatedAt, concept.UpdatedAt)
	return err
}

// GetByID retrieves a concept by ID
func (r *ConceptRepository) GetByID(ctx context.Context, id uuid.UUID) (*memory.Concept, error) {
	query := `
		SELECT id, tenant_id, story_id, name, type, metadata, created_at, updated_at
		FROM concepts
		WHERE id = $1
	`
	var concept memory.Concept
	var conceptType string
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&concept.ID, &concept.TenantID, &concept.StoryID, &concept.Name, &conceptType, &metadataJSON, &concept.CreatedAt, &concept.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	concept.Type = memory.ConceptType(conceptType)
	concept.Metadata = map[string]interface{}{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &concept.Metadata); err != nil {
			return nil, err
		}
	}
	return &concept, nil
}

// FindByName finds a concept by story, name and type (case-insensitive)
func (r *ConceptRepository) FindByName(ctx context.Context, storyID uuid.UUID, name string, conceptType memory.ConceptType) (*memory.Concept, error) {
	query := `
		SELECT id, tenant_id, story_id, name, type, metadata, created_at, updated_at
		FROM concepts
		WHERE story_id = $1 AND LOWER(name) = LOWER($2) AND type = $3
	`
	var concept memory.Concept
	var ct string
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, storyID, name, string(conceptType)).Scan(
		&concept.ID, &concept.TenantID, &concept.StoryID, &concept.Name, &ct, &metadataJSON, &concept.CreatedAt, &concept.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	concept.Type = memory.ConceptType(ct)
	concept.Metadata = map[string]interface{}{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &concept.Metadata); err != nil {
			return nil, err
		}
	}
	return &concept, nil
}

// List lists concepts by story and type
func (r *ConceptRepository) List(ctx context.Context, storyID uuid.UUID, conceptType memory.ConceptType) ([]*memory.Concept, error) {
	query := `
		SELECT id, tenant_id, story_id, name, type, metadata, created_at, updated_at
		FROM concepts
		WHERE story_id = $1 AND type = $2
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, storyID, string(conceptType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var concepts []*memory.Concept
	for rows.Next() {
		var concept memory.Concept
		var ct string
		var metadataJSON []byte
		if err := rows.Scan(&concept.ID, &concept.TenantID, &concept.StoryID, &concept.Name, &ct, &metadataJSON, &concept.CreatedAt, &concept.UpdatedAt); err != nil {
			return nil, err
		}
		concept.Type = memory.ConceptType(ct)
		concept.Metadata = map[string]interface{}{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &concept.Metadata); err != nil {
				return nil, err
			}
		}
		concepts = append(concepts, &concept)
	}

	return concepts, rows.Err()
}

// Update updates an existing concept
func (r *ConceptRepository) Update(ctx context.Context, concept *memory.Concept) error {
	metadataJSON, err := json.Marshal(concept.Metadata)
	if err != nil {
		return err
	}
	query := `
		UPDATE concepts
		SET name = $1, metadata = $2, updated_at = $3
		WHERE id = $4
	`
	_, err = r.db.Exec(ctx, query, concept.Name, metadataJSON, concept.UpdatedAt, concept.ID)
	return err
}

// Delete deletes a concept by ID
func (r *ConceptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM concepts WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
