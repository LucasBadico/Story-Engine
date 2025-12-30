//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Note: These tests depend on generated proto code.
// Run `make proto-gen` before running tests.

func TestStoryHandler_CreateStory(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := storypb.NewStoryServiceClient(conn)

	t.Run("successful creation with tenant_id in metadata", func(t *testing.T) {
		// TODO: After proto generation:
		// // First create a tenant
		// tenantClient := tenantpb.NewTenantServiceClient(conn)
		// tenantResp, _ := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		//     Name: "Test Tenant",
		// })
		// 
		// // Create story with tenant_id in metadata
		// ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		// req := &storypb.CreateStoryRequest{
		//     Title: "Test Story",
		// }
		// resp, err := client.CreateStory(ctx, req)
		// if err != nil {
		//     t.Fatalf("unexpected error: %v", err)
		// }
		// if resp.Story.Title != "Test Story" {
		//     t.Errorf("expected title to be 'Test Story', got '%s'", resp.Story.Title)
		// }
		
		t.Skip("Skipping until proto files are generated")
	})

	t.Run("without tenant_id", func(t *testing.T) {
		// TODO: After proto generation:
		// req := &storypb.CreateStoryRequest{
		//     Title: "Test Story",
		// }
		// _, err := client.CreateStory(context.Background(), req)
		// if err == nil {
		//     t.Fatal("expected error for missing tenant_id")
		// }
		// 
		// s, _ := status.FromError(err)
		// if s.Code() != codes.Unauthenticated {
		//     t.Errorf("expected Unauthenticated, got %v", s.Code())
		// }
		
		t.Skip("Skipping until proto files are generated")
	})
}

func TestStoryHandler_GetStory(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := storypb.NewStoryServiceClient(conn)

	t.Run("existing story", func(t *testing.T) {
		// TODO: Create story first, then get it
		t.Skip("Skipping until proto files are generated")
	})

	t.Run("non-existing story", func(t *testing.T) {
		// TODO: After proto generation:
		// req := &storypb.GetStoryRequest{
		//     Id: uuid.New().String(),
		// }
		// _, err := client.GetStory(context.Background(), req)
		// if err == nil {
		//     t.Fatal("expected error for non-existing story")
		// }
		// 
		// s, _ := status.FromError(err)
		// if s.Code() != codes.NotFound {
		//     t.Errorf("expected NotFound, got %v", s.Code())
		// }
		
		t.Skip("Skipping until proto files are generated")
	})
}

func TestStoryHandler_ListStories(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := storypb.NewStoryServiceClient(conn)

	t.Run("with pagination", func(t *testing.T) {
		// TODO: Create multiple stories, then list with pagination
		t.Skip("Skipping until proto files are generated")
	})
}

func TestStoryHandler_CloneStory(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := storypb.NewStoryServiceClient(conn)

	t.Run("successful clone", func(t *testing.T) {
		// TODO: Create story, clone it, verify version number increment
		t.Skip("Skipping until proto files are generated")
	})
}

func TestStoryHandler_ListStoryVersions(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := storypb.NewStoryServiceClient(conn)

	t.Run("list versions", func(t *testing.T) {
		// TODO: Create story, clone it multiple times, list all versions
		t.Skip("Skipping until proto files are generated")
	})
}

