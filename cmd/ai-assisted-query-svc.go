package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/jaeger-ai-assist-prototype/internal"
	"github.com/jaeger-ai-assist-prototype/internal/ai"
	"github.com/jaeger-ai-assist-prototype/internal/llm"
	"github.com/jaeger-ai-assist-prototype/internal/llm/langchain"
	"github.com/jaeger-ai-assist-prototype/internal/synthetic"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")

	explainTraceIdx := flag.Int("explaintrace", -1, "explain a whole trace by index")
	explainSpanTraceIdx := flag.Int("explainspan-trace", -1, "trace index for span explanation")
	explainSpanIdx := flag.Int("explainspan", -1, "span index within trace")

	flag.Parse()

	if flag.NArg() != 1 && *explainTraceIdx < 0 && *explainSpanIdx < 0 {
		fmt.Println(`usage:
  ai-query -config config.yaml "natural language query"
  ai-query -config config.yaml --explaintrace 1 
  ai-query -config config.yaml --explainspan-trace 1 --explainspan 4 
  `)
		os.Exit(1)
	}

	queryText := ""
	if flag.NArg() == 1 {
		queryText = flag.Arg(0)
	}

	// --- Load config ---
	cfg, err := langchain.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// --- LLM factory ---
	model, err := llm.NewLLM(cfg.LLM)
	if err != nil {
		log.Fatalf("LLM init failed: %v", err)
	}

	extractor := langchain.NewSearchExtractor(model)

	// --- Synthetic backend ---
	traces, err := synthetic.LoadTracesFromFile("./traces_bench.json")
	if err != nil {
		log.Fatal(err)
	}
	reader := synthetic.NewSyntheticTraceReader(traces)
	querySvc := internal.NewQueryService(reader)

	// --- AI query service ---
	aiSvc := &ai.AIQueryService{
		LLM:   extractor,
		Query: querySvc,
	}

	ctx := context.Background()

	// CASE 1: Explain whole trace
	if *explainTraceIdx >= 0 {
		if *explainTraceIdx >= len(traces) {
			log.Fatalf("trace index out of range")
		}

		fmt.Printf("=== EXPLAIN TRACE %d ===\n", *explainTraceIdx)
		explanation, err := aiSvc.ExplainTrace(ctx, traces[*explainTraceIdx])
		if err != nil {
			log.Fatalf("explain trace failed: %v", err)
		}

		fmt.Println(explanation)
		return
	}

	// CASE 2: Explain single span
	if *explainSpanIdx >= 0 {
		if *explainSpanTraceIdx < 0 {
			log.Fatalf("--explainspan requires --explainspan-trace")
		}

		if *explainSpanTraceIdx >= len(traces) {
			log.Fatalf("trace index out of range")
		}

		trace := traces[*explainSpanTraceIdx]

		span, svcName := pickSpan(trace, *explainSpanIdx)
		if span == nil {
			log.Fatalf("span index out of range")
		}

		fmt.Printf("=== EXPLAIN SPAN %d (TRACE %d) SERVICE: %s ===\n",
			*explainSpanIdx, *explainSpanTraceIdx, svcName)

		// Pass svcName to the service
		explanation, err := aiSvc.ExplainSpan(ctx, *span, svcName)
		if err != nil {
			log.Fatalf("explain span failed: %v", err)
		}

		fmt.Println(explanation)
		return
	}

	// CASE 3: Default → Natural language search
	result, err := aiSvc.Search(ctx, queryText)
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	fmt.Println("=== SEARCH RESULTS ===")
	fmt.Printf("Traces returned: %d\n\n", len(result.Traces))

	for i, trace := range result.Traces {
		fmt.Printf("Trace #%d\n", i+1)
		printTraceSummary(trace)
		fmt.Println()
	}
}

func pickSpan(t ptrace.Traces, idx int) (*ptrace.Span, string) {
	count := 0
	rss := t.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		serviceName := "unknown"
		if sn, ok := rs.Resource().Attributes().Get("service.name"); ok {
			serviceName = sn.Str()
		}

		ss := rs.ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				if count == idx {
					s := spans.At(k)
					return &s, serviceName
				}
				count++
			}
		}
	}
	return nil, ""
}

func printTraceSummary(t ptrace.Traces) {
	rs := t.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				fmt.Printf(
					"trace=%s span=%s service=%s error=%v\n",
					span.TraceID().String(),
					span.Name(),
					getService(span),
					isError(span),
				)
			}
		}
	}
}
func getService(span ptrace.Span) string {
	if v, ok := span.Attributes().Get("service.name"); ok {
		return v.Str()
	}
	return ""
}

func isError(span ptrace.Span) bool {
	if v, ok := span.Attributes().Get("error"); ok {
		return v.Bool()
	}
	return false
}
