package ingest

import (
	"context"

	"github.com/google/uuid"
)

// MockIngestStoryUseCase is a mock implementation of IngestStoryUseCase for testing
type MockIngestStoryUseCase struct {
	Called bool
	Error  error
	Output *IngestStoryOutput
}

func NewMockIngestStoryUseCase() *MockIngestStoryUseCase {
	return &MockIngestStoryUseCase{
		Output: &IngestStoryOutput{
			DocumentID: uuid.New(),
			ChunkCount: 1,
		},
	}
}

func (m *MockIngestStoryUseCase) Execute(ctx context.Context, input IngestStoryInput) (*IngestStoryOutput, error) {
	m.Called = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Output, nil
}

// MockIngestChapterUseCase is a mock implementation of IngestChapterUseCase for testing
type MockIngestChapterUseCase struct {
	Called bool
	Error  error
	Output *IngestChapterOutput
}

func NewMockIngestChapterUseCase() *MockIngestChapterUseCase {
	return &MockIngestChapterUseCase{
		Output: &IngestChapterOutput{
			DocumentID: uuid.New(),
			ChunkCount: 1,
		},
	}
}

func (m *MockIngestChapterUseCase) Execute(ctx context.Context, input IngestChapterInput) (*IngestChapterOutput, error) {
	m.Called = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Output, nil
}

// MockIngestProseBlockUseCase is a mock implementation of IngestProseBlockUseCase for testing
type MockIngestProseBlockUseCase struct {
	Called bool
	Error  error
	Output *IngestProseBlockOutput
}

func NewMockIngestProseBlockUseCase() *MockIngestProseBlockUseCase {
	return &MockIngestProseBlockUseCase{
		Output: &IngestProseBlockOutput{
			DocumentID: uuid.New(),
			ChunkCount: 1,
		},
	}
}

func (m *MockIngestProseBlockUseCase) Execute(ctx context.Context, input IngestProseBlockInput) (*IngestProseBlockOutput, error) {
	m.Called = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Output, nil
}

