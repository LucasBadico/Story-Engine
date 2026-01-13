package events

import (
	"context"
	"time"
)

type ExtractionEventLogger interface {
	Emit(ctx context.Context, event ExtractionEvent)
}

type ExtractionEvent struct {
	Type      string                 `json:"type"`
	Phase     string                 `json:"phase,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type NoopExtractionEventLogger struct{}

func (n NoopExtractionEventLogger) Emit(_ context.Context, _ ExtractionEvent) {}

func NormalizeEventLogger(logger ExtractionEventLogger) ExtractionEventLogger {
	if logger == nil {
		return NoopExtractionEventLogger{}
	}
	return logger
}

func EmitEvent(ctx context.Context, logger ExtractionEventLogger, event ExtractionEvent) {
	if logger == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	logger.Emit(ctx, event)
}
