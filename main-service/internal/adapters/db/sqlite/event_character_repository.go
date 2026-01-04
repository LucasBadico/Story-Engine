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

var _ repositories.EventCharacterRepository = (*EventCharacterRepository)(nil)

// EventCharacterRepository implements the event-character repository interface for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = ?", ec.EventID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO event_characters (id, tenant_id, event_id, character_id, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var role sql.NullString
	if ec.Role != nil {
		role = sql.NullString{String: *ec.Role, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		ec.ID.String(),
		tenantID.String(),
		ec.EventID.String(),
		ec.CharacterID.String(),
		role,
		ec.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event-character relationship by ID
func (r *EventCharacterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE tenant_id = ? AND id = ?
	`
	var ec world.EventCharacter
	var idStr, eventIDStr, characterIDStr, createdAtStr string
	var role sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &eventIDStr, &characterIDStr, &role, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_character",
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
	ec.ID = parsedID

	parsedEventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return nil, err
	}
	ec.EventID = parsedEventID

	parsedCharacterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		return nil, err
	}
	ec.CharacterID = parsedCharacterID

	// Parse nullable string
	if role.Valid {
		ec.Role = &role.String
	}

	// Parse timestamp
	ec.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ec, nil
}

// ListByEvent lists event-character relationships for an event
func (r *EventCharacterRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventCharacter, error) {
	query := `
		SELECT id, event_id, character_id, role, created_at
		FROM event_characters
		WHERE tenant_id = ? AND event_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), eventID.String())
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
		WHERE tenant_id = ? AND character_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), characterID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventCharacters(rows)
}

// Delete deletes an event-character relationship
func (r *EventCharacterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByEventAndCharacter deletes an event-character relationship
func (r *EventCharacterRepository) DeleteByEventAndCharacter(ctx context.Context, tenantID, eventID, characterID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = ? AND event_id = ? AND character_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String(), characterID.String())
	return err
}

// DeleteByEvent deletes all event-character relationships for an event
func (r *EventCharacterRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_characters WHERE tenant_id = ? AND event_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String())
	return err
}

func (r *EventCharacterRepository) scanEventCharacters(rows *sql.Rows) ([]*world.EventCharacter, error) {
	eventCharacters := make([]*world.EventCharacter, 0)
	for rows.Next() {
		var ec world.EventCharacter
		var idStr, eventIDStr, characterIDStr, createdAtStr string
		var role sql.NullString

		err := rows.Scan(
			&idStr, &eventIDStr, &characterIDStr, &role, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ec.ID = parsedID

		parsedEventID, err := uuid.Parse(eventIDStr)
		if err != nil {
			return nil, err
		}
		ec.EventID = parsedEventID

		parsedCharacterID, err := uuid.Parse(characterIDStr)
		if err != nil {
			return nil, err
		}
		ec.CharacterID = parsedCharacterID

		// Parse nullable string
		if role.Valid {
			ec.Role = &role.String
		}

		// Parse timestamp
		ec.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		eventCharacters = append(eventCharacters, &ec)
	}
	return eventCharacters, rows.Err()
}

