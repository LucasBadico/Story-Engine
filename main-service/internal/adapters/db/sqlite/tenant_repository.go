package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.TenantRepository = (*TenantRepository)(nil)

// TenantRepository implements the tenant repository interface for SQLite
type TenantRepository struct {
	db *DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, t *tenant.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, status, active_llm_profile_id, created_at, updated_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var activeLLMProfileID sql.NullString
	if t.ActiveLLMProfileID != nil {
		activeLLMProfileID = sql.NullString{String: t.ActiveLLMProfileID.String(), Valid: true}
	}

	var createdBy sql.NullString
	if t.CreatedBy != nil {
		createdBy = sql.NullString{String: t.CreatedBy.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		t.ID.String(),
		t.Name,
		string(t.Status),
		activeLLMProfileID,
		t.CreatedAt.Format(time.RFC3339),
		t.UpdatedAt.Format(time.RFC3339),
		createdBy,
	)
	if err != nil {
		// Check for unique constraint violation (SQLite error code 2067 or constraint name)
		if err.Error() == "UNIQUE constraint failed: tenants.name" {
			return &platformerrors.AlreadyExistsError{
				Resource: "tenant",
				Field:    "name",
				Value:    t.Name,
			}
		}
		return err
	}
	return nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, status, active_llm_profile_id, created_at, updated_at, created_by
		FROM tenants
		WHERE id = ?
	`
	var t tenant.Tenant
	var idStr, activeLLMProfileIDStr, createdByStr, createdAtStr, updatedAtStr string
	var activeLLMProfileID, createdBy sql.NullString

	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&idStr, &t.Name, &t.Status, &activeLLMProfileID, &createdAtStr, &updatedAtStr, &createdBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "tenant",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Parse UUID
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	t.ID = parsedID

	// Parse timestamps
	t.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUIDs
	if activeLLMProfileID.Valid {
		if id, err := uuid.Parse(activeLLMProfileID.String); err == nil {
			t.ActiveLLMProfileID = &id
		}
	}
	if createdBy.Valid {
		if id, err := uuid.Parse(createdBy.String); err == nil {
			t.CreatedBy = &id
		}
	}

	return &t, nil
}

// GetByName retrieves a tenant by name
func (r *TenantRepository) GetByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, status, active_llm_profile_id, created_at, updated_at, created_by
		FROM tenants
		WHERE name = ?
	`
	var t tenant.Tenant
	var idStr, createdAtStr, updatedAtStr string
	var activeLLMProfileID, createdBy sql.NullString

	err := r.db.QueryRow(ctx, query, name).Scan(
		&idStr, &t.Name, &t.Status, &activeLLMProfileID, &createdAtStr, &updatedAtStr, &createdBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "tenant",
				ID:       name,
			}
		}
		return nil, err
	}

	// Parse UUID
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	t.ID = parsedID

	// Parse timestamps
	t.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUIDs
	if activeLLMProfileID.Valid {
		if id, err := uuid.Parse(activeLLMProfileID.String); err == nil {
			t.ActiveLLMProfileID = &id
		}
	}
	if createdBy.Valid {
		if id, err := uuid.Parse(createdBy.String); err == nil {
			t.CreatedBy = &id
		}
	}

	return &t, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	query := `
		UPDATE tenants
		SET name = ?, status = ?, active_llm_profile_id = ?, updated_at = ?
		WHERE id = ?
	`

	var activeLLMProfileID sql.NullString
	if t.ActiveLLMProfileID != nil {
		activeLLMProfileID = sql.NullString{String: t.ActiveLLMProfileID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		t.Name,
		string(t.Status),
		activeLLMProfileID,
		t.UpdatedAt.Format(time.RFC3339),
		t.ID.String(),
	)
	return err
}

// Delete deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = ?`
	_, err := r.db.Exec(ctx, query, id.String())
	return err
}

// List lists tenants with pagination
func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]*tenant.Tenant, error) {
	query := `
		SELECT id, name, status, active_llm_profile_id, created_at, updated_at, created_by
		FROM tenants
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		var idStr, createdAtStr, updatedAtStr string
		var activeLLMProfileID, createdBy sql.NullString

		err := rows.Scan(&idStr, &t.Name, &t.Status, &activeLLMProfileID, &createdAtStr, &updatedAtStr, &createdBy)
		if err != nil {
			return nil, err
		}

		// Parse UUID
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		t.ID = parsedID

		// Parse timestamps
		t.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable UUIDs
		if activeLLMProfileID.Valid {
			if id, err := uuid.Parse(activeLLMProfileID.String); err == nil {
				t.ActiveLLMProfileID = &id
			}
		}
		if createdBy.Valid {
			if id, err := uuid.Parse(createdBy.String); err == nil {
				t.CreatedBy = &id
			}
		}

		tenants = append(tenants, &t)
	}

	return tenants, rows.Err()
}

