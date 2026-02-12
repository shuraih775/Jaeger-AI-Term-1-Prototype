package synthetic

import (
	"context"

	"iter"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/jaeger-ai-assist-prototype/internal"
)

type SyntheticTraceReader struct {
	traces []ptrace.Traces
}

func NewSyntheticTraceReader(traces []ptrace.Traces) *SyntheticTraceReader {
	return &SyntheticTraceReader{traces: traces}
}

func (r *SyntheticTraceReader) FindTraces(
	ctx context.Context,
	query internal.TraceQueryParams,
) iter.Seq2[[]ptrace.Traces, error] {
	return func(yield func([]ptrace.Traces, error) bool) {
		for _, t := range r.traces {
			if ctx.Err() != nil {
				yield(nil, ctx.Err())
				return
			}

			if query.ServiceName != "" && !internal.TraceMatchesService(t, query.ServiceName) {
				continue
			}

			if !internal.TraceMatchesMinDuration(t, query.DurationMin) {
				continue
			}

			if !yield([]ptrace.Traces{t}, nil) {
				return
			}
		}
	}
}
