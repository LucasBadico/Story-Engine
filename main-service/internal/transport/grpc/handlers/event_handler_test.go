//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	eventpb "github.com/story-engine/main-service/proto/event"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestEventHandler_CreateEvent(t *testing.T) {
	conn, cleanup := setupTestServerWithEvent(t)
	defer cleanup()

	eventClient := eventpb.NewEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Event",
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

		req := &eventpb.CreateEventRequest{
			WorldId:     worldResp.World.Id,
			Name:        "Test Event",
			Description: stringPtr("A test event"),
		}
		resp, err := eventClient.CreateEvent(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Event.Name != "Test Event" {
			t.Errorf("expected name 'Test Event', got '%s'", resp.Event.Name)
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

		req := &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "",
		}
		_, err = eventClient.CreateEvent(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestEventHandler_GetEvent(t *testing.T) {
	conn, cleanup := setupTestServerWithEvent(t)
	defer cleanup()

	eventClient := eventpb.NewEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("existing event", func(t *testing.T) {
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

		createResp, err := eventClient.CreateEvent(ctx, &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "Get Test Event",
		})
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		getReq := &eventpb.GetEventRequest{
			Id: createResp.Event.Id,
		}
		getResp, err := eventClient.GetEvent(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Event.Id != createResp.Event.Id {
			t.Errorf("expected ID %s, got %s", createResp.Event.Id, getResp.Event.Id)
		}
	})

	t.Run("non-existing event", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &eventpb.GetEventRequest{
			Id: uuid.New().String(),
		}
		_, err = eventClient.GetEvent(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing event")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestEventHandler_ListEvents(t *testing.T) {
	conn, cleanup := setupTestServerWithEvent(t)
	defer cleanup()

	eventClient := eventpb.NewEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("list events", func(t *testing.T) {
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

		event1, err := eventClient.CreateEvent(ctx, &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "Event 1",
		})
		if err != nil {
			t.Fatalf("failed to create event 1: %v", err)
		}

		event2, err := eventClient.CreateEvent(ctx, &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "Event 2",
		})
		if err != nil {
			t.Fatalf("failed to create event 2: %v", err)
		}

		listReq := &eventpb.ListEventsRequest{
			WorldId: worldResp.World.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := eventClient.ListEvents(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 events, got %d", listResp.TotalCount)
		}

		foundEvent1 := false
		foundEvent2 := false
		for _, e := range listResp.Events {
			if e.Id == event1.Event.Id {
				foundEvent1 = true
			}
			if e.Id == event2.Event.Id {
				foundEvent2 = true
			}
		}

		if !foundEvent1 {
			t.Error("event 1 not found in list")
		}
		if !foundEvent2 {
			t.Error("event 2 not found in list")
		}
	})
}

func TestEventHandler_UpdateEvent(t *testing.T) {
	conn, cleanup := setupTestServerWithEvent(t)
	defer cleanup()

	eventClient := eventpb.NewEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("update event", func(t *testing.T) {
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

		createResp, err := eventClient.CreateEvent(ctx, &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "Original Event",
		})
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		updateReq := &eventpb.UpdateEventRequest{
			Id:   createResp.Event.Id,
			Name: stringPtr("Updated Event"),
		}
		updateResp, err := eventClient.UpdateEvent(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Event.Name != "Updated Event" {
			t.Errorf("expected name 'Updated Event', got '%s'", updateResp.Event.Name)
		}
	})
}

func TestEventHandler_DeleteEvent(t *testing.T) {
	conn, cleanup := setupTestServerWithEvent(t)
	defer cleanup()

	eventClient := eventpb.NewEventServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("delete event", func(t *testing.T) {
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

		createResp, err := eventClient.CreateEvent(ctx, &eventpb.CreateEventRequest{
			WorldId: worldResp.World.Id,
			Name:    "Event to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		deleteReq := &eventpb.DeleteEventRequest{
			Id: createResp.Event.Id,
		}
		_, err = eventClient.DeleteEvent(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &eventpb.GetEventRequest{
			Id: createResp.Event.Id,
		}
		_, err = eventClient.GetEvent(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted event")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with event handler
func setupTestServerWithEvent(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	eventCharacterRepo := postgres.NewEventCharacterRepository(db)
	eventLocationRepo := postgres.NewEventLocationRepository(db)
	eventArtifactRepo := postgres.NewEventArtifactRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	createEventUseCase := eventapp.NewCreateEventUseCase(eventRepo, worldRepo, auditLogRepo, log)
	getEventUseCase := eventapp.NewGetEventUseCase(eventRepo, log)
	listEventsUseCase := eventapp.NewListEventsUseCase(eventRepo, log)
	updateEventUseCase := eventapp.NewUpdateEventUseCase(eventRepo, auditLogRepo, log)
	deleteEventUseCase := eventapp.NewDeleteEventUseCase(eventRepo, eventCharacterRepo, eventLocationRepo, eventArtifactRepo, auditLogRepo, log)
	characterRepo := postgres.NewCharacterRepository(db)
	addCharacterToEventUseCase := eventapp.NewAddCharacterToEventUseCase(eventRepo, characterRepo, eventCharacterRepo, log)
	removeCharacterFromEventUseCase := eventapp.NewRemoveCharacterFromEventUseCase(eventCharacterRepo, log)
	getEventCharactersUseCase := eventapp.NewGetEventCharactersUseCase(eventCharacterRepo, log)
	locationRepo := postgres.NewLocationRepository(db)
	addLocationToEventUseCase := eventapp.NewAddLocationToEventUseCase(eventRepo, locationRepo, eventLocationRepo, log)
	removeLocationFromEventUseCase := eventapp.NewRemoveLocationFromEventUseCase(eventLocationRepo, log)
	getEventLocationsUseCase := eventapp.NewGetEventLocationsUseCase(eventLocationRepo, log)
	artifactRepo := postgres.NewArtifactRepository(db)
	addArtifactToEventUseCase := eventapp.NewAddArtifactToEventUseCase(eventRepo, artifactRepo, eventArtifactRepo, log)
	removeArtifactFromEventUseCase := eventapp.NewRemoveArtifactFromEventUseCase(eventArtifactRepo, log)
	getEventArtifactsUseCase := eventapp.NewGetEventArtifactsUseCase(eventArtifactRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	eventHandler := NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterToEventUseCase, removeCharacterFromEventUseCase, getEventCharactersUseCase, addLocationToEventUseCase, removeLocationFromEventUseCase, getEventLocationsUseCase, addArtifactToEventUseCase, removeArtifactFromEventUseCase, getEventArtifactsUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler: tenantHandler,
		WorldHandler:  worldHandler,
		EventHandler:  eventHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

