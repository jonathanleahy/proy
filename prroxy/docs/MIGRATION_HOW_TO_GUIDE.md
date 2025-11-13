# How to Migrate a Legacy API: A Step-by-Step Engineering Guide

## Table of Contents
1. [Introduction - Why This Approach?](#introduction)
2. [Understanding the Challenge](#understanding-the-challenge)
3. [The Proxy-Based Migration Strategy](#the-proxy-based-migration-strategy)
4. [Detailed Implementation Steps](#detailed-implementation-steps)
5. [Troubleshooting Common Issues](#troubleshooting-common-issues)
6. [Best Practices and Tips](#best-practices-and-tips)

## Introduction - Why This Approach?

Imagine you're tasked with rewriting a legacy API that's been running in production for years. The original developers are gone, documentation is minimal, and you can't afford any downtime. This guide shows you how to systematically migrate such a system using a proxy-based testing approach that ensures 100% compatibility before switching over.

### What Makes This Different?

Traditional migration approaches often involve:
- Guessing at implementation details
- Hoping your tests cover real-world scenarios
- Discovering issues only after deployment

Our approach uses **actual production data patterns** to validate the new system, without ever exposing sensitive data or impacting production.

## Understanding the Challenge

### The Legacy System Problem

Let's say you have:
- **CRMv1**: A Groovy-based REST API
- **Limited docs**: You know what endpoints exist, but not all the business logic
- **Complex orchestration**: Each request triggers 6-7 calls to external services
- **Production data**: You can't directly access customer databases due to security

### Why Standard Testing Falls Short

```
Traditional Approach:
Developer → Reads Code → Writes Tests → Hopes It Works

Reality:
- Code doesn't tell the full story
- Edge cases aren't documented
- External service behaviors aren't predictable
```

## The Proxy-Based Migration Strategy

### Core Concept

Instead of guessing how the legacy system works, we **record exactly what it does** in production, then use those recordings to validate our new implementation.

```
[Client Request] → [Proxy] → [Legacy API] → [External Services]
                      ↓
                  [Recording]
                      ↓
              [Replay for Testing]
```

### The Three-Phase Approach

1. **Record Phase**: Capture real production interactions
2. **Develop Phase**: Build new system using recordings as test data
3. **Validate Phase**: Ensure perfect compatibility before deployment

## Detailed Implementation Steps

### Step 1: Set Up the HTTP Testing Proxy

First, we need a proxy that can record and replay HTTP interactions.

#### 1.1 Install the Proxy

```bash
# Clone the proxy repository
git clone https://github.com/your-org/prroxy.git
cd prroxy/proxy

# Build and run
make build
make run
```

#### 1.2 Understanding Proxy Modes

The proxy has two modes:

**Record Mode**:
- Forwards requests to the real API
- Saves both the request and response
- Stores them organized by service

**Playback Mode**:
- Intercepts requests
- Returns previously recorded responses
- No external calls needed

#### 1.3 Configure the Legacy System

Modify your legacy API to route external calls through the proxy:

```groovy
// Before (direct call)
String apiUrl = "https://external-api.com/data"

// After (through proxy)
String apiUrl = "http://proxy:8080/proxy?target=external-api.com/data"
```

### Step 2: Gather Production Data Patterns

This is where it gets interesting. We need test data, but we can't access production databases directly.

#### 2.1 The Data Discovery Process

Since we can't query the database directly, we need to be creative:

```python
# Example: Generate common name combinations to find valid data
common_first_names = ["John", "Jane", "Bob", "Alice", "Tom"]
common_last_names = ["Smith", "Jones", "Brown", "Davis", "Wilson"]

test_queries = []
for first in common_first_names:
    for last in common_last_names:
        # Try 3-letter combinations (if the system accepts them)
        test_queries.append(first[:3])
        test_queries.append(last[:3])
```

#### 2.2 Record Everything

Switch the proxy to record mode and run your test queries:

```bash
# Enable record mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"record"}'

# Run test queries through the legacy system
for query in test_queries:
    curl "http://legacy-api/search?name=${query}"
done
```

#### 2.3 Organize Your Recordings

Recordings are automatically organized:
```
recordings/
├── external_api_com/
│   ├── request_hash_1.json  # Contains request/response pair
│   ├── request_hash_2.json
│   └── request_hash_3.json
└── another_service_com/
    └── request_hash_4.json
```

### Step 3: Build the New API

Now we start building CRMv2 in Golang, but with a twist - we test against real data from day one.

#### 3.1 Create the Basic Structure

```go
// main.go
package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    router := mux.NewRouter()

    // Start with endpoint stubs
    router.HandleFunc("/customer/{id}", GetCustomer).Methods("GET")
    router.HandleFunc("/search", SearchCustomers).Methods("GET")
    router.HandleFunc("/session", CreateSession).Methods("POST")

    http.ListenAndServe(":8080", router)
}
```

#### 3.2 Implement One Endpoint at a Time

Start with the simplest endpoint:

```go
func GetCustomer(w http.ResponseWriter, r *http.Request) {
    // Initial implementation - just structure
    vars := mux.Vars(r)
    customerID := vars["id"]

    // TODO: Implement actual logic
    json.NewEncoder(w).Encode(map[string]string{
        "id": customerID,
        "status": "not_implemented",
    })
}
```

#### 3.3 Test Against Recordings

Switch proxy to playback mode:

```bash
# Enable playback mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"playback"}'
```

Now your new API can call "external services" using recorded data:

```go
func GetCustomer(w http.ResponseWriter, r *http.Request) {
    // This will hit the proxy, which returns recorded data
    resp, err := http.Get("http://proxy:8080/proxy?target=external-api.com/customer/" + customerID)
    // Process response...
}
```

### Step 4: Systematic Validation

This is the critical step - ensuring the new API behaves identically to the old one.

#### 4.1 Create a Comparison Tool

Build a tool that runs the same request against both APIs:

```python
# compare_apis.py
import json
import requests
from deepdiff import DeepDiff

def compare_endpoints(endpoint, params):
    # Get response from legacy API
    legacy_response = requests.get(f"http://legacy-api{endpoint}", params=params)

    # Get response from new API
    new_response = requests.get(f"http://new-api{endpoint}", params=params)

    # Compare responses
    diff = DeepDiff(legacy_response.json(), new_response.json(),
                    ignore_order=True,
                    exclude_paths=["root['timestamp']"])  # Ignore fields that should differ

    return diff

# Run comparison for all test cases
test_cases = load_test_cases()  # Your 100+ test cases per endpoint
mismatches = []

for test in test_cases:
    diff = compare_endpoints(test['endpoint'], test['params'])
    if diff:
        mismatches.append({
            'test': test,
            'difference': diff
        })
```

#### 4.2 Handle Mismatches

When you find differences, you need to determine if they're bugs or expected:

```python
# Categories of differences:
EXPECTED_DIFFERENCES = {
    'travel_card_name': 'Different API source in new version',
    'timestamp': 'Generated at request time',
    'session_id': 'New format in v2'
}

def analyze_mismatch(diff):
    for field in diff.get('values_changed', {}):
        if field in EXPECTED_DIFFERENCES:
            print(f"Expected difference in {field}: {EXPECTED_DIFFERENCES[field]}")
        else:
            print(f"UNEXPECTED difference in {field} - needs investigation")
```

### Step 5: Add Comprehensive Testing

#### 5.1 Unit Tests for Business Logic

```go
func TestCustomerSearch(t *testing.T) {
    // Test with recorded data patterns
    testCases := []struct {
        name     string
        query    string
        expected int
    }{
        {"Three letter search", "Joh", 15},
        {"Full name search", "John Smith", 1},
        {"No results", "Zzzzz", 0},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            results := SearchCustomers(tc.query)
            assert.Equal(t, tc.expected, len(results))
        })
    }
}
```

#### 5.2 Contract Tests

These ensure the API contract doesn't change:

```go
func TestAPIContract(t *testing.T) {
    // Use actual recorded request/response pairs
    recordings := LoadRecordings("customer_endpoint")

    for _, recording := range recordings {
        response := CallNewAPI(recording.Request)

        // Verify structure matches
        assert.Equal(t, recording.Response.StatusCode, response.StatusCode)
        assert.JSONEq(t, recording.Response.Body, response.Body)
    }
}
```

### Step 6: Production Deployment with Shadow Mode

Before fully switching over, run both systems in parallel:

#### 6.1 Implement Shadow Processing

```go
func ShadowMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Copy the request
        reqCopy := cloneRequest(r)

        // Call legacy API (primary)
        legacyResp := callLegacyAPI(reqCopy)

        // Call new API (shadow) - async
        go func() {
            newResp := callNewAPI(reqCopy)

            // Compare responses
            if !responsesMatch(legacyResp, newResp) {
                logMismatch(r, legacyResp, newResp)
            }

            // Log performance metrics
            logPerformance(legacyResp.Duration, newResp.Duration)
        }()

        // Return legacy response to client
        writeResponse(w, legacyResp)
    })
}
```

#### 6.2 Monitor and Iterate

```bash
# Check shadow mode mismatches
tail -f /var/log/shadow-mismatches.log

# View performance comparison
curl http://monitoring/api/performance-comparison
```

### Step 7: The Cutover

Once shadow mode shows no mismatches:

1. **Gradual rollout**: Route percentage of traffic to new API
2. **Monitor closely**: Watch for any unexpected behaviors
3. **Have rollback ready**: One command to switch back

```bash
# Gradual traffic shift
for percentage in 10 20 50 80 100; do
    echo "Routing $percentage% to new API"
    kubectl set env deployment/api-gateway NEW_API_WEIGHT=$percentage
    sleep 3600  # Wait 1 hour between increases

    # Check metrics
    if [ $(get_error_rate) -gt 0.01 ]; then
        echo "Error rate too high, rolling back"
        kubectl set env deployment/api-gateway NEW_API_WEIGHT=0
        exit 1
    fi
done
```

## Troubleshooting Common Issues

### Issue 1: Can't Access Production Data

**Problem**: Security policies prevent direct database access.

**Solution**: Use the creative data discovery approach:
- Generate common patterns (names, IDs, etc.)
- Use any available UI or admin tools
- Record everything you can find

### Issue 2: External Services Behave Differently

**Problem**: External APIs return different data over time.

**Solution**:
- Record multiple responses for the same request
- Build logic to handle variations
- Use the most recent recording during testing

### Issue 3: Performance Differences

**Problem**: New API is faster (cached) or slower than legacy.

**Solution**:
```go
// Run tests in both orders to account for caching
func RunPerformanceTest() {
    // Test 1: Legacy first (warms cache)
    legacyTime1 := timeAPICall(legacyAPI)
    newTime1 := timeAPICall(newAPI)

    // Test 2: New API first
    clearCache()
    newTime2 := timeAPICall(newAPI)
    legacyTime2 := timeAPICall(legacyAPI)

    // Average both scenarios
    avgLegacy := (legacyTime1 + legacyTime2) / 2
    avgNew := (newTime1 + newTime2) / 2
}
```

### Issue 4: Handling Sensitive Data

**Problem**: Can't log production data due to PII concerns.

**Solution**: Implement PII redaction:
```go
func redactPII(data interface{}) interface{} {
    // Redact known PII fields
    piiFields := []string{"ssn", "email", "phone", "address"}

    jsonData := toJSON(data)
    for _, field := range piiFields {
        jsonData = regexp.MustCompile(
            fmt.Sprintf(`"%s":\s*"[^"]*"`, field),
        ).ReplaceAll(jsonData, fmt.Sprintf(`"%s":"[REDACTED]"`, field))
    }

    return fromJSON(jsonData)
}
```

## Best Practices and Tips

### 1. Start Small
- Begin with the simplest endpoint
- Get the full process working end-to-end
- Then scale to other endpoints

### 2. Document Everything
```yaml
# endpoint-docs.yaml
/customer/{id}:
  legacy_behavior:
    - Returns 404 if customer not found
    - Includes deprecated 'legacy_id' field
    - Calls services: [service-a, service-b, service-c]
  new_behavior:
    - Maintains 404 behavior for compatibility
    - Still includes 'legacy_id' (marked deprecated)
    - Optimized to call only: [service-a, service-c]
  known_differences:
    - timestamp: Different format
    - metadata.version: "2.0" instead of "1.0"
```

### 3. Automate Validation
Create scripts that run continuously:
```bash
#!/bin/bash
# continuous-validation.sh

while true; do
    echo "Running validation suite..."
    python compare_apis.py --all-endpoints

    if [ $? -ne 0 ]; then
        send_alert "API mismatch detected"
    fi

    sleep 300  # Run every 5 minutes
done
```

### 4. Use AI Wisely

When using AI to help with migration:

```python
# Provide clear context to AI
prompt = f"""
I need to implement the {endpoint_name} endpoint in Go.

Legacy behavior (from recordings):
- Input: {sample_input}
- Output: {sample_output}
- External calls: {external_services}

Requirements:
- Must match output exactly (ignore fields: {ignored_fields})
- Must handle these edge cases: {edge_cases}
- Must include comprehensive error handling

Please implement this endpoint with full test coverage.
"""
```

### 5. Keep Code Maintainable

Even with AI assistance, enforce quality:
```go
// Run these checks automatically
- gofmt: Format code
- golint: Catch style issues
- go vet: Find suspicious constructs
- gocyclo: Check cyclomatic complexity
- go test -cover: Ensure 80%+ coverage
```

### 6. Plan for Rollback

Always have an escape route:
```yaml
# rollback-plan.yaml
triggers:
  - error_rate > 1%
  - p95_latency > 2s
  - customer_complaints > 0

steps:
  1. Switch load balancer to legacy API
  2. Alert on-call engineer
  3. Preserve logs for investigation
  4. Document failure reason

automation:
  script: ./emergency-rollback.sh
  time_to_rollback: < 30 seconds
```

## Conclusion

This migration approach might seem complex at first, but it provides:
- **Confidence**: You're testing against real production patterns
- **Safety**: Shadow mode catches issues before they impact users
- **Speed**: AI assistance + recorded data = faster development
- **Quality**: Comprehensive validation ensures compatibility

Remember: The goal isn't to rewrite the legacy system quickly - it's to replace it safely with something better. This methodology ensures you achieve both.

## Quick Reference Checklist

```markdown
□ Set up proxy tool
□ Configure legacy API to use proxy
□ Record production interactions
□ Create new API structure
□ Implement endpoints one by one
□ Validate against recordings
□ Add comprehensive tests
□ Run shadow mode processing
□ Monitor metrics and mismatches
□ Gradual production cutover
□ Celebrate successful migration! 🎉
```

---

*This guide is based on real-world experience migrating production APIs. Your specific situation may require adaptations, but the core principles remain the same: record, replay, validate, deploy.*