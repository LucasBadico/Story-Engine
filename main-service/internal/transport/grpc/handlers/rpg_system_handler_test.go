//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRPGSystemHandler_CreateRPGSystem(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGSystem(t)
	defer cleanup()

	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for RPG System",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10, "dexterity": 8}`)
		req := &rpgsystempb.CreateRPGSystemRequest{
			Name:            "D&D 5e",
			Description:     stringPtr("Dungeons & Dragons 5th Edition"),
			BaseStatsSchema: string(baseStatsSchema),
		}
		resp, err := rpgSystemClient.CreateRPGSystem(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.RpgSystem.Name != "D&D 5e" {
			t.Errorf("expected name 'D&D 5e', got '%s'", resp.RpgSystem.Name)
		}
		if resp.RpgSystem.TenantId == nil || *resp.RpgSystem.TenantId != tenantResp.Tenant.Id {
			t.Errorf("expected tenant_id %s, got %v", tenantResp.Tenant.Id, resp.RpgSystem.TenantId)
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
		req := &rpgsystempb.CreateRPGSystemRequest{
			Name:            "",
			BaseStatsSchema: `{"strength": 10}`,
		}
		_, err = rpgSystemClient.CreateRPGSystem(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})

	t.Run("empty base_stats_schema", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant 3",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Test System",
			BaseStatsSchema: "",
		}
		_, err = rpgSystemClient.CreateRPGSystem(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty base_stats_schema")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestRPGSystemHandler_GetRPGSystem(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGSystem(t)
	defer cleanup()

	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing RPG system", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		createResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Get Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		getReq := &rpgsystempb.GetRPGSystemRequest{
			Id: createResp.RpgSystem.Id,
		}
		getResp, err := rpgSystemClient.GetRPGSystem(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.RpgSystem.Id != createResp.RpgSystem.Id {
			t.Errorf("expected ID %s, got %s", createResp.RpgSystem.Id, getResp.RpgSystem.Id)
		}
	})

	t.Run("non-existing RPG system", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &rpgsystempb.GetRPGSystemRequest{
			Id: uuid.New().String(),
		}
		_, err = rpgSystemClient.GetRPGSystem(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing RPG system")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestRPGSystemHandler_ListRPGSystems(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGSystem(t)
	defer cleanup()

	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list RPG systems", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema1 := json.RawMessage(`{"strength": 10}`)
		system1, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "System 1",
			BaseStatsSchema: string(baseStatsSchema1),
		})
		if err != nil {
			t.Fatalf("failed to create system 1: %v", err)
		}

		baseStatsSchema2 := json.RawMessage(`{"magic": 10}`)
		system2, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "System 2",
			BaseStatsSchema: string(baseStatsSchema2),
		})
		if err != nil {
			t.Fatalf("failed to create system 2: %v", err)
		}

		listReq := &rpgsystempb.ListRPGSystemsRequest{
			TenantId: &tenantResp.Tenant.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := rpgSystemClient.ListRPGSystems(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 systems, got %d", listResp.TotalCount)
		}

		foundSystem1 := false
		foundSystem2 := false
		for _, s := range listResp.RpgSystems {
			if s.Id == system1.RpgSystem.Id {
				foundSystem1 = true
			}
			if s.Id == system2.RpgSystem.Id {
				foundSystem2 = true
			}
		}

		if !foundSystem1 {
			t.Error("system 1 not found in list")
		}
		if !foundSystem2 {
			t.Error("system 2 not found in list")
		}
	})
}

func TestRPGSystemHandler_UpdateRPGSystem(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGSystem(t)
	defer cleanup()

	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update RPG system", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		createResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Original System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		updateReq := &rpgsystempb.UpdateRPGSystemRequest{
			Id:   createResp.RpgSystem.Id,
			Name: stringPtr("Updated System"),
		}
		updateResp, err := rpgSystemClient.UpdateRPGSystem(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.RpgSystem.Name != "Updated System" {
			t.Errorf("expected name 'Updated System', got '%s'", updateResp.RpgSystem.Name)
		}
	})
}

func TestRPGSystemHandler_DeleteRPGSystem(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGSystem(t)
	defer cleanup()

	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete RPG system", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		createResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "System to Delete",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		deleteReq := &rpgsystempb.DeleteRPGSystemRequest{
			Id: createResp.RpgSystem.Id,
		}
		_, err = rpgSystemClient.DeleteRPGSystem(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &rpgsystempb.GetRPGSystemRequest{
			Id: createResp.RpgSystem.Id,
		}
		_, err = rpgSystemClient.GetRPGSystem(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted RPG system")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with RPG system handler
func setupTestServerWithRPGSystem(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		RPGSystemHandler: rpgSystemHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

