package tools

import (
	"context"
	"encoding/json"

	"github.com/krewire/krewire/pkg/ai"
)

// Tool defines an executable agent tool.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(ctx context.Context, args json.RawMessage) (any, error)
}

// ParallelTool is an optional interface tools can implement to declare
// whether they are safe for concurrent execution.
type ParallelTool interface {
	ParallelSafe() bool
}

// IsParallelSafe returns true if the tool implements ParallelTool and reports true.
func IsParallelSafe(t Tool) bool {
	if pt, ok := t.(ParallelTool); ok {
		return pt.ParallelSafe()
	}
	return false
}

// ToAITool converts a Tool into the lower-level ai.Tool specification.
func ToAITool(t Tool) ai.Tool {
	return ai.Tool{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters:  t.Parameters(),
		Handler: func(ctx context.Context, raw []byte) (any, error) {
			return t.Execute(ctx, raw)
		},
	}
}
