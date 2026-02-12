package ai

import "context"

type LLM interface {
	ExtractSearchIR(ctx context.Context, input string) (SearchIR, error)
	ExplainTrace(ctx context.Context, context string) (string, error)
	ExplainSpan(ctx context.Context, context string) (string, error)
}
