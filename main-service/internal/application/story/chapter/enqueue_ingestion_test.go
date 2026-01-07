package chapter

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateChapterUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &CreateChapterUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}

func TestUpdateChapterUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &UpdateChapterUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}
