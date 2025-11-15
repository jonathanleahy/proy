# REST API v2 - Go Implementation

Modern REST API built with Go using Hexagonal Architecture and Vertical Slicing.

## 🎯 Project Status

**Current Phase**: In Development (TDD Approach)

### ✅ Completed
- Project initialization and tooling setup
- Common packages (errors, httpclient, response) - 80%+ coverage
- User domain (models, client, service) - 86.7% coverage
- golangci-lint configuration
- Air hot reload configuration
- Mockery integration for test mocks

### 🚧 In Progress
- User domain handler implementation
- Person domain implementation
- Wire dependency injection setup
- Main server implementation

### 📋 TODO
- Complete user and person handlers
- Integration with Wire DI
- E2E testing with test suite
- Start/shutdown scripts
- Health endpoint

## 🏗️ Architecture

### Vertical Slicing
Each domain contains all layers together:

```
internal/
├── user/              # User domain slice
│   ├── models.go         # Domain models
│   ├── client.go         # External API client
│   ├── service.go        # Business logic
│   ├── handler.go        # HTTP handlers
│   ├── *_test.go         # Tests
│   └── mocks/            # Generated mocks
├── person/            # Person domain slice
└── common/            # Shared utilities
    ├── errors/           # Error types
    ├── httpclient/       # HTTP client with proxy
    └── response/         # HTTP response helpers
```

### Technology Stack

- **Language**: Go 1.24
- **HTTP**: stdlib `net/http`
- **Logging**: Zap
- **DI**: Wire (compile-time)
- **Testing**: testify + mockery
- **Linting**: golangci-lint
- **Hot Reload**: Air

## 🧪 Testing

### Coverage Requirements
- **Target**: 80%+ coverage
- **Current**:
  - errors: 100%
  - httpclient: 81.8%
  - response: 91.7%
  - user: 86.7%

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -cover

# Specific package
go test github.com/jonathanleahy/prroxy/rest-v2/internal/user -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### TDD Workflow

1. **RED**: Write failing test
2. **GREEN**: Implement minimal code to pass
3. **REFACTOR**: Clean up while keeping tests green
4. **COMMIT**: Commit working code with tests

## 📦 Dependencies

```bash
# Core dependencies
go get -u go.uber.org/zap
go get -u github.com/stretchr/testify
go get -u github.com/google/wire/cmd/wire

# Development tools
go install github.com/vektra/mockery/v2@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/cosmtrek/air@latest
```

## 🚀 Development

### Build

```bash
# Build binary
go build -o bin/server ./cmd/server

# Run directly
go run ./cmd/server
```

### Hot Reload

```bash
# Start with Air (auto-reload on changes)
air
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## 📝 API Endpoints

### User Domain

- `GET /api/user/:id` - Get user by ID
- `GET /api/user/:id/summary` - Get user summary with posts
- `POST /api/user/:id/report` - Get comprehensive user report

### Person Domain

- `GET /api/person?surname=X&dob=YYYY-MM-DD` - Find person (exact match)
- `GET /api/people?surname=X` - Find people by surname
- `GET /api/people?dob=YYYY-MM-DD` - Find people by DOB

## 🔧 Configuration

Environment variables:

```bash
# Server
PORT=3004

# External Services
PROXY_URL=http://0.0.0.0:8099/proxy
JSONPLACEHOLDER_TARGET=https://jsonplaceholder.typicode.com
EXTERNAL_USER_TARGET=http://0.0.0.0:3006
```

## 📚 Documentation

- [SPEC.md](./SPEC.md) - Complete technical specification
- [DEVELOPER_PROFILE.md](./DEVELOPER_PROFILE.md) - Development standards and practices

## 🤝 Development Principles

1. **Test-Driven Development**: Write tests first, always
2. **Clean Code**: Functions < 40 lines, clear naming
3. **High Coverage**: 80%+ test coverage minimum
4. **Documentation**: Godoc for all exported types
5. **Error Handling**: Comprehensive error wrapping with context

## 📊 Test Coverage by Package

| Package | Coverage | Status |
|---------|----------|--------|
| common/errors | 100.0% | ✅ |
| common/response | 91.7% | ✅ |
| user | 86.7% | ✅ |
| common/httpclient | 81.8% | ✅ |

---

**Built with TDD & Best Practices** 🚀
