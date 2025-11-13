# API Migration Testing Process - Complete Documentation

## Overview
This document describes the complete process for testing API migration from REST v1 to REST v2 using proxy-based recording and playback.

## Architecture Components

### 1. REST v1 (Legacy API)
- **Technology**: Node.js with TypeScript and Express
- **Location**: `http://0.0.0.0:3000`
- **External Dependencies**: JSONPlaceholder API
- **Endpoints**:
  - `GET /api/user/:id` - Fetch user data
  - `GET /api/user/:id/summary` - User summary with posts
  - `POST /api/user/:id/report` - Comprehensive user report

### 2. REST v2 (New API)
- **Technology**: Go (partial implementation)
- **Location**: `http://0.0.0.0:8080`
- **Goal**: Must return identical responses to REST v1

### 3. HTTP Proxy
- **Technology**: Go
- **Location**: `http://0.0.0.0:8080`
- **Modes**:
  - **Record Mode**: Captures all HTTP interactions
  - **Playback Mode**: Returns recorded responses
- **Storage**: JSON files in `recordings/` directory

### 4. Test Infrastructure
- **Test Cases**: JSON file with all endpoint definitions
- **Test Runner**: Script to execute test cases
- **Comparator**: Tool to compare v1 vs v2 responses
- **Reporter**: Generates comparison reports

## Step-by-Step Process

### Phase 1: Preparation

#### Step 1.1: Create Test Data Directory
```bash
mkdir -p test-data
```
**Purpose**: Central location for all test artifacts

#### Step 1.2: Define Test Cases
Create `test-data/test-cases.json` with all endpoints to test:
```json
{
  "test_cases": [
    {
      "id": "user_1",
      "name": "Get User 1",
      "endpoint": "/api/user/1",
      "method": "GET",
      "headers": {},
      "request_body": null,
      "expected_response": null
    },
    // ... more test cases
  ]
}
```
**Purpose**: Standardized test definitions for consistency

### Phase 2: Configure REST v1 for Proxy

#### Step 2.1: Modify REST v1 Service
Edit `rest-v1/src/services/userService.ts`:
```typescript
// Before:
const BASE_URL = 'https://jsonplaceholder.typicode.com';

// After:
const PROXY_URL = process.env.PROXY_URL || 'http://0.0.0.0:8080/proxy';
const BASE_URL = `${PROXY_URL}?target=jsonplaceholder.typicode.com`;
```
**Purpose**: Route all external API calls through the proxy

#### Step 2.2: Update REST v1 Environment
Create/modify `rest-v1/.env`:
```env
PROXY_URL=http://0.0.0.0:8080/proxy
PORT=3000
```
**Purpose**: Configure proxy location

### Phase 3: Recording

#### Step 3.1: Start Proxy in Record Mode
```bash
# Terminal 1
./build/proxy
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"record"}'
```
**Purpose**: Prepare proxy to capture all interactions

#### Step 3.2: Start REST v1
```bash
# Terminal 2
cd rest-v1
npm install
npm start
```
**Purpose**: Start legacy API with proxy configuration

#### Step 3.3: Execute Test Cases and Record
Run script to:
1. Read test cases from JSON
2. Execute each test against REST v1
3. Save REST v1 responses to test cases file
4. Proxy automatically records external API calls

```bash
# Terminal 3
./scripts/execute-and-record-v1.sh
```

**What Gets Recorded**:
- REST v1 responses → saved in test-cases.json
- External API calls → saved in proxy recordings/

### Phase 4: Playback Setup

#### Step 4.1: Switch Proxy to Playback Mode
```bash
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"playback"}'
```
**Purpose**: Proxy will now return recorded data instead of making real calls

#### Step 4.2: Configure REST v2 for Proxy
Ensure REST v2 uses the proxy for external calls:
```go
// In REST v2 service code
proxyURL := os.Getenv("PROXY_URL")
if proxyURL == "" {
    proxyURL = "http://0.0.0.0:8080/proxy"
}
baseURL := fmt.Sprintf("%s?target=jsonplaceholder.typicode.com", proxyURL)
```

### Phase 5: Testing REST v2

#### Step 5.1: Start REST v2
```bash
# Terminal 2 (stop REST v1 first)
cd rest-v2
./rest-v2
```
**Purpose**: Start new API implementation

#### Step 5.2: Execute Test Cases Against REST v2
Run script to:
1. Read test cases with v1 responses
2. Execute same tests against REST v2
3. Compare responses
4. Generate reports

```bash
# Terminal 3
./scripts/execute-and-compare-v2.sh
```

### Phase 6: Comparison & Reporting

#### Step 6.1: Response Comparison
For each test case:
1. Compare HTTP status codes
2. Compare response headers
3. Deep compare JSON response bodies
4. Identify differences

#### Step 6.2: Generate Summary Report
Create `test-data/summary-report.json`:
```json
{
  "timestamp": "2024-11-03T15:30:00Z",
  "total_tests": 15,
  "passed": 13,
  "failed": 2,
  "results": [
    {"endpoint": "/api/user/1", "status": "PASS"},
    {"endpoint": "/api/user/2", "status": "PASS"},
    {"endpoint": "/api/user/1/report", "status": "FAIL", "reason": "response mismatch"}
  ]
}
```

#### Step 6.3: Generate Detailed Reports
For each failed test, create detailed report:
`test-data/detailed-user_1_report.json`:
```json
{
  "test_id": "user_1_report",
  "endpoint": "/api/user/1/report",
  "status": "FAIL",
  "differences": [
    {
      "path": "$.stats.timestamp",
      "v1_value": "2024-11-03T10:00:00Z",
      "v2_value": "2024-11-03T10:00:00.000Z",
      "type": "format_difference"
    }
  ]
}
```

## File Structure
```
prroxy/
├── test-data/
│   ├── test-cases.json           # Test definitions with v1 responses
│   ├── summary-report.json       # Overall comparison summary
│   ├── detailed-user_1.json      # Detailed report for failed test
│   └── detailed-user_1_report.json
├── recordings/                    # Proxy recordings (auto-generated)
│   └── jsonplaceholder_typicode_com/
│       ├── <hash>.json
│       └── ...
├── scripts/
│   ├── execute-and-record-v1.sh  # Record v1 responses
│   └── execute-and-compare-v2.sh # Test v2 and compare
```

## Success Criteria

### Migration is Successful When:
1. ✅ All test cases pass (100% match)
2. ✅ Response times are comparable or better
3. ✅ No external API calls during v2 testing (fully offline)
4. ✅ All edge cases handled identically

### Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| Port conflicts | Ensure only one API runs at a time |
| Missing recordings | Re-run recording phase |
| Response mismatches | Check field-level differences in detailed reports |
| Timestamp differences | Add to ignore list in comparison logic |
| Order differences in arrays | Update comparison to be order-agnostic |

## Benefits of This Approach

1. **Real Data**: Tests use actual production patterns
2. **Offline Testing**: No external dependencies during testing
3. **Reproducible**: Same data every time
4. **Safe**: No production impact
5. **Comprehensive**: Catches all differences
6. **Automated**: Minimal manual intervention

## Next Steps After Successful Testing

1. **Performance Testing**: Compare response times under load
2. **Shadow Testing**: Run both APIs in parallel in staging
3. **Gradual Rollout**: Route percentage of traffic to v2
4. **Monitor**: Track error rates and performance
5. **Full Migration**: Switch all traffic to v2
6. **Decommission**: Remove v1 after stability period