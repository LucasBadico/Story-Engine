package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestFactionUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	factionID := uuid.New()
	worldID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	world := &grpcclient.World{
		ID:          worldID,
		TenantID:    tenantID,
		Name:        "Test World",
		Description: "A test world",
		Genre:       "fantasy",
	}
	mockClient.AddWorld(world)

	faction := &grpcclient.Faction{
		ID:          factionID,
		WorldID:     worldID,
		Name:        "Test Faction",
		Description: "A test faction",
		Beliefs:     "Test beliefs",
	}
	mockClient.AddFaction(faction)

	useCase := NewIngestFactionUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestFactionInput{
		TenantID:  tenantID,
		FactionID: factionID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
	if output.ChunkCount == 0 {
		t.Error("Expected ChunkCount > 0")
	}
}

func TestIngestFactionUseCase_Execute_FactionNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	factionID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestFactionUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestFactionInput{
		TenantID:  tenantID,
		FactionID: factionID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when faction not found")
	}
}

func TestIngestFactionUseCase_Execute_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	factionID := uuid.New()
	worldID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	world := &grpcclient.World{
		ID:       worldID,
		TenantID: tenantID,
		Name:     "Test World",
	}
	mockClient.AddWorld(world)

	faction := &grpcclient.Faction{
		ID:      factionID,
		WorldID: worldID,
		Name:    "Test Faction",
	}
	mockClient.AddFaction(faction)

	// Create existing document
	existingDoc := memory.NewDocument(tenantID, memory.SourceTypeFaction, factionID, "Old Title", "Old Content")
	existingDoc.ID = uuid.New()
	mockDocRepo.Create(ctx, existingDoc)

	useCase := NewIngestFactionUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestFactionInput{
		TenantID:  tenantID,
		FactionID: factionID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID != existingDoc.ID {
		t.Errorf("Expected DocumentID %v, got %v", existingDoc.ID, output.DocumentID)
	}
}

func TestIngestFactionUseCase_Execute_EmbeddingError(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	factionID := uuid.New()
	worldID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	world := &grpcclient.World{
		ID:       worldID,
		TenantID: tenantID,
		Name:     "Test World",
	}
	mockClient.AddWorld(world)

	faction := &grpcclient.Faction{
		ID:      factionID,
		WorldID: worldID,
		Name:    "Test Faction",
	}
	mockClient.AddFaction(faction)

	// Set embedding error
	mockEmbedder.SetEmbedFunc(func(text string) ([]float32, error) {
		return nil, errors.New("embedding error")
	})

	useCase := NewIngestFactionUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestFactionInput{
		TenantID:  tenantID,
		FactionID: factionID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when embedding fails")
	}
}

func TestIngestFactionUseCase_buildFactionContent(t *testing.T) {
	worldID := uuid.New()
	world := &grpcclient.World{
		ID:   worldID,
		Name: "Test World",
	}

	factionType := "religion"
	faction := &grpcclient.Faction{
		Name:        "Test Faction",
		Type:        &factionType,
		Description: "Test description",
		Beliefs:     "Test beliefs",
		Structure:   "Test structure",
		Symbols:     "Test symbols",
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestFactionUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildFactionContent(faction, world)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !strings.Contains(content, "Test Faction") {
		t.Error("Expected content to contain faction name")
	}
	if !strings.Contains(content, "religion") {
		t.Error("Expected content to contain faction type")
	}
	if !strings.Contains(content, "Test description") {
		t.Error("Expected content to contain faction description")
	}
	if !strings.Contains(content, "Test World") {
		t.Error("Expected content to contain world name")
	}
}

