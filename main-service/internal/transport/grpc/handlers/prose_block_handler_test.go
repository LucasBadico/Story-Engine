//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	prosepb "github.com/story-engine/main-service/proto/prose"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Helper functions for pointer conversions
func ptrString(s string) *string { return &s }
func ptrInt32(i int32) *int32    { return &i }

func TestProseBlockHandler_CreateProseBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for ProseBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("successful creation", func(t *testing.T) {
		req := &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Kind:      "final",
			Content:   "This is the prose content for the chapter.",
		}
		resp, err := proseClient.CreateProseBlock(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ProseBlock.Kind != "final" {
			t.Errorf("expected kind 'final', got '%s'", resp.ProseBlock.Kind)
		}
		if resp.ProseBlock.Content != "This is the prose content for the chapter." {
			t.Errorf("unexpected content: %s", resp.ProseBlock.Content)
		}
		if resp.ProseBlock.WordCount != 8 {
			t.Errorf("expected word_count 8, got %d", resp.ProseBlock.WordCount)
		}
	})

	t.Run("default kind", func(t *testing.T) {
		req := &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(2),
			Kind:      "", // should default to "final"
			Content:   "Some content",
		}
		resp, err := proseClient.CreateProseBlock(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ProseBlock.Kind != "final" {
			t.Errorf("expected default kind 'final', got '%s'", resp.ProseBlock.Kind)
		}
	})

	t.Run("invalid chapter_id", func(t *testing.T) {
		req := &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(uuid.New().String()),
			OrderNum:  ptrInt32(1),
			Kind:      "final",
			Content:   "Content",
		}
		_, err := proseClient.CreateProseBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid chapter_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("empty content", func(t *testing.T) {
		req := &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Kind:      "final",
			Content:   "",
		}
		_, err := proseClient.CreateProseBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty content")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("invalid order_num", func(t *testing.T) {
		req := &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(0),
			Kind:      "final",
			Content:   "Content",
		}
		_, err := proseClient.CreateProseBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid order_num")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestProseBlockHandler_GetProseBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Get ProseBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("existing prose block", func(t *testing.T) {
		createResp, _ := proseClient.CreateProseBlock(ctx, &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Kind:      "final",
			Content:   "Test content",
		})

		getResp, err := proseClient.GetProseBlock(ctx, &prosepb.GetProseBlockRequest{
			Id: createResp.ProseBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getResp.ProseBlock.Id != createResp.ProseBlock.Id {
			t.Errorf("expected ID %s, got %s", createResp.ProseBlock.Id, getResp.ProseBlock.Id)
		}
	})

	t.Run("non-existing prose block", func(t *testing.T) {
		_, err := proseClient.GetProseBlock(ctx, &prosepb.GetProseBlockRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing prose block")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestProseBlockHandler_UpdateProseBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update ProseBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("successful update", func(t *testing.T) {
		createResp, _ := proseClient.CreateProseBlock(ctx, &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Kind:      "draft",
			Content:   "Original content",
		})

		newContent := "Updated content with more words"
		newKind := "final"
		updateResp, err := proseClient.UpdateProseBlock(ctx, &prosepb.UpdateProseBlockRequest{
			Id:      createResp.ProseBlock.Id,
			Content: &newContent,
			Kind:    &newKind,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.ProseBlock.Content != "Updated content with more words" {
			t.Errorf("expected updated content, got '%s'", updateResp.ProseBlock.Content)
		}
		if updateResp.ProseBlock.Kind != "final" {
			t.Errorf("expected kind 'final', got '%s'", updateResp.ProseBlock.Kind)
		}
		if updateResp.ProseBlock.WordCount != 5 {
			t.Errorf("expected word_count 5, got %d", updateResp.ProseBlock.WordCount)
		}
	})
}

func TestProseBlockHandler_DeleteProseBlock(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete ProseBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("successful delete", func(t *testing.T) {
		createResp, _ := proseClient.CreateProseBlock(ctx, &prosepb.CreateProseBlockRequest{
			ChapterId: ptrString(chapterResp.Chapter.Id),
			OrderNum:  ptrInt32(1),
			Kind:      "final",
			Content:   "Content to delete",
		})

		deleteResp, err := proseClient.DeleteProseBlock(ctx, &prosepb.DeleteProseBlockRequest{
			Id: createResp.ProseBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})
}

func TestProseBlockHandler_ListProseBlocksByChapter(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List ProseBlocks",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})

	t.Run("list prose blocks by chapter", func(t *testing.T) {
		// Create multiple prose blocks
		for i := 1; i <= 3; i++ {
			_, err := proseClient.CreateProseBlock(ctx, &prosepb.CreateProseBlockRequest{
				ChapterId: ptrString(chapterResp.Chapter.Id),
				OrderNum:  ptrInt32(int32(i)),
				Kind:      "final",
				Content:   "Prose block content",
			})
			if err != nil {
				t.Fatalf("failed to create prose block %d: %v", i, err)
			}
		}

		listResp, err := proseClient.ListProseBlocksByChapter(ctx, &prosepb.ListProseBlocksByChapterRequest{
			ChapterId: chapterResp.Chapter.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.ProseBlocks) != 3 {
			t.Errorf("expected 3 prose blocks, got %d", len(listResp.ProseBlocks))
		}
	})
}

func TestProseBlockReferenceHandler_CreateAndList(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	proseClient := prosepb.NewProseBlockServiceClient(conn)
	refClient := prosepb.NewProseBlockReferenceServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for ProseBlockReference",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Chapter One",
	})
	proseResp, _ := proseClient.CreateProseBlock(ctx, &prosepb.CreateProseBlockRequest{
		ChapterId: ptrString(chapterResp.Chapter.Id),
		OrderNum:  ptrInt32(1),
		Kind:      "final",
		Content:   "Test prose content",
	})

	t.Run("create and list references", func(t *testing.T) {
		// Create a reference (using a fake scene ID for testing)
		fakeSceneID := uuid.New().String()
		createResp, err := refClient.CreateProseBlockReference(ctx, &prosepb.CreateProseBlockReferenceRequest{
			ProseBlockId: proseResp.ProseBlock.Id,
			EntityType:   "scene",
			EntityId:     fakeSceneID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if createResp.Reference.EntityType != "scene" {
			t.Errorf("expected entity_type 'scene', got '%s'", createResp.Reference.EntityType)
		}

		// List references by prose block
		listResp, err := refClient.ListProseBlockReferencesByProseBlock(ctx, &prosepb.ListProseBlockReferencesByProseBlockRequest{
			ProseBlockId: proseResp.ProseBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.References) != 1 {
			t.Errorf("expected 1 reference, got %d", len(listResp.References))
		}

		// Delete reference
		deleteResp, err := refClient.DeleteProseBlockReference(ctx, &prosepb.DeleteProseBlockReferenceRequest{
			Id: createResp.Reference.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleteResp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("invalid prose block for reference", func(t *testing.T) {
		_, err := refClient.CreateProseBlockReference(ctx, &prosepb.CreateProseBlockReferenceRequest{
			ProseBlockId: uuid.New().String(),
			EntityType:   "scene",
			EntityId:     uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for invalid prose_block_id")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}


