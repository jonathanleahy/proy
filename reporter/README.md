# Reporter - API Comparison Tool

A Go-based tool for comparing two API implementations (v1 vs v2) by calling the same endpoints and reporting differences in responses and performance.

## Features

- ✅ **Endpoint Comparison** - Compare responses from two different base URLs
- ✅ **Performance Measurement** - Track and compare response times
- ✅ **Deep JSON Comparison** - Identify exact differences in response bodies
- ✅ **Ignore Fields** - Exclude timestamp/dynamic fields from comparison
- ✅ **Multiple Iterations** - Run tests multiple times for reliable averages
- ✅ **Multiple Output Formats** - JSON and Markdown reports
- ✅ **80%+ Test Coverage** - Built with TDD methodology

## Use Case

Perfect for API migration scenarios where you need to ensure:
- New implementation matches old implementation exactly
- Performance is equal or better
- All endpoints behave identically

## Installation

```bash
cd reporter
go build -o reporter cmd/reporter/main.go
```

## Usage

### Basic Usage

```bash
./reporter --config config.json
```

### Options

```bash
./reporter \
  --config config.json \
  --format markdown \
  --output report.md
```

**Flags**:
- `--config` - Path to configuration file (default: config.json)
- `--format` - Output format: `json` or `markdown` (default: markdown)
- `--output` - Output file path (default: stdout)

## Configuration

Create a `config.json` file:

```json
{
  "base_url_v1": "http://0.0.0.0:3000",
  "base_url_v2": "http://0.0.0.0:8080",
  "iterations": 5,
  "ignore_fields": ["generatedAt", "timestamp"],
  "endpoints": [
    {
      "path": "/api/user/1",
      "method": "GET"
    },
    {
      "path": "/api/user/1/report",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "includeCompleted": true,
        "maxPosts": 2
      }
    }
  ]
}
```

### Configuration Fields

- **base_url_v1** (required) - Base URL for version 1 API
- **base_url_v2** (required) - Base URL for version 2 API
- **iterations** (optional) - Number of times to call each endpoint (default: 1)
- **ignore_fields** (optional) - Array of JSON field paths to ignore in comparison
- **endpoints** (required) - Array of endpoints to test

### Endpoint Configuration

Each endpoint can have:
- **path** (required) - Endpoint path (e.g., "/api/user/1")
- **method** (optional) - HTTP method (default: "GET")
- **headers** (optional) - Map of HTTP headers
- **query_params** (optional) - Map of query parameters
- **body** (optional) - Request body (for POST/PUT)

## Output Formats

### Markdown Format

```markdown
# API Comparison Report

**Total Endpoints**: 3
**Matched**: 2
**Failed**: 1
**Total Duration**: 1.5s

## Endpoint Results

### GET /api/user/1

- **Status**: MATCH
- **V1 Avg Time**: 120ms
- **V2 Avg Time**: 95ms
- **Status Codes**: V1=200, V2=200

### POST /api/user/1/report

- **Status**: MISMATCH
- **V1 Avg Time**: 250ms
- **V2 Avg Time**: 200ms
- **Status Codes**: V1=200, V2=200

**Differences**:
1. `todos.completed[0]`: "Todo 1" → "Todo 2" (value_mismatch)
```

### JSON Format

```json
{
  "TotalEndpoints": 2,
  "MatchedEndpoints": 2,
  "FailedEndpoints": 0,
  "TotalDuration": "1.5s",
  "Endpoints": [
    {
      "Path": "/api/user/1",
      "Method": "GET",
      "Match": true,
      "V1AvgTime": "120ms",
      "V2AvgTime": "95ms",
      "StatusCodeV1": 200,
      "StatusCodeV2": 200,
      "Differences": []
    }
  ]
}
```

## Example Workflow

### Step 1: Start Both APIs

```bash
# Terminal 1: Start REST v1
cd rest-v1
npm run dev  # Runs on http://0.0.0.0:3000

# Terminal 2: Start REST v2
cd rest-v2
npm run dev  # Runs on http://0.0.0.0:8080
```

### Step 2: Create Config

```bash
cp config.example.json config.json
# Edit config.json with your endpoints
```

### Step 3: Run Reporter

```bash
./reporter --config config.json --output report.md
```

### Step 4: Review Results

```bash
cat report.md
```

## Ignore Fields

Use `ignore_fields` to exclude dynamic/timestamp fields:

```json
{
  "ignore_fields": [
    "generatedAt",           // Top-level field
    "user.createdAt",        // Nested field
    "stats.timestamp"        // Nested field
  ]
}
```

## Difference Types

The comparer identifies different types of mismatches:

- **value_mismatch** - Values differ
- **type_mismatch** - Types differ (e.g., string vs number)
- **missing_in_v2** - Field exists in v1 but not v2
- **extra_in_v2** - Field exists in v2 but not v1

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

```
reporter/
├── cmd/reporter/           # Main application
│   └── main.go
├── internal/
│   ├── config/             # Configuration loader
│   │   ├── types.go
│   │   ├── loader.go
│   │   └── loader_test.go
│   ├── client/             # HTTP client with timing
│   │   ├── types.go
│   │   ├── client.go
│   │   └── client_test.go
│   ├── comparer/           # JSON response comparer
│   │   ├── types.go
│   │   ├── comparer.go
│   │   └── comparer_test.go
│   └── reporter/           # Report generator
│       ├── types.go
│       └── reporter.go
├── config.example.json     # Example configuration
├── go.mod
└── README.md
```

## Integration with Migration Workflow

This reporter fits into the migration workflow:

1. **Record Mode**: Use the proxy to record production interactions
2. **Develop v2**: Build new API using TDD
3. **Compare**: Use reporter to compare v1 and v2 responses
4. **Fix Mismatches**: Iterate until all endpoints match
5. **Deploy**: Confident migration with verified compatibility

## Exit Codes

- `0` - All endpoints matched
- `1` - One or more endpoints failed or had differences

## Advanced Usage

### Comparing Specific Endpoints

Create a minimal config with just the endpoints you want to test:

```json
{
  "base_url_v1": "http://0.0.0.0:3000",
  "base_url_v2": "http://0.0.0.0:8080",
  "endpoints": [
    {"path": "/api/user/1", "method": "GET"}
  ]
}
```

### Performance Testing

Use multiple iterations to get reliable averages:

```json
{
  "iterations": 10,
  "endpoints": [...]
}
```

### CI/CD Integration

```bash
# In your CI pipeline
./reporter --config config.json --format json --output report.json

# Check exit code
if [ $? -ne 0 ]; then
  echo "API comparison failed!"
  exit 1
fi
```

## Limitations

- Response bodies must be valid JSON
- Binary responses are not supported
- WebSocket endpoints are not supported
- Authentication is done via headers only

## Contributing

This tool was built using TDD methodology. When contributing:
1. Write tests first
2. Implement functionality
3. Ensure 80%+ coverage
4. Update documentation

## License

Internal use only. Part of the Prroxy migration testing system.
