package repositories

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
)

// EntityRelationRepository defines the interface for entity relation persistence
type EntityRelationRepository interface {
	// CRUD
	Create(ctx context.Context, r *relation.EntityRelation) error
	CreateWithMirror(ctx context.Context, r *relation.EntityRelation) (*relation.EntityRelation, error) // Creates both relation and mirror
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*relation.EntityRelation, error)
	Update(ctx context.Context, r *relation.EntityRelation) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error // Also deletes mirror if exists

	// List with cursor pagination
	ListBySource(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, opts ListOptions) (*ListResult, error)
	ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, opts ListOptions) (*ListResult, error)
	ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, opts ListOptions) (*ListResult, error)

	// Maintenance - delete relations when entity is deleted
	DeleteByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) error
}

// Cursor-based pagination
type ListOptions struct {
	Cursor         *string // Opaque cursor from previous response
	Limit          int     // Default 50, max 100
	RelationType   *string
	OrderBy        string // "created_at"
	OrderDir       string // "asc", "desc"
	ExcludeMirrors bool   // If true, only returns primary relations (id < mirror_id OR mirror_id IS NULL)
}

type ListResult struct {
	Items      []*relation.EntityRelation
	NextCursor *string
	HasMore    bool
}

// Cursor encoding/decoding
type Cursor struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// EncodeCursor encodes a cursor to a base64 string
func EncodeCursor(c Cursor) string {
	data, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeCursor decodes a cursor from a base64 string
func DecodeCursor(s string) (*Cursor, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
