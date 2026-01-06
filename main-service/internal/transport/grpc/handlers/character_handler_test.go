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
	characterrelationshipapp "github.com/story-engine/main-service/internal/application/world/character_relationship"
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
		getResp, err := characterClient.GetCharacter(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Character.Id != createResp.Character.Id {
			t.Errorf("expected ID %s, got %s", createResp.Character.Id, getResp.Character.Id)
		}
	})

	t.Run("non-existing character", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Non-existing Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		req := &characterpb.GetCharacterRequest{
			Id: uuid.New().String(),
		}
		_, err = characterClient.GetCharacter(ctx, req)
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
		_, err = characterClient.GetCharacter(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted character")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestCharacterHandler_GetCharacterEvents(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("get events for character with no events", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Events Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Events Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		characterResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character With Events",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		getReq := &characterpb.GetCharacterEventsRequest{
			CharacterId: characterResp.Character.Id,
		}
		getResp, err := characterClient.GetCharacterEvents(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(getResp.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(getResp.Events))
		}
	})

	t.Run("invalid character_id", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Invalid Events Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		getReq := &characterpb.GetCharacterEventsRequest{
			CharacterId: "invalid-uuid",
		}
		_, err = characterClient.GetCharacterEvents(ctx, getReq)
		if err == nil {
			t.Fatal("expected error for invalid character_id")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestCharacterHandler_CreateCharacterRelationship(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Relationship Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Relationship Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character 1",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character 2",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createReq := &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "ally",
			Description:      "Best friends",
			Bidirectional:    true,
		}
		createResp, err := characterClient.CreateCharacterRelationship(ctx, createReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createResp.Relationship.RelationshipType != "ally" {
			t.Errorf("expected relationship_type 'ally', got '%s'", createResp.Relationship.RelationshipType)
		}
		if createResp.Relationship.Description != "Best friends" {
			t.Errorf("expected description 'Best friends', got '%s'", createResp.Relationship.Description)
		}
		if !createResp.Relationship.Bidirectional {
			t.Error("expected bidirectional to be true")
		}
	})

	t.Run("same character relationship fails", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Same Char Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Same Char Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		charResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Solo Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		createReq := &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      charResp.Character.Id,
			OtherCharacterId: charResp.Character.Id,
			RelationshipType: "self",
		}
		_, err = characterClient.CreateCharacterRelationship(ctx, createReq)
		if err == nil {
			t.Fatal("expected error for same character relationship")
		}
	})

	t.Run("empty relationship_type fails", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Empty Type Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Empty Type Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character A",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Character B",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createReq := &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "",
		}
		_, err = characterClient.CreateCharacterRelationship(ctx, createReq)
		if err == nil {
			t.Fatal("expected error for empty relationship_type")
		}
	})
}

func TestCharacterHandler_GetCharacterRelationship(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("existing relationship", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Get Rel Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Get Rel Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Char 1",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Char 2",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createResp, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "rival",
		})
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		getReq := &characterpb.GetCharacterRelationshipRequest{
			Id: createResp.Relationship.Id,
		}
		getResp, err := characterClient.GetCharacterRelationship(ctx, getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Relationship.Id != createResp.Relationship.Id {
			t.Errorf("expected ID %s, got %s", createResp.Relationship.Id, getResp.Relationship.Id)
		}
		if getResp.Relationship.RelationshipType != "rival" {
			t.Errorf("expected relationship_type 'rival', got '%s'", getResp.Relationship.RelationshipType)
		}
	})

	t.Run("non-existing relationship", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "NonExist Rel Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

		getReq := &characterpb.GetCharacterRelationshipRequest{
			Id: uuid.New().String(),
		}
		_, err = characterClient.GetCharacterRelationship(ctx, getReq)
		if err == nil {
			t.Fatal("expected error for non-existing relationship")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})
}

func TestCharacterHandler_ListCharacterRelationships(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("list relationships", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "List Rel Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "List Rel Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Main Character",
		})
		if err != nil {
			t.Fatalf("failed to create main character: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Friend",
		})
		if err != nil {
			t.Fatalf("failed to create friend: %v", err)
		}

		char3Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Enemy",
		})
		if err != nil {
			t.Fatalf("failed to create enemy: %v", err)
		}

		// Create relationships with main character
		rel1, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "friend",
		})
		if err != nil {
			t.Fatalf("failed to create relationship 1: %v", err)
		}

		rel2, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char3Resp.Character.Id,
			RelationshipType: "enemy",
		})
		if err != nil {
			t.Fatalf("failed to create relationship 2: %v", err)
		}

		listReq := &characterpb.ListCharacterRelationshipsRequest{
			CharacterId: char1Resp.Character.Id,
		}
		listResp, err := characterClient.ListCharacterRelationships(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listResp.Relationships) != 2 {
			t.Errorf("expected 2 relationships, got %d", len(listResp.Relationships))
		}

		foundRel1 := false
		foundRel2 := false
		for _, r := range listResp.Relationships {
			if r.Id == rel1.Relationship.Id {
				foundRel1 = true
			}
			if r.Id == rel2.Relationship.Id {
				foundRel2 = true
			}
		}

		if !foundRel1 {
			t.Error("relationship 1 not found in list")
		}
		if !foundRel2 {
			t.Error("relationship 2 not found in list")
		}
	})

	t.Run("list with no relationships", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Empty List Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Empty List Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		charResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Lonely Character",
		})
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		listReq := &characterpb.ListCharacterRelationshipsRequest{
			CharacterId: charResp.Character.Id,
		}
		listResp, err := characterClient.ListCharacterRelationships(ctx, listReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listResp.Relationships) != 0 {
			t.Errorf("expected 0 relationships, got %d", len(listResp.Relationships))
		}
	})
}

func TestCharacterHandler_UpdateCharacterRelationship(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("update relationship", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Update Rel Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Update Rel Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Char A",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Char B",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createResp, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "acquaintance",
			Bidirectional:    false,
		})
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		newType := "friend"
		newDesc := "They became friends"
		newBidi := true
		updateReq := &characterpb.UpdateCharacterRelationshipRequest{
			Id:               createResp.Relationship.Id,
			RelationshipType: &newType,
			Description:      &newDesc,
			Bidirectional:    &newBidi,
		}
		updateResp, err := characterClient.UpdateCharacterRelationship(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updateResp.Relationship.RelationshipType != "friend" {
			t.Errorf("expected relationship_type 'friend', got '%s'", updateResp.Relationship.RelationshipType)
		}
		if updateResp.Relationship.Description != "They became friends" {
			t.Errorf("expected description 'They became friends', got '%s'", updateResp.Relationship.Description)
		}
		if !updateResp.Relationship.Bidirectional {
			t.Error("expected bidirectional to be true")
		}
	})

	t.Run("partial update", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Partial Update Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Partial Update Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Hero",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Villain",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createResp, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "enemy",
			Description:      "Original description",
		})
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		// Only update description
		newDesc := "Updated description only"
		updateReq := &characterpb.UpdateCharacterRelationshipRequest{
			Id:          createResp.Relationship.Id,
			Description: &newDesc,
		}
		updateResp, err := characterClient.UpdateCharacterRelationship(ctx, updateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check relationship_type is unchanged
		if updateResp.Relationship.RelationshipType != "enemy" {
			t.Errorf("expected relationship_type 'enemy', got '%s'", updateResp.Relationship.RelationshipType)
		}
		if updateResp.Relationship.Description != "Updated description only" {
			t.Errorf("expected description 'Updated description only', got '%s'", updateResp.Relationship.Description)
		}
	})
}

func TestCharacterHandler_DeleteCharacterRelationship(t *testing.T) {
	conn, cleanup := setupTestServerWithCharacter(t)
	defer cleanup()

	characterClient := characterpb.NewCharacterServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)

	t.Run("delete relationship", func(t *testing.T) {
		tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
			Name: "Delete Rel Test Tenant",
		})
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)
		worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
			Name: "Delete Rel Test World",
		})
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		char1Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Del Char 1",
		})
		if err != nil {
			t.Fatalf("failed to create character 1: %v", err)
		}

		char2Resp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
			WorldId: worldResp.World.Id,
			Name:    "Del Char 2",
		})
		if err != nil {
			t.Fatalf("failed to create character 2: %v", err)
		}

		createResp, err := characterClient.CreateCharacterRelationship(ctx, &characterpb.CreateCharacterRelationshipRequest{
			CharacterId:      char1Resp.Character.Id,
			OtherCharacterId: char2Resp.Character.Id,
			RelationshipType: "temporary",
		})
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		deleteReq := &characterpb.DeleteCharacterRelationshipRequest{
			Id: createResp.Relationship.Id,
		}
		_, err = characterClient.DeleteCharacterRelationship(ctx, deleteReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify relationship is deleted
		getReq := &characterpb.GetCharacterRelationshipRequest{
			Id: createResp.Relationship.Id,
		}
		_, err = characterClient.GetCharacterRelationship(ctx, getReq)
		if err == nil {
			t.Fatal("expected error when getting deleted relationship")
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
	characterRelationshipRepo := postgres.NewCharacterRelationshipRepository(db)
	eventReferenceRepo := postgres.NewEventReferenceRepository(db)
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
	getCharacterEventsUseCase := characterapp.NewGetCharacterEventsUseCase(eventReferenceRepo, log)
	traitRepo := postgres.NewTraitRepository(db)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	
	// Character relationship use cases
	createCharacterRelationshipUseCase := characterrelationshipapp.NewCreateCharacterRelationshipUseCase(characterRelationshipRepo, characterRepo, log)
	getCharacterRelationshipUseCase := characterrelationshipapp.NewGetCharacterRelationshipUseCase(characterRelationshipRepo, log)
	listCharacterRelationshipsUseCase := characterrelationshipapp.NewListCharacterRelationshipsUseCase(characterRelationshipRepo, log)
	updateCharacterRelationshipUseCase := characterrelationshipapp.NewUpdateCharacterRelationshipUseCase(characterRelationshipRepo, log)
	deleteCharacterRelationshipUseCase := characterrelationshipapp.NewDeleteCharacterRelationshipUseCase(characterRelationshipRepo, log)

	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(
		createCharacterUseCase, 
		getCharacterUseCase, 
		listCharactersUseCase, 
		updateCharacterUseCase, 
		deleteCharacterUseCase, 
		getCharacterTraitsUseCase, 
		getCharacterEventsUseCase,
		addTraitToCharacterUseCase, 
		updateCharacterTraitUseCase, 
		removeTraitFromCharacterUseCase,
		createCharacterRelationshipUseCase,
		getCharacterRelationshipUseCase,
		listCharacterRelationshipsUseCase,
		updateCharacterRelationshipUseCase,
		deleteCharacterRelationshipUseCase,
		log,
	)

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

