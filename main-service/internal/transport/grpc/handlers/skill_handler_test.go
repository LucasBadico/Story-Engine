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
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	skillpb "github.com/story-engine/main-service/proto/skill"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSkillHandler_CreateSkill(t *testing.T) {
	conn, cleanup := setupTestServerWithSkill(t)
	defer cleanup()

	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for Skill",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		req := &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Fireball",
			Description: stringPtr("A fireball spell"),
		}
		resp, err := skillClient.CreateSkill(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Skill.Name != "Fireball" {
			t.Errorf("expected name 'Fireball', got '%s'", resp.Skill.Name)
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
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Test System 2",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		req := &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "",
		}
		_, err = skillClient.CreateSkill(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestSkillHandler_GetSkill(t *testing.T) {
	conn, cleanup := setupTestServerWithSkill(t)
	defer cleanup()

	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing skill", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Get Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		createResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Get Test Skill",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		getReq := &skillpb.GetSkillRequest{
			Id: createResp.Skill.Id,
		}
		getResp, err := skillClient.GetSkill(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Skill.Id != createResp.Skill.Id {
			t.Errorf("expected ID %s, got %s", createResp.Skill.Id, getResp.Skill.Id)
		}
	})

	t.Run("non-existing skill", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &skillpb.GetSkillRequest{
			Id: uuid.New().String(),
		}
		_, err = skillClient.GetSkill(ctx, req)
		if err == nil {
			t.Fatal("expected error for non-existing skill")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestSkillHandler_ListSkills(t *testing.T) {
	conn, cleanup := setupTestServerWithSkill(t)
	defer cleanup()

	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list skills", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "List Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		skill1, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Skill 1",
		})
		if err != nil {
			t.Fatalf("failed to create skill 1: %v", err)
		}

		skill2, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Skill 2",
		})
		if err != nil {
			t.Fatalf("failed to create skill 2: %v", err)
		}

		listReq := &skillpb.ListSkillsRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := skillClient.ListSkills(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 skills, got %d", listResp.TotalCount)
		}

		foundSkill1 := false
		foundSkill2 := false
		for _, s := range listResp.Skills {
			if s.Id == skill1.Skill.Id {
				foundSkill1 = true
			}
			if s.Id == skill2.Skill.Id {
				foundSkill2 = true
			}
		}

		if !foundSkill1 {
			t.Error("skill 1 not found in list")
		}
		if !foundSkill2 {
			t.Error("skill 2 not found in list")
		}
	})
}

func TestSkillHandler_UpdateSkill(t *testing.T) {
	conn, cleanup := setupTestServerWithSkill(t)
	defer cleanup()

	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update skill", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Update Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		createResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Original Skill",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		updateReq := &skillpb.UpdateSkillRequest{
			Id:   createResp.Skill.Id,
			Name: stringPtr("Updated Skill"),
		}
		updateResp, err := skillClient.UpdateSkill(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Skill.Name != "Updated Skill" {
			t.Errorf("expected name 'Updated Skill', got '%s'", updateResp.Skill.Name)
		}
	})
}

func TestSkillHandler_DeleteSkill(t *testing.T) {
	conn, cleanup := setupTestServerWithSkill(t)
	defer cleanup()

	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete skill", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Delete Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		createResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Skill to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		deleteReq := &skillpb.DeleteSkillRequest{
			Id: createResp.Skill.Id,
		}
		_, err = skillClient.DeleteSkill(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &skillpb.GetSkillRequest{
			Id: createResp.Skill.Id,
		}
		_, err = skillClient.GetSkill(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted skill")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

// Helper function to create a test server with skill handler
func setupTestServerWithSkill(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
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

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	skillHandler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		RPGSystemHandler: rpgSystemHandler,
		SkillHandler:     skillHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

