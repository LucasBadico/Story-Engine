package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.EventReferenceRepository = (*EventReferenceRepository)(nil)

// EventReferenceRepository implements the event reference repository interface for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = ?", er.EventID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "event",
				ID:       er.EventID.String(),
			}
		}
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO event_references (id, tenant_id, event_id, entity_type, entity_id, relationship_type, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var relationshipType sql.NullString
	if er.RelationshipType != nil {
		relationshipType = sql.NullString{String: *er.RelationshipType, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		er.ID.String(),
		tenantID.String(),
		er.EventID.String(),
		er.EntityType,
		er.EntityID.String(),
		relationshipType,
		er.Notes,
		er.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event reference by ID
func (r *EventReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventReference, error) {
	query := `
		SELECT id, event_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM event_references
		WHERE tenant_id = ? AND id = ?
	`
	var er world.EventReference
	var idStr, eventIDStr, entityIDStr, createdAtStr string
	var relationshipType sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &eventIDStr, &er.EntityType, &entityIDStr, &relationshipType, &er.Notes, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_reference",
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
	er.ID = parsedID

	parsedEventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return nil, err
	}
	er.EventID = parsedEventID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	er.EntityID = parsedEntityID

	// Parse nullable string
	if relationshipType.Valid {
		er.RelationshipType = &relationshipType.String
	}

	// Parse timestamp
	er.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &er, nil
}

// ListByEvent lists event references for an event
func (r *EventReferenceRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventReference, error) {
	query := `
		SELECT id, event_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM event_references
		WHERE tenant_id = ? AND event_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), eventID.String())
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
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), entityType, entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventReferences(rows)
}

// Update updates an event reference
func (r *EventReferenceRepository) Update(ctx context.Context, er *world.EventReference) error {
	// Get tenant_id from event
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = ?", er.EventID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		UPDATE event_references
		SET relationship_type = ?, notes = ?
		WHERE tenant_id = ? AND id = ?
	`

	var relationshipType sql.NullString
	if er.RelationshipType != nil {
		relationshipType = sql.NullString{String: *er.RelationshipType, Valid: true}
	}

	result, err := r.db.Exec(ctx, query, relationshipType, er.Notes, tenantID.String(), er.ID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "event_reference",
			ID:       er.ID.String(),
		}
	}
	return nil
}

// Delete deletes an event reference
func (r *EventReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByEventAndEntity deletes a specific event reference
func (r *EventReferenceRepository) DeleteByEventAndEntity(ctx context.Context, tenantID, eventID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = ? AND event_id = ? AND entity_type = ? AND entity_id = ?`
	result, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String(), entityType, entityID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "event_reference",
			ID:       eventID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByEvent deletes all event references for an event
func (r *EventReferenceRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_references WHERE tenant_id = ? AND event_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String())
	return err
}

func (r *EventReferenceRepository) scanEventReferences(rows *sql.Rows) ([]*world.EventReference, error) {
	references := make([]*world.EventReference, 0)
	for rows.Next() {
		var er world.EventReference
		var idStr, eventIDStr, entityIDStr, createdAtStr string
		var relationshipType sql.NullString

		err := rows.Scan(
			&idStr, &eventIDStr, &er.EntityType, &entityIDStr, &relationshipType, &er.Notes, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		er.ID = parsedID

		parsedEventID, err := uuid.Parse(eventIDStr)
		if err != nil {
			return nil, err
		}
		er.EventID = parsedEventID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		er.EntityID = parsedEntityID

		// Parse nullable string
		if relationshipType.Valid {
			er.RelationshipType = &relationshipType.String
		}

		// Parse timestamp
		er.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &er)
	}

	return references, rows.Err()
}

