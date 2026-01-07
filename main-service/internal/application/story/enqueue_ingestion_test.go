package story

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateStoryUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &CreateStoryUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}

func TestUpdateStoryUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &UpdateStoryUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}
