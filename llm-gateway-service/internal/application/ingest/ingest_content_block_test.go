package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestContentBlockUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "This is content.",
	}
	mockClient.AddContentBlock(contentBlock)

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
	if output.ChunkCount != 1 {
		t.Errorf("Expected ChunkCount 1, got %d", output.ChunkCount)
	}
}

func TestIngestContentBlockUseCase_Execute_ContentBlockNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when content block not found")
	}
}

func TestIngestContentBlockUseCase_Execute_WithBeatReference(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID
	sceneID := uuid.New()
	beatID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "This is content.",
	}
	mockClient.AddContentBlock(contentBlock)

	scene := &grpcclient.Scene{
		ID:        sceneID,
		StoryID:   uuid.New(),
		ChapterID: chapterIDPtr,
		OrderNum:  1,
		Goal:      "Test goal",
		TimeRef:   "Morning",
	}
	mockClient.AddScene(scene)

	beat := &grpcclient.Beat{
		ID:       beatID,
		SceneID:  sceneID,
		OrderNum: 1,
		Type:     "setup",
		Intent:   "Introduce character",
		Outcome:  "Character introduced",
	}
	mockClient.AddBeat(beat)

	ref := &grpcclient.ContentBlockReference{
		ID:            uuid.New(),
		ContentBlockID: contentBlockID,
		EntityType:    "beat",
		EntityID:      beatID,
	}
	mockClient.AddContentBlockReference(ref)

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}

	// Verify chunk was created with metadata
	chunks, _ := mockChunkRepo.ListByDocument(ctx, output.DocumentID)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.BeatID == nil || *chunk.BeatID != beatID {
		t.Error("Expected BeatID to be set")
	}
	if chunk.BeatType == nil || *chunk.BeatType != "setup" {
		t.Error("Expected BeatType to be set")
	}
	if chunk.SceneID == nil || *chunk.SceneID != sceneID {
		t.Error("Expected SceneID to be set")
	}
}

func TestIngestContentBlockUseCase_Execute_WithCharacterReferences(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID
	characterID1 := uuid.New()
	characterID2 := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "John and Mary walked together.",
	}
	mockClient.AddContentBlock(contentBlock)

	char1 := &grpcclient.Character{
		ID:   characterID1,
		Name: "John",
	}
	mockClient.AddCharacter(char1)

	char2 := &grpcclient.Character{
		ID:   characterID2,
		Name: "Mary",
	}
	mockClient.AddCharacter(char2)

	ref1 := &grpcclient.ContentBlockReference{
		ID:            uuid.New(),
		ContentBlockID: contentBlockID,
		EntityType:    "character",
		EntityID:      characterID1,
	}
	mockClient.AddContentBlockReference(ref1)

	ref2 := &grpcclient.ContentBlockReference{
		ID:            uuid.New(),
		ContentBlockID: contentBlockID,
		EntityType:    "character",
		EntityID:      characterID2,
	}
	mockClient.AddContentBlockReference(ref2)

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	chunks, _ := mockChunkRepo.ListByDocument(ctx, output.DocumentID)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if len(chunk.Characters) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(chunk.Characters))
	}
}

func TestIngestContentBlockUseCase_Execute_ImageType(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	imageURL := "https://example.com/image.jpg"
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "image",
		Kind:      "final",
		Content:   imageURL,
		Metadata: map[string]interface{}{
			"description": "A beautiful landscape",
			"alt":         "Mountain landscape at sunset",
		},
	}
	mockClient.AddContentBlock(contentBlock)

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
	if output.ChunkCount != 1 {
		t.Errorf("Expected ChunkCount 1, got %d", output.ChunkCount)
	}

	// Verify chunk content includes image URL and metadata
	chunks, _ := mockChunkRepo.ListByDocument(ctx, output.DocumentID)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if !strings.Contains(chunk.Content, imageURL) {
		t.Error("Expected chunk content to contain image URL")
	}
	if !strings.Contains(chunk.Content, "A beautiful landscape") {
		t.Error("Expected chunk content to contain image description")
	}
	if chunk.ContentType == nil || *chunk.ContentType != "image" {
		t.Error("Expected ContentType to be 'image'")
	}
}

func TestIngestContentBlockUseCase_Execute_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "video",
		Kind:      "final",
		Content:   "https://example.com/video.mp4",
	}
	mockClient.AddContentBlock(contentBlock)

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID != uuid.Nil {
		t.Error("Expected DocumentID to be Nil for unsupported type")
	}
	if output.ChunkCount != 0 {
		t.Error("Expected ChunkCount 0 for unsupported type")
	}
}

func TestIngestContentBlockUseCase_Execute_EmbeddingError(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	chapterIDPtr := &chapterID

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	orderNum := 1
	contentBlock := &grpcclient.ContentBlock{
		ID:        contentBlockID,
		ChapterID: chapterIDPtr,
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "This is content.",
	}
	mockClient.AddContentBlock(contentBlock)

	// Set embedding error
	mockEmbedder.SetEmbedFunc(func(text string) ([]float32, error) {
		return nil, errors.New("embedding error")
	})

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestContentBlockInput{
		TenantID:       tenantID,
		ContentBlockID: contentBlockID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when embedding fails")
	}
}

func TestIngestContentBlockUseCase_buildEnrichedContent(t *testing.T) {
	contentBlock := &grpcclient.ContentBlock{
		Type:    "text",
		Kind:    "final",
		Content: "This is content.",
	}

	metadata := &ContentBlockMetadata{
		BeatType:     stringPtr("setup"),
		BeatIntent:   stringPtr("Introduce character"),
		Characters:   []string{"John", "Mary"},
		LocationName: stringPtr("Library"),
		Timeline:     stringPtr("Morning"),
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestContentBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildEnrichedContent(contentBlock, metadata)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !strings.Contains(content, "This is content") {
		t.Error("Expected content to contain content block content")
	}
	if !strings.Contains(content, "setup") {
		t.Error("Expected content to contain beat type")
	}
	if !strings.Contains(content, "John") || !strings.Contains(content, "Mary") {
		t.Error("Expected content to contain characters")
	}
	if !strings.Contains(content, "Library") {
		t.Error("Expected content to contain location")
	}
	if !strings.Contains(content, "Morning") {
		t.Error("Expected content to contain timeline")
	}
}

func stringPtr(s string) *string {
	return &s
}

