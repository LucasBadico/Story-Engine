//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	inventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	"github.com/story-engine/main-service/internal/core/rpg"
	"github.com/story-engine/main-service/internal/platform/logger"
	characterpb "github.com/story-engine/main-service/proto/character"
	inventorypb "github.com/story-engine/main-service/proto/inventory"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestInventoryHandler_AddItemToInventory(t *testing.T) {
	conn, db, cleanup := setupTestServerWithInventory(t)
	defer cleanup()

	inventoryClient := inventorypb.NewInventoryServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful add item", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Inventory",
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

		// Create RPG system for inventory item
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		// Create inventory item using use case
		rpgSystemID, _ := uuid.Parse(rpgSystemResp.RpgSystem.Id)
		item, err := rpg.NewInventoryItem(rpgSystemID, "Test Item")
		if err != nil {
			t.Fatalf("failed to create inventory item: %v", err)
		}

		// Create item via repository using the same DB from setup
		itemRepo := postgres.NewInventoryItemRepository(db)
		if err := itemRepo.Create(ctx, item); err != nil {
			t.Fatalf("failed to save inventory item: %v", err)
		}

		req := &inventorypb.AddItemToInventoryRequest{
			CharacterId: characterResp.Character.Id,
			ItemId:      item.ID.String(),
			Quantity:    int32Ptr(1),
		}
		resp, err := inventoryClient.AddItemToInventory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterInventory.CharacterId != characterResp.Character.Id {
			t.Errorf("expected character_id %s, got %s", characterResp.Character.Id, resp.CharacterInventory.CharacterId)
		}
		if resp.CharacterInventory.ItemId != item.ID.String() {
			t.Errorf("expected item_id %s, got %s", item.ID.String(), resp.CharacterInventory.ItemId)
		}
	})
}

func TestInventoryHandler_ListInventory(t *testing.T) {
	conn, db, cleanup := setupTestServerWithInventory(t)
	defer cleanup()

	inventoryClient := inventorypb.NewInventoryServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list inventory", func(t *testing.T) {
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

		characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "List Test Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Create RPG system for inventory item
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "List Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		// Create inventory item
		rpgSystemID, _ := uuid.Parse(rpgSystemResp.RpgSystem.Id)
		item, err := rpg.NewInventoryItem(rpgSystemID, "List Test Item")
		if err != nil {
			t.Fatalf("failed to create inventory item: %v", err)
		}

		// Save item to database using the same DB from setup
		itemRepo := postgres.NewInventoryItemRepository(db)
		if err := itemRepo.Create(ctx, item); err != nil {
			t.Fatalf("failed to save inventory item: %v", err)
		}

		_, err = inventoryClient.AddItemToInventory(ctx, &inventorypb.AddItemToInventoryRequest{
			CharacterId: characterResp.Character.Id,
			ItemId:      item.ID.String(),
			Quantity:    int32Ptr(1),
		})
		if err != nil {
			t.Fatalf("failed to add item: %v", err)
		}

		listReq := &inventorypb.ListInventoryRequest{
			CharacterId: characterResp.Character.Id,
		}
		listResp, err := inventoryClient.ListInventory(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listResp.CharacterInventory) < 1 {
			t.Errorf("expected at least 1 item, got %d", len(listResp.CharacterInventory))
		}
	})
}

// Helper function to create a test server with inventory handler
func setupTestServerWithInventory(t *testing.T) (*grpc.ClientConn, *postgres.DB, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	inventoryRepo := postgres.NewCharacterInventoryRepository(db)
	inventoryItemRepo := postgres.NewInventoryItemRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
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
	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	addItemUseCase := inventoryapp.NewAddItemToInventoryUseCase(inventoryRepo, characterRepo, inventoryItemRepo, log)
	updateItemUseCase := inventoryapp.NewUpdateCharacterInventoryUseCase(inventoryRepo, log)
	deleteItemUseCase := inventoryapp.NewDeleteCharacterInventoryUseCase(inventoryRepo, log)
	listInventoryUseCase := inventoryapp.NewListCharacterInventoryUseCase(inventoryRepo, log)
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log), characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), characterTraitRepo, log), characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, postgres.NewTraitRepository(db), log), characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log), log)
	artifactHandler := NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	inventoryHandler := NewInventoryHandler(addItemUseCase, updateItemUseCase, deleteItemUseCase, listInventoryUseCase, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		WorldHandler:     worldHandler,
		CharacterHandler: characterHandler,
		ArtifactHandler:  artifactHandler,
		InventoryHandler: inventoryHandler,
		RPGSystemHandler: rpgSystemHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, db, cleanup
}

