# REST API Testing: v1 vs v2

Test and compare two REST API implementations using proxy recording and automated comparison.

## 📋 Documentation

**Start here:**
- **[FIX-PROCESS.md](FIX-PROCESS.md)** - Systematic process for fixing each endpoint (branch → test → fix → verify)
- [TESTING-WORKFLOW.md](TESTING-WORKFLOW.md) - Complete workflow documentation

## Quick Start

```bash
# For AI assistants and automated workflows
./initialize-workflow.sh
./run-reporter.sh config.comprehensive.json

# Or use manual mode for more control
PROXY_MODE=record ./start.sh
./run-reporter.sh config.person-lookup.json
./run-reporter.sh config.person-by-surname.json
./run-reporter.sh config.person-by-dob.json

# Playback mode (uses cached responses)
./remove.sh
PROXY_MODE=playback ./start.sh
./run-reporter.sh config.comprehensive.json

# View latest results
cat reports/report_*.md | head -50
```

**Test Modes**:
- **test-record.sh**: Captures external API calls and stores in `recordings/` - use first time or to refresh data
- **test-playback.sh**: Uses cached responses (no external calls) - faster, deterministic results

Both scripts:
1. Clean up any existing services
2. Start services in the appropriate mode
3. Run comparison tests
4. Show results summary
5. Optionally keep services running for manual testing

Results are saved in `reports/` with timestamps. Recordings are in `recordings/`.

**Documentation:**
- [TESTING-WORKFLOW.md](TESTING-WORKFLOW.md) - Detailed workflow documentation
- [FIX-PROCESS.md](FIX-PROCESS.md) - Systematic process for fixing each failing endpoint (branch → test → fix → verify)

## Multi-Machine Configuration

The project supports OS-specific configuration files for working across different machines:

**Auto-Detection (Recommended):**
- Linux: Uses `env.linux` automatically
- macOS: Uses `env.darwin` automatically
- Fallback: Uses `env` if OS-specific file not found

**Setup:**
```bash
# On Linux machine
cp env.example env.linux
# Edit env.linux with your Linux paths

# On Mac machine
cp env.example env.darwin
# Edit env.darwin with your Mac paths
```

**Files:**
- `env.linux` - Linux-specific configuration (tracked in git)
- `env.darwin` - macOS-specific configuration (tracked in git)
- `env` - Generic fallback (gitignored, created from env.example)
- `env.example` - Template for new environments

All scripts automatically detect your OS and load the appropriate config. No manual switching needed!

## Groovy/Spring API Setup

If your REST v1 service is a Groovy/Spring Boot application (runs with `./gradlew run`), you need to create a start script:

```bash
# Copy the template
cp templates/crm-api-start.sh.template ~/work/git-other/crm-api/start.sh

# Make it executable
chmod +x ~/work/git-other/crm-api/start.sh
```

The template handles:
- Port configuration from environment variables
- Running through Gradle with `./gradlew run --console=plain`
- Proper output for logging

See [templates/README.md](templates/README.md) for more details and customization options.

## Configuration for Different APIs

Edit `start.sh` to point to your APIs:

```bash
# Your API locations and ports
PRROXY_BASE="~/work/personal-ooo/test/prroxy"
REST_V1_DIR="$PRROXY_BASE/rest-v1"        # Change to your v1 API path
REST_V2_DIR="$PRROXY_BASE/rest-v2"        # Change to your v2 API path
PROXY_DIR="$PRROXY_BASE/proxy"            # Proxy location

# Config files where URLs are defined (for proxy routing)
REST_V1_CONFIG="$REST_V1_DIR/src/services/userService.ts"  # Your v1 config file
REST_V2_CONFIG="$REST_V2_DIR/cmd/server/main.go"           # Your v2 config file
```

Edit `remove.sh` to match your ports:

```bash
# Your API ports
REST_V1_PORT=3002    # Change to your v1 port
REST_V2_PORT=3004    # Change to your v2 port
```

## Test Configuration

### Available Configs

**config.person-lookup.json** (default):
- Tests person lookup endpoints with full search (surname + dob)
- 25 endpoints - returns single person object
- Makes external calls to rest-external-user through proxy
- Example: `./test-record.sh config.person-lookup.json`

**config.person-by-surname.json**:
- Tests partial search by surname only
- 5 endpoints - returns array of matching people
- Makes external calls to rest-external-user through proxy
- Example: `./test-record.sh config.person-by-surname.json`

**config.person-by-dob.json**:
- Tests partial search by date of birth only
- 5 endpoints - returns array of matching people
- Makes external calls to rest-external-user through proxy
- Example: `./test-record.sh config.person-by-dob.json`

**config.user-endpoints.json**:
- Tests endpoints that make external API calls
- 10 endpoints calling jsonplaceholder.typicode.com
- Use this to test proxy record/playback functionality
- Example: `./test-record.sh config.user-endpoints.json`

**config.comprehensive.json**:
- **Complete test suite** - All endpoints from rest-v1
- 40 endpoints total:
  - 3 simple user fetches (GET /api/user/:id)
  - 3 user summaries (GET /api/user/:id/summary)
  - 3 complex reports (POST /api/user/:id/report)
  - 25 person lookups (GET /api/person)
  - 6 people searches (GET /api/people)
- Tests both JSONPlaceholder and rest-external-user
- Ideal for comprehensive API validation
- Example: `./test-record.sh config.comprehensive.json`

### Create Custom Config

```json
{
  "base_url_v1": "http://0.0.0.0:3002",    // Your v1 base URL
  "base_url_v2": "http://0.0.0.0:3004",    // Your v2 base URL
  "iterations": 1,
  "endpoints": [
    {
      "path": "/api/person?surname=Smith&dob=1990-01-01",
      "method": "GET"
    }
  ]
}
```

Then run: `./run-reporter.sh your-config.json`

## What It Does

1. **Proxy**: Routes API calls and caches responses in `recordings/`
2. **Reporter**: Calls both APIs, compares responses, reports differences
3. **Results**: Saved in `reports/` with pass/fail for each endpoint
