package scene

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateSceneUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &CreateSceneUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}

func TestUpdateSceneUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &UpdateSceneUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}
