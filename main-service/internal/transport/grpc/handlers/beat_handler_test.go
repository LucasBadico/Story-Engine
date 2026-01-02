//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	beatpb "github.com/story-engine/main-service/proto/beat"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestBeatHandler_CreateBeat(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Beat",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
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

	t.Run("successful creation", func(t *testing.T) {
		req := &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "setup",
			Intent:   "Introduce the protagonist",
			Outcome:  "Reader knows the main character",
		}
		resp, err := beatClient.CreateBeat(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Beat.Type != "setup" {
			t.Errorf("expected type 'setup', got '%s'", resp.Beat.Type)
		}
		if resp.Beat.Intent != "Introduce the protagonist" {
			t.Errorf("expected intent 'Introduce the protagonist', got '%s'", resp.Beat.Intent)
		}
	})

	t.Run("invalid scene_id", func(t *testing.T) {
		req := &beatpb.CreateBeatRequest{
			SceneId:  uuid.New().String(),
			OrderNum: 1,
			Type:     "setup",
		}
		_, err := beatClient.CreateBeat(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid scene_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		req := &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "invalid_type",
		}
		_, err := beatClient.CreateBeat(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid type")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("missing type", func(t *testing.T) {
		req := &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "",
		}
		_, err := beatClient.CreateBeat(ctx, req)
		if err == nil {
			t.Fatal("expected error for missing type")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("invalid order_num", func(t *testing.T) {
		req := &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 0,
			Type:     "setup",
		}
		_, err := beatClient.CreateBeat(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid order_num")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestBeatHandler_GetBeat(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get Beat",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
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

	t.Run("existing beat", func(t *testing.T) {
		createResp, err := beatClient.CreateBeat(ctx, &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "setup",
		})
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		getResp, err := beatClient.GetBeat(ctx, &beatpb.GetBeatRequest{
			Id: createResp.Beat.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getResp.Beat.Id != createResp.Beat.Id {
			t.Errorf("expected ID %s, got %s", createResp.Beat.Id, getResp.Beat.Id)
		}
	})

	t.Run("non-existing beat", func(t *testing.T) {
		_, err := beatClient.GetBeat(ctx, &beatpb.GetBeatRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing beat")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestBeatHandler_UpdateBeat(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update Beat",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
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

	t.Run("successful update", func(t *testing.T) {
		createResp, err := beatClient.CreateBeat(ctx, &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "setup",
			Intent:   "Original intent",
		})
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		newIntent := "Updated intent"
		newOutcome := "New outcome"
		updateResp, err := beatClient.UpdateBeat(ctx, &beatpb.UpdateBeatRequest{
			Id:      createResp.Beat.Id,
			Intent:  &newIntent,
			Outcome: &newOutcome,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.Beat.Intent != "Updated intent" {
			t.Errorf("expected intent 'Updated intent', got '%s'", updateResp.Beat.Intent)
		}
		if updateResp.Beat.Outcome != "New outcome" {
			t.Errorf("expected outcome 'New outcome', got '%s'", updateResp.Beat.Outcome)
		}
	})
}

func TestBeatHandler_MoveBeat(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Move Beat",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	scene1Resp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
		StoryId:  storyResp.Story.Id,
		OrderNum: 1,
	})
	if err != nil {
		t.Fatalf("failed to create scene1: %v", err)
	}
	scene2Resp, err := sceneClient.CreateScene(ctx, &scenepb.CreateSceneRequest{
		StoryId:  storyResp.Story.Id,
		OrderNum: 2,
	})
	if err != nil {
		t.Fatalf("failed to create scene2: %v", err)
	}

	t.Run("move to different scene", func(t *testing.T) {
		createResp, err := beatClient.CreateBeat(ctx, &beatpb.CreateBeatRequest{
			SceneId:  scene1Resp.Scene.Id,
			OrderNum: 1,
			Type:     "setup",
		})
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		moveResp, err := beatClient.MoveBeat(ctx, &beatpb.MoveBeatRequest{
			Id:      createResp.Beat.Id,
			SceneId: scene2Resp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if moveResp.Beat.SceneId != scene2Resp.Scene.Id {
			t.Errorf("expected scene_id %s, got %s", scene2Resp.Scene.Id, moveResp.Beat.SceneId)
		}
	})
}

func TestBeatHandler_ListBeatsByScene(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List Beats",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
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

	t.Run("list beats by scene", func(t *testing.T) {
		beatTypes := []string{"setup", "turn", "reveal"}
		for i, beatType := range beatTypes {
			_, err := beatClient.CreateBeat(ctx, &beatpb.CreateBeatRequest{
				SceneId:  sceneResp.Scene.Id,
				OrderNum: int32(i + 1),
				Type:     beatType,
			})
			if err != nil {
				t.Fatalf("failed to create beat %d: %v", i, err)
			}
		}

		listResp, err := beatClient.ListBeatsByScene(ctx, &beatpb.ListBeatsBySceneRequest{
			SceneId: sceneResp.Scene.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.Beats) != 3 {
			t.Errorf("expected 3 beats, got %d", len(listResp.Beats))
		}
	})
}

func TestBeatHandler_DeleteBeat(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	beatClient := beatpb.NewBeatServiceClient(conn)
	sceneClient := scenepb.NewSceneServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete Beat",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
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

	t.Run("successful delete", func(t *testing.T) {
		createResp, err := beatClient.CreateBeat(ctx, &beatpb.CreateBeatRequest{
			SceneId:  sceneResp.Scene.Id,
			OrderNum: 1,
			Type:     "setup",
		})
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		deleteResp, err := beatClient.DeleteBeat(ctx, &beatpb.DeleteBeatRequest{
			Id: createResp.Beat.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})
}


