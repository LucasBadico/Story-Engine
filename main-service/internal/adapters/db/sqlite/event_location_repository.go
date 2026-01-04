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

var _ repositories.EventLocationRepository = (*EventLocationRepository)(nil)

// EventLocationRepository implements the event-location repository interface for SQLite
type EventLocationRepository struct {
	db *DB
}

// NewEventLocationRepository creates a new event-location repository
func NewEventLocationRepository(db *DB) *EventLocationRepository {
	return &EventLocationRepository{db: db}
}

// Create creates a new event-location relationship
func (r *EventLocationRepository) Create(ctx context.Context, el *world.EventLocation) error {
	// Get tenant_id from event
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = ?", el.EventID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO event_locations (id, tenant_id, event_id, location_id, significance, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var significance sql.NullString
	if el.Significance != nil {
		significance = sql.NullString{String: *el.Significance, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		el.ID.String(),
		tenantID.String(),
		el.EventID.String(),
		el.LocationID.String(),
		significance,
		el.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event-location relationship by ID
func (r *EventLocationRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE tenant_id = ? AND id = ?
	`
	var el world.EventLocation
	var idStr, eventIDStr, locationIDStr, createdAtStr string
	var significance sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &eventIDStr, &locationIDStr, &significance, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_location",
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
	el.ID = parsedID

	parsedEventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return nil, err
	}
	el.EventID = parsedEventID

	parsedLocationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		return nil, err
	}
	el.LocationID = parsedLocationID

	// Parse nullable string
	if significance.Valid {
		el.Significance = &significance.String
	}

	// Parse timestamp
	el.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &el, nil
}

// ListByEvent lists event-location relationships for an event
func (r *EventLocationRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE tenant_id = ? AND event_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), eventID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventLocations(rows)
}

// ListByLocation lists event-location relationships for a location
func (r *EventLocationRepository) ListByLocation(ctx context.Context, tenantID, locationID uuid.UUID) ([]*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE tenant_id = ? AND location_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), locationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventLocations(rows)
}

// Delete deletes an event-location relationship
func (r *EventLocationRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByEventAndLocation deletes an event-location relationship
func (r *EventLocationRepository) DeleteByEventAndLocation(ctx context.Context, tenantID, eventID, locationID uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE tenant_id = ? AND event_id = ? AND location_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String(), locationID.String())
	return err
}

// DeleteByEvent deletes all event-location relationships for an event
func (r *EventLocationRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE tenant_id = ? AND event_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String())
	return err
}

func (r *EventLocationRepository) scanEventLocations(rows *sql.Rows) ([]*world.EventLocation, error) {
	eventLocations := make([]*world.EventLocation, 0)
	for rows.Next() {
		var el world.EventLocation
		var idStr, eventIDStr, locationIDStr, createdAtStr string
		var significance sql.NullString

		err := rows.Scan(
			&idStr, &eventIDStr, &locationIDStr, &significance, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		el.ID = parsedID

		parsedEventID, err := uuid.Parse(eventIDStr)
		if err != nil {
			return nil, err
		}
		el.EventID = parsedEventID

		parsedLocationID, err := uuid.Parse(locationIDStr)
		if err != nil {
			return nil, err
		}
		el.LocationID = parsedLocationID

		// Parse nullable string
		if significance.Valid {
			el.Significance = &significance.String
		}

		// Parse timestamp
		el.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		eventLocations = append(eventLocations, &el)
	}
	return eventLocations, rows.Err()
}

