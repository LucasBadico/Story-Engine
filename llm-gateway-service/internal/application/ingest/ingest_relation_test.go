package ingest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestRelationUseCase_Execute_SuccessRelationship(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	worldID := uuid.New()
	relationID := uuid.New()
	charID := uuid.New()
	factionID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	mockClient.AddCharacter(&grpcclient.Character{
		ID:          charID,
		WorldID:     worldID,
		Name:        "Ari",
		Description: "A devoted knight.",
	})
	mockClient.AddFaction(&grpcclient.Faction{
		ID:          factionID,
		WorldID:     worldID,
		Name:        "Order of the Sun",
		Description: "A holy order.",
	})
	mockClient.AddRelation(&grpcclient.EntityRelation{
		ID:           relationID,
		TenantID:     tenantID,
		WorldID:      worldID,
		SourceType:   "character",
		SourceID:     charID,
		TargetType:   "faction",
		TargetID:     factionID,
		RelationType: "member_of",
		Summary:      "Ari belongs to the Order of the Sun.",
	})

	useCase := NewIngestRelationUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	output, err := useCase.Execute(ctx, IngestRelationInput{
		TenantID:   tenantID,
		RelationID: relationID,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
	if output.ChunkCount == 0 {
		t.Error("Expected ChunkCount > 0")
	}

	doc, err := mockDocRepo.GetBySource(ctx, tenantID, memory.SourceTypeRelation, relationID)
	if err != nil {
		t.Fatalf("Expected document to exist: %v", err)
	}
	if doc.Metadata["relation_kind"] != "relationship" {
		t.Fatalf("Expected relation_kind relationship, got %q", doc.Metadata["relation_kind"])
	}
	if doc.Metadata["source_name"] != "Ari" {
		t.Fatalf("Expected source_name Ari, got %q", doc.Metadata["source_name"])
	}
	if doc.Metadata["target_name"] != "Order of the Sun" {
		t.Fatalf("Expected target_name Order of the Sun, got %q", doc.Metadata["target_name"])
	}
}

func TestIngestRelationUseCase_Execute_Citation(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	relationID := uuid.New()
	sceneID := uuid.New()
	charID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	mockClient.AddScene(&grpcclient.Scene{
		ID:       sceneID,
		StoryID:  uuid.New(),
		OrderNum: 1,
		Goal:     "Battle at the gate",
	})
	mockClient.AddCharacter(&grpcclient.Character{
		ID:          charID,
		WorldID:     uuid.New(),
		Name:        "Ari",
		Description: "A devoted knight.",
	})
	mockClient.AddRelation(&grpcclient.EntityRelation{
		ID:           relationID,
		TenantID:     tenantID,
		WorldID:      uuid.New(),
		SourceType:   "scene",
		SourceID:     sceneID,
		TargetType:   "character",
		TargetID:     charID,
		RelationType: "mentions",
		Summary:      "Ari is referenced during the battle.",
	})

	useCase := NewIngestRelationUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	_, err := useCase.Execute(ctx, IngestRelationInput{
		TenantID:   tenantID,
		RelationID: relationID,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	doc, err := mockDocRepo.GetBySource(ctx, tenantID, memory.SourceTypeRelation, relationID)
	if err != nil {
		t.Fatalf("Expected document to exist: %v", err)
	}
	if doc.Metadata["relation_kind"] != "citation" {
		t.Fatalf("Expected relation_kind citation, got %q", doc.Metadata["relation_kind"])
	}
}
