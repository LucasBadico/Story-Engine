package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/story-engine/main-service/internal/core/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.EntityRelationRepository = (*EntityRelationRepository)(nil)

// EntityRelationRepository implements the entity relation repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err = r.db.Exec(ctx, query,
		rel.ID, rel.TenantID, rel.WorldID,
		rel.SourceType, rel.SourceID,
		rel.TargetType, rel.TargetID,
		rel.RelationType,
		rel.ContextType, rel.ContextID,
		attributesJSON,
		rel.Summary,
		rel.MirrorID, rel.CreatedByUserID,
		rel.CreatedAt, rel.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
	defer tx.Rollback(ctx)

	// Create main relation
	if err := r.createInTx(ctx, tx, rel); err != nil {
		return nil, err
	}

	// Create mirror relation
	if err := r.createInTx(ctx, tx, mirror); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return mirror, nil
}

// createInTx creates a relation within a transaction
func (r *EntityRelationRepository) createInTx(ctx context.Context, tx pgx.Tx, rel *relation.EntityRelation) error {
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err = tx.Exec(ctx, query,
		rel.ID, rel.TenantID, rel.WorldID,
		rel.SourceType, rel.SourceID,
		rel.TargetType, rel.TargetID,
		rel.RelationType,
		rel.ContextType, rel.ContextID,
		attributesJSON,
		rel.Summary,
		rel.MirrorID, rel.CreatedByUserID,
		rel.CreatedAt, rel.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
		WHERE tenant_id = $1 AND id = $2
	`
	var rel relation.EntityRelation
	var contextType sql.NullString
	var contextID sql.NullString
	var mirrorID sql.NullString
	var createdByUserID sql.NullString
	var attributesJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&rel.ID, &rel.TenantID, &rel.WorldID,
		&rel.SourceType, &rel.SourceID,
		&rel.TargetType, &rel.TargetID,
		&rel.RelationType,
		&contextType, &contextID,
		&attributesJSON,
		&rel.Summary,
		&mirrorID, &createdByUserID,
		&rel.CreatedAt, &rel.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "entity_relation",
				ID:       id.String(),
			}
		}
		return nil, err
	}

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

	// Parse JSON fields
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &rel.Attributes); err != nil {
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

	query := `
		UPDATE entity_relations
		SET source_type = $2, source_id = $3,
		    target_type = $4, target_id = $5,
		    relation_type = $6, context_type = $7, context_id = $8,
		    attributes = $9, summary = $10, mirror_id = $11, updated_at = $12
		WHERE tenant_id = $13 AND id = $1
	`
	_, err = r.db.Exec(ctx, query,
		rel.ID,
		rel.SourceType, rel.SourceID,
		rel.TargetType, rel.TargetID,
		rel.RelationType, rel.ContextType, rel.ContextID,
		attributesJSON, rel.Summary, rel.MirrorID,
		rel.UpdatedAt, rel.TenantID)
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
	defer tx.Rollback(ctx)

	// Delete main relation
	query := `DELETE FROM entity_relations WHERE tenant_id = $1 AND id = $2`
	_, err = tx.Exec(ctx, query, tenantID, id)
	if err != nil {
		return err
	}

	// Delete mirror if exists
	if rel.MirrorID != nil {
		_, err = tx.Exec(ctx, query, tenantID, *rel.MirrorID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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

	where := []string{"tenant_id = $1", "source_type = $2", "source_id = $3"}
	args := []interface{}{tenantID, sourceType, sourceID}
	argIdx := 4

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, fmt.Sprintf("relation_type = $%d", argIdx))
		args = append(args, *opts.RelationType)
		argIdx++
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
		where = append(where, fmt.Sprintf("(created_at, id) > ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursor.CreatedAt, cursor.ID)
		argIdx += 2
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
		LIMIT $%d
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir, argIdx)
	args = append(args, limit+1) // Fetch one extra to check if there's more

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

	where := []string{"tenant_id = $1", "target_type = $2", "target_id = $3"}
	args := []interface{}{tenantID, targetType, targetID}
	argIdx := 4

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, fmt.Sprintf("relation_type = $%d", argIdx))
		args = append(args, *opts.RelationType)
		argIdx++
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
		where = append(where, fmt.Sprintf("(created_at, id) > ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursor.CreatedAt, cursor.ID)
		argIdx += 2
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
		LIMIT $%d
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir, argIdx)
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

	where := []string{"tenant_id = $1", "world_id = $2"}
	args := []interface{}{tenantID, worldID}
	argIdx := 3

	// Apply filters
	if opts.RelationType != nil {
		where = append(where, fmt.Sprintf("relation_type = $%d", argIdx))
		args = append(args, *opts.RelationType)
		argIdx++
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
		where = append(where, fmt.Sprintf("(created_at, id) > ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursor.CreatedAt, cursor.ID)
		argIdx += 2
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
		LIMIT $%d
	`, strings.Join(where, " AND "), orderBy, orderDir, orderDir, argIdx)
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
		WHERE tenant_id = $1 AND (
			(source_type = $2 AND source_id = $3) OR
			(target_type = $2 AND target_id = $3)
		)
	`
	_, err := r.db.Exec(ctx, query, tenantID, entityType, entityID)
	return err
}

// scanEntityRelations scans rows into EntityRelation structs
func (r *EntityRelationRepository) scanEntityRelations(rows pgx.Rows) ([]*relation.EntityRelation, error) {
	relations := make([]*relation.EntityRelation, 0)
	for rows.Next() {
		var rel relation.EntityRelation
		var contextType sql.NullString
		var contextID sql.NullString
		var mirrorID sql.NullString
		var createdByUserID sql.NullString
		var attributesJSON []byte

		err := rows.Scan(
			&rel.ID, &rel.TenantID, &rel.WorldID,
			&rel.SourceType, &rel.SourceID,
			&rel.TargetType, &rel.TargetID,
			&rel.RelationType,
			&contextType, &contextID,
			&attributesJSON,
			&rel.Summary,
			&mirrorID, &createdByUserID,
			&rel.CreatedAt, &rel.UpdatedAt)
		if err != nil {
			return nil, err
		}

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

		// Parse JSON fields
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &rel.Attributes); err != nil {
				return nil, err
			}
		} else {
			rel.Attributes = make(map[string]interface{})
		}

		relations = append(relations, &rel)
	}

	return relations, rows.Err()
}
