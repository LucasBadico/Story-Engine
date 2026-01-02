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
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	"github.com/story-engine/main-service/internal/platform/logger"
	characterpb "github.com/story-engine/main-service/proto/character"
	characterskillpb "github.com/story-engine/main-service/proto/character_skill"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	skillpb "github.com/story-engine/main-service/proto/skill"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestCharacterSkillHandler_AddCharacterSkill(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacterSkill(t)
	defer cleanup()

	characterSkillClient := characterskillpb.NewCharacterSkillServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful add skill", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Character Skill",
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

		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		skillResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Fireball",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		req := &characterskillpb.AddCharacterSkillRequest{
			CharacterId: characterResp.Character.Id,
			SkillId:     skillResp.Skill.Id,
		}
		resp, err := characterSkillClient.AddCharacterSkill(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterSkill.CharacterId != characterResp.Character.Id {
			t.Errorf("expected character_id %s, got %s", characterResp.Character.Id, resp.CharacterSkill.CharacterId)
		}
		if resp.CharacterSkill.SkillId != skillResp.Skill.Id {
			t.Errorf("expected skill_id %s, got %s", skillResp.Skill.Id, resp.CharacterSkill.SkillId)
		}
	})
}

func TestCharacterSkillHandler_ListCharacterSkills(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacterSkill(t)
	defer cleanup()

	characterSkillClient := characterskillpb.NewCharacterSkillServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list character skills", func(t *testing.T) {
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

		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "List Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		skillResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "List Test Skill",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		_, err = characterSkillClient.AddCharacterSkill(ctx, &characterskillpb.AddCharacterSkillRequest{
			CharacterId: characterResp.Character.Id,
			SkillId:     skillResp.Skill.Id,
		})
		if err != nil {
			t.Fatalf("failed to add skill: %v", err)
		}

		listReq := &characterskillpb.ListCharacterSkillsRequest{
			CharacterId: characterResp.Character.Id,
		}
		listResp, err := characterSkillClient.ListCharacterSkills(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listResp.CharacterSkills) < 1 {
			t.Errorf("expected at least 1 skill, got %d", len(listResp.CharacterSkills))
		}
	})
}

// Helper function to create a test server with character skill handler
func setupTestServerWithCharacterSkill(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	characterSkillRepo := postgres.NewCharacterSkillRepository(db)
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
	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	learnSkillUseCase := characterskillapp.NewLearnSkillUseCase(characterSkillRepo, characterRepo, skillRepo, log)
	updateCharacterSkillUseCase := characterskillapp.NewUpdateCharacterSkillUseCase(characterSkillRepo, skillRepo, log)
	deleteCharacterSkillUseCase := characterskillapp.NewDeleteCharacterSkillUseCase(characterSkillRepo, log)
	listCharacterSkillsUseCase := characterskillapp.NewListCharacterSkillsUseCase(characterSkillRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log), characterapp.NewAddTraitToCharacterUseCase(characterRepo, postgres.NewTraitRepository(db), characterTraitRepo, log), characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, postgres.NewTraitRepository(db), log), characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log), log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	skillHandler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	characterSkillHandler := NewCharacterSkillHandler(learnSkillUseCase, updateCharacterSkillUseCase, deleteCharacterSkillUseCase, listCharacterSkillsUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:       tenantHandler,
		WorldHandler:       worldHandler,
		CharacterHandler:   characterHandler,
		RPGSystemHandler:  rpgSystemHandler,
		SkillHandler:       skillHandler,
		CharacterSkillHandler: characterSkillHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

func int32Ptr(i int32) *int32 {
	return &i
}

