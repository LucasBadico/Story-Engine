package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// MockMainServiceClient is a mock implementation of MainServiceClient for testing
type MockMainServiceClient struct {
	stories        map[uuid.UUID]*Story
	chapters       map[uuid.UUID]*Chapter
	scenes         map[uuid.UUID]*Scene
	beats          map[uuid.UUID]*Beat
	worlds         map[uuid.UUID]*World
	factions       map[uuid.UUID]*Faction
	lores          map[uuid.UUID]*Lore
	contentBlocks  map[uuid.UUID]*ContentBlock
	contentAnchors map[uuid.UUID][]*ContentAnchor
	characters     map[uuid.UUID]*Character
	locations      map[uuid.UUID]*Location
	events         map[uuid.UUID]*Event
	artifacts      map[uuid.UUID]*Artifact
	relations      map[uuid.UUID]*EntityRelation

	// Error functions
	getStoryErr   func(uuid.UUID) error
	getChapterErr func(uuid.UUID) error
	getSceneErr   func(uuid.UUID) error
	getBeatErr    func(uuid.UUID) error
}

// NewMockMainServiceClient creates a new mock main service client
func NewMockMainServiceClient() *MockMainServiceClient {
	return &MockMainServiceClient{
		stories:        make(map[uuid.UUID]*Story),
		chapters:       make(map[uuid.UUID]*Chapter),
		scenes:         make(map[uuid.UUID]*Scene),
		beats:          make(map[uuid.UUID]*Beat),
		worlds:         make(map[uuid.UUID]*World),
		factions:       make(map[uuid.UUID]*Faction),
		lores:          make(map[uuid.UUID]*Lore),
		contentBlocks:  make(map[uuid.UUID]*ContentBlock),
		contentAnchors: make(map[uuid.UUID][]*ContentAnchor),
		characters:     make(map[uuid.UUID]*Character),
		locations:      make(map[uuid.UUID]*Location),
		events:         make(map[uuid.UUID]*Event),
		artifacts:      make(map[uuid.UUID]*Artifact),
		relations:      make(map[uuid.UUID]*EntityRelation),
	}
}

// AddStory adds a story to the mock
func (m *MockMainServiceClient) AddStory(story *Story) {
	m.stories[story.ID] = story
}

// AddChapter adds a chapter to the mock
func (m *MockMainServiceClient) AddChapter(chapter *Chapter) {
	m.chapters[chapter.ID] = chapter
}

// AddScene adds a scene to the mock
func (m *MockMainServiceClient) AddScene(scene *Scene) {
	m.scenes[scene.ID] = scene
}

// AddBeat adds a beat to the mock
func (m *MockMainServiceClient) AddBeat(beat *Beat) {
	m.beats[beat.ID] = beat
}

// SetGetStoryError sets an error function for GetStory
func (m *MockMainServiceClient) SetGetStoryError(fn func(uuid.UUID) error) {
	m.getStoryErr = fn
}

// SetGetChapterError sets an error function for GetChapter
func (m *MockMainServiceClient) SetGetChapterError(fn func(uuid.UUID) error) {
	m.getChapterErr = fn
}

// SetGetSceneError sets an error function for GetScene
func (m *MockMainServiceClient) SetGetSceneError(fn func(uuid.UUID) error) {
	m.getSceneErr = fn
}

// SetGetBeatError sets an error function for GetBeat
func (m *MockMainServiceClient) SetGetBeatError(fn func(uuid.UUID) error) {
	m.getBeatErr = fn
}

// GetStory retrieves a story by ID
func (m *MockMainServiceClient) GetStory(ctx context.Context, storyID uuid.UUID) (*Story, error) {
	if m.getStoryErr != nil {
		if err := m.getStoryErr(storyID); err != nil {
			return nil, err
		}
	}
	story, ok := m.stories[storyID]
	if !ok {
		return nil, fmt.Errorf("story not found: %s", storyID)
	}
	return story, nil
}

// GetChapter retrieves a chapter by ID
func (m *MockMainServiceClient) GetChapter(ctx context.Context, chapterID uuid.UUID) (*Chapter, error) {
	if m.getChapterErr != nil {
		if err := m.getChapterErr(chapterID); err != nil {
			return nil, err
		}
	}
	chapter, ok := m.chapters[chapterID]
	if !ok {
		return nil, fmt.Errorf("chapter not found: %s", chapterID)
	}
	return chapter, nil
}

// GetScene retrieves a scene by ID
func (m *MockMainServiceClient) GetScene(ctx context.Context, sceneID uuid.UUID) (*Scene, error) {
	if m.getSceneErr != nil {
		if err := m.getSceneErr(sceneID); err != nil {
			return nil, err
		}
	}
	scene, ok := m.scenes[sceneID]
	if !ok {
		return nil, fmt.Errorf("scene not found: %s", sceneID)
	}
	return scene, nil
}

// GetBeat retrieves a beat by ID
func (m *MockMainServiceClient) GetBeat(ctx context.Context, beatID uuid.UUID) (*Beat, error) {
	if m.getBeatErr != nil {
		if err := m.getBeatErr(beatID); err != nil {
			return nil, err
		}
	}
	beat, ok := m.beats[beatID]
	if !ok {
		return nil, fmt.Errorf("beat not found: %s", beatID)
	}
	return beat, nil
}

// AddWorld adds a world to the mock
func (m *MockMainServiceClient) AddWorld(world *World) {
	m.worlds[world.ID] = world
}

// AddFaction adds a faction to the mock
func (m *MockMainServiceClient) AddFaction(faction *Faction) {
	m.factions[faction.ID] = faction
}

// AddLore adds a lore to the mock
func (m *MockMainServiceClient) AddLore(lore *Lore) {
	m.lores[lore.ID] = lore
}

// GetWorld retrieves a world by ID
func (m *MockMainServiceClient) GetWorld(ctx context.Context, worldID uuid.UUID) (*World, error) {
	world, ok := m.worlds[worldID]
	if !ok {
		return nil, fmt.Errorf("world not found: %s", worldID)
	}
	return world, nil
}

// GetFaction retrieves a faction by ID
func (m *MockMainServiceClient) GetFaction(ctx context.Context, factionID uuid.UUID) (*Faction, error) {
	faction, ok := m.factions[factionID]
	if !ok {
		return nil, fmt.Errorf("faction not found: %s", factionID)
	}
	return faction, nil
}

// GetLore retrieves a lore by ID
func (m *MockMainServiceClient) GetLore(ctx context.Context, loreID uuid.UUID) (*Lore, error) {
	lore, ok := m.lores[loreID]
	if !ok {
		return nil, fmt.Errorf("lore not found: %s", loreID)
	}
	return lore, nil
}

// AddContentBlock adds a content block to the mock
func (m *MockMainServiceClient) AddContentBlock(contentBlock *ContentBlock) {
	m.contentBlocks[contentBlock.ID] = contentBlock
}

// AddContentAnchor adds a content anchor to the mock
func (m *MockMainServiceClient) AddContentAnchor(anchor *ContentAnchor) {
	if m.contentAnchors[anchor.ContentBlockID] == nil {
		m.contentAnchors[anchor.ContentBlockID] = []*ContentAnchor{}
	}
	m.contentAnchors[anchor.ContentBlockID] = append(m.contentAnchors[anchor.ContentBlockID], anchor)
}

// GetContentBlock retrieves a content block by ID
func (m *MockMainServiceClient) GetContentBlock(ctx context.Context, contentBlockID uuid.UUID) (*ContentBlock, error) {
	contentBlock, ok := m.contentBlocks[contentBlockID]
	if !ok {
		return nil, fmt.Errorf("content block not found: %s", contentBlockID)
	}
	return contentBlock, nil
}

func (m *MockMainServiceClient) ListContentBlocksByChapter(ctx context.Context, chapterID uuid.UUID) ([]*ContentBlock, error) {
	// Return all content blocks for simplicity in tests
	blocks := []*ContentBlock{}
	for _, cb := range m.contentBlocks {
		if cb.ChapterID != nil && *cb.ChapterID == chapterID {
			blocks = append(blocks, cb)
		}
	}
	return blocks, nil
}

func (m *MockMainServiceClient) ListContentAnchors(ctx context.Context, contentBlockID uuid.UUID) ([]*ContentAnchor, error) {
	anchors, ok := m.contentAnchors[contentBlockID]
	if !ok {
		return []*ContentAnchor{}, nil
	}
	return anchors, nil
}

func (m *MockMainServiceClient) ListContentBlocksByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*ContentBlock, error) {
	blocks := []*ContentBlock{}
	seen := make(map[uuid.UUID]struct{})

	for _, anchors := range m.contentAnchors {
		for _, anchor := range anchors {
			if anchor == nil {
				continue
			}
			if anchor.EntityType != entityType || anchor.EntityID != entityID {
				continue
			}
			if _, ok := seen[anchor.ContentBlockID]; ok {
				continue
			}
			block, exists := m.contentBlocks[anchor.ContentBlockID]
			if !exists {
				continue
			}
			seen[anchor.ContentBlockID] = struct{}{}
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

// AddCharacter adds a character to the mock
func (m *MockMainServiceClient) AddCharacter(character *Character) {
	m.characters[character.ID] = character
}

func (m *MockMainServiceClient) GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error) {
	character, ok := m.characters[characterID]
	if !ok {
		return nil, fmt.Errorf("character not found: %s", characterID)
	}
	return character, nil
}

func (m *MockMainServiceClient) GetLocation(ctx context.Context, locationID uuid.UUID) (*Location, error) {
	location, ok := m.locations[locationID]
	if !ok {
		return nil, fmt.Errorf("location not found: %s", locationID)
	}
	return location, nil
}

func (m *MockMainServiceClient) GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error) {
	event, ok := m.events[eventID]
	if !ok {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	return event, nil
}

func (m *MockMainServiceClient) GetArtifact(ctx context.Context, artifactID uuid.UUID) (*Artifact, error) {
	artifact, ok := m.artifacts[artifactID]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", artifactID)
	}
	return artifact, nil
}

func (m *MockMainServiceClient) GetCharacterTraits(ctx context.Context, characterID uuid.UUID) ([]*CharacterTrait, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockMainServiceClient) GetEventCharacters(ctx context.Context, eventID uuid.UUID) ([]*EventCharacter, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockMainServiceClient) GetEventLocations(ctx context.Context, eventID uuid.UUID) ([]*EventLocation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockMainServiceClient) GetEventArtifacts(ctx context.Context, eventID uuid.UUID) ([]*EventArtifact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockMainServiceClient) ListSceneReferences(ctx context.Context, sceneID uuid.UUID) ([]*SceneReference, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddLocation adds a location to the mock
func (m *MockMainServiceClient) AddLocation(location *Location) {
	m.locations[location.ID] = location
}

// AddEvent adds an event to the mock
func (m *MockMainServiceClient) AddEvent(event *Event) {
	m.events[event.ID] = event
}

// AddArtifact adds an artifact to the mock
func (m *MockMainServiceClient) AddArtifact(artifact *Artifact) {
	m.artifacts[artifact.ID] = artifact
}

// AddRelation adds a relation to the mock
func (m *MockMainServiceClient) AddRelation(relation *EntityRelation) {
	m.relations[relation.ID] = relation
}

func (m *MockMainServiceClient) GetRelation(ctx context.Context, relationID uuid.UUID) (*EntityRelation, error) {
	relation, ok := m.relations[relationID]
	if !ok {
		return nil, fmt.Errorf("relation not found: %s", relationID)
	}
	return relation, nil
}
