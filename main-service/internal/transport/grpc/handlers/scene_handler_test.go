//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/story"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	characterpb "github.com/story-engine/main-service/proto/character"
	locationpb "github.com/story-engine/main-service/proto/location"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Scene",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get Scene",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("existing scene", func(t *testing.T) {
		createResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
			Goal:     "Get Test Scene",
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update Scene",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		createResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
			Goal:     "Original Goal",
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Move Scene",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	chapter1Resp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})
	if err != nil {
		t.Fatalf("failed to create chapter1: %v", err)
	}
	chapter2Resp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  2,
		Title:   "Chapter Two",
	})
	if err != nil {
		t.Fatalf("failed to create chapter2: %v", err)
	}

	t.Run("move to different chapter", func(t *testing.T) {
		chapter1ID := chapter1Resp.Chapter.Id
		createResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:   storyResp.Story.Id,
			ChapterId: &chapter1ID,
			OrderNum:  1,
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

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
		createResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:   storyResp.Story.Id,
			ChapterId: &chapter1ID,
			OrderNum:  2,
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List Scenes",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

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
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete Scene",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		createResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 1,
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

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

func TestSceneHandler_AddSceneReference(t *testing.T) {
	conn, cleanup := setupTestServerWithSceneReferences(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	locationClient := locationpb.NewLocationServiceClient(conn)
	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Scene References",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{Name: "Test World"})
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	sceneResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
		StoryId:  storyResp.Story.Id,
		OrderNum: 1,
	})
	if err != nil {
		t.Fatalf("failed to create scene: %v", err)
	}

	t.Run("add character reference", func(t *testing.T) {
		characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		addResp, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "character",
			EntityId:   characterResp.Character.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addResp.Reference.EntityType != "character" {
			t.Errorf("expected entity_type 'character', got '%s'", addResp.Reference.EntityType)
		}
		if addResp.Reference.EntityId != characterResp.Character.Id {
			t.Errorf("expected entity_id %s, got %s", characterResp.Character.Id, addResp.Reference.EntityId)
		}
	})

	t.Run("add location reference", func(t *testing.T) {
		locationResp, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Location",
		})
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		addResp, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "location",
			EntityId:   locationResp.Location.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addResp.Reference.EntityType != "location" {
			t.Errorf("expected entity_type 'location', got '%s'", addResp.Reference.EntityType)
		}
	})

	t.Run("add artifact reference", func(t *testing.T) {
		artifactResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Artifact",
		})
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		addResp, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "artifact",
			EntityId:   artifactResp.Artifact.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addResp.Reference.EntityType != "artifact" {
			t.Errorf("expected entity_type 'artifact', got '%s'", addResp.Reference.EntityType)
		}
	})

	t.Run("invalid entity type", func(t *testing.T) {
		_, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "invalid",
			EntityId:   uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for invalid entity type")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("invalid scene_id", func(t *testing.T) {
		_, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    uuid.New().String(),
			EntityType: "character",
			EntityId:   uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for invalid scene_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestSceneHandler_RemoveSceneReference(t *testing.T) {
	conn, cleanup := setupTestServerWithSceneReferences(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Remove Reference",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{Name: "Test World"})
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	sceneResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
		StoryId:  storyResp.Story.Id,
		OrderNum: 1,
	})
	if err != nil {
		t.Fatalf("failed to create scene: %v", err)
	}
	characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
		WorldId: worldResp.World.Id,
		Name:    "Test Character",
	})
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful remove", func(t *testing.T) {
		// Add reference first
		_, err := sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "character",
			EntityId:   characterResp.Character.Id,
		})
		if err != nil {
			t.Fatalf("failed to add reference: %v", err)
		}

		// Remove reference
		_, err = sceneClient.RemoveSceneReference(ctx, &scenepb.RemoveSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "character",
			EntityId:   characterResp.Character.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's removed
		getResp, err := sceneClient.GetSceneReferences(ctx, &scenepb.GetSceneReferencesRequest{
			SceneId: sceneResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(getResp.References) != 0 {
			t.Errorf("expected 0 references, got %d", len(getResp.References))
		}
	})

	t.Run("remove non-existing reference", func(t *testing.T) {
		_, err := sceneClient.RemoveSceneReference(ctx, &scenepb.RemoveSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "character",
			EntityId:   uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing reference")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestSceneHandler_GetSceneReferences(t *testing.T) {
	conn, cleanup := setupTestServerWithSceneReferences(t)
	defer cleanup()

	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	locationClient := locationpb.NewLocationServiceClient(conn)
	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get References",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{Name: "Test World"})
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	sceneResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
		StoryId:  storyResp.Story.Id,
		OrderNum: 1,
	})
	if err != nil {
		t.Fatalf("failed to create scene: %v", err)
	}

	t.Run("get multiple references", func(t *testing.T) {
		characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		locationResp, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Location",
		})
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}
		artifactResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Artifact",
		})

		// Add references
		_, _ = sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "character",
			EntityId:   characterResp.Character.Id,
		})
		_, _ = sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "location",
			EntityId:   locationResp.Location.Id,
		})
		_, _ = sceneClient.AddSceneReference(ctx, &scenepb.AddSceneReferenceRequest{
			SceneId:    sceneResp.Scene.Id,
			EntityType: "artifact",
			EntityId:   artifactResp.Artifact.Id,
		})

		// Get references
		getResp, err := sceneClient.GetSceneReferences(ctx, &scenepb.GetSceneReferencesRequest{
			SceneId: sceneResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(getResp.References) != 3 {
			t.Errorf("expected 3 references, got %d", len(getResp.References))
		}
		if getResp.TotalCount != 3 {
			t.Errorf("expected total_count 3, got %d", getResp.TotalCount)
		}
	})

	t.Run("get empty references", func(t *testing.T) {
		newSceneResp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
			StoryId:  storyResp.Story.Id,
			OrderNum: 2,
		})
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		getResp, err := sceneClient.GetSceneReferences(ctx, &scenepb.GetSceneReferencesRequest{
			SceneId: newSceneResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(getResp.References) != 0 {
			t.Errorf("expected 0 references, got %d", len(getResp.References))
		}
	})

	t.Run("invalid scene_id", func(t *testing.T) {
		_, err := sceneClient.GetSceneReferences(ctx, &scenepb.GetSceneReferencesRequest{
			SceneId: "not-a-uuid",
		})
		if err == nil {
			t.Fatal("expected error for invalid scene_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

// setupTestServerWithSceneReferences creates a test server with Scene, Character, Location, and Artifact handlers
func setupTestServerWithSceneReferences(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	// Initialize use cases
	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, sceneRepo, postgres.NewBeatRepository(db), postgres.NewContentBlockRepository(db), auditLogRepo, postgres.NewTransactionRepository(db), log)
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addSceneReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getSceneReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)

	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, postgres.NewCharacterTraitRepository(db), worldRepo, auditLogRepo, log)

	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, auditLogRepo, log)

	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)

	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)

	// Create handlers
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	storyHandler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, versionGraphUseCase, log)
	chapterHandler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	sceneHandler := NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addSceneReferenceUC, removeSceneReferenceUC, getSceneReferencesUC, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, characterapp.NewGetCharacterTraitsUseCase(postgres.NewCharacterTraitRepository(db), log), characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), postgres.NewCharacterTraitRepository(db), log), characterapp.NewUpdateCharacterTraitUseCase(postgres.NewCharacterTraitRepository(db), postgres.NewTraitRepository(db), log), characterapp.NewRemoveTraitFromCharacterUseCase(postgres.NewCharacterTraitRepository(db), log), log)
	locationHandler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, locationapp.NewGetChildrenUseCase(locationRepo, log), locationapp.NewGetAncestorsUseCase(locationRepo, log), locationapp.NewGetDescendantsUseCase(locationRepo, log), locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log), log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log), artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log), artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log), log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		WorldHandler:    worldHandler,
		StoryHandler:     storyHandler,
		ChapterHandler:  chapterHandler,
		SceneHandler:     sceneHandler,
		CharacterHandler: characterHandler,
		LocationHandler:  locationHandler,
		ArtifactHandler:  artifactHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}


