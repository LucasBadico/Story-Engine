//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestChapterHandler_CreateChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup: create tenant and story
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Chapter",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
		Title: "Test Story for Chapter",
	})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		req := &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Chapter One",
		}
		resp, err := chapterClient.CreateChapter(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Chapter.Title != "Chapter One" {
			t.Errorf("expected title 'Chapter One', got '%s'", resp.Chapter.Title)
		}
		if resp.Chapter.Number != 1 {
			t.Errorf("expected number 1, got %d", resp.Chapter.Number)
		}
		if resp.Chapter.StoryId != storyResp.Story.Id {
			t.Errorf("expected story_id %s, got %s", storyResp.Story.Id, resp.Chapter.StoryId)
		}
	})

	t.Run("invalid story_id", func(t *testing.T) {
		req := &chapterpb.CreateChapterRequest{
			StoryId: uuid.New().String(),
			Number:  1,
			Title:   "Chapter One",
		}
		_, err := chapterClient.CreateChapter(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid story_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		req := &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  0,
			Title:   "Chapter Zero",
		}
		_, err := chapterClient.CreateChapter(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid number")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("empty title", func(t *testing.T) {
		req := &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "",
		}
		_, err := chapterClient.CreateChapter(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestChapterHandler_GetChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get Chapter",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("existing chapter", func(t *testing.T) {
		createResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Get Test Chapter",
		})
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		getResp, err := chapterClient.GetChapter(ctx, &chapterpb.GetChapterRequest{
			Id: createResp.Chapter.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getResp.Chapter.Id != createResp.Chapter.Id {
			t.Errorf("expected ID %s, got %s", createResp.Chapter.Id, getResp.Chapter.Id)
		}
	})

	t.Run("non-existing chapter", func(t *testing.T) {
		_, err := chapterClient.GetChapter(ctx, &chapterpb.GetChapterRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing chapter")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		_, err := chapterClient.GetChapter(ctx, &chapterpb.GetChapterRequest{
			Id: "not-a-uuid",
		})
		if err == nil {
			t.Fatal("expected error for invalid ID")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestChapterHandler_UpdateChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update Chapter",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("successful update", func(t *testing.T) {
		createResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Original Title",
		})

		newTitle := "Updated Title"
		newNumber := int32(2)
		updateResp, err := chapterClient.UpdateChapter(ctx, &chapterpb.UpdateChapterRequest{
			Id:     createResp.Chapter.Id,
			Title:  &newTitle,
			Number: &newNumber,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.Chapter.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", updateResp.Chapter.Title)
		}
		if updateResp.Chapter.Number != 2 {
			t.Errorf("expected number 2, got %d", updateResp.Chapter.Number)
		}
	})

	t.Run("non-existing chapter", func(t *testing.T) {
		newTitle := "Updated Title"
		_, err := chapterClient.UpdateChapter(ctx, &chapterpb.UpdateChapterRequest{
			Id:    uuid.New().String(),
			Title: &newTitle,
		})
		if err == nil {
			t.Fatal("expected error for non-existing chapter")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestChapterHandler_DeleteChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete Chapter",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("successful delete", func(t *testing.T) {
		createResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Delete Test Chapter",
		})

		deleteResp, err := chapterClient.DeleteChapter(ctx, &chapterpb.DeleteChapterRequest{
			Id: createResp.Chapter.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}

		// Verify it's deleted
		_, err = chapterClient.GetChapter(ctx, &chapterpb.GetChapterRequest{
			Id: createResp.Chapter.Id,
		})
		if err == nil {
			t.Fatal("expected error for deleted chapter")
		}
	})

	t.Run("non-existing chapter", func(t *testing.T) {
		_, err := chapterClient.DeleteChapter(ctx, &chapterpb.DeleteChapterRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing chapter")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestChapterHandler_ListChaptersByStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List Chapters",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})

	t.Run("list chapters", func(t *testing.T) {
		// Create multiple chapters
		for i := 1; i <= 3; i++ {
			_, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
				StoryId: storyResp.Story.Id,
				Number:  int32(i),
				Title:   "Chapter " + string(rune('0'+i)),
			})
			if err != nil {
				t.Fatalf("failed to create chapter %d: %v", i, err)
			}
		}

		listResp, err := chapterClient.ListChaptersByStory(ctx, &chapterpb.ListChaptersByStoryRequest{
			StoryId: storyResp.Story.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.Chapters) != 3 {
			t.Errorf("expected 3 chapters, got %d", len(listResp.Chapters))
		}
		if listResp.TotalCount != 3 {
			t.Errorf("expected total_count 3, got %d", listResp.TotalCount)
		}
	})

	t.Run("empty list for new story", func(t *testing.T) {
		newStoryResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Empty Story"})
		listResp, err := chapterClient.ListChaptersByStory(ctx, &chapterpb.ListChaptersByStoryRequest{
			StoryId: newStoryResp.Story.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.Chapters) != 0 {
			t.Errorf("expected 0 chapters, got %d", len(listResp.Chapters))
		}
	})
}


