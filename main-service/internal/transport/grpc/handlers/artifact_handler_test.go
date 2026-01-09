//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	commonpb "github.com/story-engine/main-service/proto/common"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestArtifactHandler_CreateArtifact(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifact(t)
	defer cleanup()

	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Artifact",
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

		req := &artifactpb.CreateArtifactRequest{
			WorldId:     worldResp.World.Id,
			Name:        "Test Artifact",
			Description: "A test artifact",
		}
		resp, err := artifactClient.CreateArtifact(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Artifact.Name != "Test Artifact" {
			t.Errorf("expected name 'Test Artifact', got '%s'", resp.Artifact.Name)
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

		req := &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "",
		}
		_, err = artifactClient.CreateArtifact(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestArtifactHandler_GetArtifact(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifact(t)
	defer cleanup()

	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("existing artifact", func(t *testing.T) {
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

		createResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Get Test Artifact",
		})
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		getReq := &artifactpb.GetArtifactRequest{
			Id: createResp.Artifact.Id,
		}
		getResp, err := artifactClient.GetArtifact(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Artifact.Id != createResp.Artifact.Id {
			t.Errorf("expected ID %s, got %s", createResp.Artifact.Id, getResp.Artifact.Id)
		}
	})

	t.Run("non-existing artifact", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &artifactpb.GetArtifactRequest{
			Id: uuid.New().String(),
		}
		_, err = artifactClient.GetArtifact(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing artifact")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestArtifactHandler_ListArtifacts(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifact(t)
	defer cleanup()

	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("list artifacts", func(t *testing.T) {
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

		artifact1, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Artifact 1",
		})
		if err != nil {
			t.Fatalf("failed to create artifact 1: %v", err)
		}

		artifact2, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Artifact 2",
		})
		if err != nil {
			t.Fatalf("failed to create artifact 2: %v", err)
		}

		listReq := &artifactpb.ListArtifactsRequest{
			WorldId: worldResp.World.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := artifactClient.ListArtifacts(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 artifacts, got %d", listResp.TotalCount)
		}

		foundArtifact1 := false
		foundArtifact2 := false
		for _, a := range listResp.Artifacts {
			if a.Id == artifact1.Artifact.Id {
				foundArtifact1 = true
			}
			if a.Id == artifact2.Artifact.Id {
				foundArtifact2 = true
			}
		}

		if !foundArtifact1 {
			t.Error("artifact 1 not found in list")
		}
		if !foundArtifact2 {
			t.Error("artifact 2 not found in list")
		}
	})
}

func TestArtifactHandler_UpdateArtifact(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifact(t)
	defer cleanup()

	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("update artifact", func(t *testing.T) {
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

		createResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Original Artifact",
		})
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		updateReq := &artifactpb.UpdateArtifactRequest{
			Id:   createResp.Artifact.Id,
			Name: stringPtr("Updated Artifact"),
		}
		updateResp, err := artifactClient.UpdateArtifact(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Artifact.Name != "Updated Artifact" {
			t.Errorf("expected name 'Updated Artifact', got '%s'", updateResp.Artifact.Name)
		}
	})
}

func TestArtifactHandler_DeleteArtifact(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifact(t)
	defer cleanup()

	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("delete artifact", func(t *testing.T) {
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

		createResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Artifact to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		deleteReq := &artifactpb.DeleteArtifactRequest{
			Id: createResp.Artifact.Id,
		}
		_, err = artifactClient.DeleteArtifact(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &artifactpb.GetArtifactRequest{
			Id: createResp.Artifact.Id,
		}
		_, err = artifactClient.GetArtifact(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted artifact")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with artifact handler
func setupTestServerWithArtifact(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	// Entity relations use cases for artifact references
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, createRelationUseCase, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, createRelationUseCase, listRelationsBySourceUseCase, deleteRelationUseCase, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(listRelationsBySourceUseCase, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:  tenantHandler,
		WorldHandler:   worldHandler,
		ArtifactHandler: artifactHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

