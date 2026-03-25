package importer

import "github.com/google/uuid"

// ImportChapterInput represents the input for importing a chapter from an export folder
type ImportChapterInput struct {
	FolderPath    string
	TenantID      uuid.UUID
	WorldID       uuid.UUID
	StoryID       uuid.UUID
	ChapterNumber int
	ChapterID     *uuid.UUID // optional, if chapter already exists
}

// ImportReport holds the results of an import operation
type ImportReport struct {
	Created []ImportedEntity
	Updated []ImportedEntity
	Skipped []ImportedEntity
	Failed  []FailedEntity
}

// ImportedEntity represents a successfully imported or skipped entity
type ImportedEntity struct {
	Type string // "character", "location", "artifact", "character_state", etc.
	Name string
	ID   uuid.UUID
}

// FailedEntity represents an entity that failed to import
type FailedEntity struct {
	Type   string
	Name   string
	Reason string
}

// NewImportReport creates a new ImportReport with initialized slices
func NewImportReport() *ImportReport {
	return &ImportReport{
		Created: []ImportedEntity{},
		Updated: []ImportedEntity{},
		Skipped: []ImportedEntity{},
		Failed:  []FailedEntity{},
	}
}

// AddCreated records a successfully created entity
func (r *ImportReport) AddCreated(entityType, name string, id uuid.UUID) {
	r.Created = append(r.Created, ImportedEntity{Type: entityType, Name: name, ID: id})
}

// AddSkipped records a skipped entity (already exists)
func (r *ImportReport) AddSkipped(entityType, name string, id uuid.UUID) {
	r.Skipped = append(r.Skipped, ImportedEntity{Type: entityType, Name: name, ID: id})
}

// AddFailed records a failed entity import
func (r *ImportReport) AddFailed(entityType, name, reason string) {
	r.Failed = append(r.Failed, FailedEntity{Type: entityType, Name: name, Reason: reason})
}
