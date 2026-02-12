package langchain

const SearchExtractionPrompt = `
Extract trace filters into JSON. Use the "Explanation" to reason before outputting JSON.
Rule: Do NOT convert units (s, ms, m). Extract durations and times exactly as written.

<Examples>
# 1. Latency Bounds
Input: "latency longer than 2s"
Explanation: ">" maps to min_duration_ms; value is "2s". No units are processed.
Output: {"min_duration_ms": "2s", "max_duration_ms": null, "service": null, "operation": null, "start_time": null, "end_time": null, "tags": {}}

Input: "shorter than 500ms"
Explanation: "shorter" maps to max_duration_ms; value is "500ms".
Output: {"min_duration_ms": null, "max_duration_ms": "500ms", "service": null, "operation": null, "start_time": null, "end_time": null, "tags": {}}

# 2. Time Ranges
Input: "since yesterday"
Explanation: "since" indicates a start point; end_time defaults to "now".
Output: {"start_time": "yesterday", "end_time": "now", "service": null, "operation": null, "tags": {}}

Input: "between 2pm and 4pm"
Explanation: "between" provides both a start_time ("2pm") and an end_time ("4pm").
Output: {"start_time": "2pm", "end_time": "4pm", "service": null, "operation": null, "tags": {}}

# 3. Identity Logic (Service vs Operation)
Input: "logs from payment-service"
Explanation: "payment-service" is a noun identifying the system (service).
Output: {"service": "payment-service", "operation": null, "tags": {}}

Input: "calls to GetUser"
Explanation: "GetUser" is a verb/action identifying the function (operation).
Output: {"service": null, "operation": "GetUser", "tags": {}}

Input: "login in auth-api"
Explanation: "login" is the operation (verb); "auth-api" is the service (noun).
Output: {"service": "auth-api", "operation": "login", "tags": {}}

# 4. HTTP Method vs Operation
Input: "GET requests for GetItems"
Explanation: "GET" is an HTTP method (tag); "GetItems" is the function name (operation).
Output: {"service": null, "operation": "GetItems", "tags": {"http.method": "GET"}}

# 5. Status Codes and Errors
Input: "500 errors in payments"
Explanation: "500" is a status code; "errors" triggers error:true; "payments" is the service.
Output: {"service": "payments", "operation": null, "tags": {"http.status_code": "500", "error": "true"}}

# 6. Complex Master Example
Input: "Show me 500 errors in orders-api for GetCart > 1.5s from 2 hours ago till 1h ago"
Explanation: "500" is status code; "orders-api" is service; "GetCart" is operation; "> 1.5s" is min_duration_ms; "2h ago" is start; "1h ago" is end.
Output: {
  "service": "orders-api",
  "operation": "GetCart",
  "min_duration_ms": "1.5s",
  "max_duration_ms": null,
  "start_time": "2h ago",
  "end_time": "1h ago",
  "tags": {"http.status_code": "500", "error": "true"}
}
Input: "Find traces where latency > 20 ms for GET requests for GetItems operation from two hours ago"
Explanation: "> 20ms" is min_duration; "GET" is an HTTP method tag; "GetItems" is the operation; "two hours ago" is start_time.
Output: {
  "service": null,
  "operation": "GetItems",
  "min_duration_ms": "20ms",
  "max_duration_ms": null,
  "start_time": "2h ago",
  "end_time": "now",
  "tags": {"http.method": "GET"}
}
</Examples>

<Task>
User Input: {{.Input}}
</Task>
`

const TraceExplainPrompt = `
You are a distributed tracing assistant.

Given the pruned trace summary below, provide a clear explanation in 3-5 sentences that covers:
1. What the trace is doing end-to-end (high-level flow).
2. Whether the trace appears normal or problematic.
3. If there is an error, where it most likely originated.
4. One or two reasonable next debugging steps (only if there is a problem).

Rules:
- Do NOT hallucinate details that are not present in the summary.
- If the information is insufficient to draw conclusions, say: "Insufficient data".
- If there is no error, explicitly state: "No clear error observed."

Trace Context:
{{.Context}}
`

const SpanExplainPrompt = `
You are a distributed tracing assistant.

Given the span details below, explain in 2-4 sentences:
1. What this span represents in the system.
2. How it relates to the surrounding request (based only on the given data).
3. If the span is in error, why that might have happened.
4. If the span is NOT in error, what "normal" behavior this likely represents.

Rules:
- Do NOT hallucinate missing details.
- If the information is insufficient, say: "Insufficient data".
- If there is no error, explicitly state: "No error observed on this span."

Span Context:
{{.Context}}
`
