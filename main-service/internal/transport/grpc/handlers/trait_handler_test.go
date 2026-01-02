//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	traitpb "github.com/story-engine/main-service/proto/trait"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTraitHandler_CreateTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithTrait(t)
	defer cleanup()

	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Trait",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &traitpb.CreateTraitRequest{
			Name:        "Brave",
			Category:    "Personality",
			Description: "A brave character trait",
		}
		resp, err := traitClient.CreateTrait(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Trait.Name != "Brave" {
			t.Errorf("expected name 'Brave', got '%s'", resp.Trait.Name)
		}
		if resp.Trait.Category != "Personality" {
			t.Errorf("expected category 'Personality', got '%s'", resp.Trait.Category)
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
		req := &traitpb.CreateTraitRequest{
			Name: "",
		}
		_, err = traitClient.CreateTrait(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestTraitHandler_GetTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithTrait(t)
	defer cleanup()

	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing trait", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Get Test Trait",
		})
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		getReq := &traitpb.GetTraitRequest{
			Id: createResp.Trait.Id,
		}
		getResp, err := traitClient.GetTrait(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Trait.Id != createResp.Trait.Id {
			t.Errorf("expected ID %s, got %s", createResp.Trait.Id, getResp.Trait.Id)
		}
	})

	t.Run("non-existing trait", func(t *testing.T) {
		req := &traitpb.GetTraitRequest{
			Id: uuid.New().String(),
		}
		_, err := traitClient.GetTrait(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for non-existing trait")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestTraitHandler_ListTraits(t *testing.T) {
	conn, cleanup := setupTestServerWithTrait(t)
	defer cleanup()

	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list traits", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		trait1, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Trait 1",
		})
		if err != nil {
			t.Fatalf("failed to create trait 1: %v", err)
		}

		trait2, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Trait 2",
		})
		if err != nil {
			t.Fatalf("failed to create trait 2: %v", err)
		}

		listReq := &traitpb.ListTraitsRequest{
			TenantId: tenantResp.Tenant.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := traitClient.ListTraits(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 traits, got %d", listResp.TotalCount)
		}

		foundTrait1 := false
		foundTrait2 := false
		for _, tr := range listResp.Traits {
			if tr.Id == trait1.Trait.Id {
				foundTrait1 = true
			}
			if tr.Id == trait2.Trait.Id {
				foundTrait2 = true
			}
		}

		if !foundTrait1 {
			t.Error("trait 1 not found in list")
		}
		if !foundTrait2 {
			t.Error("trait 2 not found in list")
		}
	})
}

func TestTraitHandler_UpdateTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithTrait(t)
	defer cleanup()

	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update trait", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Original Trait",
		})
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		updateReq := &traitpb.UpdateTraitRequest{
			Id:   createResp.Trait.Id,
			Name: stringPtr("Updated Trait"),
		}
		updateResp, err := traitClient.UpdateTrait(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Trait.Name != "Updated Trait" {
			t.Errorf("expected name 'Updated Trait', got '%s'", updateResp.Trait.Name)
		}
	})
}

func TestTraitHandler_DeleteTrait(t *testing.T) {
	conn, cleanup := setupTestServerWithTrait(t)
	defer cleanup()

	traitClient := traitpb.NewTraitServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete trait", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		createResp, err := traitClient.CreateTrait(ctx, &traitpb.CreateTraitRequest{
			Name: "Trait to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		deleteReq := &traitpb.DeleteTraitRequest{
			Id: createResp.Trait.Id,
		}
		_, err = traitClient.DeleteTrait(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &traitpb.GetTraitRequest{
			Id: createResp.Trait.Id,
		}
		_, err = traitClient.GetTrait(context.Background(), getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted trait")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with trait handler
func setupTestServerWithTrait(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	traitRepo := postgres.NewTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	traitHandler := NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler: tenantHandler,
		TraitHandler:  traitHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

