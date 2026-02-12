package synthetic

import (
	"encoding/json"
	"fmt"
	"os"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

// LoadTracesFromFile reads the benchmark JSON and returns a slice of ptrace.Traces
func LoadTracesFromFile(path string) ([]ptrace.Traces, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// The file is a JSON array of OTel trace objects
	var rawTraces []json.RawMessage
	if err := json.Unmarshal(data, &rawTraces); err != nil {
		return nil, fmt.Errorf("failed to unmarshal array: %w", err)
	}

	unmarshaler := &ptrace.JSONUnmarshaler{}
	out := make([]ptrace.Traces, 0, len(rawTraces))

	for i, raw := range rawTraces {
		t, err := unmarshaler.UnmarshalTraces(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal trace at index %d: %w", i, err)
		}
		out = append(out, t)
	}

	return out, nil
}

// import (
// 	"encoding/json"
// 	"math/rand"
// 	"os"
// 	"time"

// 	"go.opentelemetry.io/collector/pdata/pcommon"
// 	"go.opentelemetry.io/collector/pdata/ptrace"
// )

// func main() {
// 	// Fixed seed for reproducibility
// 	r := rand.New(rand.NewSource(42))
// 	count := 100 // Total number of traces to generate

// 	marshaler := ptrace.JSONMarshaler{}
// 	var allTraces []interface{}

// 	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

// 	for i := 0; i < count; i++ {
// 		td := ptrace.NewTraces()
// 		tid := fixedTraceID(i + 1)

// 		// Rotate through 3 stories
// 		switch i % 3 {
// 		case 0:
// 			generateCheckoutStory(td, tid, baseTime.Add(time.Duration(i)*time.Second), r)
// 		case 1:
// 			generateSearchStory(td, tid, baseTime.Add(time.Duration(i)*time.Second), r)
// 		case 2:
// 			generateCatalogStory(td, tid, baseTime.Add(time.Duration(i)*time.Second), r)
// 		}

// 		// Convert to raw JSON map for storage
// 		buf, _ := marshaler.MarshalTraces(td)
// 		var raw map[string]interface{}
// 		json.Unmarshal(buf, &raw)
// 		allTraces = append(allTraces, raw)
// 	}

// 	f, _ := os.Create("traces_bench.json")
// 	encoder := json.NewEncoder(f)
// 	encoder.SetIndent("", "  ")
// 	encoder.Encode(allTraces)
// }

// // --- Story Logic ---

// func generateCheckoutStory(td ptrace.Traces, tid pcommon.TraceID, start time.Time, r *rand.Rand) {
// 	fe := addSpan(td, "frontend", "POST /checkout", tid, pcommon.SpanID{}, ptrace.SpanKindServer, start, 300)
// 	fe.Attributes().PutInt("http.status_code", 402)

// 	pay := addSpan(td, "payment-svc", "Authorize", tid, fe.SpanID(), ptrace.SpanKindClient, start.Add(50*time.Millisecond), 150)
// 	pay.Status().SetCode(ptrace.StatusCodeError)
// 	pay.Status().SetMessage("insufficient_funds")
// 	pay.Attributes().PutInt("http.status_code", 402)
// }

// func generateSearchStory(td ptrace.Traces, tid pcommon.TraceID, start time.Time, r *rand.Rand) {
// 	fe := addSpan(td, "frontend", "GET /search", tid, pcommon.SpanID{}, ptrace.SpanKindServer, start, 50)
// 	setHTTP(fe, "GET", 200)

// 	db := addSpan(td, "search-db", "SELECT products", tid, fe.SpanID(), ptrace.SpanKindClient, start.Add(5*time.Millisecond), 20)
// 	db.Attributes().PutStr("db.system", "postgres")
// }

// func generateCatalogStory(td ptrace.Traces, tid pcommon.TraceID, start time.Time, r *rand.Rand) {
// 	fe := addSpan(td, "frontend", "GET /items", tid, pcommon.SpanID{}, ptrace.SpanKindServer, start, 100)
// 	setHTTP(fe, "GET", 200)

// 	cat := addSpan(td, "catalog-svc", "GetItems", tid, fe.SpanID(), ptrace.SpanKindServer, start.Add(10*time.Millisecond), 80)

// 	// Simulate a slow DB call in the catalog
// 	db := addSpan(td, "catalog-db", "FETCH", tid, cat.SpanID(), ptrace.SpanKindClient, start.Add(20*time.Millisecond), 60)
// 	db.Attributes().PutBool("slow_query", true)
// }

// // --- Helpers ---

// func addSpan(t ptrace.Traces, svc, name string, tid pcommon.TraceID, pid pcommon.SpanID, kind ptrace.SpanKind, start time.Time, durMs int) ptrace.Span {
// 	rs := t.ResourceSpans().AppendEmpty()
// 	rs.Resource().Attributes().PutStr("service.name", svc)
// 	s := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
// 	s.SetTraceID(tid)
// 	s.SetSpanID(fixedSpanID(uint64(rand.Int63())))
// 	if !pid.IsEmpty() {
// 		s.SetParentSpanID(pid)
// 	}
// 	s.SetName(name)
// 	s.SetKind(kind)
// 	s.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
// 	s.SetEndTimestamp(pcommon.NewTimestampFromTime(start.Add(time.Duration(durMs) * time.Millisecond)))
// 	return s
// }

// func setHTTP(s ptrace.Span, method string, code int) {
// 	s.Attributes().PutStr("http.method", method)
// 	s.Attributes().PutInt("http.status_code", int64(code))
// }

// func fixedTraceID(i int) pcommon.TraceID {
// 	var b [16]byte
// 	b[15] = byte(i)
// 	return pcommon.TraceID(b)
// }

// func fixedSpanID(i uint64) pcommon.SpanID {
// 	var b [8]byte
// 	for j := 0; j < 8; j++ {
// 		b[j] = byte(i >> (j * 8))
// 	}
// 	return pcommon.SpanID(b)
// }
