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

var _ repositories.EventRepository = (*EventRepository)(nil)

// EventRepository implements the event repository interface for SQLite
type EventRepository struct {
	db *DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create creates a new event
func (r *EventRepository) Create(ctx context.Context, e *world.Event) error {
	query := `
		INSERT INTO events (id, tenant_id, world_id, name, type, description, timeline, importance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var eventType sql.NullString
	if e.Type != nil {
		eventType = sql.NullString{String: *e.Type, Valid: true}
	}

	var description sql.NullString
	if e.Description != nil {
		description = sql.NullString{String: *e.Description, Valid: true}
	}

	var timeline sql.NullString
	if e.Timeline != nil {
		timeline = sql.NullString{String: *e.Timeline, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		e.ID.String(),
		e.TenantID.String(),
		e.WorldID.String(),
		e.Name,
		eventType,
		description,
		timeline,
		e.Importance,
		e.CreatedAt.Format(time.RFC3339),
		e.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, created_at, updated_at
		FROM events
		WHERE tenant_id = ? AND id = ?
	`
	var e world.Event
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var eventType, description, timeline sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &e.Name, &eventType, &description, &timeline, &e.Importance, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event",
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
	e.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	e.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	e.WorldID = parsedWorldID

	// Parse timestamps
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	e.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable strings
	if eventType.Valid {
		e.Type = &eventType.String
	}
	if description.Valid {
		e.Description = &description.String
	}
	if timeline.Valid {
		e.Timeline = &timeline.String
	}

	return &e, nil
}

// ListByWorld lists events for a world
func (r *EventRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, created_at, updated_at
		FROM events
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// Update updates an event
func (r *EventRepository) Update(ctx context.Context, e *world.Event) error {
	query := `
		UPDATE events
		SET name = ?, type = ?, description = ?, timeline = ?, importance = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var eventType sql.NullString
	if e.Type != nil {
		eventType = sql.NullString{String: *e.Type, Valid: true}
	}

	var description sql.NullString
	if e.Description != nil {
		description = sql.NullString{String: *e.Description, Valid: true}
	}

	var timeline sql.NullString
	if e.Timeline != nil {
		timeline = sql.NullString{String: *e.Timeline, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		e.Name,
		eventType,
		description,
		timeline,
		e.Importance,
		e.UpdatedAt.Format(time.RFC3339),
		e.TenantID.String(),
		e.ID.String(),
	)
	return err
}

// Delete deletes an event
func (r *EventRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM events WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

func (r *EventRepository) scanEvents(rows *sql.Rows) ([]*world.Event, error) {
	events := make([]*world.Event, 0)
	for rows.Next() {
		var e world.Event
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var eventType, description, timeline sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &e.Name, &eventType, &description, &timeline, &e.Importance, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		e.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		e.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		e.WorldID = parsedWorldID

		// Parse timestamps
		e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		e.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable strings
		if eventType.Valid {
			e.Type = &eventType.String
		}
		if description.Valid {
			e.Description = &description.String
		}
		if timeline.Valid {
			e.Timeline = &timeline.String
		}

		events = append(events, &e)
	}

	return events, rows.Err()
}

