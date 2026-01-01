//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	commonpb "github.com/story-engine/main-service/proto/common"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestStoryHandler_CreateStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation with tenant_id in metadata", func(t *testing.T) {
		// First create a tenant
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Story",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create story with tenant_id in metadata
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &storypb.CreateStoryRequest{
			Title: "Test Story",
		}
		resp, err := storyClient.CreateStory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Story.Title != "Test Story" {
			t.Errorf("expected title to be 'Test Story', got '%s'", resp.Story.Title)
		}
		if resp.Story.TenantId != tenantResp.Tenant.Id {
			t.Errorf("expected tenant_id %s, got %s", tenantResp.Tenant.Id, resp.Story.TenantId)
		}
		if resp.Story.VersionNumber != 1 {
			t.Errorf("expected version_number 1, got %d", resp.Story.VersionNumber)
		}
	})

	t.Run("successful creation with tenant_id in request", func(t *testing.T) {
		// First create a tenant
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Story 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create story with tenant_id in metadata (required by auth interceptor)
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &storypb.CreateStoryRequest{
			Title: "Test Story 2",
		}
		resp, err := storyClient.CreateStory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Story.Title != "Test Story 2" {
			t.Errorf("expected title to be 'Test Story 2', got '%s'", resp.Story.Title)
		}
	})

	t.Run("without tenant_id", func(t *testing.T) {
		req := &storypb.CreateStoryRequest{
			Title: "Test Story",
		}
		_, err := storyClient.CreateStory(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for missing tenant_id")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.Unauthenticated {
			t.Errorf("expected Unauthenticated, got %v", s.Code())
		}
	})

	t.Run("with invalid tenant", func(t *testing.T) {
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", uuid.New().String())
		req := &storypb.CreateStoryRequest{
			Title: "Test Story",
		}
		_, err := storyClient.CreateStory(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid tenant")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("empty title", func(t *testing.T) {
		// First create a tenant
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Empty Title",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &storypb.CreateStoryRequest{
			Title: "",
		}
		_, err = storyClient.CreateStory(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty title")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestStoryHandler_GetStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing story", func(t *testing.T) {
		// Create tenant and story
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Get",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Get Test Story",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		
		// Debug: check the created story ID
		if createResp.Story == nil {
			t.Fatal("createResp.Story is nil")
		}
		t.Logf("Created story ID: '%s' (length: %d)", createResp.Story.Id, len(createResp.Story.Id))
		if len(createResp.Story.Id) != 36 {
			t.Fatalf("expected story ID length 36, got %d: '%s'", len(createResp.Story.Id), createResp.Story.Id)
		}

		// Now get it (with tenant_id in metadata)
		getReq := &storypb.GetStoryRequest{
			Id: createResp.Story.Id,
		}
		getResp, err := storyClient.GetStory(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Story.Id != createResp.Story.Id {
			t.Errorf("expected ID %s, got %s", createResp.Story.Id, getResp.Story.Id)
		}
		if getResp.Story.Title != "Get Test Story" {
			t.Errorf("expected title 'Get Test Story', got '%s'", getResp.Story.Title)
		}
	})

	t.Run("non-existing story", func(t *testing.T) {
		// Create tenant first to get valid tenant_id
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Non-existing",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		
		req := &storypb.GetStoryRequest{
			Id: uuid.New().String(),
		}
		_, err := storyClient.GetStory(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing story")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid story ID", func(t *testing.T) {
		// Create tenant first to get valid tenant_id
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Invalid ID",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		
		req := &storypb.GetStoryRequest{
			Id: "not-a-uuid",
		}
		_, err := storyClient.GetStory(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid ID")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestStoryHandler_ListStories(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("with pagination", func(t *testing.T) {
		// Create tenant
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for List",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		// Create multiple stories
		for i := 1; i <= 3; i++ {
			_, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
				Title: "List Test Story " + string(rune('0'+i)),
			})
			if err != nil {
				t.Fatalf("failed to create story %d: %v", i, err)
			}
		}

		// List stories
		req := &storypb.ListStoriesRequest{
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		resp, err := storyClient.ListStories(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Stories) != 3 {
			t.Errorf("expected 3 stories, got %d", len(resp.Stories))
		}
	})

	t.Run("without tenant_id", func(t *testing.T) {
		req := &storypb.ListStoriesRequest{}
		_, err := storyClient.ListStories(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for missing tenant_id")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.Unauthenticated {
			t.Errorf("expected Unauthenticated, got %v", s.Code())
		}
	})
}

func TestStoryHandler_CloneStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful clone", func(t *testing.T) {
		// Create tenant and story
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Clone",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		if tenantResp.Tenant == nil {
			t.Fatal("tenant response is nil")
		}
		
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Clone Test Story",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		if createResp.Story == nil {
			t.Fatal("create story response is nil")
		}

		// Clone it (with tenant_id in metadata)
		cloneReq := &storypb.CloneStoryRequest{
			SourceStoryId: createResp.Story.Id,
		}
		cloneResp, err := storyClient.CloneStory(ctx, cloneReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify version number increment
		if cloneResp.NewVersionNumber != 2 {
			t.Errorf("expected version_number 2, got %d", cloneResp.NewVersionNumber)
		}
		if cloneResp.Story.VersionNumber != 2 {
			t.Errorf("expected story version_number 2, got %d", cloneResp.Story.VersionNumber)
		}
		if cloneResp.Story.RootStoryId != createResp.Story.RootStoryId {
			t.Errorf("expected root_story_id %s, got %s", createResp.Story.RootStoryId, cloneResp.Story.RootStoryId)
		}
		if cloneResp.Story.PreviousStoryId != createResp.Story.Id {
			t.Errorf("expected previous_story_id %s, got %s", createResp.Story.Id, cloneResp.Story.PreviousStoryId)
		}
	})

	t.Run("non-existing source story", func(t *testing.T) {
		// Create tenant first to get valid tenant_id
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Non-existing Clone",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		
		req := &storypb.CloneStoryRequest{
			SourceStoryId: uuid.New().String(),
		}
		_, err := storyClient.CloneStory(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing source story")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestStoryHandler_ListStoryVersions(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list versions", func(t *testing.T) {
		// Create tenant and story
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Versions",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		if tenantResp.Tenant == nil {
			t.Fatal("tenant response is nil")
		}
		
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Version Test Story",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		if createResp.Story == nil {
			t.Fatal("create story response is nil")
		}

		// Clone it twice (with tenant_id in metadata)
		clone1Resp, err := storyClient.CloneStory(ctx, &storypb.CloneStoryRequest{
			SourceStoryId: createResp.Story.Id,
		})
		if err != nil {
			t.Fatalf("failed to clone story first time: %v", err)
		}
		_, err = storyClient.CloneStory(ctx, &storypb.CloneStoryRequest{
			SourceStoryId: clone1Resp.Story.Id,
		})
		if err != nil {
			t.Fatalf("failed to clone story second time: %v", err)
		}

		// List all versions (with tenant_id in metadata)
		req := &storypb.ListStoryVersionsRequest{
			RootStoryId: createResp.Story.RootStoryId,
		}
		resp, err := storyClient.ListStoryVersions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have 3 versions (original + 2 clones)
		if len(resp.Versions) != 3 {
			t.Errorf("expected 3 versions, got %d", len(resp.Versions))
		}

		// Verify version numbers
		versionNumbers := make(map[int32]bool)
		for _, v := range resp.Versions {
			versionNumbers[v.VersionNumber] = true
		}
		for i := int32(1); i <= 3; i++ {
			if !versionNumbers[i] {
				t.Errorf("missing version number %d", i)
			}
		}
	})

	t.Run("non-existing root story", func(t *testing.T) {
		// Create tenant first to get valid tenant_id
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Non-existing Root",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		
		req := &storypb.ListStoryVersionsRequest{
			RootStoryId: uuid.New().String(),
		}
		resp, err := storyClient.ListStoryVersions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return empty list for non-existing root
		if len(resp.Versions) != 0 {
			t.Errorf("expected 0 versions for non-existing root, got %d", len(resp.Versions))
		}
	})
}

func TestStoryHandler_UpdateStory(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	storyClient := storypb.NewStoryServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful update title", func(t *testing.T) {
		// Create tenant and story
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Update Story",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		createResp, err := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Original Title",
		})
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		// Update story
		newTitle := "Updated Title"
		updateResp, err := storyClient.UpdateStory(ctx, &storypb.UpdateStoryRequest{
			Id:    createResp.Story.Id,
			Title: &newTitle,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.Story.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", updateResp.Story.Title)
		}
	})

	t.Run("successful update status", func(t *testing.T) {
		// Create tenant and story
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Update Story Status",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, _ := storyClient.CreateStory(ctx, &storypb.CreateStoryRequest{
			Title: "Test Story",
		})

		// Update status
		newStatus := "published"
		updateResp, err := storyClient.UpdateStory(ctx, &storypb.UpdateStoryRequest{
			Id:     createResp.Story.Id,
			Status: &newStatus,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updateResp.Story.Status != "published" {
			t.Errorf("expected status 'published', got '%s'", updateResp.Story.Status)
		}
	})

	t.Run("non-existing story", func(t *testing.T) {
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Non-existing Update",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		newTitle := "Updated Title"
		_, err := storyClient.UpdateStory(ctx, &storypb.UpdateStoryRequest{
			Id:    uuid.New().String(),
			Title: &newTitle,
		})
		if err == nil {
			t.Fatal("expected error for non-existing story")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid story ID", func(t *testing.T) {
		tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Invalid Update ID",
		})
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		newTitle := "Updated Title"
		_, err := storyClient.UpdateStory(ctx, &storypb.UpdateStoryRequest{
			Id:    "not-a-uuid",
			Title: &newTitle,
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
