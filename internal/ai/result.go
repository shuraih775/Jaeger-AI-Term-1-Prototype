package ai

import "go.opentelemetry.io/collector/pdata/ptrace"

type SearchResult struct {
	Traces []ptrace.Traces
}
