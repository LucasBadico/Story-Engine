package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.TenantRepository = (*TenantRepository)(nil)

// TenantRepository implements the tenant repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		t.ID, t.Name, string(t.Status), t.ActiveLLMProfileID, t.CreatedAt, t.UpdatedAt, t.CreatedBy)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, status, active_llm_profile_id, created_at, updated_at, created_by
		FROM tenants
		WHERE id = $1
	`
	var t tenant.Tenant
	var activeLLMProfileID sql.NullString
	var createdBy sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Status, &activeLLMProfileID, &t.CreatedAt, &t.UpdatedAt, &createdBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

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
		WHERE name = $1
	`
	var t tenant.Tenant
	var activeLLMProfileID sql.NullString
	var createdBy sql.NullString

	err := r.db.QueryRow(ctx, query, name).Scan(
		&t.ID, &t.Name, &t.Status, &activeLLMProfileID, &t.CreatedAt, &t.UpdatedAt, &createdBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

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
		SET name = $2, status = $3, active_llm_profile_id = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, t.ID, t.Name, string(t.Status), t.ActiveLLMProfileID, t.UpdatedAt)
	return err
}

// Delete deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// List lists tenants with pagination
func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]*tenant.Tenant, error) {
	query := `
		SELECT id, name, status, active_llm_profile_id, created_at, updated_at, created_by
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		var activeLLMProfileID sql.NullString
		var createdBy sql.NullString

		err := rows.Scan(&t.ID, &t.Name, &t.Status, &activeLLMProfileID, &t.CreatedAt, &t.UpdatedAt, &createdBy)
		if err != nil {
			return nil, err
		}

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

