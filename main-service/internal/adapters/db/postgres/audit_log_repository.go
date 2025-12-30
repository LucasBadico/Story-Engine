package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.AuditLogRepository = (*AuditLogRepository)(nil)

// AuditLogRepository implements the audit log repository interface
type AuditLogRepository struct {
	db *DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log
func (r *AuditLogRepository) Create(ctx context.Context, log *audit.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, tenant_id, actor_user_id, action, entity_type, entity_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var metadataJSON []byte
	if log.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := r.db.Exec(ctx, query,
		log.ID, log.TenantID, log.ActorUserID, string(log.Action),
		string(log.EntityType), log.EntityID, metadataJSON, log.CreatedAt)
	return err
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*audit.AuditLog, error) {
	query := `
		SELECT id, tenant_id, actor_user_id, action, entity_type, entity_id, metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`
	var a audit.AuditLog
	var actorUserID sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.TenantID, &actorUserID, &a.Action, &a.EntityType, &a.EntityID, &metadataJSON, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("audit log not found")
		}
		return nil, err
	}

	if actorUserID.Valid {
		if id, err := uuid.Parse(actorUserID.String); err == nil {
			a.ActorUserID = &id
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
			return nil, err
		}
	}

	return &a, nil
}

// ListByTenant lists audit logs for a tenant
func (r *AuditLogRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	query := `
		SELECT id, tenant_id, actor_user_id, action, entity_type, entity_id, metadata, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByEntity lists audit logs for an entity
func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType audit.EntityType, entityID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	query := `
		SELECT id, tenant_id, actor_user_id, action, entity_type, entity_id, metadata, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByActor lists audit logs for an actor
func (r *AuditLogRepository) ListByActor(ctx context.Context, actorUserID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	query := `
		SELECT id, tenant_id, actor_user_id, action, entity_type, entity_id, metadata, created_at
		FROM audit_logs
		WHERE actor_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, actorUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

func (r *AuditLogRepository) scanAuditLogs(rows pgx.Rows) ([]*audit.AuditLog, error) {
	var logs []*audit.AuditLog
	for rows.Next() {
		var a audit.AuditLog
		var actorUserID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&a.ID, &a.TenantID, &actorUserID, &a.Action, &a.EntityType, &a.EntityID, &metadataJSON, &a.CreatedAt)
		if err != nil {
			return nil, err
		}

		if actorUserID.Valid {
			if id, err := uuid.Parse(actorUserID.String); err == nil {
				a.ActorUserID = &id
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
				return nil, err
			}
		}

		logs = append(logs, &a)
	}

	return logs, rows.Err()
}

