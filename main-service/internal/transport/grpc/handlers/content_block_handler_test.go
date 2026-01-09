//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Helper functions for pointer conversions
func ptrString(s string) *string { return &s }
func ptrInt32(i int32) *int32    { return &i }

func TestContentBlockHandler_CreateContentBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for ContentBlock",
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

	t.Run("successful creation", func(t *testing.T) {
		req := &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "final",
			Content:   "This is the prose content for the chapter.",
		}
		resp, err := contentClient.CreateContentBlock(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ContentBlock.Kind != "final" {
			t.Errorf("expected kind 'final', got '%s'", resp.ContentBlock.Kind)
		}
		if resp.ContentBlock.Content != "This is the prose content for the chapter." {
			t.Errorf("unexpected content: %s", resp.ContentBlock.Content)
		}
		if resp.ContentBlock.Type != "text" {
			t.Errorf("expected type 'text', got '%s'", resp.ContentBlock.Type)
		}
	})

	t.Run("default kind", func(t *testing.T) {
		req := &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(2),
			Type:      "text",
			Kind:      "", // should default to "final"
			Content:   "Some content",
		}
		resp, err := contentClient.CreateContentBlock(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ContentBlock.Kind != "final" {
			t.Errorf("expected default kind 'final', got '%s'", resp.ContentBlock.Kind)
		}
	})

	t.Run("invalid chapter_id", func(t *testing.T) {
		req := &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(uuid.New().String()),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "final",
			Content:   "Content",
		}
		_, err := contentClient.CreateContentBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid chapter_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("empty content", func(t *testing.T) {
		req := &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "final",
			Content:   "",
		}
		_, err := contentClient.CreateContentBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty content")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("invalid order_num", func(t *testing.T) {
		req := &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(0),
			Type:      "text",
			Kind:      "final",
			Content:   "Content",
		}
		_, err := contentClient.CreateContentBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid order_num")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestContentBlockHandler_GetContentBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get ContentBlock",
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

	t.Run("existing content block", func(t *testing.T) {
		createResp, err := contentClient.CreateContentBlock(ctx, &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "final",
			Content:   "Test content",
		})
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		getResp, err := contentClient.GetContentBlock(ctx, &contentblockpb.GetContentBlockRequest{
			Id: createResp.ContentBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getResp.ContentBlock.Id != createResp.ContentBlock.Id {
			t.Errorf("expected ID %s, got %s", createResp.ContentBlock.Id, getResp.ContentBlock.Id)
		}
	})

	t.Run("non-existing content block", func(t *testing.T) {
		_, err := contentClient.GetContentBlock(ctx, &contentblockpb.GetContentBlockRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing content block")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestContentBlockHandler_UpdateContentBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update ContentBlock",
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

	t.Run("successful update", func(t *testing.T) {
		createResp, err := contentClient.CreateContentBlock(ctx, &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "draft",
			Content:   "Original content",
		})
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		newContent := "Updated content with more words"
		newKind := "final"
		updateResp, err := contentClient.UpdateContentBlock(ctx, &contentblockpb.UpdateContentBlockRequest{
			Id:      createResp.ContentBlock.Id,
			Content: &newContent,
			Kind:    &newKind,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.ContentBlock.Content != "Updated content with more words" {
			t.Errorf("expected updated content, got '%s'", updateResp.ContentBlock.Content)
		}
		if updateResp.ContentBlock.Kind != "final" {
			t.Errorf("expected kind 'final', got '%s'", updateResp.ContentBlock.Kind)
		}
		if updateResp.ContentBlock.Type != "text" {
			t.Errorf("expected type 'text', got '%s'", updateResp.ContentBlock.Type)
		}
	})
}

func TestContentBlockHandler_DeleteContentBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete ContentBlock",
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

	t.Run("successful delete", func(t *testing.T) {
		createResp, err := contentClient.CreateContentBlock(ctx, &contentblockpb.CreateContentBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Type:      "text",
			Kind:      "final",
			Content:   "Content to delete",
		})
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		deleteResp, err := contentClient.DeleteContentBlock(ctx, &contentblockpb.DeleteContentBlockRequest{
			Id: createResp.ContentBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})
}

func TestContentBlockHandler_ListContentBlocksByChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List ContentBlocks",
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

	t.Run("list content blocks by chapter", func(t *testing.T) {
		// Create multiple content blocks
		for i := 1; i <= 3; i++ {
			_, err := contentClient.CreateContentBlock(ctx, &contentblockpb.CreateContentBlockRequest{
				ChapterId: ptrString(chapterResp.Chapter.Id),
				OrderNum:  ptrInt32(int32(i)),
				Type:      "text",
				Kind:      "final",
				Content:   "Content block content",
			})
			if err != nil {
				t.Fatalf("failed to create content block %d: %v", i, err)
			}
		}

		listResp, err := contentClient.ListContentBlocksByChapter(ctx, &contentblockpb.ListContentBlocksByChapterRequest{
			ChapterId: chapterResp.Chapter.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.ContentBlocks) != 3 {
			t.Errorf("expected 3 content blocks, got %d", len(listResp.ContentBlocks))
		}
	})
}

func TestContentAnchorHandler_CreateAndList(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	contentClient := contentblockpb.NewContentBlockServiceClient(conn)
	refClient := contentblockpb.NewContentAnchorServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for ContentAnchor",
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
	contentResp, err := contentClient.CreateContentBlock(ctx, &contentblockpb.CreateContentBlockRequest{
		ChapterId: ptrString(chapterResp.Chapter.Id),
		OrderNum:  ptrInt32(1),
		Type:      "text",
		Kind:      "final",
		Content:   "Test content",
	})
	if err != nil {
		t.Fatalf("failed to create content block: %v", err)
	}

	t.Run("create and list anchors", func(t *testing.T) {
		// Create an anchor (using a fake scene ID for testing)
		fakeSceneID := uuid.New().String()
		createResp, err := refClient.CreateContentAnchor(ctx, &contentblockpb.CreateContentAnchorRequest{
			ContentBlockId: contentResp.ContentBlock.Id,
			EntityType:     "scene",
			EntityId:       fakeSceneID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if createResp.Anchor.EntityType != "scene" {
			t.Errorf("expected entity_type 'scene', got '%s'", createResp.Anchor.EntityType)
		}

		// List anchors by content block
		listResp, err := refClient.ListContentAnchorsByContentBlock(ctx, &contentblockpb.ListContentAnchorsByContentBlockRequest{
			ContentBlockId: contentResp.ContentBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.Anchors) != 1 {
			t.Errorf("expected 1 anchor, got %d", len(listResp.Anchors))
		}

		// Delete anchor
		deleteResp, err := refClient.DeleteContentAnchor(ctx, &contentblockpb.DeleteContentAnchorRequest{
			Id: createResp.Anchor.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("invalid content block for anchor", func(t *testing.T) {
		_, err := refClient.CreateContentAnchor(ctx, &contentblockpb.CreateContentAnchorRequest{
			ContentBlockId: uuid.New().String(),
			EntityType:     "scene",
			EntityId:       uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for invalid content_block_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}
