//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	characterpb "github.com/story-engine/main-service/proto/character"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestCharacterHandler_CreateCharacter(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Character",
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

		req := &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Test Character",
		}
		resp, err := characterClient.CreateCharacter(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Character.Name != "Test Character" {
			t.Errorf("expected name 'Test Character', got '%s'", resp.Character.Name)
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

		req := &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "",
		}
		_, err = characterClient.CreateCharacter(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestCharacterHandler_GetCharacter(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("existing character", func(t *testing.T) {
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

		createResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Get Test Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		getReq := &characterpb.GetCharacterRequest{
			Id: createResp.Character.Id,
		}
		getResp, err := characterClient.GetCharacter(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Character.Id != createResp.Character.Id {
			t.Errorf("expected ID %s, got %s", createResp.Character.Id, getResp.Character.Id)
		}
	})

	t.Run("non-existing character", func(t *testing.T) {
		req := &characterpb.GetCharacterRequest{
			Id: uuid.New().String(),
		}
		_, err := characterClient.GetCharacter(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for non-existing character")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestCharacterHandler_ListCharacters(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("list characters", func(t *testing.T) {
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

		character1, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character 1",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		character2, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character 2",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		listReq := &characterpb.ListCharactersRequest{
			WorldId: worldResp.World.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := characterClient.ListCharacters(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 characters, got %d", listResp.TotalCount)
		}

		foundCharacter1 := false
		foundCharacter2 := false
		for _, c := range listResp.Characters {
			if c.Id == character1.Character.Id {
				foundCharacter1 = true
			}
			if c.Id == character2.Character.Id {
				foundCharacter2 = true
			}
		}

		if !foundCharacter1 {
			t.Error("character 1 not found in list")
		}
		if !foundCharacter2 {
			t.Error("character 2 not found in list")
		}
	})
}

func TestCharacterHandler_UpdateCharacter(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("update character", func(t *testing.T) {
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

		createResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Original Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		updateReq := &characterpb.UpdateCharacterRequest{
			Id:   createResp.Character.Id,
			Name: stringPtr("Updated Character"),
		}
		updateResp, err := characterClient.UpdateCharacter(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Character.Name != "Updated Character" {
			t.Errorf("expected name 'Updated Character', got '%s'", updateResp.Character.Name)
		}
	})
}

func TestCharacterHandler_DeleteCharacter(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("delete character", func(t *testing.T) {
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

		createResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		deleteReq := &characterpb.DeleteCharacterRequest{
			Id: createResp.Character.Id,
		}
		_, err = characterClient.DeleteCharacter(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &characterpb.GetCharacterRequest{
			Id: createResp.Character.Id,
		}
		_, err = characterClient.GetCharacter(context.Background(), getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted character")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with character handler
func setupTestServerWithCharacter(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	getCharacterTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	traitRepo := postgres.NewTraitRepository(db)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, getCharacterTraitsUseCase, addTraitToCharacterUseCase, updateCharacterTraitUseCase, removeTraitFromCharacterUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		WorldHandler:     worldHandler,
		CharacterHandler: characterHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

