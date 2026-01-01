package ingest

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestStoryUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	story := &grpcclient.Story{
		ID:       storyID,
		TenantID: tenantID,
		Title:    "Test Story",
		Status:   "draft",
	}
	mockClient.AddStory(story)

	useCase := NewIngestStoryUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestStoryInput{
		TenantID: tenantID,
		StoryID:  storyID,
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

func TestIngestStoryUseCase_Execute_StoryNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestStoryUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestStoryInput{
		TenantID: tenantID,
		StoryID:  storyID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when story not found")
	}
}

func TestIngestStoryUseCase_Execute_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	story := &grpcclient.Story{
		ID:       storyID,
		TenantID: tenantID,
		Title:    "Test Story",
		Status:   "draft",
	}
	mockClient.AddStory(story)

	// Create existing document
	existingDoc := memory.NewDocument(tenantID, memory.SourceTypeStory, storyID, "Old Title", "Old Content")
	existingDoc.ID = uuid.New()
	mockDocRepo.Create(ctx, existingDoc)

	useCase := NewIngestStoryUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestStoryInput{
		TenantID: tenantID,
		StoryID:  storyID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID != existingDoc.ID {
		t.Errorf("Expected DocumentID %v, got %v", existingDoc.ID, output.DocumentID)
	}
}

func TestIngestStoryUseCase_Execute_EmbeddingError(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	story := &grpcclient.Story{
		ID:       storyID,
		TenantID: tenantID,
		Title:    "Test Story",
		Status:   "draft",
	}
	mockClient.AddStory(story)

	// Set embedding error
	mockEmbedder.SetEmbedFunc(func(text string) ([]float32, error) {
		return nil, errors.New("embedding error")
	})

	useCase := NewIngestStoryUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestStoryInput{
		TenantID: tenantID,
		StoryID:  storyID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when embedding fails")
	}
}

func TestIngestStoryUseCase_buildStoryContent(t *testing.T) {
	story := &grpcclient.Story{
		Title:  "Test Story",
		Status: "draft",
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestStoryUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildStoryContent(story)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !contains(content, "Test Story") {
		t.Error("Expected content to contain story title")
	}
	if !contains(content, "draft") {
		t.Error("Expected content to contain story status")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

