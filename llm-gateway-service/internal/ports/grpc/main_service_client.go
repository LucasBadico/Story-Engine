package grpc

import (
	"context"

	"github.com/google/uuid"
)

// MainServiceClient defines the interface for communicating with main-service via gRPC
type MainServiceClient interface {
	// GetStory retrieves a story by ID
	GetStory(ctx context.Context, storyID uuid.UUID) (*Story, error)
	
	// GetChapter retrieves a chapter by ID
	GetChapter(ctx context.Context, chapterID uuid.UUID) (*Chapter, error)
	
	// GetScene retrieves a scene by ID
	GetScene(ctx context.Context, sceneID uuid.UUID) (*Scene, error)
	
	// GetBeat retrieves a beat by ID
	GetBeat(ctx context.Context, beatID uuid.UUID) (*Beat, error)
	
	// GetContentBlock retrieves a content block by ID
	GetContentBlock(ctx context.Context, contentBlockID uuid.UUID) (*ContentBlock, error)
	
	// ListContentBlocksByChapter lists content blocks for a chapter
	ListContentBlocksByChapter(ctx context.Context, chapterID uuid.UUID) ([]*ContentBlock, error)

	// ListContentAnchors lists anchors for a content block
	ListContentAnchors(ctx context.Context, contentBlockID uuid.UUID) ([]*ContentAnchor, error)
	// ListContentBlocksByEntity lists content blocks referencing an entity
	ListContentBlocksByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*ContentBlock, error)

	// World entities
	GetWorld(ctx context.Context, worldID uuid.UUID) (*World, error)
	GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error)
	GetLocation(ctx context.Context, locationID uuid.UUID) (*Location, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error)
	GetArtifact(ctx context.Context, artifactID uuid.UUID) (*Artifact, error)
	GetFaction(ctx context.Context, factionID uuid.UUID) (*Faction, error)
	GetLore(ctx context.Context, loreID uuid.UUID) (*Lore, error)

	// Character relations
	GetCharacterTraits(ctx context.Context, characterID uuid.UUID) ([]*CharacterTrait, error)

	// Event relations
	GetEventCharacters(ctx context.Context, eventID uuid.UUID) ([]*EventCharacter, error)
	GetEventLocations(ctx context.Context, eventID uuid.UUID) ([]*EventLocation, error)
	GetEventArtifacts(ctx context.Context, eventID uuid.UUID) ([]*EventArtifact, error)

	// Scene references (relates story with world)
	ListSceneReferences(ctx context.Context, sceneID uuid.UUID) ([]*SceneReference, error)
}

// Story represents a story from main-service
type Story struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Title     string
	Status    string
	CreatedAt int64
	UpdatedAt int64
}

// Chapter represents a chapter from main-service
type Chapter struct {
	ID        uuid.UUID
	StoryID   uuid.UUID
	Number    int
	Title     string
	Status    string
	CreatedAt int64
	UpdatedAt int64
}

// Scene represents a scene from main-service
type Scene struct {
	ID             uuid.UUID
	StoryID        uuid.UUID
	ChapterID      *uuid.UUID
	OrderNum       int
	Goal           string
	TimeRef        string
	POVCharacterID *uuid.UUID
	LocationID     *uuid.UUID
	CreatedAt      int64
	UpdatedAt      int64
}

// Beat represents a beat from main-service
type Beat struct {
	ID        uuid.UUID
	SceneID   uuid.UUID
	OrderNum  int
	Type      string
	Intent    string
	Outcome   string
	CreatedAt int64
	UpdatedAt int64
}

// ContentBlock represents a content block from main-service
type ContentBlock struct {
	ID        uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  *int
	Type      string // text, image, video, audio, embed, link
	Kind      string // final, alt_a, alt_b, cleaned, localized, draft, thumbnail
	Content   string // text content or URL depending on type
	Metadata  map[string]interface{} // type-specific metadata
	CreatedAt int64
	UpdatedAt int64
}

// ContentAnchor represents a text anchor from a content block to an entity
type ContentAnchor struct {
	ID            uuid.UUID
	ContentBlockID uuid.UUID
	EntityType    string // "scene", "beat", "chapter", "character", "location", "artifact", "event", "world", etc.
	EntityID      uuid.UUID
	CreatedAt     int64
}

// World represents a world from main-service
type World struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Genre       string
	IsImplicit  bool
	CreatedAt   int64
	UpdatedAt   int64
}

// Character represents a character from main-service
type Character struct {
	ID            uuid.UUID
	WorldID       uuid.UUID
	ArchetypeID   *uuid.UUID
	Name          string
	Description   string
	CreatedAt     int64
	UpdatedAt     int64
}

// CharacterTrait represents a character trait
type CharacterTrait struct {
	ID          uuid.UUID
	CharacterID uuid.UUID
	TraitID     uuid.UUID
	TraitName   string
	Value       string
	Notes       string
	CreatedAt   int64
	UpdatedAt   int64
}

// Location represents a location from main-service
type Location struct {
	ID             uuid.UUID
	WorldID        uuid.UUID
	ParentID       *uuid.UUID
	Name           string
	Type           string
	Description    string
	HierarchyLevel int
	CreatedAt      int64
	UpdatedAt      int64
}

// Event represents an event from main-service
type Event struct {
	ID          uuid.UUID
	WorldID     uuid.UUID
	Name        string
	Type        *string
	Description *string
	Timeline    *string
	Importance  int
	CreatedAt   int64
	UpdatedAt   int64
}

// EventCharacter represents the relationship between an event and a character
type EventCharacter struct {
	ID          uuid.UUID
	EventID     uuid.UUID
	CharacterID uuid.UUID
	Role        *string
	CreatedAt   int64
}

// EventLocation represents the relationship between an event and a location
type EventLocation struct {
	ID           uuid.UUID
	EventID      uuid.UUID
	LocationID   uuid.UUID
	Significance *string
	CreatedAt    int64
}

// EventArtifact represents the relationship between an event and an artifact
type EventArtifact struct {
	ID         uuid.UUID
	EventID    uuid.UUID
	ArtifactID uuid.UUID
	Role       *string
	CreatedAt  int64
}

// Artifact represents an artifact from main-service
type Artifact struct {
	ID          uuid.UUID
	WorldID     uuid.UUID
	Name        string
	Description string
	Rarity      string
	CreatedAt   int64
	UpdatedAt   int64
}

// SceneReference represents a reference from a scene to an entity
type SceneReference struct {
	ID         uuid.UUID
	SceneID    uuid.UUID
	EntityType string // "character", "location", "artifact"
	EntityID   uuid.UUID
	CreatedAt  int64
}

// Faction represents a faction from main-service
type Faction struct {
	ID             uuid.UUID
	WorldID        uuid.UUID
	ParentID       *uuid.UUID
	Name           string
	Type           *string
	Description    string
	Beliefs        string
	Structure      string
	Symbols        string
	HierarchyLevel int
	CreatedAt      int64
	UpdatedAt      int64
}

// Lore represents a lore from main-service
type Lore struct {
	ID             uuid.UUID
	WorldID        uuid.UUID
	ParentID       *uuid.UUID
	Name           string
	Category       *string
	Description    string
	Rules          string
	Limitations    string
	Requirements   string
	HierarchyLevel int
	CreatedAt      int64
	UpdatedAt      int64
}
