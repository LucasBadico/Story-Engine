package beat

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateBeatUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &CreateBeatUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}

func TestUpdateBeatUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &UpdateBeatUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}
