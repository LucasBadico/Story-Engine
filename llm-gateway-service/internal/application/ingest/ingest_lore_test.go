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

func TestIngestLoreUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	loreID := uuid.New()
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

	lore := &grpcclient.Lore{
		ID:          loreID,
		WorldID:     worldID,
		Name:        "Test Lore",
		Description: "A test lore",
		Rules:       "Test rules",
	}
	mockClient.AddLore(lore)

	useCase := NewIngestLoreUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestLoreInput{
		TenantID: tenantID,
		LoreID:   loreID,
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

func TestIngestLoreUseCase_Execute_LoreNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	loreID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestLoreUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestLoreInput{
		TenantID: tenantID,
		LoreID:   loreID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when lore not found")
	}
}

func TestIngestLoreUseCase_Execute_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	loreID := uuid.New()
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

	lore := &grpcclient.Lore{
		ID:      loreID,
		WorldID: worldID,
		Name:    "Test Lore",
	}
	mockClient.AddLore(lore)

	// Create existing document
	existingDoc := memory.NewDocument(tenantID, memory.SourceTypeLore, loreID, "Old Title", "Old Content")
	existingDoc.ID = uuid.New()
	mockDocRepo.Create(ctx, existingDoc)

	useCase := NewIngestLoreUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestLoreInput{
		TenantID: tenantID,
		LoreID:   loreID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID != existingDoc.ID {
		t.Errorf("Expected DocumentID %v, got %v", existingDoc.ID, output.DocumentID)
	}
}

func TestIngestLoreUseCase_Execute_EmbeddingError(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	loreID := uuid.New()
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

	lore := &grpcclient.Lore{
		ID:      loreID,
		WorldID: worldID,
		Name:    "Test Lore",
	}
	mockClient.AddLore(lore)

	// Set embedding error
	mockEmbedder.SetEmbedFunc(func(text string) ([]float32, error) {
		return nil, errors.New("embedding error")
	})

	useCase := NewIngestLoreUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestLoreInput{
		TenantID: tenantID,
		LoreID:   loreID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when embedding fails")
	}
}

func TestIngestLoreUseCase_buildLoreContent(t *testing.T) {
	worldID := uuid.New()
	world := &grpcclient.World{
		ID:   worldID,
		Name: "Test World",
	}

	category := "magic"
	lore := &grpcclient.Lore{
		Name:         "Test Lore",
		Category:     &category,
		Description:  "Test description",
		Rules:        "Test rules",
		Limitations:  "Test limitations",
		Requirements: "Test requirements",
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestLoreUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildLoreContent(lore, world)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !strings.Contains(content, "Test Lore") {
		t.Error("Expected content to contain lore name")
	}
	if !strings.Contains(content, "magic") {
		t.Error("Expected content to contain lore category")
	}
	if !strings.Contains(content, "Test description") {
		t.Error("Expected content to contain lore description")
	}
	if !strings.Contains(content, "Test World") {
		t.Error("Expected content to contain world name")
	}
}

