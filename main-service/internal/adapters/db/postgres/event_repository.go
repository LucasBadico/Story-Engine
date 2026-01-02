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

var _ repositories.EventRepository = (*EventRepository)(nil)

// EventRepository implements the event repository interface
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
		INSERT INTO events (id, world_id, name, type, description, timeline, importance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		e.ID, e.WorldID, e.Name, e.Type, e.Description, e.Timeline, e.Importance, e.CreatedAt, e.UpdatedAt)
	return err
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Event, error) {
	query := `
		SELECT id, world_id, name, type, description, timeline, importance, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	var e world.Event

	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.WorldID, &e.Name, &e.Type, &e.Description, &e.Timeline, &e.Importance, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &e, nil
}

// ListByWorld lists events for a world
func (r *EventRepository) ListByWorld(ctx context.Context, worldID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, world_id, name, type, description, timeline, importance, created_at, updated_at
		FROM events
		WHERE world_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, worldID)
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
		SET name = $2, type = $3, description = $4, timeline = $5, importance = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, e.ID, e.Name, e.Type, e.Description, e.Timeline, e.Importance, e.UpdatedAt)
	return err
}

// Delete deletes an event
func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *EventRepository) scanEvents(rows pgx.Rows) ([]*world.Event, error) {
	events := make([]*world.Event, 0)
	for rows.Next() {
		var e world.Event
		err := rows.Scan(
			&e.ID, &e.WorldID, &e.Name, &e.Type, &e.Description, &e.Timeline, &e.Importance, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}


