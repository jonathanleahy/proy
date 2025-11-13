# REST API v2 - Golang with Hexagonal Architecture

A REST API built in Go using **Hexagonal Architecture** (Ports and Adapters), featuring TDD and BDD methodologies.

## Architecture

This project implements **Hexagonal Architecture** (also known as Ports and Adapters):

```
┌─────────────────────────────────────────────────────────────┐
│                      Inbound Adapters                        │
│                   (HTTP, CLI, gRPC, etc.)                    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         HTTP Handler (Gin Framework)                  │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │ depends on                          │
│                       ↓                                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Inbound Ports                            │  │
│  │           (Service Interfaces)                        │  │
│  └────────────────────┬─────────────────────────────────┘  │
└───────────────────────┼─────────────────────────────────────┘
                        │
         ┌──────────────┴──────────────┐
         │                              │
         │      Domain Layer            │
         │   (Business Logic)           │
         │                              │
         │  - Health Entity             │
         │  - Health Service            │
         │                              │
         └──────────────┬───────────────┘
                        │ depends on
                        ↓
         ┌──────────────────────────────┐
         │    Outbound Ports            │
         │  (Repository Interfaces,     │
         │   External Service Interfaces)│
         └──────────────┬───────────────┘
                        │
┌───────────────────────┼─────────────────────────────────────┐
│                       ↓                                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │        Outbound Adapters                              │  │
│  │  (Database, External APIs, Message Queues, etc.)     │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│                   Outbound Adapters                          │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Domain Layer** (core) - Contains business logic, no external dependencies
2. **Ports** - Interfaces that define contracts
   - **Inbound Ports**: What the application offers (use cases/services)
   - **Outbound Ports**: What the application needs (repositories, external services)
3. **Adapters** - Implementations of ports
   - **Inbound Adapters**: HTTP handlers, CLI, gRPC servers
   - **Outbound Adapters**: Database clients, HTTP clients, message queue publishers
4. **Dependency Inversion** - Core depends on abstractions, not concrete implementations

## Features

- ✅ **Hexagonal Architecture** - Clean separation of concerns
- ✅ **Health Endpoint** - Simple health check endpoint
- ✅ **TDD Methodology** - Unit tests with testify
- ✅ **BDD Methodology** - Integration tests with Ginkgo/Gomega
- ✅ **100% Domain & Adapter Coverage** - Comprehensive test coverage
- ✅ **Dependency Injection** - Proper DI through constructor injection
- ✅ **Gin Framework** - Fast, lightweight HTTP router
- ✅ **JSON Responses** - RESTful JSON API

## Project Structure

```
rest-v2/
├── cmd/server/                    # Application entry point
│   ├── main.go                   # Wiring & DI configuration
│   └── main_test.go              # Application tests
│
├── internal/
│   ├── domain/                   # Domain Layer (Business Logic)
│   │   └── health/
│   │       ├── health.go         # Domain entity
│   │       ├── health_test.go    # Domain tests
│   │       ├── service.go        # Domain service (implements port)
│   │       └── service_test.go   # Service tests
│   │
│   ├── ports/                    # Ports (Interfaces)
│   │   └── inbound/
│   │       └── health_service.go # Service interface (inbound port)
│   │
│   └── adapters/                 # Adapters (Implementations)
│       └── inbound/
│           └── http/             # HTTP adapter
│               ├── health_handler.go      # Gin HTTP handler
│               └── health_handler_test.go # Handler tests
│
├── tests/integration/            # BDD integration tests
│   ├── health_suite_test.go     # Ginkgo test suite
│   └── health_test.go           # BDD specs
│
├── go.mod
└── README.md
```

### Architecture Layers Explained

#### Domain Layer (`internal/domain/`)
- **Pure business logic** - No dependencies on frameworks or external libraries
- Contains entities, value objects, and domain services
- This is the **heart** of the application
- Example: `health.Health` entity, `health.Service` domain service

#### Ports Layer (`internal/ports/`)
- **Interfaces** that define contracts
- **Inbound ports** (`inbound/`): Define what the application offers (use cases)
- **Outbound ports** (`outbound/`): Define what the application needs (repositories, APIs)
- Example: `inbound.HealthService` interface

#### Adapters Layer (`internal/adapters/`)
- **Implementations** of ports
- **Inbound adapters** (`inbound/`): HTTP handlers, CLI, gRPC servers
- **Outbound adapters** (`outbound/`): Database repos, HTTP clients, caches
- Example: `http.HealthHandler` implements HTTP interface, uses `HealthService` port

#### Application Layer (`cmd/server/`)
- **Wiring & Dependency Injection**
- Creates instances and connects them
- Configuration and startup logic

## Requirements

- Go 1.21 or higher

## Installation

```bash
cd rest-v2
go mod download
```

## Running the Application

### Default (Port 8080)

```bash
go run cmd/server/main.go
```

### Custom Port

```bash
PORT=3000 go run cmd/server/main.go
```

### Build Binary

```bash
go build -o rest-v2 cmd/server/main.go
./rest-v2
```

## API Endpoints

### Health Check

Returns the health status of the API.

**Request:**
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-02T00:45:30.123456Z",
  "version": "2.0.0"
}
```

**Example:**
```bash
curl http://0.0.0.0:8080/health
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Run Unit Tests Only

```bash
# Domain tests
go test ./internal/domain/health/...

# Adapter tests
go test ./internal/adapters/inbound/http/...
```

### Run BDD Integration Tests Only

```bash
go test ./tests/integration/...
```

### Verbose Test Output

```bash
go test -v ./...
```

### Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Coverage

- **domain/health**: 100.0% ✅
- **adapters/inbound/http**: 100.0% ✅
- **cmd/server**: 41.7%
- **Overall**: Exceeds 80% requirement for business logic

## Development

### TDD Workflow

1. Write failing test first
2. Implement minimal code to pass
3. Refactor
4. Repeat

### BDD Workflow (Ginkgo/Gomega)

1. Describe behavior in plain English
2. Write specs using Given/When/Then pattern
3. Implement handlers
4. Verify behavior

### Hexagonal Architecture Workflow

1. **Start with Domain** - Define entities and business logic
2. **Create Ports** - Define interfaces for what you need
3. **Implement Domain Services** - Business logic using ports
4. **Create Adapters** - Implement ports for specific technologies
5. **Wire in Main** - Connect everything with dependency injection

## Benefits of Hexagonal Architecture

### 1. **Testability**
- Domain logic can be tested without any infrastructure
- Easy to mock dependencies through ports
- Fast unit tests (no database, HTTP, etc.)

### 2. **Flexibility**
- Swap implementations without changing business logic
- Example: Change from Gin to Echo framework - only adapters change
- Example: Change from PostgreSQL to MongoDB - only adapters change

### 3. **Maintainability**
- Clear separation of concerns
- Business logic isolated from technical details
- Easy to understand and modify

### 4. **Technology Independence**
- Domain layer has no framework dependencies
- Can change frameworks without touching business logic
- Future-proof architecture

## Example: Dependency Injection

```go
// main.go - Wiring the hexagonal architecture

// 1. Create domain service (business logic)
healthService := health.NewService("2.0.0")

// 2. Create HTTP adapter (depends on service through port)
healthHandler := http.NewHealthHandler(healthService)

// 3. Configure routes
router.GET("/health", healthHandler.GetHealth)
```

The handler depends on the **port** (interface), not the concrete service. This allows us to:
- Swap implementations easily
- Test with mocks
- Add new adapters (CLI, gRPC) without changing domain

## Dependencies

- **gin-gonic/gin** - HTTP web framework (adapter layer)
- **stretchr/testify** - Testing toolkit (TDD)
- **onsi/ginkgo/v2** - BDD testing framework
- **onsi/gomega** - Matcher/assertion library for BDD

## Migration from v1

This is REST API v2, a Golang rewrite of the Node.js REST v1 using Hexagonal Architecture.

Key differences:
- **Language**: Go instead of Node.js/TypeScript
- **Architecture**: Hexagonal (Ports & Adapters)
- **Framework**: Gin instead of Express
- **Testing**: Ginkgo/Gomega (BDD) + testify (TDD)
- **Design**: Clean Architecture principles

## Integration with Prroxy

This API is designed to work with the Prroxy migration testing system:

1. **Proxy Tool**: Record production traffic
2. **REST v1**: Original Node.js implementation
3. **REST v2**: New Golang implementation (this project) ← **Hexagonal Architecture**
4. **Reporter**: Compare v1 and v2 responses

## Adding New Features

### Example: Adding a new endpoint

1. **Domain Layer**: Create entity and service
```go
// internal/domain/user/user.go
type User struct { ... }

// internal/domain/user/service.go
func (s *Service) GetUser(id string) (*User, error) { ... }
```

2. **Port**: Define interface
```go
// internal/ports/inbound/user_service.go
type UserService interface {
    GetUser(id string) (*domain.User, error)
}
```

3. **Adapter**: Implement HTTP handler
```go
// internal/adapters/inbound/http/user_handler.go
type UserHandler struct {
    userService inbound.UserService
}
```

4. **Wire**: Connect in main.go
```go
userService := user.NewService()
userHandler := http.NewUserHandler(userService)
router.GET("/user/:id", userHandler.GetUser)
```

## CI/CD Integration

The API returns standard HTTP status codes:
- `200 OK` - Service is healthy
- `404 Not Found` - Route doesn't exist
- `500 Internal Server Error` - Service error

Perfect for integration with:
- Docker health checks
- Kubernetes liveness/readiness probes
- Load balancer health checks

## Docker Support

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o rest-v2 cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/rest-v2 .
EXPOSE 8080
CMD ["./rest-v2"]
```

## Environment Variables

- `PORT` - Server port (default: 8080)

## Further Reading

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Ports and Adapters Pattern](https://herbertograca.com/2017/09/14/ports-adapters-architecture/)

## License

Internal use only. Part of the Prroxy migration testing system.
