//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	characterpb "github.com/story-engine/main-service/proto/character"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	pb "github.com/story-engine/main-service/proto/relationship_state"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRelationshipStateHandler_CreateAndGet(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewRelationshipStateServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)

	// Setup: tenant → world → 2 characters + story → chapter
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{Name: "Test World"})
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	char1, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
		WorldId: worldResp.World.Id, Name: "Haruka",
	})
	if err != nil {
		t.Fatalf("failed to create char1: %v", err)
	}
	char2, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
		WorldId: worldResp.World.Id, Name: "Tetsuo",
	})
	if err != nil {
		t.Fatalf("failed to create char2: %v", err)
	}

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id, Number: 1, Title: "Chapter 1",
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	t.Run("create with entities", func(t *testing.T) {
		resp, err := client.CreateRelationshipState(ctx, &pb.CreateRelationshipStateRequest{
			ChapterId:          chapterResp.Chapter.Id,
			ChapterNumber:      1,
			TensionDescription: "Surface-level intimacy masking deep resentment",
			HasTimelineDay:     true,
			TimelineDay:        -40,
			TimelineConfidence: "explicit",
			Entities: []*pb.EntityLink{
				{Role: "source", EntityType: "character", EntityId: char1.Character.Id},
				{Role: "target", EntityType: "character", EntityId: char2.Character.Id},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.RelationshipState.TensionDescription != "Surface-level intimacy masking deep resentment" {
			t.Errorf("unexpected tension: %s", resp.RelationshipState.TensionDescription)
		}
		if len(resp.RelationshipState.Entities) != 2 {
			t.Errorf("expected 2 entities, got %d", len(resp.RelationshipState.Entities))
		}

		// Get
		getResp, err := client.GetRelationshipState(ctx, &pb.GetRelationshipStateRequest{Id: resp.RelationshipState.Id})
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if getResp.RelationshipState.TimelineDay != -40 {
			t.Errorf("expected timeline_day -40, got %d", getResp.RelationshipState.TimelineDay)
		}
	})

	t.Run("get non-existing", func(t *testing.T) {
		_, err := client.GetRelationshipState(ctx, &pb.GetRelationshipStateRequest{Id: uuid.New().String()})
		if err == nil {
			t.Fatal("expected error")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestRelationshipStateHandler_ListByChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewRelationshipStateServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id, Number: 1,
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	// Create 2 relationship states for same chapter
	for i := 0; i < 2; i++ {
		_, err := client.CreateRelationshipState(ctx, &pb.CreateRelationshipStateRequest{
			ChapterId:          chapterResp.Chapter.Id,
			ChapterNumber:      1,
			TensionDescription: "Tension pair",
		})
		if err != nil {
			t.Fatalf("failed to create state %d: %v", i, err)
		}
	}

	resp, err := client.ListRelationshipStatesByChapter(ctx, &pb.ListRelationshipStatesByChapterRequest{
		ChapterId: chapterResp.Chapter.Id,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.RelationshipStates) != 2 {
		t.Errorf("expected 2 states, got %d", len(resp.RelationshipStates))
	}
}

func TestRelationshipStateHandler_Delete(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewRelationshipStateServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id, Number: 1,
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	createResp, err := client.CreateRelationshipState(ctx, &pb.CreateRelationshipStateRequest{
		ChapterId:          chapterResp.Chapter.Id,
		ChapterNumber:      1,
		TensionDescription: "to delete",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = client.DeleteRelationshipState(ctx, &pb.DeleteRelationshipStateRequest{Id: createResp.RelationshipState.Id})
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = client.GetRelationshipState(ctx, &pb.GetRelationshipStateRequest{Id: createResp.RelationshipState.Id})
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}
