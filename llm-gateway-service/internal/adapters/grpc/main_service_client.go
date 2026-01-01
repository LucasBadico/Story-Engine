package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	storypb "github.com/story-engine/main-service/proto/story"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	scenepb "github.com/story-engine/main-service/proto/scene"
	beatpb "github.com/story-engine/main-service/proto/beat"
	prosepb "github.com/story-engine/main-service/proto/prose"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ grpcclient.MainServiceClient = (*MainServiceClient)(nil)

// MainServiceClient implements the gRPC client interface for main-service
type MainServiceClient struct {
	storyClient              storypb.StoryServiceClient
	chapterClient            chapterpb.ChapterServiceClient
	sceneClient              scenepb.SceneServiceClient
	beatClient               beatpb.BeatServiceClient
	proseClient              prosepb.ProseBlockServiceClient
	proseReferenceClient     prosepb.ProseBlockReferenceServiceClient
	conn                     *grpclib.ClientConn
}

// NewMainServiceClient creates a new gRPC client for main-service
func NewMainServiceClient(addr string) (*MainServiceClient, error) {
	conn, err := grpclib.NewClient(addr, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main-service: %w", err)
	}

	return &MainServiceClient{
		storyClient:          storypb.NewStoryServiceClient(conn),
		chapterClient:       chapterpb.NewChapterServiceClient(conn),
		sceneClient:         scenepb.NewSceneServiceClient(conn),
		beatClient:          beatpb.NewBeatServiceClient(conn),
		proseClient:         prosepb.NewProseBlockServiceClient(conn),
		proseReferenceClient: prosepb.NewProseBlockReferenceServiceClient(conn),
		conn:                 conn,
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

// GetProseBlock retrieves a prose block by ID
func (c *MainServiceClient) GetProseBlock(ctx context.Context, proseBlockID uuid.UUID) (*grpcclient.ProseBlock, error) {
	resp, err := c.proseClient.GetProseBlock(ctx, &prosepb.GetProseBlockRequest{
		Id: proseBlockID.String(),
	})
	if err != nil {
		return nil, err
	}

	return protoToProseBlock(resp.ProseBlock), nil
}

// ListProseBlocksByChapter lists prose blocks for a chapter
func (c *MainServiceClient) ListProseBlocksByChapter(ctx context.Context, chapterID uuid.UUID) ([]*grpcclient.ProseBlock, error) {
	resp, err := c.proseClient.ListProseBlocksByChapter(ctx, &prosepb.ListProseBlocksByChapterRequest{
		ChapterId: chapterID.String(),
	})
	if err != nil {
		return nil, err
	}

	proseBlocks := make([]*grpcclient.ProseBlock, len(resp.ProseBlocks))
	for i, pb := range resp.ProseBlocks {
		proseBlocks[i] = protoToProseBlock(pb)
	}

	return proseBlocks, nil
}

// ListProseBlockReferences lists references for a prose block
func (c *MainServiceClient) ListProseBlockReferences(ctx context.Context, proseBlockID uuid.UUID) ([]*grpcclient.ProseBlockReference, error) {
	resp, err := c.proseReferenceClient.ListProseBlockReferencesByProseBlock(ctx, &prosepb.ListProseBlockReferencesByProseBlockRequest{
		ProseBlockId: proseBlockID.String(),
	})
	if err != nil {
		return nil, err
	}

	references := make([]*grpcclient.ProseBlockReference, len(resp.References))
	for i, ref := range resp.References {
		references[i] = protoToProseBlockReference(ref)
	}

	return references, nil
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
	if s.LocationId != nil {
		locID, _ := uuid.Parse(*s.LocationId)
		scene.LocationID = &locID
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

func protoToProseBlock(pb *prosepb.ProseBlock) *grpcclient.ProseBlock {
	proseID, _ := uuid.Parse(pb.Id)
	chapterID, _ := uuid.Parse(pb.ChapterId)

	return &grpcclient.ProseBlock{
		ID:        proseID,
		ChapterID: chapterID,
		OrderNum:  int(pb.OrderNum),
		Kind:      pb.Kind,
		Content:   pb.Content,
		WordCount: int(pb.WordCount),
		CreatedAt: pb.CreatedAt.Seconds,
		UpdatedAt: pb.UpdatedAt.Seconds,
	}
}

func protoToProseBlockReference(ref *prosepb.ProseBlockReference) *grpcclient.ProseBlockReference {
	refID, _ := uuid.Parse(ref.Id)
	proseBlockID, _ := uuid.Parse(ref.ProseBlockId)
	entityID, _ := uuid.Parse(ref.EntityId)

	return &grpcclient.ProseBlockReference{
		ID:          refID,
		ProseBlockID: proseBlockID,
		EntityType:  ref.EntityType,
		EntityID:    entityID,
		CreatedAt:   ref.CreatedAt.Seconds,
	}
}

