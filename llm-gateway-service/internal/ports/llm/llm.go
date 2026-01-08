package llm

import "context"

// RouterModel generates structured responses for routing decisions.
type RouterModel interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
