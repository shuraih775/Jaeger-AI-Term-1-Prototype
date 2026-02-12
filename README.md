- Unit conversion is offloaded to go code
- Should be able to make more than one query if multiple operation name and services are given in the natural language input
- in dates if end_time is missing we use now by default


one of the propmts that worked fine

```
const SearchExtractionPrompt = `
<Task>
You extract structured trace search filters from user input.
Output ONLY a valid JSON object. Do not include markdown formatting or explanations.
Do not guess what user might want without strong evidence in user input.
Use <MappingLogic> to map each word in input to json schema/output
Use <Thinking> to understand how to think
</Task>

<MappingLogic>:
- for 3-DIGIT NUMBERS with no unit and related to  close to "error" (e.g. "500", "404") -> MAP TO tags["http.status_code"]
- for "GET","POST","PUT","PATCH" close to "http" or "method" or "requests" -> MAP TO tags["http.method"]
- ">", "above", "longer" -> min_duration_ms
- "<", "below", "shorter" -> max_duration_ms
- verbs ("login", "get_*", "set_*", "Get*", "Set*" etc.) -> operation
- nouns ( names like payments, mysql, redis, orders, users) -> service
- any date like (hours, mins,days) with since, from should be -> start_time
- any date like (hours, mins,days) with till, to should be -> end_time
</MappingLogic>



<Examples>
Input: "show me errors in the payments service"
Output: {"service": "payments", "operation": null, "min_duration_ms": null, "max_duration_ms": null, "start_time": null, "end_time": null, "tags": {"error": "true"}}

Input: "find get_user calls taking longer than 500ms"
Output: {"service": null, "operation": "get_user", "min_duration_ms": "500ms", "max_duration_ms": null, "start_time": null, "end_time": null, "tags": {}}

Input: "find GetUsers calls taking lesser than 500ms"
Output: {"service": null, "operation": "GetUsers", "min_duration_ms": null, "max_duration_ms": "500ms", "start_time": null, "end_time": null, "tags": {}}

Input: "find traces with latency > 2ms"
Output: {"service": null, "operation": null, "min_duration_ms": "2ms", "max_duration_ms": null, "start_time": null, "end_time": null, "tags": {}}

Input: "Show me 500 errors from payment-service > 2s"
Output: {"service": null, "operation": null, "min_duration_ms": "2s", "max_duration_ms": null, "start_time": null, "end_time": null, "tags": {"http.status_code":"500"}}

Input: "Show me 500 errors from payment-service > 2s on tuesday"
Output: {"service": null, "operation": null, "min_duration_ms": "2s", "max_duration_ms": null, "start_time": tuesday, "end_time": tuesday, "tags": {"http.status_code":"500"}}

Input: "Show me 500 errors from two hours ago in payment-service > 2s "
Output: {"service": null, "operation": null, "min_duration_ms": "2s", "max_duration_ms": null, "start_time": "2h", "end_time": "now", "tags": {"http.status_code":"500"}}

Input: "Show me 500 errors in payment-service with latency > 2s that occured yesterday"
Output: {"service": null, "operation": null, "min_duration_ms": "2s", "max_duration_ms": null, "start_time": "yesterday", "end_time": "yesterday", "tags": {"http.status_code":"500"}}

Input: "Show me GET method errors from user-service > 10s"
Output: {"service": null, "operation": null, "min_duration_ms": "10s", "max_duration_ms": null, "start_time": null, "end_time": null, "tags": {"http.method":"GET"}}
</Examples>

<Input>
{{.Input}}
</Input>
`
```

const SearchExtractionPrompt = `
<Task>
You extract structured trace search filters from user input.
Output ONLY a valid JSON object. Do not include markdown formatting or explanations.
Do not guess what user might want without strong evidence in user input.
Use <MappingLogic> to map each word in input to json schema/output
</Task>

<MappingLogic>:
- for 3-DIGIT NUMBERS with no unit and related to  close to "error" (e.g. "500", "404") -> MAP TO tags["http.status_code"]
- for "GET","POST","PUT","PATCH" close to "http" or "method" or "requests" -> MAP TO tags["http.method"]
- ">", "above", "longer" -> min_duration_ms
- "<", "below", "shorter" -> max_duration_ms
- verbs ("login", "get_*", "set_*", "Get*", "Set*" etc.) -> operation
- nouns ( names like payments, mysql, redis, orders, users) -> service
</MappingLogic>



<Examples>
Input: "Show me GET errors in payments from 2 hours ago"
Explanation: "GET" is an HTTP method; "errors" triggers error:true; "payments" is a service; "2 hours ago" is a relative start time.
Output: {"service": "payments", "operation": null, "min_duration_ms": null, "max_duration_ms": null, "start_time": "2h", "end_time": "now", "tags": {"error": "true", "http.method": "GET"}}

Input: "GetUsers calls in auth-service > 500ms on tuesday"
Explanation: "GetUsers" is the action (operation); "auth-service" is the system (service); ">" represents lower bound "500ms" is numerical value; "tuesday" is a specific day for both start and end.
Output: {"service": "auth-service", "operation": "GetUsers", "min_duration_ms": "500ms", "max_duration_ms": null, "start_time": "tuesday", "end_time": "tuesday", "tags": {}}

Input: "find 500 errors < 2s between yesterday and now"
Explanation: "500" is a status code; "< 2s" is an upper latency bound; "yesterday" is the start; "now" is the end.
Output: {"service": null, "operation": null, "min_duration_ms": null, "max_duration_ms": "2s", "start_time": "yesterday", "end_time": "now", "tags": {"http.status_code": "500"}}
</Examples>

<Input>
{{.Input}}
</Input>
`