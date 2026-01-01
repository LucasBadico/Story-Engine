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

var _ repositories.EventLocationRepository = (*EventLocationRepository)(nil)

// EventLocationRepository implements the event-location repository interface
type EventLocationRepository struct {
	db *DB
}

// NewEventLocationRepository creates a new event-location repository
func NewEventLocationRepository(db *DB) *EventLocationRepository {
	return &EventLocationRepository{db: db}
}

// Create creates a new event-location relationship
func (r *EventLocationRepository) Create(ctx context.Context, el *world.EventLocation) error {
	query := `
		INSERT INTO event_locations (id, event_id, location_id, significance, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id, location_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, el.ID, el.EventID, el.LocationID, el.Significance, el.CreatedAt)
	return err
}

// GetByID retrieves an event-location relationship by ID
func (r *EventLocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE id = $1
	`
	var el world.EventLocation

	err := r.db.QueryRow(ctx, query, id).Scan(
		&el.ID, &el.EventID, &el.LocationID, &el.Significance, &el.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_location",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &el, nil
}

// ListByEvent lists event-location relationships for an event
func (r *EventLocationRepository) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE event_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventLocations(rows)
}

// ListByLocation lists event-location relationships for a location
func (r *EventLocationRepository) ListByLocation(ctx context.Context, locationID uuid.UUID) ([]*world.EventLocation, error) {
	query := `
		SELECT id, event_id, location_id, significance, created_at
		FROM event_locations
		WHERE location_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventLocations(rows)
}

// Delete deletes an event-location relationship
func (r *EventLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByEventAndLocation deletes an event-location relationship
func (r *EventLocationRepository) DeleteByEventAndLocation(ctx context.Context, eventID, locationID uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE event_id = $1 AND location_id = $2`
	_, err := r.db.Exec(ctx, query, eventID, locationID)
	return err
}

// DeleteByEvent deletes all event-location relationships for an event
func (r *EventLocationRepository) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM event_locations WHERE event_id = $1`
	_, err := r.db.Exec(ctx, query, eventID)
	return err
}

func (r *EventLocationRepository) scanEventLocations(rows pgx.Rows) ([]*world.EventLocation, error) {
	eventLocations := make([]*world.EventLocation, 0)
	for rows.Next() {
		var el world.EventLocation
		err := rows.Scan(
			&el.ID, &el.EventID, &el.LocationID, &el.Significance, &el.CreatedAt)
		if err != nil {
			return nil, err
		}
		eventLocations = append(eventLocations, &el)
	}
	return eventLocations, rows.Err()
}

