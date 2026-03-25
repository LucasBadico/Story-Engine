//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pb "github.com/story-engine/main-service/proto/timeline_event"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTimelineEventHandler_CreateAndGet(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewTimelineEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)

	// Setup: tenant → story → chapter
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{Name: "Test Tenant"})
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
		Number:  0,
		Title:   "Prologue",
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	t.Run("create with timeline", func(t *testing.T) {
		resp, err := client.CreateTimelineEvent(ctx, &pb.CreateTimelineEventRequest{
			StoryId:            storyResp.Story.Id,
			ChapterId:          chapterResp.Chapter.Id,
			ChapterNumber:      0,
			Description:        "Collaring ceremony at Rouge Seasons Hotel",
			HasTimelineDay:     true,
			TimelineDay:        0,
			TimelineConfidence: "explicit",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TimelineEvent.Description != "Collaring ceremony at Rouge Seasons Hotel" {
			t.Errorf("unexpected description: %s", resp.TimelineEvent.Description)
		}
		if resp.TimelineEvent.TimelineDay != 0 {
			t.Errorf("expected timeline_day 0, got %d", resp.TimelineEvent.TimelineDay)
		}

		// Get
		getResp, err := client.GetTimelineEvent(ctx, &pb.GetTimelineEventRequest{Id: resp.TimelineEvent.Id})
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if getResp.TimelineEvent.Id != resp.TimelineEvent.Id {
			t.Error("ID mismatch")
		}
	})

	t.Run("get non-existing", func(t *testing.T) {
		_, err := client.GetTimelineEvent(ctx, &pb.GetTimelineEventRequest{Id: uuid.New().String()})
		if err == nil {
			t.Fatal("expected error")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestTimelineEventHandler_ListByStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewTimelineEventServiceClient(conn)
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

	// Create chapters + events at ch 1, 3, 5
	for _, ch := range []int32{1, 3, 5} {
		chResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  ch,
		})
		if err != nil {
			t.Fatalf("failed to create chapter %d: %v", ch, err)
		}
		_, err = client.CreateTimelineEvent(ctx, &pb.CreateTimelineEventRequest{
			StoryId:       storyResp.Story.Id,
			ChapterId:     chResp.Chapter.Id,
			ChapterNumber: ch,
			Description:   "Event at chapter",
		})
		if err != nil {
			t.Fatalf("failed to create event at chapter %d: %v", ch, err)
		}
	}

	t.Run("as_of_chapter=4 returns 2 events", func(t *testing.T) {
		resp, err := client.ListTimelineEventsByStory(ctx, &pb.ListTimelineEventsByStoryRequest{
			StoryId:     storyResp.Story.Id,
			AsOfChapter: 4,
			Limit:       50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.TimelineEvents) != 2 {
			t.Errorf("expected 2 events (ch 1, 3), got %d", len(resp.TimelineEvents))
		}
	})

	t.Run("as_of_chapter=10 returns all 3", func(t *testing.T) {
		resp, err := client.ListTimelineEventsByStory(ctx, &pb.ListTimelineEventsByStoryRequest{
			StoryId:     storyResp.Story.Id,
			AsOfChapter: 10,
			Limit:       50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.TimelineEvents) != 3 {
			t.Errorf("expected 3 events, got %d", len(resp.TimelineEvents))
		}
	})
}

func TestTimelineEventHandler_Delete(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewTimelineEventServiceClient(conn)
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
	chResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id, Number: 1,
	})
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	createResp, err := client.CreateTimelineEvent(ctx, &pb.CreateTimelineEventRequest{
		StoryId:       storyResp.Story.Id,
		ChapterId:     chResp.Chapter.Id,
		ChapterNumber: 1,
		Description:   "to delete",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = client.DeleteTimelineEvent(ctx, &pb.DeleteTimelineEventRequest{Id: createResp.TimelineEvent.Id})
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = client.GetTimelineEvent(ctx, &pb.GetTimelineEventRequest{Id: createResp.TimelineEvent.Id})
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}
