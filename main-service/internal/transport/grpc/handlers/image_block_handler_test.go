//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/application/story"
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	"github.com/story-engine/main-service/internal/platform/logger"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	imageblockpb "github.com/story-engine/main-service/proto/image_block"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
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
	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	storyHandler := NewStoryHandler(createStoryUseCase, story.NewCloneStoryUseCase(storyRepo, chapterRepo, postgres.NewSceneRepository(db), postgres.NewBeatRepository(db), postgres.NewProseBlockRepository(db), auditLogRepo, postgres.NewTransactionRepository(db), log), story.NewGetStoryVersionGraphUseCase(storyRepo, log), storyRepo, log)
	chapterHandler := NewChapterHandler(chapterRepo, storyRepo, log)
	imageBlockHandler := NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		StoryHandler:     storyHandler,
		ChapterHandler:  chapterHandler,
		ImageBlockHandler: imageBlockHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

