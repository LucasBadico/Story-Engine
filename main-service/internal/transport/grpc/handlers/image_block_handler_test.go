//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	imageblockpb "github.com/story-engine/main-service/proto/image_block"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestImageBlockHandler_CreateImageBlock(t *testing.T) {
	conn, cleanup := setupTestServerWithImageBlock(t)
	defer cleanup()

	imageBlockClient := imageblockpb.NewImageBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Image Block",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Test Story",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Test Chapter",
		})
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		req := &imageblockpb.CreateImageBlockRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
			Kind:      "final",
			ImageUrl:  "https://example.com/image.jpg",
			AltText:   stringPtr("Test image"),
		}
		resp, err := imageBlockClient.CreateImageBlock(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ImageBlock.ImageUrl != "https://example.com/image.jpg" {
			t.Errorf("expected image_url 'https://example.com/image.jpg', got '%s'", resp.ImageBlock.ImageUrl)
		}
	})

	t.Run("empty image_url", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Test Story 2",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Test Chapter 2",
		})
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		req := &imageblockpb.CreateImageBlockRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
			Kind:      "final",
			ImageUrl:  "",
		}
		_, err = imageBlockClient.CreateImageBlock(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty image_url")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestImageBlockHandler_GetImageBlock(t *testing.T) {
	conn, cleanup := setupTestServerWithImageBlock(t)
	defer cleanup()

	imageBlockClient := imageblockpb.NewImageBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing image block", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		storyResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Get Test Story",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		chapterResp, err := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  1,
			Title:   "Get Test Chapter",
		})
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		createResp, err := imageBlockClient.CreateImageBlock(ctx, &imageblockpb.CreateImageBlockRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
			Kind:      "final",
			ImageUrl:  "https://example.com/get-image.jpg",
		})
		if err != nil {
			t.Fatalf("failed to create image block: %v", err)
		}

		getReq := &imageblockpb.GetImageBlockRequest{
			Id: createResp.ImageBlock.Id,
		}
		getResp, err := imageBlockClient.GetImageBlock(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.ImageBlock.Id != createResp.ImageBlock.Id {
			t.Errorf("expected ID %s, got %s", createResp.ImageBlock.Id, getResp.ImageBlock.Id)
		}
	})
}

func TestImageBlockHandler_UpdateImageBlock(t *testing.T) {
	conn, cleanup := setupTestServerWithImageBlock(t)
	defer cleanup()

	imageBlockClient := imageblockpb.NewImageBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Update ImageBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Test Chapter",
	})

	t.Run("successful update", func(t *testing.T) {
		createResp, _ := imageBlockClient.CreateImageBlock(ctx, &imageblockpb.CreateImageBlockRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
			Kind:      "final",
			ImageUrl:  "https://example.com/original.jpg",
			AltText:   stringPtr("Original alt text"),
		})

		newImageURL := "https://example.com/updated.jpg"
		newAltText := "Updated alt text"
		newCaption := "Updated caption"
		newWidth := int32(800)
		newHeight := int32(600)

		updateResp, err := imageBlockClient.UpdateImageBlock(ctx, &imageblockpb.UpdateImageBlockRequest{
			Id:       createResp.ImageBlock.Id,
			ImageUrl: &newImageURL,
			AltText:  &newAltText,
			Caption:  &newCaption,
			Width:    &newWidth,
			Height:   &newHeight,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.ImageBlock.ImageUrl != "https://example.com/updated.jpg" {
			t.Errorf("expected image_url 'https://example.com/updated.jpg', got '%s'", updateResp.ImageBlock.ImageUrl)
		}
		if updateResp.ImageBlock.AltText == nil || *updateResp.ImageBlock.AltText != "Updated alt text" {
			t.Errorf("expected alt_text 'Updated alt text', got %v", updateResp.ImageBlock.AltText)
		}
		if updateResp.ImageBlock.Caption == nil || *updateResp.ImageBlock.Caption != "Updated caption" {
			t.Errorf("expected caption 'Updated caption', got %v", updateResp.ImageBlock.Caption)
		}
	})

	t.Run("non-existing image block", func(t *testing.T) {
		newImageURL := "https://example.com/updated.jpg"
		_, err := imageBlockClient.UpdateImageBlock(ctx, &imageblockpb.UpdateImageBlockRequest{
			Id:       uuid.New().String(),
			ImageUrl: &newImageURL,
		})
		if err == nil {
			t.Fatal("expected error for non-existing image block")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid image block ID", func(t *testing.T) {
		newImageURL := "https://example.com/updated.jpg"
		_, err := imageBlockClient.UpdateImageBlock(ctx, &imageblockpb.UpdateImageBlockRequest{
			Id:       "not-a-uuid",
			ImageUrl: &newImageURL,
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

func TestImageBlockHandler_DeleteImageBlock(t *testing.T) {
	conn, cleanup := setupTestServerWithImageBlock(t)
	defer cleanup()

	imageBlockClient := imageblockpb.NewImageBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for Delete ImageBlock",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Test Chapter",
	})

	t.Run("successful delete", func(t *testing.T) {
		createResp, _ := imageBlockClient.CreateImageBlock(ctx, &imageblockpb.CreateImageBlockRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
			Kind:      "final",
			ImageUrl:  "https://example.com/delete.jpg",
		})

		_, err := imageBlockClient.DeleteImageBlock(ctx, &imageblockpb.DeleteImageBlockRequest{
			Id: createResp.ImageBlock.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's deleted
		_, err = imageBlockClient.GetImageBlock(ctx, &imageblockpb.GetImageBlockRequest{
			Id: createResp.ImageBlock.Id,
		})
		if err == nil {
			t.Fatal("expected error for deleted image block")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("non-existing image block", func(t *testing.T) {
		_, err := imageBlockClient.DeleteImageBlock(ctx, &imageblockpb.DeleteImageBlockRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing image block")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestImageBlockHandler_ListImageBlocks(t *testing.T) {
	conn, cleanup := setupTestServerWithImageBlock(t)
	defer cleanup()

	imageBlockClient := imageblockpb.NewImageBlockServiceClient(conn)
	chapterClient := chapterpb.NewChapterServiceClient(conn)
	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant for List ImageBlocks",
	})
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	storyResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{Title: "Test Story"})
	chapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
		StoryId: storyResp.Story.Id,
		Number:  1,
		Title:   "Test Chapter",
	})

	t.Run("list image blocks by chapter", func(t *testing.T) {
		// Create multiple image blocks
		for i := 1; i <= 3; i++ {
			_, err := imageBlockClient.CreateImageBlock(ctx, &imageblockpb.CreateImageBlockRequest{
				ChapterId: stringPtr(chapterResp.Chapter.Id),
				Kind:      "final",
				ImageUrl:  "https://example.com/image" + string(rune('0'+i)) + ".jpg",
			})
			if err != nil {
				t.Fatalf("failed to create image block %d: %v", i, err)
			}
		}

		listResp, err := imageBlockClient.ListImageBlocks(ctx, &imageblockpb.ListImageBlocksRequest{
			ChapterId: stringPtr(chapterResp.Chapter.Id),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.ImageBlocks) != 3 {
			t.Errorf("expected 3 image blocks, got %d", len(listResp.ImageBlocks))
		}
		if listResp.TotalCount != 3 {
			t.Errorf("expected total_count 3, got %d", listResp.TotalCount)
		}
	})

	t.Run("empty list for new chapter", func(t *testing.T) {
		newChapterResp, _ := chapterClient.CreateChapter(ctx, &chapterpb.CreateChapterRequest{
			StoryId: storyResp.Story.Id,
			Number:  2,
			Title:   "Empty Chapter",
		})

		listResp, err := imageBlockClient.ListImageBlocks(ctx, &imageblockpb.ListImageBlocksRequest{
			ChapterId: stringPtr(newChapterResp.Chapter.Id),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.ImageBlocks) != 0 {
			t.Errorf("expected 0 image blocks, got %d", len(listResp.ImageBlocks))
		}
	})
}

// Helper function to create a test server with image block handler
func setupTestServerWithImageBlock(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	imageBlockRepo := postgres.NewImageBlockRepository(db)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	worldRepo := postgres.NewWorldRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	cloneStoryUseCase := story.NewCloneStoryUseCase(storyRepo, chapterRepo, postgres.NewSceneRepository(db), postgres.NewBeatRepository(db), postgres.NewProseBlockRepository(db), auditLogRepo, postgres.NewTransactionRepository(db), log)
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)
	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	storyHandler := NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, versionGraphUseCase, log)
	chapterHandler := NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	imageBlockHandler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:     tenantHandler,
		StoryHandler:      storyHandler,
		ChapterHandler:    chapterHandler,
		ImageBlockHandler: imageBlockHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}
