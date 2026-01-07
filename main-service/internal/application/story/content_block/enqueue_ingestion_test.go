package content_block

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateContentBlockUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &CreateContentBlockUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}

func TestUpdateContentBlockUseCase_EnqueueIngestion_NilQueue(t *testing.T) {
	uc := &UpdateContentBlockUseCase{
		ingestionQueue: nil,
		logger:         &logger.NoOpLogger{},
	}

	uc.enqueueIngestion(context.Background(), uuid.New(), uuid.New())
}
