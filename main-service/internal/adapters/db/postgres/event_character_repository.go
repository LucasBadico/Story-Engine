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

var _ repositories.EventCharacterRepository = (*EventCharacterRepository)(nil)

// EventCharacterRepository implements the event-character repository interface
type EventCharacterRepository struct {
	db *DB
}

// NewEventCharacterRepository creates a new event-character repository
func NewEventCharacterRepository(db *DB) *EventCharacterRepository {
	return &EventCharacterRepository{db: db}
}

// Create creates a new event-character relationship
func (r *EventCharacterRepository) Create(ctx context.Context, ec *world.EventCharacter) error {
	query := `
		INSERT INTO event_characters (id, event_id, character_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id, character_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, ec.ID, ec.EventID, ec.CharacterID, ec.Role, ec.CreatedAt)
	return err
}

// GetByID retrieves an event-character relationship by ID
func (r *EventCharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE id = $1
	`
	var ec world.EventCharacter

	err := r.db.QueryRow(ctx, query, id).Scan(
		&ec.ID, &ec.EventID, &ec.CharacterID, &ec.Role, &ec.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_character",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &ec, nil
}

// ListByEvent lists event-character relationships for an event
func (r *EventCharacterRepository) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE event_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventCharacters(rows)
}

// ListByCharacter lists event-character relationships for a character
func (r *EventCharacterRepository) ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE character_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventCharacters(rows)
}

// Delete deletes an event-character relationship
func (r *EventCharacterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByEventAndCharacter deletes an event-character relationship
func (r *EventCharacterRepository) DeleteByEventAndCharacter(ctx context.Context, eventID, characterID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE event_id = $1 AND character_id = $2`
	_, err := r.db.Exec(ctx, query, eventID, characterID)
	return err
}

// DeleteByEvent deletes all event-character relationships for an event
func (r *EventCharacterRepository) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE event_id = $1`
	_, err := r.db.Exec(ctx, query, eventID)
	return err
}

func (r *EventCharacterRepository) scanEventCharacters(rows pgx.Rows) ([]*world.EventCharacter, error) {
	eventCharacters := make([]*world.EventCharacter, 0)
	for rows.Next() {
		var ec world.EventCharacter
		err := rows.Scan(
			&ec.ID, &ec.EventID, &ec.CharacterID, &ec.Role, &ec.CreatedAt)
		if err != nil {
			return nil, err
		}
		eventCharacters = append(eventCharacters, &ec)
	}
	return eventCharacters, rows.Err()
}

