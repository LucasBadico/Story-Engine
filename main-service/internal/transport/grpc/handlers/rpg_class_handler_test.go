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
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	"github.com/story-engine/main-service/internal/platform/logger"
	commonpb "github.com/story-engine/main-service/proto/common"
	rpgclasspb "github.com/story-engine/main-service/proto/rpg_class"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	skillpb "github.com/story-engine/main-service/proto/skill"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRPGClassHandler_CreateRPGClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Test Tenant for RPG Class",
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

		req := &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Warrior",
			Description: stringPtr("A warrior class"),
		}
		resp, err := rpgClassClient.CreateRPGClass(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.RpgClass.Name != "Warrior" {
			t.Errorf("expected name 'Warrior', got '%s'", resp.RpgClass.Name)
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

		req := &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "",
		}
		_, err = rpgClassClient.CreateRPGClass(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestRPGClassHandler_GetRPGClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing RPG class", func(t *testing.T) {
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

		createResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Get Test Class",
		})
		if err != nil {
			t.Fatalf("failed to create RPG class: %v", err)
		}

		getReq := &rpgclasspb.GetRPGClassRequest{
			Id: createResp.RpgClass.Id,
		}
		getResp, err := rpgClassClient.GetRPGClass(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.RpgClass.Id != createResp.RpgClass.Id {
			t.Errorf("expected ID %s, got %s", createResp.RpgClass.Id, getResp.RpgClass.Id)
		}
	})

	t.Run("non-existing RPG class", func(t *testing.T) {
		req := &rpgclasspb.GetRPGClassRequest{
			Id: uuid.New().String(),
		}
		_, err := rpgClassClient.GetRPGClass(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for non-existing RPG class")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestRPGClassHandler_ListRPGClasses(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("list RPG classes", func(t *testing.T) {
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

		class1, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Class 1",
		})
		if err != nil {
			t.Fatalf("failed to create class 1: %v", err)
		}

		class2, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Class 2",
		})
		if err != nil {
			t.Fatalf("failed to create class 2: %v", err)
		}

		listReq := &rpgclasspb.ListRPGClassesRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Pagination: &commonpb.PaginationRequest{
				Limit:  10,
				Offset: 0,
			},
		}
		listResp, err := rpgClassClient.ListRPGClasses(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if listResp.TotalCount < 2 {
			t.Errorf("expected at least 2 classes, got %d", listResp.TotalCount)
		}

		foundClass1 := false
		foundClass2 := false
		for _, c := range listResp.RpgClasses {
			if c.Id == class1.RpgClass.Id {
				foundClass1 = true
			}
			if c.Id == class2.RpgClass.Id {
				foundClass2 = true
			}
		}

		if !foundClass1 {
			t.Error("class 1 not found in list")
		}
		if !foundClass2 {
			t.Error("class 2 not found in list")
		}
	})
}

func TestRPGClassHandler_UpdateRPGClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("update RPG class", func(t *testing.T) {
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

		createResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Original Class",
		})
		if err != nil {
			t.Fatalf("failed to create RPG class: %v", err)
		}

		updateReq := &rpgclasspb.UpdateRPGClassRequest{
			Id:   createResp.RpgClass.Id,
			Name: stringPtr("Updated Class"),
		}
		updateResp, err := rpgClassClient.UpdateRPGClass(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.RpgClass.Name != "Updated Class" {
			t.Errorf("expected name 'Updated Class', got '%s'", updateResp.RpgClass.Name)
		}
	})
}

func TestRPGClassHandler_DeleteRPGClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("delete RPG class", func(t *testing.T) {
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

		createResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Class to Delete",
		})
		if err != nil {
			t.Fatalf("failed to create RPG class: %v", err)
		}

		deleteReq := &rpgclasspb.DeleteRPGClassRequest{
			Id: createResp.RpgClass.Id,
		}
		_, err = rpgClassClient.DeleteRPGClass(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		getReq := &rpgclasspb.GetRPGClassRequest{
			Id: createResp.RpgClass.Id,
		}
		_, err = rpgClassClient.GetRPGClass(context.Background(), getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted RPG class")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestRPGClassHandler_AddSkillToClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	t.Run("add skill to class", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Add Skill Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		baseStatsSchema := json.RawMessage(`{"strength": 10}`)
		rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
			Name:            "Add Skill Test System",
			BaseStatsSchema: string(baseStatsSchema),
		})
		if err != nil {
			t.Fatalf("failed to create RPG system: %v", err)
		}

		classResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Warrior",
		})
		if err != nil {
			t.Fatalf("failed to create RPG class: %v", err)
		}

		skillResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Sword Mastery",
		})
		if err != nil {
			t.Fatalf("failed to create skill: %v", err)
		}

		addReq := &rpgclasspb.AddSkillToRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skillResp.Skill.Id,
		}
		_, err = rpgClassClient.AddSkillToRPGClass(ctx, addReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRPGClassHandler_RemoveSkillFromClass(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Remove Skill Test Tenant",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	baseStatsSchema := json.RawMessage(`{"strength": 10}`)
	rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
		Name:            "Remove Skill Test System",
		BaseStatsSchema: string(baseStatsSchema),
	})
	if err != nil {
		t.Fatalf("failed to create rpg system: %v", err)
	}
	classResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
		RpgSystemId: rpgSystemResp.RpgSystem.Id,
		Name:        "Warrior",
	})
	if err != nil {
		t.Fatalf("failed to create rpg class: %v", err)
	}
	skillResp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
		RpgSystemId: rpgSystemResp.RpgSystem.Id,
		Name:        "Sword Mastery",
	})
	if err != nil {
		t.Fatalf("failed to create skill: %v", err)
	}

	t.Run("successful remove", func(t *testing.T) {
		// Add skill first
		_, err := rpgClassClient.AddSkillToRPGClass(ctx, &rpgclasspb.AddSkillToRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skillResp.Skill.Id,
		})
		if err != nil {
			t.Fatalf("failed to add skill: %v", err)
		}

		// Remove skill
		_, err = rpgClassClient.RemoveSkillFromRPGClass(ctx, &rpgclasspb.RemoveSkillFromRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skillResp.Skill.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's removed
		listResp, err := rpgClassClient.ListRPGClassSkills(ctx, &rpgclasspb.ListRPGClassSkillsRequest{
			RpgClassId: classResp.RpgClass.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.SkillIds) != 0 {
			t.Errorf("expected 0 skills, got %d", len(listResp.SkillIds))
		}
	})

	t.Run("remove non-existing skill", func(t *testing.T) {
		_, err := rpgClassClient.RemoveSkillFromRPGClass(ctx, &rpgclasspb.RemoveSkillFromRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing skill")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestRPGClassHandler_ListRPGClassSkills(t *testing.T) {
	conn, cleanup := setupTestServerWithRPGClass(t)
	defer cleanup()

	rpgClassClient := rpgclasspb.NewRPGClassServiceClient(conn)
	skillClient := skillpb.NewSkillServiceClient(conn)
	rpgSystemClient := rpgsystempb.NewRPGSystemServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	// Setup
	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "List Skills Test Tenant",
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
	baseStatsSchema := json.RawMessage(`{"strength": 10}`)
	rpgSystemResp, err := rpgSystemClient.CreateRPGSystem(ctx, &rpgsystempb.CreateRPGSystemRequest{
		Name:            "List Skills Test System",
		BaseStatsSchema: string(baseStatsSchema),
	})
	if err != nil {
		t.Fatalf("failed to create rpg system: %v", err)
	}
	classResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
		RpgSystemId: rpgSystemResp.RpgSystem.Id,
		Name:        "Warrior",
	})
	if err != nil {
		t.Fatalf("failed to create rpg class: %v", err)
	}

	t.Run("list multiple skills", func(t *testing.T) {
		// Create and add multiple skills
		skill1Resp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Sword Mastery",
		})
		if err != nil {
			t.Fatalf("failed to create skill1: %v", err)
		}
		skill2Resp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Shield Defense",
		})
		if err != nil {
			t.Fatalf("failed to create skill2: %v", err)
		}
		skill3Resp, err := skillClient.CreateSkill(ctx, &skillpb.CreateSkillRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Combat Tactics",
		})
		if err != nil {
			t.Fatalf("failed to create skill3: %v", err)
		}

		_, _ = rpgClassClient.AddSkillToRPGClass(ctx, &rpgclasspb.AddSkillToRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skill1Resp.Skill.Id,
		})
		_, _ = rpgClassClient.AddSkillToRPGClass(ctx, &rpgclasspb.AddSkillToRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skill2Resp.Skill.Id,
		})
		_, _ = rpgClassClient.AddSkillToRPGClass(ctx, &rpgclasspb.AddSkillToRPGClassRequest{
			RpgClassId: classResp.RpgClass.Id,
			SkillId:    skill3Resp.Skill.Id,
		})

		// List skills
		listResp, err := rpgClassClient.ListRPGClassSkills(ctx, &rpgclasspb.ListRPGClassSkillsRequest{
			RpgClassId: classResp.RpgClass.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.SkillIds) != 3 {
			t.Errorf("expected 3 skills, got %d", len(listResp.SkillIds))
		}
		if listResp.TotalCount != 3 {
			t.Errorf("expected total_count 3, got %d", listResp.TotalCount)
		}
	})

	t.Run("empty list for new class", func(t *testing.T) {
		newClassResp, err := rpgClassClient.CreateRPGClass(ctx, &rpgclasspb.CreateRPGClassRequest{
			RpgSystemId: rpgSystemResp.RpgSystem.Id,
			Name:        "Mage",
		})
		if err != nil {
			t.Fatalf("failed to create rpg class: %v", err)
		}

		listResp, err := rpgClassClient.ListRPGClassSkills(ctx, &rpgclasspb.ListRPGClassSkillsRequest{
			RpgClassId: newClassResp.RpgClass.Id,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(listResp.SkillIds) != 0 {
			t.Errorf("expected 0 skills, got %d", len(listResp.SkillIds))
		}
	})
}

// Helper function to create a test server with RPG class handler
func setupTestServerWithRPGClass(t *testing.T) (*grpc.ClientConn, func()) {
	db, cleanupDB := postgres.SetupTestDB(t)

	tenantRepo := postgres.NewTenantRepository(db)
	rpgSystemRepo := postgres.NewRPGSystemRepository(db)
	skillRepo := postgres.NewSkillRepository(db)
	rpgClassRepo := postgres.NewRPGClassRepository(db)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(db)
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
	createRPGClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getRPGClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listRPGClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateRPGClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteRPGClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillToClassUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	removeSkillFromClassUseCase := rpgclassapp.NewRemoveSkillFromClassUseCase(rpgClassSkillRepo, rpgClassRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	rpgSystemHandler := NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	skillHandler := NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	rpgClassHandler := NewRPGClassHandler(createRPGClassUseCase, getRPGClassUseCase, listRPGClassesUseCase, updateRPGClassUseCase, deleteRPGClassUseCase, addSkillToClassUseCase, listClassSkillsUseCase, removeSkillFromClassUseCase, log)

	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:    tenantHandler,
		RPGSystemHandler: rpgSystemHandler,
		SkillHandler:     skillHandler,
		RPGClassHandler:  rpgClassHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}

