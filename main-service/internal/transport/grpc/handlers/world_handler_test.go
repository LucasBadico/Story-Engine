//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestWorldHandler_CreateWorld(t *testing.T) {
	conn, cleanup := setupTestServerWithWorld(t)
	defer cleanup()

	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation with tenant_id in metadata", func(t *testing.T) {
		// First create a tenant
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for World",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create world with tenant_id in metadata
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.CreateWorldRequest{
			Name:        "Test World",
			Description: "A test world",
			Genre:       "Fantasy",
		}
		resp, err := worldClient.CreateWorld(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.World.Name != "Test World" {
			t.Errorf("expected name to be 'Test World', got '%s'", resp.World.Name)
		}
		if resp.World.TenantId != tenantResp.Tenant.Id {
			t.Errorf("expected tenant_id %s, got %s", tenantResp.Tenant.Id, resp.World.TenantId)
		}
		if resp.World.Description != "A test world" {
			t.Errorf("expected description 'A test world', got '%s'", resp.World.Description)
		}
		if resp.World.Genre != "Fantasy" {
			t.Errorf("expected genre 'Fantasy', got '%s'", resp.World.Genre)
		}
	})

	t.Run("successful creation with tenant_id in request", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.CreateWorldRequest{
			Name:        "Test World 2",
			Description: "Another test world",
		}
		resp, err := worldClient.CreateWorld(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.World.Name != "Test World 2" {
			t.Errorf("expected name to be 'Test World 2', got '%s'", resp.World.Name)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant 3",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.CreateWorldRequest{
			Name: "",
		}
		_, err = worldClient.CreateWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("invalid tenant_id", func(t *testing.T) {
		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", "not-a-uuid")
		req := &worldpb.CreateWorldRequest{
			Name: "Test World",
		}
		_, err := worldClient.CreateWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid tenant_id")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestWorldHandler_GetWorld(t *testing.T) {
	conn, cleanup := setupTestServerWithWorld(t)
	defer cleanup()

	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing world", func(t *testing.T) {
		// Create tenant and world
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Get Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Get world
		getReq := &worldpb.GetWorldRequest{
			Id: createResp.World.Id,
		}
		getResp, err := worldClient.GetWorld(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.World.Id != createResp.World.Id {
			t.Errorf("expected ID %s, got %s", createResp.World.Id, getResp.World.Id)
		}
		if getResp.World.Name != "Get Test World" {
			t.Errorf("expected name 'Get Test World', got '%s'", getResp.World.Name)
		}
	})

	t.Run("non-existing world", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.GetWorldRequest{
			Id: uuid.New().String(),
		}
		_, err = worldClient.GetWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing world")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid world ID", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Invalid ID Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.GetWorldRequest{
			Id: "not-a-uuid",
		}
		_, err = worldClient.GetWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for invalid ID")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestWorldHandler_ListWorlds(t *testing.T) {
	conn, cleanup := setupTestServerWithWorld(t)
	defer cleanup()

	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list worlds for tenant", func(t *testing.T) {
		// Create tenant
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		// Create multiple worlds
		world1, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "World 1",
		})
		if err != nil {
			t.Fatalf("failed to create world 1: %v", err)
		}

		world2, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "World 2",
		})
		if err != nil {
			t.Fatalf("failed to create world 2: %v", err)
		}

		// List worlds
		listReq := &worldpb.ListWorldsRequest{
			TenantId: tenantResp.Tenant.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := worldClient.ListWorlds(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 worlds, got %d", listResp.TotalCount)
		}

		foundWorld1 := false
		foundWorld2 := false
		for _, w := range listResp.Worlds {
			if w.Id == world1.World.Id {
				foundWorld1 = true
			}
			if w.Id == world2.World.Id {
				foundWorld2 = true
			}
		}

		if !foundWorld1 {
			t.Error("world 1 not found in list")
		}
		if !foundWorld2 {
			t.Error("world 2 not found in list")
		}
	})
}

func TestWorldHandler_UpdateWorld(t *testing.T) {
	conn, cleanup := setupTestServerWithWorld(t)
	defer cleanup()

	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update world name", func(t *testing.T) {
		// Create tenant and world
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Original World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Update world
		updateReq := &worldpb.UpdateWorldRequest{
			Id:   createResp.World.Id,
			Name: stringPtr("Updated World"),
		}
		updateResp, err := worldClient.UpdateWorld(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.World.Name != "Updated World" {
			t.Errorf("expected name 'Updated World', got '%s'", updateResp.World.Name)
		}
		if updateResp.World.Id != createResp.World.Id {
			t.Errorf("expected ID %s, got %s", createResp.World.Id, updateResp.World.Id)
		}
	})

	t.Run("non-existing world", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.UpdateWorldRequest{
			Id:   uuid.New().String(),
			Name: stringPtr("Non-existent"),
		}
		_, err = worldClient.UpdateWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing world")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestWorldHandler_DeleteWorld(t *testing.T) {
	conn, cleanup := setupTestServerWithWorld(t)
	defer cleanup()

	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete existing world", func(t *testing.T) {
		// Create tenant and world
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "World to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Delete world
		deleteReq := &worldpb.DeleteWorldRequest{
			Id: createResp.World.Id,
		}
		_, err = worldClient.DeleteWorld(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify world is deleted
		getReq := &worldpb.GetWorldRequest{
			Id: createResp.World.Id,
		}
		_, err = worldClient.GetWorld(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted world")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("delete non-existing world", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &worldpb.DeleteWorldRequest{
			Id: uuid.New().String(),
		}
		_, err = worldClient.DeleteWorld(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing world")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with world handler
func setupTestServerWithWorld(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	// Initialize use cases
	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)

	// Create handlers
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)

	// Use the testing package's SetupTestServerWithHandlers
	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler: tenantHandler,
		WorldHandler:  worldHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

func stringPtr(s string) *string {
	return &s
}

