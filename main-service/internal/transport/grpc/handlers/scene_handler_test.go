//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSceneHandler_CreateScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Scene",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("successful creation with chapter", func(t *testing.T) {
		chapterID := chapterResp.Chapter.Id
		req := &scenepb.CreateSceneRequest{
			StoryId:   storyResp.Story.Id,
			ChapterId: &chapterID,
			OrderNum:  1,
			Goal:      "Test scene goal",
			TimeRef:   "Morning",
		}
		resp, err := sceneClient.CreateScene(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Scene.Goal != "Test scene goal" {
			t.Errorf("expected goal 'Test scene goal', got '%s'", resp.Scene.Goal)
		}
		if resp.Scene.OrderNum != 1 {
			t.Errorf("expected order_num 1, got %d", resp.Scene.OrderNum)
		}
	})

	t.Run("successful creation without chapter", func(t *testing.T) {
		req := &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
			Goal:     "Orphan scene",
		}
		resp, err := sceneClient.CreateScene(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Scene.ChapterId != nil {
			t.Errorf("expected nil chapter_id, got %s", *resp.Scene.ChapterId)
		}
	})

	t.Run("invalid story_id", func(t *testing.T) {
		req := &scenepb.CreateSceneRequest{
			StoryId:  uuid.New().String(),
			OrderNum: 1,
		}
		_, err := sceneClient.CreateScene(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid story_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid order_num", func(t *testing.T) {
		req := &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 0,
		}
		_, err := sceneClient.CreateScene(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid order_num")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestSceneHandler_GetScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get Scene",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("existing scene", func(t *testing.T) {
		createResp, _ := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
			Goal:     "Get Test Scene",
		})

		getResp, err := sceneClient.GetScene(ctx, &scenepb.GetSceneRequest{
			Id: createResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getResp.Scene.Id != createResp.Scene.Id {
			t.Errorf("expected ID %s, got %s", createResp.Scene.Id, getResp.Scene.Id)
		}
	})

	t.Run("non-existing scene", func(t *testing.T) {
		_, err := sceneClient.GetScene(ctx, &scenepb.GetSceneRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing scene")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestSceneHandler_UpdateScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update Scene",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("successful update", func(t *testing.T) {
		createResp, _ := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
			Goal:     "Original Goal",
		})

		newGoal := "Updated Goal"
		newTimeRef := "Evening"
		updateResp, err := sceneClient.UpdateScene(ctx, &scenepb.UpdateSceneRequest{
			Id:      createResp.Scene.Id,
			Goal:    &newGoal,
			TimeRef: &newTimeRef,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.Scene.Goal != "Updated Goal" {
			t.Errorf("expected goal 'Updated Goal', got '%s'", updateResp.Scene.Goal)
		}
		if updateResp.Scene.TimeRef != "Evening" {
			t.Errorf("expected time_ref 'Evening', got '%s'", updateResp.Scene.TimeRef)
		}
	})
}

func TestSceneHandler_MoveScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Move Scene",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapter1Resp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})
	chapter2Resp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  2,
		Title:   "Chapter Two",
	})

	t.Run("move to different chapter", func(t *testing.T) {
		chapter1ID := chapter1Resp.Chapter.Id
		createResp, _ := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:   storyResp.Story.Id,
			ChapterId: &chapter1ID,
			OrderNum:  1,
		})

		chapter2ID := chapter2Resp.Chapter.Id
		moveResp, err := sceneClient.MoveScene(ctx, &scenepb.MoveSceneRequest{
			Id:        createResp.Scene.Id,
			ChapterId: &chapter2ID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if moveResp.Scene.ChapterId == nil || *moveResp.Scene.ChapterId != chapter2ID {
			t.Errorf("expected chapter_id %s, got %v", chapter2ID, moveResp.Scene.ChapterId)
		}
	})

	t.Run("move to no chapter", func(t *testing.T) {
		chapter1ID := chapter1Resp.Chapter.Id
		createResp, _ := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:   storyResp.Story.Id,
			ChapterId: &chapter1ID,
			OrderNum:  2,
		})

		moveResp, err := sceneClient.MoveScene(ctx, &scenepb.MoveSceneRequest{
			Id: createResp.Scene.Id,
			// ChapterId not set = move to no chapter
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if moveResp.Scene.ChapterId != nil {
			t.Errorf("expected nil chapter_id, got %v", moveResp.Scene.ChapterId)
		}
	})
}

func TestSceneHandler_ListScenesByStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List Scenes",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("list scenes by story", func(t *testing.T) {
		// Create multiple scenes
		for i := 1; i <= 3; i++ {
			_, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
				StoryId:  storyResp.Story.Id,
				OrderNum: int32(i),
				Goal:     "Scene goal",
			})
			if err != nil {
				t.Fatalf("failed to create scene %d: %v", i, err)
			}
		}

		listResp, err := sceneClient.ListScenesByStory(ctx, &scenepb.ListScenesByStoryRequest{
			StoryId: storyResp.Story.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.Scenes) != 3 {
			t.Errorf("expected 3 scenes, got %d", len(listResp.Scenes))
		}
	})
}

func TestSceneHandler_DeleteScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete Scene",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("successful delete", func(t *testing.T) {
		createResp, _ := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
		})

		deleteResp, err := sceneClient.DeleteScene(ctx, &scenepb.DeleteSceneRequest{
			Id: createResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})
}


