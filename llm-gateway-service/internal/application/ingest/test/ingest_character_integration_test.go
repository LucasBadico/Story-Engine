//go:build integration

package ingest_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/adapters/db/postgres"
	grpcadapter "github.com/story-engine/llm-gateway-service/internal/adapters/grpc"
	"github.com/story-engine/llm-gateway-service/internal/application/ingest"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	characterpb "github.com/story-engine/main-service/proto/character"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type recordingSummaryGenerator struct {
	lastInput ingest.GenerateSummaryInput
}

func (g *recordingSummaryGenerator) Execute(ctx context.Context, input ingest.GenerateSummaryInput) (ingest.GenerateSummaryOutput, error) {
	g.lastInput = input
	return ingest.GenerateSummaryOutput{
		Summaries: []string{"Summary: " + input.Name},
	}, nil
}

func TestIngestCharacterUseCase_Integration_SummaryChunk(t *testing.T) {
	if strings.TrimSpace(os.Getenv("MAIN_SERVICE_TESTS_ENABLED")) == "" {
		t.Skip("MAIN_SERVICE_TESTS_ENABLED not set; skipping main-service integration tests")
	}

	cfg := config.Load()
	addr := strings.TrimSpace(cfg.MainService.GRPCAddr)
	if addr == "" {
		t.Skip("MAIN_SERVICE_GRPC_ADDR not set; skipping")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Skipf("main-service not reachable at %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	tenantClient := tenantpb.NewTenantServiceClient(conn)
	worldClient := worldpb.NewWorldServiceClient(conn)
	characterClient := characterpb.NewCharacterServiceClient(conn)
	contentBlockClient := contentblockpb.NewContentBlockServiceClient(conn)
	contentBlockRefClient := contentblockpb.NewContentBlockReferenceServiceClient(conn)

	tenantName := "Ingest Integration Tenant " + uuid.NewString()
	tenantID := strings.TrimSpace(os.Getenv("TEST_TENANT_ID"))
	if tenantID == "" {
		tenantResp, err := tenantClient.CreateTenant(ctx, &tenantpb.CreateTenantRequest{
			Name: tenantName,
		})
		if err != nil {
			t.Fatalf("create tenant: %v", err)
		}
		tenantID = tenantResp.Tenant.Id
	}

	tenantCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("tenant_id", tenantID))

	worldResp, err := worldClient.CreateWorld(tenantCtx, &worldpb.CreateWorldRequest{
		TenantId: tenantID,
		Name:     "Integration World",
	})
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	characterResp, err := characterClient.CreateCharacter(tenantCtx, &characterpb.CreateCharacterRequest{
		WorldId:     worldResp.World.Id,
		Name:        "Aria",
		Description: "A mage of the Crimson Order.",
	})
	if err != nil {
		t.Fatalf("create character: %v", err)
	}

	contentBlockText := "Aria protected the Obsidian Tower during the siege."
	contentBlockResp, err := contentBlockRefClient.ListContentBlocksByEntity(tenantCtx, &contentblockpb.ListContentBlocksByEntityRequest{
		EntityType: "character",
		EntityId:   characterResp.Character.Id,
	})
	if err != nil {
		t.Fatalf("list content blocks by entity: %v", err)
	}

	var contentBlockID string
	for _, block := range contentBlockResp.ContentBlocks {
		if block != nil && strings.TrimSpace(block.Content) != "" {
			contentBlockID = block.Id
			break
		}
	}

	if contentBlockID == "" {
		orderNum := int32(1)
		createdResp, err := contentBlockClient.CreateContentBlock(tenantCtx, &contentblockpb.CreateContentBlockRequest{
			OrderNum: &orderNum,
			Type:     "text",
			Kind:     "draft",
			Content:  contentBlockText,
		})
		if err != nil {
			t.Fatalf("create content block: %v", err)
		}
		contentBlockID = createdResp.ContentBlock.Id

		_, err = contentBlockRefClient.CreateContentBlockReference(tenantCtx, &contentblockpb.CreateContentBlockReferenceRequest{
			ContentBlockId: contentBlockID,
			EntityType:     "character",
			EntityId:       characterResp.Character.Id,
		})
		if err != nil {
			t.Fatalf("create content block reference: %v", err)
		}
	}

	db, cleanup := postgres.SetupTestDB(t)
	t.Cleanup(cleanup)

	documentRepo := postgres.NewDocumentRepository(db)
	chunkRepo := postgres.NewChunkRepository(db)

	mainClient, err := grpcadapter.NewMainServiceClient(addr)
	if err != nil {
		t.Fatalf("create main service client: %v", err)
	}
	t.Cleanup(func() { _ = mainClient.Close() })

	embedder := embeddings.NewMockEmbedder(768) // TODO: add an dinamic variable/central variable
	summaryGen := &recordingSummaryGenerator{}

	useCase := ingest.NewIngestCharacterUseCase(mainClient, documentRepo, chunkRepo, embedder, logger.New())
	useCase.SetSummaryGenerator(summaryGen)

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		t.Fatalf("parse tenant id: %v", err)
	}
	characterUUID, err := uuid.Parse(characterResp.Character.Id)
	if err != nil {
		t.Fatalf("parse character id: %v", err)
	}

	output, err := useCase.Execute(tenantCtx, ingest.IngestCharacterInput{
		TenantID:    tenantUUID,
		CharacterID: characterUUID,
	})
	if err != nil {
		t.Fatalf("ingest character: %v", err)
	}

	if len(summaryGen.lastInput.Contents) == 0 {
		t.Fatalf("expected summary contents, got none")
	}
	foundBlock := false
	for _, value := range summaryGen.lastInput.Contents {
		if strings.Contains(value, contentBlockText) {
			foundBlock = true
			break
		}
	}
	if !foundBlock {
		t.Fatalf("expected summary contents to include content block text")
	}

	chunks, err := chunkRepo.ListByDocument(ctx, output.DocumentID)
	if err != nil {
		t.Fatalf("list chunks: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got none")
	}

	foundSummary := false
	for _, chunk := range chunks {
		if chunk.ChunkType != nil && *chunk.ChunkType == "summary" {
			foundSummary = true
			if chunk.EmbedText == nil || strings.TrimSpace(*chunk.EmbedText) == "" {
				t.Fatalf("summary chunk missing embed_text")
			}
			continue
		}
		if chunk.EmbedText != nil {
			t.Fatalf("raw chunk should not have embed_text")
		}
	}
	if !foundSummary {
		t.Fatalf("expected at least one summary chunk")
	}
}
