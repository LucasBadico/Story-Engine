package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	beatpb "github.com/story-engine/main-service/proto/beat"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	characterpb "github.com/story-engine/main-service/proto/character"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	eventpb "github.com/story-engine/main-service/proto/event"
	factionpb "github.com/story-engine/main-service/proto/faction"
	locationpb "github.com/story-engine/main-service/proto/location"
	lorepb "github.com/story-engine/main-service/proto/lore"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ grpcclient.MainServiceClient = (*MainServiceClient)(nil)

// MainServiceClient implements the gRPC client interface for main-service
type MainServiceClient struct {
	storyClient                 storypb.StoryServiceClient
	chapterClient               chapterpb.ChapterServiceClient
	sceneClient                 scenepb.SceneServiceClient
	beatClient                  beatpb.BeatServiceClient
	contentBlockClient          contentblockpb.ContentBlockServiceClient
	contentAnchorClient         contentblockpb.ContentAnchorServiceClient
	worldClient                 worldpb.WorldServiceClient
	characterClient             characterpb.CharacterServiceClient
	locationClient              locationpb.LocationServiceClient
	eventClient                 eventpb.EventServiceClient
	artifactClient              artifactpb.ArtifactServiceClient
	factionClient               factionpb.FactionServiceClient
	loreClient                  lorepb.LoreServiceClient
	conn                        *grpclib.ClientConn
}

// NewMainServiceClient creates a new gRPC client for main-service
func NewMainServiceClient(addr string) (*MainServiceClient, error) {
	conn, err := grpclib.NewClient(addr, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main-service: %w", err)
	}

	return &MainServiceClient{
		storyClient:                 storypb.NewStoryServiceClient(conn),
		chapterClient:               chapterpb.NewChapterServiceClient(conn),
		sceneClient:                 scenepb.NewSceneServiceClient(conn),
		beatClient:                  beatpb.NewBeatServiceClient(conn),
		contentBlockClient:          contentblockpb.NewContentBlockServiceClient(conn),
		contentAnchorClient:         contentblockpb.NewContentAnchorServiceClient(conn),
		worldClient:                 worldpb.NewWorldServiceClient(conn),
		characterClient:             characterpb.NewCharacterServiceClient(conn),
		locationClient:              locationpb.NewLocationServiceClient(conn),
		eventClient:                 eventpb.NewEventServiceClient(conn),
		artifactClient:              artifactpb.NewArtifactServiceClient(conn),
		factionClient:               factionpb.NewFactionServiceClient(conn),
		loreClient:                  lorepb.NewLoreServiceClient(conn),
		conn:                        conn,
	}, nil
}

// Close closes the gRPC connection
func (c *MainServiceClient) Close() error {
	return c.conn.Close()
}

// GetStory retrieves a story by ID
func (c *MainServiceClient) GetStory(ctx context.Context, storyID uuid.UUID) (*grpcclient.Story, error) {
	resp, err := c.storyClient.GetStory(ctx, &storypb.GetStoryRequest{
		Id: storyID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToStory(resp.Story), nil
}

// GetChapter retrieves a chapter by ID
func (c *MainServiceClient) GetChapter(ctx context.Context, chapterID uuid.UUID) (*grpcclient.Chapter, error) {
	resp, err := c.chapterClient.GetChapter(ctx, &chapterpb.GetChapterRequest{
		Id: chapterID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToChapter(resp.Chapter), nil
}

// GetScene retrieves a scene by ID
func (c *MainServiceClient) GetScene(ctx context.Context, sceneID uuid.UUID) (*grpcclient.Scene, error) {
	resp, err := c.sceneClient.GetScene(ctx, &scenepb.GetSceneRequest{
		Id: sceneID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToScene(resp.Scene), nil
}

// GetBeat retrieves a beat by ID
func (c *MainServiceClient) GetBeat(ctx context.Context, beatID uuid.UUID) (*grpcclient.Beat, error) {
	resp, err := c.beatClient.GetBeat(ctx, &beatpb.GetBeatRequest{
		Id: beatID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToBeat(resp.Beat), nil
}

// GetContentBlock retrieves a content block by ID
func (c *MainServiceClient) GetContentBlock(ctx context.Context, contentBlockID uuid.UUID) (*grpcclient.ContentBlock, error) {
	resp, err := c.contentBlockClient.GetContentBlock(ctx, &contentblockpb.GetContentBlockRequest{
		Id: contentBlockID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToContentBlock(resp.ContentBlock), nil
}

// ListContentBlocksByChapter lists content blocks for a chapter
func (c *MainServiceClient) ListContentBlocksByChapter(ctx context.Context, chapterID uuid.UUID) ([]*grpcclient.ContentBlock, error) {
	resp, err := c.contentBlockClient.ListContentBlocksByChapter(ctx, &contentblockpb.ListContentBlocksByChapterRequest{
		ChapterId: chapterID.String(),
	})
	if err != nil {
		return nil, err
	}

	contentBlocks := make([]*grpcclient.ContentBlock, len(resp.ContentBlocks))
	for i, cb := range resp.ContentBlocks {
		contentBlocks[i] = protoToContentBlock(cb)
	}

	return contentBlocks, nil
}

// ListContentAnchors lists anchors for a content block
func (c *MainServiceClient) ListContentAnchors(ctx context.Context, contentBlockID uuid.UUID) ([]*grpcclient.ContentAnchor, error) {
	resp, err := c.contentAnchorClient.ListContentAnchorsByContentBlock(ctx, &contentblockpb.ListContentAnchorsByContentBlockRequest{
		ContentBlockId: contentBlockID.String(),
	})
	if err != nil {
		return nil, err
	}

	anchors := make([]*grpcclient.ContentAnchor, len(resp.Anchors))
	for i, anchor := range resp.Anchors {
		anchors[i] = protoToContentAnchor(anchor)
	}

	return anchors, nil
}

// ListContentBlocksByEntity lists content blocks associated with an entity.
func (c *MainServiceClient) ListContentBlocksByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*grpcclient.ContentBlock, error) {
	resp, err := c.contentAnchorClient.ListContentBlocksByEntity(ctx, &contentblockpb.ListContentBlocksByEntityRequest{
		EntityType: entityType,
		EntityId:   entityID.String(),
	})
	if err != nil {
		return nil, err
	}

	contentBlocks := make([]*grpcclient.ContentBlock, len(resp.ContentBlocks))
	for i, cb := range resp.ContentBlocks {
		contentBlocks[i] = protoToContentBlock(cb)
	}

	return contentBlocks, nil
}

// Helper functions to convert proto to domain models

func protoToStory(s *storypb.Story) *grpcclient.Story {
	storyID, _ := uuid.Parse(s.Id)
	tenantID, _ := uuid.Parse(s.TenantId)

	return &grpcclient.Story{
		ID:        storyID,
		TenantID:  tenantID,
		Title:     s.Title,
		Status:    s.Status,
		CreatedAt: s.CreatedAt.Seconds,
		UpdatedAt: s.UpdatedAt.Seconds,
	}
}

func protoToChapter(c *chapterpb.Chapter) *grpcclient.Chapter {
	chapterID, _ := uuid.Parse(c.Id)
	storyID, _ := uuid.Parse(c.StoryId)

	return &grpcclient.Chapter{
		ID:        chapterID,
		StoryID:   storyID,
		Number:    int(c.Number),
		Title:     c.Title,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Seconds,
		UpdatedAt: c.UpdatedAt.Seconds,
	}
}

func protoToScene(s *scenepb.Scene) *grpcclient.Scene {
	sceneID, _ := uuid.Parse(s.Id)
	storyID, _ := uuid.Parse(s.StoryId)

	scene := &grpcclient.Scene{
		ID:        sceneID,
		StoryID:   storyID,
		OrderNum:  int(s.OrderNum),
		Goal:      s.Goal,
		TimeRef:   s.TimeRef,
		CreatedAt: s.CreatedAt.Seconds,
		UpdatedAt: s.UpdatedAt.Seconds,
	}

	if s.ChapterId != nil {
		chapterID, _ := uuid.Parse(*s.ChapterId)
		scene.ChapterID = &chapterID
	}
	if s.PovCharacterId != nil {
		povID, _ := uuid.Parse(*s.PovCharacterId)
		scene.POVCharacterID = &povID
	}

	return scene
}

func protoToBeat(b *beatpb.Beat) *grpcclient.Beat {
	beatID, _ := uuid.Parse(b.Id)
	sceneID, _ := uuid.Parse(b.SceneId)

	return &grpcclient.Beat{
		ID:        beatID,
		SceneID:   sceneID,
		OrderNum:  int(b.OrderNum),
		Type:      b.Type,
		Intent:    b.Intent,
		Outcome:   b.Outcome,
		CreatedAt: b.CreatedAt.Seconds,
		UpdatedAt: b.UpdatedAt.Seconds,
	}
}

func protoToContentBlock(cb *contentblockpb.ContentBlock) *grpcclient.ContentBlock {
	contentBlockID, _ := uuid.Parse(cb.Id)

	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		Type:      cb.Type,
		Kind:      cb.Kind,
		Content:   cb.Content,
		CreatedAt: cb.CreatedAt.Seconds,
		UpdatedAt: cb.UpdatedAt.Seconds,
	}

	if cb.ChapterId != nil {
		chapterID, _ := uuid.Parse(*cb.ChapterId)
		contentBlock.ChapterID = &chapterID
	}
	if cb.OrderNum != nil {
		orderNum := int(*cb.OrderNum)
		contentBlock.OrderNum = &orderNum
	}

	// Convert metadata from structpb.Struct to map[string]interface{}
	if cb.Metadata != nil {
		contentBlock.Metadata = cb.Metadata.AsMap()
	} else {
		contentBlock.Metadata = make(map[string]interface{})
	}

	return contentBlock
}

func protoToContentAnchor(anchor *contentblockpb.ContentAnchor) *grpcclient.ContentAnchor {
	anchorID, _ := uuid.Parse(anchor.Id)
	contentBlockID, _ := uuid.Parse(anchor.ContentBlockId)
	entityID, _ := uuid.Parse(anchor.EntityId)

	return &grpcclient.ContentAnchor{
		ID:             anchorID,
		ContentBlockID: contentBlockID,
		EntityType:     anchor.EntityType,
		EntityID:       entityID,
		CreatedAt:      anchor.CreatedAt.Seconds,
	}
}

// World entity methods

// GetWorld retrieves a world by ID
func (c *MainServiceClient) GetWorld(ctx context.Context, worldID uuid.UUID) (*grpcclient.World, error) {
	resp, err := c.worldClient.GetWorld(ctx, &worldpb.GetWorldRequest{
		Id: worldID.String(),
	})
	if err != nil {
		return nil, err
	}
	return protoToWorld(resp.World), nil
}

// GetCharacter retrieves a character by ID
func (c *MainServiceClient) GetCharacter(ctx context.Context, characterID uuid.UUID) (*grpcclient.Character, error) {
	resp, err := c.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
		Id: characterID.String(),
	})
	if err != nil {
		return nil, err
	}
	return protoToCharacter(resp.Character), nil
}

// GetLocation retrieves a location by ID
func (c *MainServiceClient) GetLocation(ctx context.Context, locationID uuid.UUID) (*grpcclient.Location, error) {
	resp, err := c.locationClient.GetLocation(ctx, &locationpb.GetLocationRequest{
		Id: locationID.String(),
	})
	if err != nil {
		return nil, err
	}
	return protoToLocation(resp.Location), nil
}

// GetEvent retrieves an event by ID
func (c *MainServiceClient) GetEvent(ctx context.Context, eventID uuid.UUID) (*grpcclient.Event, error) {
	resp, err := c.eventClient.GetEvent(ctx, &eventpb.GetEventRequest{
		Id: eventID.String(),
	})
	if err != nil {
		return nil, err
	}
	return protoToEvent(resp.Event), nil
}

// GetArtifact retrieves an artifact by ID
func (c *MainServiceClient) GetArtifact(ctx context.Context, artifactID uuid.UUID) (*grpcclient.Artifact, error) {
	resp, err := c.artifactClient.GetArtifact(ctx, &artifactpb.GetArtifactRequest{
		Id: artifactID.String(),
	})
	if err != nil {
		return nil, err
	}
	return protoToArtifact(resp.Artifact), nil
}

// GetCharacterTraits retrieves traits for a character
func (c *MainServiceClient) GetCharacterTraits(ctx context.Context, characterID uuid.UUID) ([]*grpcclient.CharacterTrait, error) {
	resp, err := c.characterClient.GetCharacterTraits(ctx, &characterpb.GetCharacterTraitsRequest{
		CharacterId: characterID.String(),
	})
	if err != nil {
		return nil, err
	}
	traits := make([]*grpcclient.CharacterTrait, len(resp.Traits))
	for i, t := range resp.Traits {
		traits[i] = protoToCharacterTrait(t)
	}
	return traits, nil
}

// GetEventCharacters retrieves characters for an event
func (c *MainServiceClient) GetEventCharacters(ctx context.Context, eventID uuid.UUID) ([]*grpcclient.EventCharacter, error) {
	refs, err := c.getEventReferences(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return mapEventCharacters(eventID, refs), nil
}

// GetEventLocations retrieves locations for an event
func (c *MainServiceClient) GetEventLocations(ctx context.Context, eventID uuid.UUID) ([]*grpcclient.EventLocation, error) {
	refs, err := c.getEventReferences(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return mapEventLocations(eventID, refs), nil
}

// GetEventArtifacts retrieves artifacts for an event
func (c *MainServiceClient) GetEventArtifacts(ctx context.Context, eventID uuid.UUID) ([]*grpcclient.EventArtifact, error) {
	refs, err := c.getEventReferences(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return mapEventArtifacts(eventID, refs), nil
}

// ListSceneReferences lists references for a scene
func (c *MainServiceClient) ListSceneReferences(ctx context.Context, sceneID uuid.UUID) ([]*grpcclient.SceneReference, error) {
	resp, err := c.sceneClient.GetSceneReferences(ctx, &scenepb.GetSceneReferencesRequest{
		SceneId: sceneID.String(),
	})
	if err != nil {
		return nil, err
	}
	refs := make([]*grpcclient.SceneReference, len(resp.References))
	for i, ref := range resp.References {
		refs[i] = protoToSceneReference(ref)
	}
	return refs, nil
}

// Helper functions to convert proto to domain models for World entities

func protoToWorld(w *worldpb.World) *grpcclient.World {
	worldID, _ := uuid.Parse(w.Id)
	tenantID, _ := uuid.Parse(w.TenantId)

	return &grpcclient.World{
		ID:          worldID,
		TenantID:    tenantID,
		Name:        w.Name,
		Description: w.Description,
		Genre:       w.Genre,
		IsImplicit:  w.IsImplicit,
		CreatedAt:   w.CreatedAt.Seconds,
		UpdatedAt:   w.UpdatedAt.Seconds,
	}
}

func protoToCharacter(ch *characterpb.Character) *grpcclient.Character {
	charID, _ := uuid.Parse(ch.Id)
	worldID, _ := uuid.Parse(ch.WorldId)

	char := &grpcclient.Character{
		ID:          charID,
		WorldID:     worldID,
		Name:        ch.Name,
		Description: ch.Description,
		CreatedAt:   ch.CreatedAt.Seconds,
		UpdatedAt:   ch.UpdatedAt.Seconds,
	}

	if ch.ArchetypeId != "" {
		archetypeID, _ := uuid.Parse(ch.ArchetypeId)
		char.ArchetypeID = &archetypeID
	}

	return char
}

func protoToCharacterTrait(ct *characterpb.CharacterTrait) *grpcclient.CharacterTrait {
	traitID, _ := uuid.Parse(ct.Id)
	charID, _ := uuid.Parse(ct.CharacterId)
	traitID2, _ := uuid.Parse(ct.TraitId)

	return &grpcclient.CharacterTrait{
		ID:          traitID,
		CharacterID: charID,
		TraitID:     traitID2,
		TraitName:   ct.TraitName,
		Value:       ct.Value,
		Notes:       ct.Notes,
		CreatedAt:   ct.CreatedAt.Seconds,
		UpdatedAt:   ct.UpdatedAt.Seconds,
	}
}

func protoToLocation(loc *locationpb.Location) *grpcclient.Location {
	locID, _ := uuid.Parse(loc.Id)
	worldID, _ := uuid.Parse(loc.WorldId)

	location := &grpcclient.Location{
		ID:             locID,
		WorldID:        worldID,
		Name:           loc.Name,
		Type:           loc.Type,
		Description:    loc.Description,
		HierarchyLevel: int(loc.HierarchyLevel),
		CreatedAt:      loc.CreatedAt.Seconds,
		UpdatedAt:      loc.UpdatedAt.Seconds,
	}

	if loc.ParentId != "" {
		parentID, _ := uuid.Parse(loc.ParentId)
		location.ParentID = &parentID
	}

	return location
}

func protoToEvent(e *eventpb.Event) *grpcclient.Event {
	eventID, _ := uuid.Parse(e.Id)
	worldID, _ := uuid.Parse(e.WorldId)

	event := &grpcclient.Event{
		ID:         eventID,
		WorldID:    worldID,
		Name:       e.Name,
		Importance: int(e.Importance),
		CreatedAt:  e.CreatedAt.Seconds,
		UpdatedAt:  e.UpdatedAt.Seconds,
	}

	if e.Type != nil {
		event.Type = e.Type
	}
	if e.Description != nil {
		event.Description = e.Description
	}
	if e.Timeline != nil {
		event.Timeline = e.Timeline
	}

	return event
}

func (c *MainServiceClient) getEventReferences(ctx context.Context, eventID uuid.UUID) ([]*eventpb.EventReference, error) {
	resp, err := c.eventClient.GetEventReferences(ctx, &eventpb.GetEventReferencesRequest{
		EventId: eventID.String(),
	})
	if err != nil {
		return nil, err
	}
	return resp.References, nil
}

func mapEventCharacters(eventID uuid.UUID, refs []*eventpb.EventReference) []*grpcclient.EventCharacter {
	eventChars := make([]*grpcclient.EventCharacter, 0)
	for _, ref := range refs {
		if ref.EntityType != "character" {
			continue
		}
		charID, err := uuid.Parse(ref.EntityId)
		if err != nil {
			continue
		}
		refID, err := uuid.Parse(ref.Id)
		if err != nil {
			continue
		}
		createdAt := int64(0)
		if ref.CreatedAt != nil {
			createdAt = ref.CreatedAt.Seconds
		}
		eventChars = append(eventChars, &grpcclient.EventCharacter{
			ID:          refID,
			EventID:     eventID,
			CharacterID: charID,
			Role:        ref.RelationshipType,
			CreatedAt:   createdAt,
		})
	}
	return eventChars
}

func mapEventLocations(eventID uuid.UUID, refs []*eventpb.EventReference) []*grpcclient.EventLocation {
	eventLocs := make([]*grpcclient.EventLocation, 0)
	for _, ref := range refs {
		if ref.EntityType != "location" {
			continue
		}
		locID, err := uuid.Parse(ref.EntityId)
		if err != nil {
			continue
		}
		refID, err := uuid.Parse(ref.Id)
		if err != nil {
			continue
		}
		createdAt := int64(0)
		if ref.CreatedAt != nil {
			createdAt = ref.CreatedAt.Seconds
		}
		eventLocs = append(eventLocs, &grpcclient.EventLocation{
			ID:           refID,
			EventID:      eventID,
			LocationID:   locID,
			Significance: ref.RelationshipType,
			CreatedAt:    createdAt,
		})
	}
	return eventLocs
}

func mapEventArtifacts(eventID uuid.UUID, refs []*eventpb.EventReference) []*grpcclient.EventArtifact {
	eventArts := make([]*grpcclient.EventArtifact, 0)
	for _, ref := range refs {
		if ref.EntityType != "artifact" {
			continue
		}
		artifactID, err := uuid.Parse(ref.EntityId)
		if err != nil {
			continue
		}
		refID, err := uuid.Parse(ref.Id)
		if err != nil {
			continue
		}
		createdAt := int64(0)
		if ref.CreatedAt != nil {
			createdAt = ref.CreatedAt.Seconds
		}
		eventArts = append(eventArts, &grpcclient.EventArtifact{
			ID:         refID,
			EventID:    eventID,
			ArtifactID: artifactID,
			Role:       ref.RelationshipType,
			CreatedAt:  createdAt,
		})
	}
	return eventArts
}

func protoToArtifact(a *artifactpb.Artifact) *grpcclient.Artifact {
	artifactID, _ := uuid.Parse(a.Id)
	worldID, _ := uuid.Parse(a.WorldId)

	return &grpcclient.Artifact{
		ID:          artifactID,
		WorldID:     worldID,
		Name:        a.Name,
		Description: a.Description,
		Rarity:      a.Rarity,
		CreatedAt:   a.CreatedAt.Seconds,
		UpdatedAt:   a.UpdatedAt.Seconds,
	}
}

func protoToSceneReference(ref *scenepb.SceneReference) *grpcclient.SceneReference {
	refID, _ := uuid.Parse(ref.Id)
	sceneID, _ := uuid.Parse(ref.SceneId)
	entityID, _ := uuid.Parse(ref.EntityId)

	return &grpcclient.SceneReference{
		ID:         refID,
		SceneID:    sceneID,
		EntityType: ref.EntityType,
		EntityID:   entityID,
		CreatedAt:  ref.CreatedAt.Seconds,
	}
}

// GetFaction retrieves a faction by ID
func (c *MainServiceClient) GetFaction(ctx context.Context, factionID uuid.UUID) (*grpcclient.Faction, error) {
	resp, err := c.factionClient.GetFaction(ctx, &factionpb.GetFactionRequest{
		Id: factionID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToFaction(resp.Faction), nil
}

// GetLore retrieves a lore by ID
func (c *MainServiceClient) GetLore(ctx context.Context, loreID uuid.UUID) (*grpcclient.Lore, error) {
	resp, err := c.loreClient.GetLore(ctx, &lorepb.GetLoreRequest{
		Id: loreID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToLore(resp.Lore), nil
}

func protoToFaction(f *factionpb.Faction) *grpcclient.Faction {
	factionID, _ := uuid.Parse(f.Id)
	worldID, _ := uuid.Parse(f.WorldId)

	faction := &grpcclient.Faction{
		ID:             factionID,
		WorldID:        worldID,
		Name:           f.Name,
		Description:    f.Description,
		Beliefs:        f.Beliefs,
		Structure:      f.Structure,
		Symbols:        f.Symbols,
		HierarchyLevel: int(f.HierarchyLevel),
		CreatedAt:      f.CreatedAt.Seconds,
		UpdatedAt:      f.UpdatedAt.Seconds,
	}

	if f.ParentId != nil && *f.ParentId != "" {
		parentID, _ := uuid.Parse(*f.ParentId)
		faction.ParentID = &parentID
	}

	if f.Type != nil && *f.Type != "" {
		faction.Type = f.Type
	}

	return faction
}

func protoToLore(l *lorepb.Lore) *grpcclient.Lore {
	loreID, _ := uuid.Parse(l.Id)
	worldID, _ := uuid.Parse(l.WorldId)

	lore := &grpcclient.Lore{
		ID:             loreID,
		WorldID:        worldID,
		Name:           l.Name,
		Description:    l.Description,
		Rules:          l.Rules,
		Limitations:    l.Limitations,
		Requirements:   l.Requirements,
		HierarchyLevel: int(l.HierarchyLevel),
		CreatedAt:      l.CreatedAt.Seconds,
		UpdatedAt:      l.UpdatedAt.Seconds,
	}

	if l.ParentId != nil && *l.ParentId != "" {
		parentID, _ := uuid.Parse(*l.ParentId)
		lore.ParentID = &parentID
	}

	if l.Category != nil && *l.Category != "" {
		lore.Category = l.Category
	}

	return lore
}
