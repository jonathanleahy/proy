# Project Overview: Prroxy - API Migration Testing System

## Table of Contents
1. [Introduction](#introduction)
2. [System Architecture](#system-architecture)
3. [Components](#components)
4. [Workflow](#workflow)
5. [Getting Started](#getting-started)
6. [Use Cases](#use-cases)
7. [Technical Details](#technical-details)

## Introduction

Prroxy is a comprehensive system designed to facilitate safe and efficient API migrations through proxy-based testing. The system demonstrates how to migrate from a legacy REST API (v1) to a new implementation (v2) while ensuring 100% compatibility and zero downtime.

### Key Innovation

Traditional API migrations are risky and time-consuming. Prroxy solves this by:
1. **Recording** real production interactions via an HTTP proxy
2. **Replaying** those interactions offline for testing
3. **Validating** new implementations against recorded real-world data
4. **Automating** the comparison and testing process

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Prroxy System                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────┐      ┌──────────────────┐                  │
│  │   REST API v1  │◄─────┤  HTTP Proxy Tool │                  │
│  │   (Legacy)     │      │  (Record/Replay) │                  │
│  └────────────────┘      └──────────────────┘                  │
│         │                         │                              │
│         │ Record Mode             │ Playback Mode               │
│         ▼                         ▼                              │
│  ┌──────────────────────────────────────────┐                  │
│  │     Recorded Interactions Storage         │                  │
│  │    (JSON files organized by service)      │                  │
│  └──────────────────────────────────────────┘                  │
│         │                         │                              │
│         │                         │                              │
│         ▼                         ▼                              │
│  ┌────────────────┐      ┌────────────────┐                    │
│  │  REST API v2   │      │  Test Suite    │                    │
│  │   (Future)     │      │  (Validation)  │                    │
│  └────────────────┘      └────────────────┘                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. HTTP Proxy Tool (`/proxy`)

**Purpose**: Record and replay HTTP interactions

**Technology Stack**:
- Language: Go
- Framework: Native Go HTTP server
- Storage: JSON-based filesystem
- Testing: TDD with 80%+ coverage

**Key Features**:
- **Record Mode**: Intercepts and saves all HTTP request/response pairs
- **Playback Mode**: Returns saved responses without external calls
- **Organization**: Automatically organizes recordings by target service
- **Web Dashboard**: Visual management interface
- **Docker Support**: Containerized deployment

**File Structure**:
```
proxy/
├── cmd/proxy/main.go           # Entry point
├── internal/
│   ├── config/                 # Configuration
│   ├── handler/                # HTTP handlers
│   ├── mode/                   # Record/Playback modes
│   ├── models/                 # Data structures
│   └── storage/                # Recording storage
├── tests/                      # BDD test scenarios
└── web/dashboard.html          # Management UI
```

### 2. REST API v1 (`/rest-v1`)

**Purpose**: Legacy API system for migration demonstration

**Technology Stack**:
- Language: TypeScript
- Framework: Express.js
- Testing: Jest + Supertest (92.59% coverage)
- Method: TDD (Test-Driven Development)

**API Endpoints**:

| Endpoint | Method | Complexity | Description |
|----------|--------|------------|-------------|
| `/api/user/:id` | GET | Simple | Fetch user from external API |
| `/api/user/:id/summary` | GET | Medium | Aggregate user + posts data |
| `/api/user/:id/report` | POST | Complex | Parallel API calls + manipulation |

**Endpoint Details**:

#### 1. Simple Endpoint - GET /api/user/:id
- Single external API call
- Data transformation (simplification)
- Basic error handling

#### 2. Medium Endpoint - GET /api/user/:id/summary
- Two sequential API calls (user, then posts)
- Data aggregation and manipulation
- Summary generation

#### 3. Complex Endpoint - POST /api/user/:id/report
- Three parallel API calls (user, posts, todos)
- Complex body parameter processing
- Statistics calculation
- Multiple data transformations

**File Structure**:
```
rest-v1/
├── src/
│   ├── routes/userRoutes.ts    # API endpoints
│   ├── services/userService.ts # Business logic
│   ├── types/index.ts          # TypeScript types
│   ├── app.ts                  # Express app setup
│   └── server.ts               # Server entry point
├── tests/                      # Comprehensive test suite
│   ├── userService.test.ts     # Service layer tests
│   ├── userRoutes.test.ts      # Route integration tests
│   └── app.test.ts             # App configuration tests
└── test-data.md                # Mock data reference
```

### 3. Documentation (`/docs`)

**Purpose**: Comprehensive documentation for different audiences

**Contents**:
- `ENGINEERING_SUMMARY.md` - Executive/manager-level overview
- `MIGRATION_HOW_TO_GUIDE.md` - Detailed engineer guide
- `PROJECT_OVERVIEW.md` - This file

## Workflow

### Phase 1: Record Production Patterns

```bash
# 1. Start the proxy in record mode
cd proxy
make run

# 2. Configure proxy to record mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"record"}'

# 3. Run REST API v1
cd rest-v1
npm run dev

# 4. Make requests through the system
curl http://0.0.0.0:3000/api/user/1
```

**What Happens**:
- REST API v1 receives request
- Makes external calls (to JSONPlaceholder API)
- All external interactions are recorded by proxy
- Recordings saved in `recordings/` folder

### Phase 2: Offline Development & Testing

```bash
# 1. Switch proxy to playback mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"playback"}'

# 2. Develop and test new API version (v2)
# External calls return recorded data
# No internet/VPN connection needed!
```

**Benefits**:
- Work offline
- Consistent test data
- Fast iteration cycles
- No rate limiting from external APIs

### Phase 3: Validation & Migration

```bash
# 1. Run test suite with recorded data
cd rest-v1
npm test

# 2. Compare v1 and v2 responses
# 3. Fix any discrepancies
# 4. Repeat until 100% match
```

## Getting Started

### Prerequisites
- Docker (recommended) OR:
  - Go 1.21+ (for proxy)
  - Node.js 18+ (for REST APIs)

### Quick Start

```bash
# 1. Clone and navigate
git clone <repository>
cd prroxy

# 2. Start the proxy
cd proxy
make docker-run
# OR
make build && make run

# 3. Start REST API v1 (in another terminal)
cd rest-v1
npm install
npm run dev

# 4. Test the system
curl http://0.0.0.0:8080/health          # Proxy health
curl http://0.0.0.0:3000/health          # API health
curl http://0.0.0.0:3000/api/user/1      # Test endpoint
```

### View Recordings

```bash
# Via web dashboard
open http://0.0.0.0:8080/admin/ui

# Via filesystem
ls recordings/
```

## Use Cases

### 1. Legacy API Migration

**Scenario**: Rewrite Groovy API in Golang

**Steps**:
1. Record production interactions with legacy API
2. Build new Golang API using TDD
3. Test new API against recorded interactions
4. Deploy with shadow processing for validation
5. Gradual cutover once validated

### 2. VPN-Isolated Development

**Scenario**: Can't access production systems from development environment

**Solution**:
1. Record interactions on company VPN
2. Save recordings offline
3. Develop on different network using playback mode
4. Return to company VPN for final validation

### 3. Integration Testing

**Scenario**: External services are flaky or rate-limited

**Solution**:
1. Record successful interactions once
2. Run integration tests using playback mode
3. Fast, reliable, repeatable tests
4. No external dependencies during testing

### 4. Performance Testing

**Scenario**: Need to isolate application performance from network latency

**Solution**:
1. Use playback mode to eliminate network calls
2. Measure pure application performance
3. Identify bottlenecks without external factors

## Technical Details

### Test Coverage

**Proxy Tool**:
- Built with TDD methodology
- BDD scenarios for user workflows
- Comprehensive error handling tests

**REST API v1**:
- **92.59%** statement coverage
- **85.71%** branch coverage
- **100%** function coverage
- **92.2%** line coverage

```
File             | % Stmts | % Branch | % Funcs | % Lines
-----------------|---------|----------|---------|--------
All files        |   92.59 |    85.71 |     100 |   92.2
src/app.ts       |     100 |      100 |     100 |    100
src/routes       |   83.87 |    66.66 |     100 |  83.87
src/services     |   97.36 |      100 |     100 |  97.05
```

### External Dependencies

**Proxy Tool**:
- No external service dependencies
- Pure Go implementation
- Minimal third-party libraries

**REST API v1**:
- **JSONPlaceholder API** (https://jsonplaceholder.typicode.com)
  - Free fake REST API for testing
  - Provides user, post, and todo data
  - Used for demonstration purposes only

### Recording Format

Recordings are stored as JSON files:

```json
{
  "request": {
    "method": "GET",
    "url": "https://jsonplaceholder.typicode.com/users/1",
    "headers": {...},
    "body": ""
  },
  "response": {
    "status": 200,
    "headers": {...},
    "body": "{...}"
  },
  "metadata": {
    "target": "jsonplaceholder.typicode.com",
    "timestamp": "2025-11-01T15:00:00Z",
    "duration": "45ms"
  }
}
```

### Performance Characteristics

**Proxy**:
- Minimal latency overhead in record mode (<5ms)
- Near-instant response in playback mode (<1ms)
- Efficient storage (gzipped JSON)

**REST API v1**:
- Endpoint 1 (Simple): ~200ms (external call dependent)
- Endpoint 2 (Medium): ~400ms (2 sequential calls)
- Endpoint 3 (Complex): ~300ms (3 parallel calls)

*All timings with proxy in playback mode*

## Directory Structure

```
prroxy/
├── proxy/                      # HTTP Proxy Tool (Go)
│   ├── cmd/                   # Application entry
│   ├── internal/              # Internal packages
│   ├── tests/                 # BDD tests
│   ├── web/                   # Dashboard UI
│   ├── Dockerfile             # Container config
│   ├── Makefile               # Build commands
│   └── README.md              # Proxy documentation
│
├── rest-v1/                   # Legacy REST API (TypeScript)
│   ├── src/                   # Source code
│   │   ├── routes/           # API endpoints
│   │   ├── services/         # Business logic
│   │   └── types/            # TypeScript types
│   ├── tests/                 # Test suite
│   ├── package.json           # Dependencies
│   ├── tsconfig.json          # TypeScript config
│   ├── jest.config.js         # Test config
│   └── README.md              # API documentation
│
├── docs/                      # Project documentation
│   ├── ENGINEERING_SUMMARY.md
│   ├── MIGRATION_HOW_TO_GUIDE.md
│   └── PROJECT_OVERVIEW.md    # This file
│
├── recordings/                # Recorded HTTP interactions
│   └── [service-name]/       # Organized by target service
│
└── README.md                  # Project root README
```

## Development Workflow

### 1. Adding New Endpoints to REST API v1

```bash
cd rest-v1

# 1. Write tests first (TDD)
# Edit: tests/userService.test.ts
# Edit: tests/userRoutes.test.ts

# 2. Run tests (should fail)
npm test

# 3. Implement functionality
# Edit: src/services/userService.ts
# Edit: src/routes/userRoutes.ts

# 4. Run tests (should pass)
npm test

# 5. Verify coverage
npm test -- --coverage
```

### 2. Testing with the Proxy

```bash
# Terminal 1: Start proxy in record mode
cd proxy
make run

# Terminal 2: Start REST API
cd rest-v1
npm run dev

# Terminal 3: Make test requests
curl http://0.0.0.0:3000/api/user/1

# Switch to playback mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"playback"}'

# Now requests use recorded data
curl http://0.0.0.0:3000/api/user/1
```

### 3. Building for Production

```bash
# Proxy
cd proxy
make build
./proxy --port=8080 --mode=playback

# REST API v1
cd rest-v1
npm run build
npm start
```

## Environment Variables

**Proxy**:
```bash
PROXY_PORT=8080
PROXY_HOST=0.0.0.0
PROXY_RECORDINGS_DIR=./recordings
PROXY_MODE=playback
PROXY_TLS_SKIP_VERIFY=true
```

**REST API v1**:
```bash
PORT=3000
NODE_ENV=production
```

## Troubleshooting

### Issue: Proxy recordings not working

**Check**:
1. Is proxy in record mode? `curl http://0.0.0.0:8080/admin/status`
2. Is the target URL correct?
3. Check logs for errors

### Issue: REST API tests failing

**Check**:
1. Dependencies installed? `npm install`
2. Correct Node version? (18+)
3. Run with verbose: `npm test -- --verbose`

### Issue: Can't connect to external APIs

**Solution**:
1. Switch proxy to playback mode
2. Ensure recordings exist for the endpoint
3. Check recordings directory permissions

## Future Enhancements

1. **REST API v2**: Implement new version in Golang
2. **Comparison Tool**: Automated v1 vs v2 validation
3. **Shadow Mode**: Run both versions in parallel
4. **Performance Metrics**: Built-in benchmarking
5. **CI/CD Integration**: Automated testing pipeline

## Contributing

This project uses:
- **TDD**: All features start with tests
- **Clean Code**: Following SOLID principles
- **Documentation**: Keep docs updated with code
- **Coverage**: Maintain 80%+ test coverage

## Support

For issues or questions:
1. Check the relevant README files
2. Review test cases for examples
3. Check the migration guide for detailed workflows

## License

Internal use only. Developed for API migration testing and demonstration.

---

**Project Status**: ✅ Active Development

**Last Updated**: November 2025

**Maintainer**: Engineering Team
