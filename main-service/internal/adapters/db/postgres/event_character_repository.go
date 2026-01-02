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
	// Get tenant_id from event
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = $1", ec.EventID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO event_characters (id, tenant_id, event_id, character_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_id, character_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, ec.ID, tenantID, ec.EventID, ec.CharacterID, ec.Role, ec.CreatedAt)
	return err
}

// GetByID retrieves an event-character relationship by ID
func (r *EventCharacterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE tenant_id = $1 AND id = $2
	`
	var ec world.EventCharacter

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
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
func (r *EventCharacterRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE tenant_id = $1 AND event_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventCharacters(rows)
}

// ListByCharacter lists event-character relationships for a character
func (r *EventCharacterRepository) ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE tenant_id = $1 AND character_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventCharacters(rows)
}

// Delete deletes an event-character relationship
func (r *EventCharacterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByEventAndCharacter deletes an event-character relationship
func (r *EventCharacterRepository) DeleteByEventAndCharacter(ctx context.Context, tenantID, eventID, characterID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = $1 AND event_id = $2 AND character_id = $3`
	_, err := r.db.Exec(ctx, query, tenantID, eventID, characterID)
	return err
}

// DeleteByEvent deletes all event-character relationships for an event
func (r *EventCharacterRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = $1 AND event_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, eventID)
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


