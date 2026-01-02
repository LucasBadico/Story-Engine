//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	traitpb "github.com/story-engine/main-service/proto/trait"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestArchetypeHandler_CreateArchetype(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Archetype",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &archetypepb.CreateArchetypeRequest{
			Name:        "Warrior",
			Description: "A warrior archetype",
		}
		resp, err := archetypeClient.CreateArchetype(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Archetype.Name != "Warrior" {
			t.Errorf("expected name 'Warrior', got '%s'", resp.Archetype.Name)
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
		req := &archetypepb.CreateArchetypeRequest{
			Name: "",
		}
		_, err = archetypeClient.CreateArchetype(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestArchetypeHandler_GetArchetype(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing archetype", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Get Test Archetype",
		})
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		getReq := &archetypepb.GetArchetypeRequest{
			Id: createResp.Archetype.Id,
		}
		getResp, err := archetypeClient.GetArchetype(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Archetype.Id != createResp.Archetype.Id {
			t.Errorf("expected ID %s, got %s", createResp.Archetype.Id, getResp.Archetype.Id)
		}
	})

	t.Run("non-existing archetype", func(t *testing.T) {
		req := &archetypepb.GetArchetypeRequest{
			Id: uuid.New().String(),
		}
		_, err := archetypeClient.GetArchetype(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for non-existing archetype")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestArchetypeHandler_ListArchetypes(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list archetypes", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		archetype1, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Archetype 1",
		})
		if err != nil {
			t.Fatalf("failed to create archetype 1: %v", err)
		}

		archetype2, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Archetype 2",
		})
		if err != nil {
			t.Fatalf("failed to create archetype 2: %v", err)
		}

		listReq := &archetypepb.ListArchetypesRequest{
			TenantId: tenantResp.Tenant.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := archetypeClient.ListArchetypes(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 archetypes, got %d", listResp.TotalCount)
		}

		foundArchetype1 := false
		foundArchetype2 := false
		for _, a := range listResp.Archetypes {
			if a.Id == archetype1.Archetype.Id {
				foundArchetype1 = true
			}
			if a.Id == archetype2.Archetype.Id {
				foundArchetype2 = true
			}
		}

		if !foundArchetype1 {
			t.Error("archetype 1 not found in list")
		}
		if !foundArchetype2 {
			t.Error("archetype 2 not found in list")
		}
	})
}

func TestArchetypeHandler_UpdateArchetype(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update archetype", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Original Archetype",
		})
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		updateReq := &archetypepb.UpdateArchetypeRequest{
			Id:   createResp.Archetype.Id,
			Name: stringPtr("Updated Archetype"),
		}
		updateResp, err := archetypeClient.UpdateArchetype(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Archetype.Name != "Updated Archetype" {
			t.Errorf("expected name 'Updated Archetype', got '%s'", updateResp.Archetype.Name)
		}
	})
}

func TestArchetypeHandler_DeleteArchetype(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete archetype", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Archetype to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		deleteReq := &archetypepb.DeleteArchetypeRequest{
			Id: createResp.Archetype.Id,
		}
		_, err = archetypeClient.DeleteArchetype(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &archetypepb.GetArchetypeRequest{
			Id: createResp.Archetype.Id,
		}
		_, err = archetypeClient.GetArchetype(context.Background(), getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted archetype")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestArchetypeHandler_AddTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("add trait to archetype", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Add Trait Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		archetypeResp, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Warrior",
		})
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		traitResp, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Brave",
		})
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		addReq := &archetypepb.AddTraitToArchetypeRequest{
			ArchetypeId: archetypeResp.Archetype.Id,
			TraitId:     traitResp.Trait.Id,
		}
		_, err = archetypeClient.AddTraitToArchetype(ctx, addReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestArchetypeHandler_RemoveTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithArchetype(t)
	defer cleanup()

	archetypeClient := archetypepb.NewArchetypeServiceClient(conn)
	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("remove trait from archetype", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Remove Trait Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		archetypeResp, err := archetypeClient.CreateArchetype(ctx, &archetypepb.CreateArchetypeRequest{
			Name: "Warrior",
		})
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		traitResp, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Brave",
		})
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		// Add trait first
		_, err = archetypeClient.AddTraitToArchetype(ctx, &archetypepb.AddTraitToArchetypeRequest{
			ArchetypeId: archetypeResp.Archetype.Id,
			TraitId:     traitResp.Trait.Id,
		})
		if err != nil {
			t.Fatalf("failed to add trait: %v", err)
		}

		// Remove trait
		removeReq := &archetypepb.RemoveTraitFromArchetypeRequest{
			ArchetypeId: archetypeResp.Archetype.Id,
			TraitId:     traitResp.Trait.Id,
		}
		_, err = archetypeClient.RemoveTraitFromArchetype(ctx, removeReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// Helper function to create a test server with archetype handler
func setupTestServerWithArchetype(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitToArchetypeUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitFromArchetypeUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, traitapp.NewListTraitsUseCase(traitRepo, log), traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log), traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log), log)
	archetypeHandler := NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		TraitHandler:    traitHandler,
		ArchetypeHandler: archetypeHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

