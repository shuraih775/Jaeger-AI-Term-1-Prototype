package ai

import (
	"testing"
	"time"

	"github.com/jaeger-ai-assist-prototype/internal"
)

func strptr(s string) *string {
	return &s
}

func int64ptr(v int64) *int64 {
	return &v
}

func TestSearchIR_ValidMapping(t *testing.T) {
	ir := SearchIR{
		Service:       strptr("payment-service"),
		Operation:     strptr("POST /charge"),
		MinDurationMs: int64ptr(2000),
		MaxDurationMs: int64ptr(5000),
		Tags: map[string]string{
			"http.status_code": "500",
			"error":            "true",
		},
	}

	if err := ValidateSearchIR(ir); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	qp, err := MapIRToQueryParams(ir)
	if err != nil {
		t.Fatalf("unexpected mapping error: %v", err)
	}

	if qp.ServiceName != "payment-service" {
		t.Fatalf("service mismatch: %s", qp.ServiceName)
	}

	if qp.OperationName != "POST /charge" {
		t.Fatalf("operation mismatch: %s", qp.OperationName)
	}

	if qp.DurationMin != 2*time.Second {
		t.Fatalf("min duration mismatch: %v", qp.DurationMin)
	}

	if qp.DurationMax != 5*time.Second {
		t.Fatalf("max duration mismatch: %v", qp.DurationMax)
	}

	if qp.Attributes.Len() != 2 {
		t.Fatalf("expected 2 attributes, got %d", qp.Attributes.Len())
	}

	v, ok := qp.Attributes.Get("http.status_code")
	if !ok || v.Str() != "500" {
		t.Fatalf("http.status_code tag missing or incorrect")
	}
}

func TestSearchIR_ValidationFailures(t *testing.T) {
	tests := []SearchIR{
		{MinDurationMs: int64ptr(-1)},
		{MaxDurationMs: int64ptr(-10)},
		{MinDurationMs: int64ptr(5000), MaxDurationMs: int64ptr(1000)},
		{StartTime: strptr("not-a-time")},
		{Tags: map[string]string{"": "500"}},
		{Tags: map[string]string{"http.status_code": ""}},
	}

	for i, ir := range tests {
		if err := ValidateSearchIR(ir); err == nil {
			t.Fatalf("test %d: expected validation error, got nil", i)
		}
	}
}

func TestSearchIR_EmptyIRProducesEmptyQuery(t *testing.T) {
	ir := SearchIR{}

	if err := ValidateSearchIR(ir); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	qp, err := MapIRToQueryParams(ir)
	if err != nil {
		t.Fatalf("unexpected mapping error: %v", err)
	}

	if qp != (internal.TraceQueryParams{}) {
		t.Fatalf("expected empty TraceQueryParams, got %+v", qp)
	}
}

func TestSearchIR_TimeParsing(t *testing.T) {
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339)

	ir := SearchIR{
		StartTime: &start,
		EndTime:   &end,
	}

	if err := ValidateSearchIR(ir); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	qp, err := MapIRToQueryParams(ir)
	if err != nil {
		t.Fatalf("unexpected mapping error: %v", err)
	}

	if qp.StartTimeMin.IsZero() || qp.StartTimeMax.IsZero() {
		t.Fatalf("expected parsed start/end times")
	}
}

func TestSearchIR_TagsAreIndependent(t *testing.T) {
	ir := SearchIR{
		Tags: map[string]string{
			"error": "true",
		},
	}

	if err := ValidateSearchIR(ir); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	qp, _ := MapIRToQueryParams(ir)

	if qp.Attributes.Len() != 1 {
		t.Fatalf("expected 1 attribute")
	}

	ir.Tags["error"] = "false"

	v, _ := qp.Attributes.Get("error")
	if v.Str() != "true" {
		t.Fatalf("mapping should not alias IR tags")
	}
}
