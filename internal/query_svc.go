package internal

import (
	"context"
	"iter"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type TraceQueryParams struct {
	ServiceName   string
	OperationName string
	Attributes    pcommon.Map
	StartTimeMin  time.Time
	StartTimeMax  time.Time
	DurationMin   time.Duration
	DurationMax   time.Duration
	SearchDepth   int
}

type TraceReader interface {
	FindTraces(
		ctx context.Context,
		query TraceQueryParams,
	) iter.Seq2[[]ptrace.Traces, error]
}

type QueryService struct {
	reader TraceReader
}

func NewQueryService(r TraceReader) *QueryService {
	return &QueryService{reader: r}
}

func (qs *QueryService) FindTraces(
	ctx context.Context,
	params TraceQueryParams,
) iter.Seq2[[]ptrace.Traces, error] {
	return qs.reader.FindTraces(ctx, params)
}

func TraceMatchesService(t ptrace.Traces, service string) bool {
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

func TraceMatchesMinDuration(t ptrace.Traces, min time.Duration) bool {
	if min == 0 {
		return true
	}
	rs := t.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				dur := span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())
				if dur >= min {
					return true
				}
			}
		}
	}
	return false
}
