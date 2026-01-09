//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	locationpb "github.com/story-engine/main-service/proto/location"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLocationHandler_CreateLocation(t *testing.T) {
	conn, cleanup := setupTestServerWithLocation(t)
	defer cleanup()

	locationClient := locationpb.NewLocationServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Location",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		req := &locationpb.CreateLocationRequest{
			WorldId:     worldResp.World.Id,
			Name:        "Test Location",
			Description: "A test location",
		}
		resp, err := locationClient.CreateLocation(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Location.Name != "Test Location" {
			t.Errorf("expected name 'Test Location', got '%s'", resp.Location.Name)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant 2",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Test World 2",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		req := &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "",
		}
		_, err = locationClient.CreateLocation(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestLocationHandler_GetLocation(t *testing.T) {
	conn, cleanup := setupTestServerWithLocation(t)
	defer cleanup()

	locationClient := locationpb.NewLocationServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("existing location", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Get Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		createResp, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Get Test Location",
		})
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		getReq := &locationpb.GetLocationRequest{
			Id: createResp.Location.Id,
		}
		getResp, err := locationClient.GetLocation(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Location.Id != createResp.Location.Id {
			t.Errorf("expected ID %s, got %s", createResp.Location.Id, getResp.Location.Id)
		}
	})

	t.Run("non-existing location", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &locationpb.GetLocationRequest{
			Id: uuid.New().String(),
		}
		_, err = locationClient.GetLocation(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing location")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestLocationHandler_ListLocations(t *testing.T) {
	conn, cleanup := setupTestServerWithLocation(t)
	defer cleanup()

	locationClient := locationpb.NewLocationServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("list locations", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "List Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		location1, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Location 1",
		})
		if err != nil {
			t.Fatalf("failed to create location 1: %v", err)
		}

		location2, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Location 2",
		})
		if err != nil {
			t.Fatalf("failed to create location 2: %v", err)
		}

		listReq := &locationpb.ListLocationsRequest{
			WorldId: worldResp.World.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := locationClient.ListLocations(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 locations, got %d", listResp.TotalCount)
		}

		foundLocation1 := false
		foundLocation2 := false
		for _, l := range listResp.Locations {
			if l.Id == location1.Location.Id {
				foundLocation1 = true
			}
			if l.Id == location2.Location.Id {
				foundLocation2 = true
			}
		}

		if !foundLocation1 {
			t.Error("location 1 not found in list")
		}
		if !foundLocation2 {
			t.Error("location 2 not found in list")
		}
	})
}

func TestLocationHandler_UpdateLocation(t *testing.T) {
	conn, cleanup := setupTestServerWithLocation(t)
	defer cleanup()

	locationClient := locationpb.NewLocationServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("update location", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Update Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		createResp, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Original Location",
		})
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		updateReq := &locationpb.UpdateLocationRequest{
			Id:   createResp.Location.Id,
			Name: stringPtr("Updated Location"),
		}
		updateResp, err := locationClient.UpdateLocation(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Location.Name != "Updated Location" {
			t.Errorf("expected name 'Updated Location', got '%s'", updateResp.Location.Name)
		}
	})
}

func TestLocationHandler_DeleteLocation(t *testing.T) {
	conn, cleanup := setupTestServerWithLocation(t)
	defer cleanup()

	locationClient := locationpb.NewLocationServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("delete location", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Delete Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		createResp, err := locationClient.CreateLocation(ctx, &locationpb.CreateLocationRequest{
			WorldId: worldResp.World.Id,
			Name:    "Location to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		deleteReq := &locationpb.DeleteLocationRequest{
			Id: createResp.Location.Id,
		}
		_, err = locationClient.DeleteLocation(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &locationpb.GetLocationRequest{
			Id: createResp.Location.Id,
		}
		_, err = locationClient.GetLocation(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted location")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with location handler
func setupTestServerWithLocation(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:   tenantHandler,
		WorldHandler:    worldHandler,
		LocationHandler: locationHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

