//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	"github.com/story-engine/main-service/internal/platform/logger"
	characterpb "github.com/story-engine/main-service/proto/character"
	characterrpgstatspb "github.com/story-engine/main-service/proto/character_rpg_stats"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestCharacterRPGStatsHandler_CreateCharacterRPGStats(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacterRPGStats(t)
	defer cleanup()

	statsClient := characterrpgstatspb.NewCharacterRPGStatsServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Character Stats",
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

		characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		baseStatsJSON := json.RawMessage(`{"strength": 15, "dexterity": 12}`)
		req := &characterrpgstatspb.CreateCharacterRPGStatsRequest{
			CharacterId: characterResp.Character.Id,
			BaseStats:   string(baseStatsJSON),
		}
		resp, err := statsClient.CreateCharacterRPGStats(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterRpgStats.CharacterId != characterResp.Character.Id {
			t.Errorf("expected character_id %s, got %s", characterResp.Character.Id, resp.CharacterRpgStats.CharacterId)
		}
	})
}

// Helper function to create a test server with character RPG stats handler
func setupTestServerWithCharacterRPGStats(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	createStatsUseCase := characterstatsapp.NewCreateCharacterStatsUseCase(characterStatsRepo, characterRepo, eventRepo, log)
	getActiveStatsUseCase := characterstatsapp.NewGetActiveCharacterStatsUseCase(characterStatsRepo, log)
	listStatsHistoryUseCase := characterstatsapp.NewListCharacterStatsHistoryUseCase(characterStatsRepo, log)
	activateVersionUseCase := characterstatsapp.NewActivateCharacterStatsVersionUseCase(characterStatsRepo, log)
	deleteAllStatsUseCase := characterstatsapp.NewDeleteAllCharacterStatsUseCase(characterStatsRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log), characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), characterTraitRepo, log), characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, postgres.NewTraitRepository(db), log), characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log), log)
	characterStatsHandler := NewCharacterRPGStatsHandler(createStatsUseCase, getActiveStatsUseCase, listStatsHistoryUseCase, activateVersionUseCase, deleteAllStatsUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:          tenantHandler,
		WorldHandler:           worldHandler,
		CharacterHandler:      characterHandler,
		CharacterRPGStatsHandler: characterStatsHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

