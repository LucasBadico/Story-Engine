package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.RelationshipStateRepository = (*RelationshipStateRepository)(nil)

// RelationshipStateRepository implements the relationship state repository interface
type RelationshipStateRepository struct {
	db *DB
}

// NewRelationshipStateRepository creates a new relationship state repository
func NewRelationshipStateRepository(db *DB) *RelationshipStateRepository {
	return &RelationshipStateRepository{db: db}
}

// Create creates a new relationship state
func (r *RelationshipStateRepository) Create(ctx context.Context, rs *world.RelationshipState) error {
	metadataJSON, err := json.Marshal(rs.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO relationship_states (
			id, tenant_id, chapter_id, chapter_number, tension_description,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		rs.ID, rs.TenantID, rs.ChapterID, rs.ChapterNumber, rs.TensionDescription,
		rs.TimelineDay, rs.TimelineConfidence, metadataJSON, rs.CreatedAt, rs.UpdatedAt)
	return err
}

// GetByID retrieves a relationship state by ID
func (r *RelationshipStateRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.RelationshipState, error) {
	query := `
		SELECT id, tenant_id, chapter_id, chapter_number, tension_description,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM relationship_states
		WHERE tenant_id = $1 AND id = $2
	`
	rs, err := r.scanRelationshipState(r.db.QueryRow(ctx, query, tenantID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "relationship_state",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Load entities
	entities, err := r.ListEntities(ctx, rs.ID)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		rs.Entities = append(rs.Entities, *e)
	}

	return rs, nil
}

// GetLatestByEntityPair retrieves the latest relationship state for an entity pair as of a given chapter
func (r *RelationshipStateRepository) GetLatestByEntityPair(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, targetType string, targetID uuid.UUID, asOfChapter int) (*world.RelationshipState, error) {
	query := `
		SELECT DISTINCT rs.id, rs.tenant_id, rs.chapter_id, rs.chapter_number, rs.tension_description,
			rs.timeline_day, rs.timeline_confidence, rs.metadata, rs.created_at, rs.updated_at
		FROM relationship_states rs
		JOIN relationship_state_entities rse1 ON rse1.relationship_state_id = rs.id
			AND rse1.role = 'source' AND rse1.entity_type = $2 AND rse1.entity_id = $3
		JOIN relationship_state_entities rse2 ON rse2.relationship_state_id = rs.id
			AND rse2.role = 'target' AND rse2.entity_type = $4 AND rse2.entity_id = $5
		WHERE rs.tenant_id = $1 AND rs.chapter_number <= $6
		ORDER BY rs.chapter_number DESC LIMIT 1
	`
	rs, err := r.scanRelationshipState(r.db.QueryRow(ctx, query, tenantID, sourceType, sourceID, targetType, targetID, asOfChapter))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "relationship_state",
				ID:       sourceID.String() + ":" + targetID.String(),
			}
		}
		return nil, err
	}

	// Load entities
	entities, err := r.ListEntities(ctx, rs.ID)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		rs.Entities = append(rs.Entities, *e)
	}

	return rs, nil
}

// ListByChapter lists relationship states for a chapter
func (r *RelationshipStateRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*world.RelationshipState, error) {
	query := `
		SELECT id, tenant_id, chapter_id, chapter_number, tension_description,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM relationship_states
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states, err := r.scanRelationshipStates(rows)
	if err != nil {
		return nil, err
	}

	// Load entities for each state
	for _, s := range states {
		entities, err := r.ListEntities(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		for _, e := range entities {
			s.Entities = append(s.Entities, *e)
		}
	}

	return states, nil
}

// Delete deletes a relationship state
func (r *RelationshipStateRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM relationship_states WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// AddEntity adds an entity link to a relationship state
func (r *RelationshipStateRepository) AddEntity(ctx context.Context, entity *world.RelationshipStateEntity) error {
	query := `
		INSERT INTO relationship_state_entities (id, relationship_state_id, role, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		entity.ID, entity.RelationshipStateID, entity.Role, entity.EntityType, entity.EntityID, entity.CreatedAt)
	return err
}

// ListEntities lists entity links for a relationship state
func (r *RelationshipStateRepository) ListEntities(ctx context.Context, relationshipStateID uuid.UUID) ([]*world.RelationshipStateEntity, error) {
	query := `
		SELECT id, relationship_state_id, role, entity_type, entity_id, created_at
		FROM relationship_state_entities
		WHERE relationship_state_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, relationshipStateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entities := make([]*world.RelationshipStateEntity, 0)
	for rows.Next() {
		var e world.RelationshipStateEntity
		if err := rows.Scan(&e.ID, &e.RelationshipStateID, &e.Role, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, err
		}
		entities = append(entities, &e)
	}
	return entities, rows.Err()
}

func (r *RelationshipStateRepository) scanRelationshipState(row pgx.Row) (*world.RelationshipState, error) {
	var rs world.RelationshipState
	var metadataJSON []byte

	err := row.Scan(
		&rs.ID, &rs.TenantID, &rs.ChapterID, &rs.ChapterNumber, &rs.TensionDescription,
		&rs.TimelineDay, &rs.TimelineConfidence, &metadataJSON, &rs.CreatedAt, &rs.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &rs.Metadata); err != nil {
			return nil, err
		}
	} else {
		rs.Metadata = map[string]interface{}{}
	}

	rs.Entities = []world.RelationshipStateEntity{}

	return &rs, nil
}

func (r *RelationshipStateRepository) scanRelationshipStates(rows pgx.Rows) ([]*world.RelationshipState, error) {
	states := make([]*world.RelationshipState, 0)
	for rows.Next() {
		var rs world.RelationshipState
		var metadataJSON []byte

		err := rows.Scan(
			&rs.ID, &rs.TenantID, &rs.ChapterID, &rs.ChapterNumber, &rs.TensionDescription,
			&rs.TimelineDay, &rs.TimelineConfidence, &metadataJSON, &rs.CreatedAt, &rs.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &rs.Metadata); err != nil {
				return nil, err
			}
		} else {
			rs.Metadata = map[string]interface{}{}
		}

		rs.Entities = []world.RelationshipStateEntity{}
		states = append(states, &rs)
	}

	return states, rows.Err()
}
