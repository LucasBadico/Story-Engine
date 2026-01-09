package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.EntityRelationRepository = (*EntityRelationRepository)(nil)

// EntityRelationRepository implements the entity relation repository interface for SQLite
type EntityRelationRepository struct {
	db *DB
}

// NewEntityRelationRepository creates a new entity relation repository
func NewEntityRelationRepository(db *DB) *EntityRelationRepository {
	return &EntityRelationRepository{db: db}
}

// Create creates a new entity relation
func (r *EntityRelationRepository) Create(ctx context.Context, rel *relation.EntityRelation) error {
	attributesJSON, err := rel.AttributesJSON()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO entity_relations (
			id, tenant_id, world_id, source_type, source_id,
			target_type, target_id, relation_type,
			context_type, context_id, attributes,
			summary, mirror_id, created_by_user_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var contextType sql.NullString
	if rel.ContextType != nil {
		contextType = sql.NullString{String: *rel.ContextType, Valid: true}
	}
	var contextID sql.NullString
	if rel.ContextID != nil {
		contextID = sql.NullString{String: rel.ContextID.String(), Valid: true}
	}
	var mirrorID sql.NullString
	if rel.MirrorID != nil {
		mirrorID = sql.NullString{String: rel.MirrorID.String(), Valid: true}
	}
	var createdByUserID sql.NullString
	if rel.CreatedByUserID != nil {
		createdByUserID = sql.NullString{String: rel.CreatedByUserID.String(), Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		rel.ID.String(),
		rel.TenantID.String(),
		rel.WorldID.String(),
		rel.SourceType, rel.SourceID.String(),
		rel.TargetType, rel.TargetID.String(),
		rel.RelationType,
		contextType, contextID,
		string(attributesJSON),
		rel.Summary,
		mirrorID, createdByUserID,
		rel.CreatedAt.Format(time.RFC3339),
		rel.UpdatedAt.Format(time.RFC3339))

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return &platformerrors.AlreadyExistsError{
				Resource: "entity_relation",
				Field:    "relation",
				Value:    fmt.Sprintf("%s-%s-%s", rel.SourceID.String(), rel.RelationType, rel.TargetID.String()),
			}
		}
		return err
	}
	return nil
}

// CreateWithMirror creates both the relation and its mirror in a single transaction
func (r *EntityRelationRepository) CreateWithMirror(ctx context.Context, rel *relation.EntityRelation) (*relation.EntityRelation, error) {
	mirror := rel.CreateMirrorRelation()

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create main relation first (without mirror_id to avoid FK constraint issue)
	originalMirrorID := rel.MirrorID
	rel.MirrorID = nil
	if err := r.createInTx(ctx, tx, rel); err != nil {
		rel.MirrorID = originalMirrorID
		return nil, err
	}

	// Create mirror relation (with mirror_id pointing to original)
	if err := r.createInTx(ctx, tx, mirror); err != nil {
		rel.MirrorID = originalMirrorID
		return nil, err
	}

	// Update original relation to set mirror_id
	updateQuery := `UPDATE entity_relations SET mirror_id = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, updateQuery, mirror.ID.String(), rel.ID.String())
	if err != nil {
		rel.MirrorID = originalMirrorID
		return nil, err
	}

	// Restore mirror_id in the relation object
	rel.MirrorID = originalMirrorID

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return mirror, nil
}

// createInTx creates a relation within a transaction
func (r *EntityRelationRepository) createInTx(ctx context.Context, tx *sql.Tx, rel *relation.EntityRelation) error {
	attributesJSON, err := rel.AttributesJSON()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO entity_relations (
			id, tenant_id, world_id, source_type, source_id,
			target_type, target_id, relation_type,
			context_type, context_id, attributes,
			summary, mirror_id, created_by_user_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var contextType sql.NullString
	if rel.ContextType != nil {
		contextType = sql.NullString{String: *rel.ContextType, Valid: true}
	}
	var contextID sql.NullString
	if rel.ContextID != nil {
		contextID = sql.NullString{String: rel.ContextID.String(), Valid: true}
	}
	var mirrorID sql.NullString
	if rel.MirrorID != nil {
		mirrorID = sql.NullString{String: rel.MirrorID.String(), Valid: true}
	}
	var createdByUserID sql.NullString
	if rel.CreatedByUserID != nil {
		createdByUserID = sql.NullString{String: rel.CreatedByUserID.String(), Valid: true}
	}

	_, err = tx.ExecContext(ctx, query,
		rel.ID.String(),
		rel.TenantID.String(),
		rel.WorldID.String(),
		rel.SourceType, rel.SourceID.String(),
		rel.TargetType, rel.TargetID.String(),
		rel.RelationType,
		contextType, contextID,
		string(attributesJSON),
		rel.Summary,
		mirrorID, createdByUserID,
		rel.CreatedAt.Format(time.RFC3339),
		rel.UpdatedAt.Format(time.RFC3339))

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return &platformerrors.AlreadyExistsError{
				Resource: "entity_relation",
				Field:    "relation",
				Value:    fmt.Sprintf("%s-%s-%s", rel.SourceID.String(), rel.RelationType, rel.TargetID.String()),
			}
		}
		return err
	}
	return nil
}

// GetByID retrieves an entity relation by ID
func (r *EntityRelationRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*relation.EntityRelation, error) {
	query := `
		SELECT id, tenant_id, world_id, source_type, source_id,
		       target_type, target_id, relation_type,
		       context_type, context_id, attributes,
		       summary, mirror_id, created_by_user_id, created_at, updated_at
		FROM entity_relations
		WHERE tenant_id = ? AND id = ?
	`
	var rel relation.EntityRelation
	var idStr, tenantIDStr, worldIDStr, sourceIDStr, targetIDStr, createdAtStr, updatedAtStr string
	var contextType, contextID, mirrorID, createdByUserID sql.NullString
	var attributesJSON sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr,
		&rel.SourceType, &sourceIDStr,
		&rel.TargetType, &targetIDStr,
		&rel.RelationType,
		&contextType, &contextID,
		&attributesJSON,
		&rel.Summary,
		&mirrorID, &createdByUserID,
		&createdAtStr, &updatedAtStr)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "entity_relation",
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
	rel.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	rel.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	rel.WorldID = parsedWorldID

	parsedSourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		return nil, err
	}
	rel.SourceID = parsedSourceID

	parsedTargetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		return nil, err
	}
	rel.TargetID = parsedTargetID

	// Parse nullable fields
	if contextType.Valid {
		rel.ContextType = &contextType.String
	}
	if contextID.Valid {
		parsedID, err := uuid.Parse(contextID.String)
		if err == nil {
			rel.ContextID = &parsedID
		}
	}
	if mirrorID.Valid {
		parsedID, err := uuid.Parse(mirrorID.String)
		if err == nil {
			rel.MirrorID = &parsedID
		}
	}
	if createdByUserID.Valid {
		parsedID, err := uuid.Parse(createdByUserID.String)
		if err == nil {
			rel.CreatedByUserID = &parsedID
		}
	}

	// Parse timestamps
	rel.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	rel.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if attributesJSON.Valid && attributesJSON.String != "" {
		if err := json.Unmarshal([]byte(attributesJSON.String), &rel.Attributes); err != nil {
			return nil, err
		}
	} else {
		rel.Attributes = make(map[string]interface{})
	}

	return &rel, nil
}

// Update updates an entity relation
func (r *EntityRelationRepository) Update(ctx context.Context, rel *relation.EntityRelation) error {
	attributesJSON, err := rel.AttributesJSON()
	if err != nil {
		return err
	}

	var contextType sql.NullString
	if rel.ContextType != nil {
		contextType = sql.NullString{String: *rel.ContextType, Valid: true}
	}
	var contextID sql.NullString
	if rel.ContextID != nil {
		contextID = sql.NullString{String: rel.ContextID.String(), Valid: true}
	}
	var mirrorID sql.NullString
	if rel.MirrorID != nil {
		mirrorID = sql.NullString{String: rel.MirrorID.String(), Valid: true}
	}

	query := `
		UPDATE entity_relations
		SET source_type = ?, source_id = ?,
		    target_type = ?, target_id = ?,
		    relation_type = ?, context_type = ?, context_id = ?,
		    attributes = ?, summary = ?, mirror_id = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err = r.db.Exec(ctx, query,
		rel.SourceType, rel.SourceID.String(),
		rel.TargetType, rel.TargetID.String(),
		rel.RelationType, contextType, contextID,
		string(attributesJSON), rel.Summary, mirrorID,
		rel.UpdatedAt.Format(time.RFC3339),
		rel.TenantID.String(),
		rel.ID.String())
	return err
}

// Delete deletes an entity relation (and its mirror if exists)
func (r *EntityRelationRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	// Get the relation to find mirror_id
	rel, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete main relation
	query := `DELETE FROM entity_relations WHERE tenant_id = ? AND id = ?`
	_, err = tx.ExecContext(ctx, query, tenantID.String(), id.String())
	if err != nil {
		return err
	}

	// Delete mirror if exists
	if rel.MirrorID != nil {
		_, err = tx.ExecContext(ctx, query, tenantID.String(), rel.MirrorID.String())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListBySource lists relations by source with cursor pagination
func (r *EntityRelationRepository) ListBySource(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, opts repositories.ListOptions) (*repositories.ListResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	where := []string{"tenant_id = ?", "source_type = ?", "source_id = ?"}
	args := []interface{}{tenantID.String(), sourceType, sourceID.String()}

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, "relation_type = ?")
		args = append(args, *opts.RelationType)
	}
	if opts.ExcludeMirrors {
		where = append(where, "(mirror_id IS NULL OR id < mirror_id)")
	}

	// Cursor pagination
	if opts.Cursor != nil {
		cursor, err := repositories.DecodeCursor(*opts.Cursor)
		if err != nil {
			return nil, err
		}
		where = append(where, "(created_at, id) > (?, ?)")
		args = append(args, cursor.CreatedAt.Format(time.RFC3339), cursor.ID.String())
	}

	orderBy := opts.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	orderDir := opts.OrderDir
	if orderDir == "" {
		orderDir = "asc"
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, world_id, source_type, source_id,
		       target_type, target_id, relation_type,
		       context_type, context_id, attributes,
		       summary, mirror_id, created_by_user_id, created_at, updated_at
		FROM entity_relations
		WHERE %s
		ORDER BY %s %s, id %s
		LIMIT ?
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := r.scanEntityRelations(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		cursor := repositories.Cursor{
			ID:        last.ID,
			CreatedAt: last.CreatedAt,
		}
		encoded := repositories.EncodeCursor(cursor)
		nextCursor = &encoded
	}

	return &repositories.ListResult{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// ListByTarget lists relations by target with cursor pagination
func (r *EntityRelationRepository) ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, opts repositories.ListOptions) (*repositories.ListResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	where := []string{"tenant_id = ?", "target_type = ?", "target_id = ?"}
	args := []interface{}{tenantID.String(), targetType, targetID.String()}

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, "relation_type = ?")
		args = append(args, *opts.RelationType)
	}
	if opts.ExcludeMirrors {
		where = append(where, "(mirror_id IS NULL OR id < mirror_id)")
	}

	// Cursor pagination
	if opts.Cursor != nil {
		cursor, err := repositories.DecodeCursor(*opts.Cursor)
		if err != nil {
			return nil, err
		}
		where = append(where, "(created_at, id) > (?, ?)")
		args = append(args, cursor.CreatedAt.Format(time.RFC3339), cursor.ID.String())
	}

	orderBy := opts.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	orderDir := opts.OrderDir
	if orderDir == "" {
		orderDir = "asc"
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, world_id, source_type, source_id,
		       target_type, target_id, relation_type,
		       context_type, context_id, attributes,
		       summary, mirror_id, created_by_user_id, created_at, updated_at
		FROM entity_relations
		WHERE %s
		ORDER BY %s %s, id %s
		LIMIT ?
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := r.scanEntityRelations(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		cursor := repositories.Cursor{
			ID:        last.ID,
			CreatedAt: last.CreatedAt,
		}
		encoded := repositories.EncodeCursor(cursor)
		nextCursor = &encoded
	}

	return &repositories.ListResult{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// ListByWorld lists relations by world with cursor pagination
func (r *EntityRelationRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, opts repositories.ListOptions) (*repositories.ListResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	where := []string{"tenant_id = ?", "world_id = ?"}
	args := []interface{}{tenantID.String(), worldID.String()}

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, "relation_type = ?")
		args = append(args, *opts.RelationType)
	}
	if opts.ExcludeMirrors {
		where = append(where, "(mirror_id IS NULL OR id < mirror_id)")
	}

	// Cursor pagination
	if opts.Cursor != nil {
		cursor, err := repositories.DecodeCursor(*opts.Cursor)
		if err != nil {
			return nil, err
		}
		where = append(where, "(created_at, id) > (?, ?)")
		args = append(args, cursor.CreatedAt.Format(time.RFC3339), cursor.ID.String())
	}

	orderBy := opts.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	orderDir := opts.OrderDir
	if orderDir == "" {
		orderDir = "asc"
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, world_id, source_type, source_id,
		       target_type, target_id, relation_type,
		       context_type, context_id, attributes,
		       summary, mirror_id, created_by_user_id, created_at, updated_at
		FROM entity_relations
		WHERE %s
		ORDER BY %s %s, id %s
		LIMIT ?
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := r.scanEntityRelations(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		cursor := repositories.Cursor{
			ID:        last.ID,
			CreatedAt: last.CreatedAt,
		}
		encoded := repositories.EncodeCursor(cursor)
		nextCursor = &encoded
	}

	return &repositories.ListResult{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// DeleteByEntity deletes all relations involving an entity (when entity is deleted)
func (r *EntityRelationRepository) DeleteByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `
		DELETE FROM entity_relations
		WHERE tenant_id = ? AND (
			(source_type = ? AND source_id = ?) OR
			(target_type = ? AND target_id = ?)
		)
	`
	_, err := r.db.Exec(ctx, query, tenantID.String(), entityType, entityID.String(), entityType, entityID.String())
	return err
}

// scanEntityRelations scans rows into EntityRelation structs
func (r *EntityRelationRepository) scanEntityRelations(rows *sql.Rows) ([]*relation.EntityRelation, error) {
	relations := make([]*relation.EntityRelation, 0)
	for rows.Next() {
		var rel relation.EntityRelation
		var idStr, tenantIDStr, worldIDStr, sourceIDStr, targetIDStr, createdAtStr, updatedAtStr string
		var contextType, contextID, mirrorID, createdByUserID sql.NullString
		var attributesJSON sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr,
			&rel.SourceType, &sourceIDStr,
			&rel.TargetType, &targetIDStr,
			&rel.RelationType,
			&contextType, &contextID,
			&attributesJSON,
			&rel.Summary,
			&mirrorID, &createdByUserID,
			&createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		rel.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		rel.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		rel.WorldID = parsedWorldID

		parsedSourceID, err := uuid.Parse(sourceIDStr)
		if err != nil {
			return nil, err
		}
		rel.SourceID = parsedSourceID

		parsedTargetID, err := uuid.Parse(targetIDStr)
		if err != nil {
			return nil, err
		}
		rel.TargetID = parsedTargetID

		// Parse nullable fields
		if contextType.Valid {
			rel.ContextType = &contextType.String
		}
		if contextID.Valid {
			parsedID, err := uuid.Parse(contextID.String)
			if err == nil {
				rel.ContextID = &parsedID
			}
		}
		if mirrorID.Valid {
			parsedID, err := uuid.Parse(mirrorID.String)
			if err == nil {
				rel.MirrorID = &parsedID
			}
		}
		if createdByUserID.Valid {
			parsedID, err := uuid.Parse(createdByUserID.String)
			if err == nil {
				rel.CreatedByUserID = &parsedID
			}
		}

		// Parse timestamps
		rel.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		rel.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if attributesJSON.Valid && attributesJSON.String != "" {
			if err := json.Unmarshal([]byte(attributesJSON.String), &rel.Attributes); err != nil {
				return nil, err
			}
		} else {
			rel.Attributes = make(map[string]interface{})
		}

		relations = append(relations, &rel)
	}

	return relations, rows.Err()
}
