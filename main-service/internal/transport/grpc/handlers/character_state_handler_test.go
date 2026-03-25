//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	characterpb "github.com/story-engine/main-service/proto/character"
	pb "github.com/story-engine/main-service/proto/character_state"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// setupCharacterStateTest creates tenant → world → character and returns ctx + IDs
func setupCharacterStateTest(t *testing.T, conn *grpc.ClientConn) (ctx context.Context, characterID string) {
	t.Helper()
	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)

	tenantResp, err := tenantClient.CreateTenant(context.Background(), &tenantpb.CreateTenantRequest{
		Name: "Test Tenant " + uuid.New().String()[:8],
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	ctx = metadata.AppendToOutgoingContext(context.Background(), "tenant_id", tenantResp.Tenant.Id)

	worldResp, err := worldClient.CreateWorld(ctx, &worldpb.CreateWorldRequest{
		Name: "Test World",
	})
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	charResp, err := characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
		WorldId:     worldResp.World.Id,
		Name:        "Haruka",
		Description: "Test character",
	})
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	return ctx, charResp.Character.Id
}

func TestCharacterStateHandler_CreateAndGet(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewCharacterStateServiceClient(conn)
	ctx, characterID := setupCharacterStateTest(t, conn)

	t.Run("create with timeline", func(t *testing.T) {
		resp, err := client.CreateCharacterState(ctx, &pb.CreateCharacterStateRequest{
			CharacterId:        characterID,
			ChapterNumber:      1,
			Emotion:            "anxious",
			Goal:               "survive the night",
			Secrets:            []string{"knows about the letter"},
			Clothes:            "white dress",
			PhysicalState:      "exhausted",
			HasTimelineDay:     true,
			TimelineDay:        -40,
			TimelineConfidence: "explicit",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterState.Emotion != "anxious" {
			t.Errorf("expected emotion 'anxious', got '%s'", resp.CharacterState.Emotion)
		}
		if resp.CharacterState.TimelineDay != -40 {
			t.Errorf("expected timeline_day -40, got %d", resp.CharacterState.TimelineDay)
		}
		if !resp.CharacterState.HasTimelineDay {
			t.Error("expected has_timeline_day to be true")
		}
		if len(resp.CharacterState.Secrets) != 1 {
			t.Errorf("expected 1 secret, got %d", len(resp.CharacterState.Secrets))
		}

		// Get it back
		getResp, err := client.GetCharacterState(ctx, &pb.GetCharacterStateRequest{
			Id: resp.CharacterState.Id,
		})
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if getResp.CharacterState.Goal != "survive the night" {
			t.Errorf("expected goal 'survive the night', got '%s'", getResp.CharacterState.Goal)
		}
	})

	t.Run("non-existing character", func(t *testing.T) {
		_, err := client.CreateCharacterState(ctx, &pb.CreateCharacterStateRequest{
			CharacterId:   uuid.New().String(),
			ChapterNumber: 1,
			Emotion:       "happy",
		})
		if err == nil {
			t.Fatal("expected error for non-existing character")
		}
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("get non-existing", func(t *testing.T) {
		_, err := client.GetCharacterState(ctx, &pb.GetCharacterStateRequest{
			Id: uuid.New().String(),
		})
		if err == nil {
			t.Fatal("expected error for non-existing state")
		}
	})
}

func TestCharacterStateHandler_GetLatest(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewCharacterStateServiceClient(conn)
	ctx, characterID := setupCharacterStateTest(t, conn)

	// Create states at chapters 1, 3, 5
	emotions := map[int32]string{1: "hopeful", 3: "cracking", 5: "broken"}
	for ch, emotion := range emotions {
		_, err := client.CreateCharacterState(ctx, &pb.CreateCharacterStateRequest{
			CharacterId:   characterID,
			ChapterNumber: ch,
			Emotion:       emotion,
		})
		if err != nil {
			t.Fatalf("failed to create state at chapter %d: %v", ch, err)
		}
	}

	t.Run("as of chapter 4 returns chapter 3", func(t *testing.T) {
		resp, err := client.GetLatestCharacterState(ctx, &pb.GetLatestCharacterStateRequest{
			CharacterId: characterID,
			AsOfChapter: 4,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterState.ChapterNumber != 3 {
			t.Errorf("expected chapter 3 (latest <= 4), got %d", resp.CharacterState.ChapterNumber)
		}
	})

	t.Run("as of chapter 5 returns chapter 5", func(t *testing.T) {
		resp, err := client.GetLatestCharacterState(ctx, &pb.GetLatestCharacterStateRequest{
			CharacterId: characterID,
			AsOfChapter: 5,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CharacterState.ChapterNumber != 5 {
			t.Errorf("expected chapter 5, got %d", resp.CharacterState.ChapterNumber)
		}
	})
}

func TestCharacterStateHandler_Delete(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := pb.NewCharacterStateServiceClient(conn)
	ctx, characterID := setupCharacterStateTest(t, conn)

	createResp, err := client.CreateCharacterState(ctx, &pb.CreateCharacterStateRequest{
		CharacterId:   characterID,
		ChapterNumber: 1,
		Emotion:       "to be deleted",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = client.DeleteCharacterState(ctx, &pb.DeleteCharacterStateRequest{
		Id: createResp.CharacterState.Id,
	})
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = client.GetCharacterState(ctx, &pb.GetCharacterStateRequest{
		Id: createResp.CharacterState.Id,
	})
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}
