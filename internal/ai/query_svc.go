package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaeger-ai-assist-prototype/internal"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type AIQueryService struct {
	LLM   LLM
	Query *internal.QueryService
}

func (s *AIQueryService) Search(
	ctx context.Context,
	text string,
) (SearchResult, error) {
	ir, err := s.LLM.ExtractSearchIR(ctx, text)
	if err != nil {
		return SearchResult{}, err
	}

	if err := ValidateSearchIR(ir); err != nil {
		return SearchResult{}, err
	}

	qp, err := MapIRToQueryParams(ir)
	if err != nil {
		return SearchResult{}, err
	}

	iter := s.Query.FindTraces(ctx, qp)

	var result SearchResult

	iter(func(batch []ptrace.Traces, err error) bool {
		if err != nil {
			result = SearchResult{}
			return false
		}

		result.Traces = append(result.Traces, batch...)
		return true
	})

	return result, nil
}

func (s *AIQueryService) ExplainTrace(
	ctx context.Context,
	trace ptrace.Traces,
) (string, error) {
	ctxData := buildTraceContext(trace)
	return s.LLM.ExplainTrace(ctx, ctxData)
}

func (s *AIQueryService) ExplainSpan(
	ctx context.Context,
	span ptrace.Span,
	serviceName string,
) (string, error) {
	ctxData := buildSpanContext(span, serviceName)
	return s.LLM.ExplainSpan(ctx, ctxData)
}

func buildTraceContext(t ptrace.Traces) string {
	b := strings.Builder{}
	b.WriteString("Trace Analysis Context:\n")

	rs := t.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		res := rs.At(i)
		svc, _ := res.Resource().Attributes().Get("service.name")

		ss := res.ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				// Basic Info
				b.WriteString(fmt.Sprintf("\n[Span] Name: %s | Service: %s | Kind: %s\n",
					span.Name(), svc.Str(), span.Kind().String()))
				b.WriteString(fmt.Sprintf("  Duration: %dms\n",
					span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime()).Milliseconds()))

				// CRITICAL: Include HTTP and Error Attributes
				span.Attributes().Range(func(k string, v pcommon.Value) bool {
					if strings.HasPrefix(k, "http.") || k == "db.system" || k == "error" {
						b.WriteString(fmt.Sprintf("  Tag: %s = %s\n", k, v.AsString()))
					}
					return true
				})

				// Status Detail
				if span.Status().Code() == ptrace.StatusCodeError {
					b.WriteString(fmt.Sprintf("  Status: ERROR (%s)\n", span.Status().Message()))
				}
			}
		}
	}
	return b.String()
}

func buildSpanContext(span ptrace.Span, serviceName string) string {
	b := strings.Builder{}
	b.WriteString("### Detailed Span Analysis\n")
	b.WriteString(fmt.Sprintf("Operation: %s\n", span.Name()))
	b.WriteString(fmt.Sprintf("Service: %s\n", serviceName))
	b.WriteString(fmt.Sprintf("Kind: %s\n", span.Kind().String()))

	duration := span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())
	b.WriteString(fmt.Sprintf("Duration: %v\n", duration))

	// 1. Attributes: Keep technical context
	b.WriteString("\nAttributes:\n")
	span.Attributes().Range(func(k string, v pcommon.Value) bool {
		// Include HTTP, DB, RPC, Messaging, and Error tags
		isTechnical := strings.HasPrefix(k, "http.") ||
			strings.HasPrefix(k, "db.") ||
			strings.HasPrefix(k, "rpc.") ||
			strings.HasPrefix(k, "messaging.") ||
			k == "error" || k == "exception.type"

		if isTechnical {
			b.WriteString(fmt.Sprintf("- %s: %s\n", k, valueToString(v)))
		}
		return true
	})

	// 2. Status: Ensure errors are loud and clear
	if span.Status().Code() == ptrace.StatusCodeError {
		b.WriteString("\n[!] Status: ERROR\n")
		b.WriteString(fmt.Sprintf("[!] Error Message: %s\n", span.Status().Message()))
	}

	// 3. Events: This is the "Logs" section from your Jaeger fixtures
	if span.Events().Len() > 0 {
		b.WriteString("\nEvents (Internal Logs):\n")
		for i := 0; i < span.Events().Len(); i++ {
			event := span.Events().At(i)
			offset := event.Timestamp().AsTime().Sub(span.StartTimestamp().AsTime())
			b.WriteString(fmt.Sprintf("- T+%v: %s\n", offset, event.Name()))

			// Include event-specific attributes (like stack traces)
			event.Attributes().Range(func(k string, v pcommon.Value) bool {
				b.WriteString(fmt.Sprintf("  └ %s: %s\n", k, v.AsString()))
				return true
			})
		}
	}

	return b.String()
}

func getService(span ptrace.Span) string {
	if v, ok := span.Attributes().Get("service.name"); ok {
		return v.Str()
	}
	return "unknown"
}

func valueToString(v pcommon.Value) string {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		return v.Str()
	case pcommon.ValueTypeBool:
		return fmt.Sprintf("%v", v.Bool())
	case pcommon.ValueTypeInt:
		return fmt.Sprintf("%d", v.Int())
	case pcommon.ValueTypeDouble:
		return fmt.Sprintf("%f", v.Double())
	default:
		return v.AsString()
	}
}
