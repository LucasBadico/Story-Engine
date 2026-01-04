package ingest

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestChapterUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	chapterID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	chapter := &grpcclient.Chapter{
		ID:      chapterID,
		StoryID: storyID,
		Number:  1,
		Title:   "Chapter 1",
		Status:  "draft",
	}
	mockClient.AddChapter(chapter)

	orderNum := 1
	chapterIDPtr := &chapterID
	contentBlock := &grpcclient.ContentBlock{
		ID:        uuid.New(),
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "This is the content.",
	}
	mockClient.AddContentBlock(contentBlock)

	useCase := NewIngestChapterUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestChapterInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
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

func TestIngestChapterUseCase_Execute_ChapterNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	chapterID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestChapterUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestChapterInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when chapter not found")
	}
}

func TestIngestChapterUseCase_Execute_NoContentBlocks(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	chapterID := uuid.New()
	storyID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	chapter := &grpcclient.Chapter{
		ID:      chapterID,
		StoryID: storyID,
		Number:  1,
		Title:   "Chapter 1",
		Status:  "draft",
	}
	mockClient.AddChapter(chapter)

	useCase := NewIngestChapterUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestChapterInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should still create document even without content blocks
	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
}

func TestIngestChapterUseCase_buildChapterContent(t *testing.T) {
	chapterID := uuid.New()
	storyID := uuid.New()

	chapter := &grpcclient.Chapter{
		ID:      chapterID,
		StoryID: storyID,
		Number:  1,
		Title:   "Chapter 1",
		Status:  "draft",
	}

	orderNum1 := 1
	orderNum2 := 2
	chapterIDPtr := &chapterID
	contentBlocks := []*grpcclient.ContentBlock{
		{
			ID:        uuid.New(),
			ChapterID: chapterIDPtr,
			OrderNum:  &orderNum1,
			Type:      "text",
			Content:   "First paragraph.",
		},
		{
			ID:        uuid.New(),
			ChapterID: chapterIDPtr,
			OrderNum:  &orderNum2,
			Type:      "text",
			Content:   "Second paragraph.",
		},
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestChapterUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildChapterContent(chapter, contentBlocks)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !strings.Contains(content, "Chapter 1") {
		t.Error("Expected content to contain chapter title")
	}
	if !strings.Contains(content, "draft") {
		t.Error("Expected content to contain chapter status")
	}
	if !strings.Contains(content, "First paragraph") {
		t.Error("Expected content to contain first content block")
	}
	if !strings.Contains(content, "Second paragraph") {
		t.Error("Expected content to contain second content block")
	}
}
