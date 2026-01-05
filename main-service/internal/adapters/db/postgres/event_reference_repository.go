package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.EventReferenceRepository = (*EventReferenceRepository)(nil)

// EventReferenceRepository implements the event reference repository interface
type EventReferenceRepository struct {
	db *DB
}

// NewEventReferenceRepository creates a new event reference repository
func NewEventReferenceRepository(db *DB) *EventReferenceRepository {
	return &EventReferenceRepository{db: db}
}

// Create creates a new event reference
func (r *EventReferenceRepository) Create(ctx context.Context, er *world.EventReference) error {
	// Get tenant_id from event
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = $1", er.EventID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "event",
				ID:       er.EventID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (event_id, entity_type, entity_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		er.ID, tenantID, er.EventID, er.EntityType, er.EntityID, er.RelationshipType, er.Notes, er.CreatedAt)
	return err
}

// GetByID retrieves an event reference by ID
func (r *EventReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventReference, error) {
	query := `
		SELECT id, event_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM event_references
		WHERE tenant_id = $1 AND id = $2
	`
	var er world.EventReference

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&er.ID, &er.EventID, &er.EntityType, &er.EntityID, &er.RelationshipType, &er.Notes, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &er, nil
}

// ListByEvent lists event references for an event
func (r *EventReferenceRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventReference, error) {
	query := `
		SELECT id, event_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM event_references
		WHERE tenant_id = $1 AND event_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventReferences(rows)
}

// ListByEntity lists event references for an entity
func (r *EventReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.EventReference, error) {
	query := `
		SELECT id, event_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM event_references
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventReferences(rows)
}

// Update updates an event reference
func (r *EventReferenceRepository) Update(ctx context.Context, er *world.EventReference) error {
	// Get tenant_id from event
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = $1", er.EventID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE event_references
		SET relationship_type = $2, notes = $3
		WHERE tenant_id = $4 AND id = $1
	`
	result, err := r.db.Exec(ctx, query, er.ID, er.RelationshipType, er.Notes, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "event_reference",
			ID:       er.ID.String(),
		}
	}
	return nil
}

// Delete deletes an event reference
func (r *EventReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByEventAndEntity deletes a specific event reference
func (r *EventReferenceRepository) DeleteByEventAndEntity(ctx context.Context, tenantID, eventID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = $1 AND event_id = $2 AND entity_type = $3 AND entity_id = $4`
	result, err := r.db.Exec(ctx, query, tenantID, eventID, entityType, entityID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "event_reference",
			ID:       eventID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByEvent deletes all event references for an event
func (r *EventReferenceRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = $1 AND event_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, eventID)
	return err
}

func (r *EventReferenceRepository) scanEventReferences(rows pgx.Rows) ([]*world.EventReference, error) {
	references := make([]*world.EventReference, 0)
	for rows.Next() {
		var er world.EventReference
		err := rows.Scan(
			&er.ID, &er.EventID, &er.EntityType, &er.EntityID, &er.RelationshipType, &er.Notes, &er.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &er)
	}
	return references, rows.Err()
}

