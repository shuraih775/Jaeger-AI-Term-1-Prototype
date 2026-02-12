package ai

import (
	"context"
	"testing"

	"github.com/jaeger-ai-assist-prototype/internal"
	"github.com/jaeger-ai-assist-prototype/internal/synthetic"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type FakeLLM struct {
	IR  SearchIR
	Err error
}

func (f *FakeLLM) ExtractSearchIR(
	ctx context.Context,
	input string,
) (SearchIR, error) {
	if f.Err != nil {
		return SearchIR{}, f.Err
	}
	return f.IR, nil
}

func TestAIQueryService_Search_ServiceFilter(t *testing.T) {
	traces := synthetic.GenerateTraces(5)
	reader := synthetic.NewSyntheticTraceReader(traces)
	querySvc := internal.NewQueryService(reader)

	fakeLLM := &FakeLLM{
		IR: SearchIR{
			Service: strptr("payment-service"),
		},
	}

	aiSvc := &AIQueryService{
		LLM:   fakeLLM,
		Query: querySvc,
	}

	result, err := aiSvc.Search(context.Background(), "ignored input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Traces) == 0 {
		t.Fatalf("expected at least one matching trace")
	}

	for _, trace := range result.Traces {
		if !traceContainsService(trace, "payment-service") {
			t.Fatalf("returned trace does not contain expected service")
		}
	}
}

func traceContainsService(t ptrace.Traces, service string) bool {
	rs := t.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				if v, ok := spans.At(k).Attributes().Get("service.name"); ok {
					if v.Str() == service {
						return true
					}
				}
			}
		}
	}
	return false
}

func TestAIQueryService_Search_InvalidIR(t *testing.T) {
	traces := synthetic.GenerateTraces(3)
	reader := synthetic.NewSyntheticTraceReader(traces)
	querySvc := internal.NewQueryService(reader)

	fakeLLM := &FakeLLM{
		IR: SearchIR{
			MinDurationMs: int64ptr(-10),
		},
	}

	aiSvc := &AIQueryService{
		LLM:   fakeLLM,
		Query: querySvc,
	}

	_, err := aiSvc.Search(context.Background(), "ignored")
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

func TestAIQueryService_Search_DurationFilter(t *testing.T) {
	traces := synthetic.GenerateTraces(5)
	reader := synthetic.NewSyntheticTraceReader(traces)
	querySvc := internal.NewQueryService(reader)

	fakeLLM := &FakeLLM{
		IR: SearchIR{
			MinDurationMs: int64ptr(50),
		},
	}

	aiSvc := &AIQueryService{
		LLM:   fakeLLM,
		Query: querySvc,
	}

	result, err := aiSvc.Search(context.Background(), "ignored")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Traces) == 0 {
		t.Fatalf("expected traces matching duration filter")
	}
}
