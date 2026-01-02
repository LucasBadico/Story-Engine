//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	"github.com/story-engine/main-service/internal/platform/logger"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	artifactrpgstatspb "github.com/story-engine/main-service/proto/artifact_rpg_stats"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestArtifactRPGStatsHandler_CreateArtifactRPGStats(t *testing.T) {
	conn, cleanup := setupTestServerWithArtifactRPGStats(t)
	defer cleanup()

	statsClient := artifactrpgstatspb.NewArtifactRPGStatsServiceClient(conn)
	artifactClient := artifactpb.NewArtifactServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Artifact Stats",
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

		artifactResp, err := artifactClient.CreateArtifact(ctx, &artifactpb.CreateArtifactRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Artifact",
		})
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		statsJSON := json.RawMessage(`{"strength": 5, "dexterity": 3}`)
		req := &artifactrpgstatspb.CreateArtifactRPGStatsRequest{
			ArtifactId: artifactResp.Artifact.Id,
			Stats:      string(statsJSON),
		}
		resp, err := statsClient.CreateArtifactRPGStats(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ArtifactRpgStats.ArtifactId != artifactResp.Artifact.Id {
			t.Errorf("expected artifact_id %s, got %s", artifactResp.Artifact.Id, resp.ArtifactRpgStats.ArtifactId)
		}
	})
}

// Helper function to create a test server with artifact RPG stats handler
func setupTestServerWithArtifactRPGStats(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(db)
	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	createStatsUseCase := artifactstatsapp.NewCreateArtifactStatsUseCase(artifactStatsRepo, artifactRepo, eventRepo, log)
	getActiveStatsUseCase := artifactstatsapp.NewGetActiveArtifactStatsUseCase(artifactStatsRepo, log)
	listStatsHistoryUseCase := artifactstatsapp.NewListArtifactStatsHistoryUseCase(artifactStatsRepo, log)
	activateVersionUseCase := artifactstatsapp.NewActivateArtifactStatsVersionUseCase(artifactStatsRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	artifactStatsHandler := NewArtifactRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listStatsHistoryUseCase, activateVersionUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:        tenantHandler,
		WorldHandler:        worldHandler,
		ArtifactHandler:     artifactHandler,
		ArtifactRPGStatsHandler: artifactStatsHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

