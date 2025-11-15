# REST API v2 - Technical Specification

## Project Overview

REST API v2 is a Go-based HTTP service that provides user and person lookup functionality. It acts as a facade that aggregates data from external services (jsonplaceholder.typicode.com and rest-external-user) and presents a unified API interface.

## Architecture

### Design Philosophy

**Vertical Slicing**: Each domain/feature lives in its own folder containing all architectural layers together:

```
internal/
├── user/           # User domain vertical slice
│   ├── handler.go      # HTTP handlers (inbound adapter)
│   ├── handler_test.go
│   ├── service.go      # Business logic (domain)
│   ├── service_test.go
│   ├── client.go       # External API client (outbound adapter)
│   ├── client_test.go
│   ├── models.go       # Domain models
│   └── mocks/          # Generated mocks
├── person/         # Person domain vertical slice
│   ├── handler.go
│   ├── handler_test.go
│   ├── service.go
│   ├── service_test.go
│   ├── client.go
│   ├── client_test.go
│   ├── models.go
│   └── mocks/
└── common/         # Shared utilities
    ├── httpclient/     # Shared HTTP client with proxy support
    ├── errors/         # Error types and utilities
    └── response/       # HTTP response helpers
```

### Hexagonal Architecture Principles

While using vertical slicing, we maintain hexagonal architecture concepts:

- **Domain Layer** (service.go): Pure business logic, no external dependencies
- **Inbound Adapters** (handler.go): HTTP handlers, request/response transformation
- **Outbound Adapters** (client.go): External API clients, data fetching
- **Models**: Domain entities shared across layers

### Dependency Flow

```
HTTP Request → Handler → Service → Client → External API
                  ↓         ↓
            Validation  Business Logic
                  ↓
            HTTP Response
```

## API Endpoints

### 1. Get User (Simple)

**Endpoint**: `GET /api/user/:id`

**Description**: Retrieves basic user information from jsonplaceholder.typicode.com

**Request**:
```
GET /api/user/1
```

**Response** (200 OK):
```json
{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid user ID format
- `404 Not Found`: User not found
- `500 Internal Server Error`: External API error

**External Calls**: 1 call to `GET https://jsonplaceholder.typicode.com/users/{id}`

---

### 2. Get User Summary (Medium Complexity)

**Endpoint**: `GET /api/user/:id/summary`

**Description**: Retrieves user information with post statistics and recent post titles

**Request**:
```
GET /api/user/1/summary
```

**Response** (200 OK):
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "postCount": 10,
  "recentPosts": [
    "sunt aut facere repellat provident",
    "qui est esse",
    "ea molestias quasi exercitationem"
  ],
  "summary": "User Leanne Graham has written 10 posts"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid user ID format
- `500 Internal Server Error`: External API error

**External Calls**:
1. `GET https://jsonplaceholder.typicode.com/users/{id}`
2. `GET https://jsonplaceholder.typicode.com/posts?userId={id}`

---

### 3. Get User Report (Complex)

**Endpoint**: `POST /api/user/:id/report`

**Description**: Generates comprehensive user report with posts and todos. Makes parallel external calls for efficiency.

**Request**:
```
POST /api/user/1/report
Content-Type: application/json

{
  "includeCompleted": true,
  "maxPosts": 5
}
```

**Request Body Parameters**:
- `includeCompleted` (boolean, optional): Whether to include completed todos. Default: true
- `maxPosts` (integer, optional): Maximum number of posts to return. Default: all posts

**Response** (200 OK):
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "stats": {
    "totalPosts": 10,
    "totalTodos": 20,
    "completedTodos": 10,
    "pendingTodos": 10,
    "completionRate": "50.0%"
  },
  "posts": [
    {
      "id": 1,
      "title": "sunt aut facere repellat provident",
      "preview": "quia et suscipit\nsuscipit..."
    }
  ],
  "todos": {
    "pending": ["delectus aut autem", "quis ut nam facilis"],
    "completed": ["fugiat veniam minus", "et porro tempora"]
  },
  "generatedAt": "2025-11-14T10:30:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid user ID or request body
- `500 Internal Server Error`: External API error

**External Calls** (parallel):
1. `GET https://jsonplaceholder.typicode.com/users/{id}`
2. `GET https://jsonplaceholder.typicode.com/posts?userId={id}`
3. `GET https://jsonplaceholder.typicode.com/todos?userId={id}`

---

### 4. Find Person (Exact Match)

**Endpoint**: `GET /api/person?surname={surname}&dob={dob}`

**Description**: Finds a single person by exact surname and date of birth match

**Request**:
```
GET /api/person?surname=Thompson&dob=1985-03-15
```

**Response** (200 OK):
```json
{
  "firstname": "Emma",
  "surname": "Thompson",
  "dob": "1985-03-15",
  "country": "United Kingdom"
}
```

**Error Responses**:
- `400 Bad Request`: Missing required parameters or invalid date format
- `404 Not Found`: No person found with given surname and DOB
- `500 Internal Server Error`: External service error

**External Calls**: 1 call to `GET http://0.0.0.0:3006/person?surname={surname}&dob={dob}` (via proxy)

**Validation**:
- Both `surname` and `dob` are required
- `dob` must be in format `YYYY-MM-DD`

---

### 5. Find People (Partial Search)

**Endpoint**: `GET /api/people?surname={surname}` or `GET /api/people?dob={dob}`

**Description**: Searches for people by surname OR date of birth (partial match)

**Request**:
```
GET /api/people?surname=Thompson
```

**Response** (200 OK):
```json
[
  {
    "firstname": "Emma",
    "surname": "Thompson",
    "dob": "1985-03-15",
    "country": "United Kingdom"
  }
]
```

**Request**:
```
GET /api/people?dob=1985-03-15
```

**Response** (200 OK):
```json
[
  {
    "firstname": "Emma",
    "surname": "Thompson",
    "dob": "1985-03-15",
    "country": "United Kingdom"
  },
  {
    "firstname": "Sebastian",
    "surname": "Müller",
    "dob": "1985-02-28",
    "country": "Austria"
  }
]
```

**Error Responses**:
- `400 Bad Request`: Missing both parameters or invalid date format
- `500 Internal Server Error`: External service error

**External Calls**: 1 call to `GET http://0.0.0.0:3006/person?surname={surname}` or `?dob={dob}` (via proxy)

**Validation**:
- At least one of `surname` or `dob` is required
- `dob` (if provided) must be in format `YYYY-MM-DD`

---

## Data Models

### User Domain

```go
// User represents basic user information from jsonplaceholder
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Website  string `json:"website"`
}

// Post represents a blog post from jsonplaceholder
type Post struct {
    UserID int    `json:"userId"`
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

// Todo represents a todo item from jsonplaceholder
type Todo struct {
    UserID    int    `json:"userId"`
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

// UserSummary represents aggregated user summary response
type UserSummary struct {
    UserID      int      `json:"userId"`
    UserName    string   `json:"userName"`
    Email       string   `json:"email"`
    PostCount   int      `json:"postCount"`
    RecentPosts []string `json:"recentPosts"`
    Summary     string   `json:"summary"`
}

// ReportRequest represents user report request parameters
type ReportRequest struct {
    IncludeCompleted bool `json:"includeCompleted"`
    MaxPosts         *int `json:"maxPosts,omitempty"`
}

// UserReport represents comprehensive user report response
type UserReport struct {
    UserID      int               `json:"userId"`
    UserName    string            `json:"userName"`
    Email       string            `json:"email"`
    Stats       ReportStats       `json:"stats"`
    Posts       []PostPreview     `json:"posts"`
    Todos       TodoGroups        `json:"todos"`
    GeneratedAt string            `json:"generatedAt"`
}

type ReportStats struct {
    TotalPosts     int    `json:"totalPosts"`
    TotalTodos     int    `json:"totalTodos"`
    CompletedTodos int    `json:"completedTodos"`
    PendingTodos   int    `json:"pendingTodos"`
    CompletionRate string `json:"completionRate"`
}

type PostPreview struct {
    ID      int    `json:"id"`
    Title   string `json:"title"`
    Preview string `json:"preview"`
}

type TodoGroups struct {
    Pending   []string `json:"pending"`
    Completed []string `json:"completed"`
}
```

### Person Domain

```go
// Person represents a person from rest-external-user service
type Person struct {
    Firstname string `json:"firstname"`
    Surname   string `json:"surname"`
    DOB       string `json:"dob"`
    Country   string `json:"country"`
}
```

---

## External Dependencies

### 1. JSONPlaceholder API

**Base URL**: `https://jsonplaceholder.typicode.com`

**Accessed via Proxy**: `http://0.0.0.0:8099/proxy?target=https://jsonplaceholder.typicode.com`

**Endpoints Used**:
- `GET /users/{id}` - Get user by ID
- `GET /posts?userId={id}` - Get user's posts
- `GET /todos?userId={id}` - Get user's todos

**Note**: All calls must go through the proxy with `Accept-Encoding: identity` header to disable compression for recording compatibility.

### 2. REST External User Service

**Base URL**: `http://0.0.0.0:3006`

**Accessed via Proxy**: `http://0.0.0.0:8099/proxy?target=http://0.0.0.0:3006`

**Endpoints Used**:
- `GET /person?surname={surname}&dob={dob}` - Find person by exact match
- `GET /person?surname={surname}` or `?dob={dob}` - Find people by partial match

---

## Configuration

### Environment Variables

```bash
# Server Configuration
PORT=3004                    # HTTP server port (default: 3004)

# Proxy Configuration
PROXY_URL=http://0.0.0.0:8099/proxy  # Proxy base URL

# External Services
JSONPLACEHOLDER_TARGET=https://jsonplaceholder.typicode.com
EXTERNAL_USER_TARGET=http://0.0.0.0:3006

# HTTP Client Configuration
HTTP_TIMEOUT=10s             # HTTP client timeout (default: 10s)
HTTP_MAX_IDLE_CONNS=100      # Max idle connections (default: 100)
HTTP_IDLE_CONN_TIMEOUT=90s   # Idle connection timeout (default: 90s)
```

### Configuration Best Practices

1. **Environment Variables**: Use environment variables for all configuration
2. **Sensible Defaults**: Provide defaults for non-critical configuration
3. **Validation**: Validate configuration at startup and fail fast if invalid
4. **No Secrets**: Never hardcode secrets or credentials

---

## Technical Implementation Details

### HTTP Framework

**stdlib `net/http`** with custom routing:

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/user/", userHandler.HandleUser)
mux.HandleFunc("/api/person", personHandler.HandlePerson)

server := &http.Server{
    Addr:         ":3004",
    Handler:      mux,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

### Dependency Injection

**Google Wire** for compile-time DI:

```go
// +build wireinject

package main

import "github.com/google/wire"

func InitializeServer() (*http.Server, error) {
    wire.Build(
        user.NewClient,
        user.NewService,
        user.NewHandler,
        person.NewClient,
        person.NewService,
        person.NewHandler,
        // ... other providers
        NewServer,
    )
    return &http.Server{}, nil
}
```

### Error Handling

**Custom error types** with proper HTTP status mapping:

```go
// common/errors/errors.go
package errors

type AppError struct {
    Code    string // Error code for client reference
    Message string // Human-readable message
    Err     error  // Underlying error
    Status  int    // HTTP status code
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// Common errors
var (
    ErrNotFound    = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
    ErrBadRequest  = &AppError{Code: "BAD_REQUEST", Message: "Invalid request", Status: 400}
    ErrInternal    = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
)
```

### HTTP Response Helpers

```go
// common/response/response.go
package response

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, err error) {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        JSON(w, appErr.Status, map[string]string{"error": appErr.Message})
        return
    }
    JSON(w, 500, map[string]string{"error": "Internal server error"})
}
```

### HTTP Client with Proxy

```go
// common/httpclient/client.go
package httpclient

type Client struct {
    httpClient *http.Client
    proxyURL   string
}

func New(proxyURL string, timeout time.Duration) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: timeout},
        proxyURL:   proxyURL,
    }
}

func (c *Client) Get(ctx context.Context, target string) (*http.Response, error) {
    // Build proxy URL: http://proxy?target={encoded_target}
    encodedTarget := url.QueryEscape(target)
    proxyURL := fmt.Sprintf("%s?target=%s", c.proxyURL, encodedTarget)

    req, err := http.NewRequestWithContext(ctx, "GET", proxyURL, nil)
    if err != nil {
        return nil, err
    }

    // Disable compression for recording compatibility
    req.Header.Set("Accept-Encoding", "identity")

    return c.httpClient.Do(req)
}
```

---

## Testing Strategy

### Test Coverage Requirements

- **Minimum Coverage**: 80%
- **Target Coverage**: 90%+
- **Critical Paths**: 100% coverage for business logic

### Test Types

1. **Unit Tests** (70% of tests):
   - Handler tests with mocked services
   - Service tests with mocked clients
   - Client tests with mocked HTTP responses
   - Table-driven tests for all functions

2. **Integration Tests** (20% of tests):
   - End-to-end tests with test server
   - Tests against actual proxy in playback mode
   - Error scenario testing

3. **E2E Tests** (10% of tests):
   - Full system tests with all dependencies
   - Test suite compatibility verification

### Test Structure

```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name        string
        userID      int
        mockReturn  *User
        mockError   error
        want        *User
        wantErr     bool
        errContains string
    }{
        {
            name:   "success - valid user",
            userID: 1,
            mockReturn: &User{
                ID:    1,
                Name:  "John Doe",
                Email: "john@example.com",
            },
            want: &User{
                ID:    1,
                Name:  "John Doe",
                Email: "john@example.com",
            },
        },
        {
            name:        "error - user not found",
            userID:      999,
            mockError:   errors.ErrNotFound,
            wantErr:     true,
            errContains: "not found",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Mocking Strategy

**Using mockery** to generate mocks:

```bash
# Generate mocks for all interfaces
mockery --name=UserClient --dir=./internal/user --output=./internal/user/mocks
mockery --name=UserService --dir=./internal/user --output=./internal/user/mocks
```

**Interface definitions**:

```go
// internal/user/service.go
type UserClient interface {
    GetUser(ctx context.Context, id int) (*User, error)
    GetPosts(ctx context.Context, userID int) ([]Post, error)
    GetTodos(ctx context.Context, userID int) ([]Todo, error)
}

type UserService interface {
    GetUser(ctx context.Context, id int) (*User, error)
    GetUserSummary(ctx context.Context, id int) (*UserSummary, error)
    GetUserReport(ctx context.Context, id int, req ReportRequest) (*UserReport, error)
}
```

---

## Development Workflow

### TDD Cycle

1. **Red**: Write failing test
2. **Green**: Implement minimal code to pass
3. **Refactor**: Clean up while keeping tests green
4. **Commit**: Commit working code with tests

### Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/user/...

# Run tests with race detector
go test -race ./...

# Lint code
golangci-lint run

# Generate mocks
go generate ./...

# Build binary
go build -o bin/server ./cmd/server

# Run server
./bin/server

# Hot reload (using air)
air
```

### File Structure

```
rest-v2/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── user/                    # User domain vertical slice
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── models.go
│   │   └── mocks/
│   ├── person/                  # Person domain vertical slice
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── models.go
│   │   └── mocks/
│   └── common/                  # Shared utilities
│       ├── httpclient/
│       │   ├── client.go
│       │   └── client_test.go
│       ├── errors/
│       │   ├── errors.go
│       │   └── errors_test.go
│       └── response/
│           ├── response.go
│           └── response_test.go
├── wire.go                      # Wire DI configuration
├── wire_gen.go                  # Generated Wire code
├── go.mod
├── go.sum
├── .air.toml                    # Air hot reload config
├── .golangci.yml               # Linting configuration
├── SPEC.md                      # This file
├── DEVELOPER_PROFILE.md         # Developer profile
└── README.md                    # Project documentation
```

---

## Success Criteria

### Functional Requirements

- ✅ All 5 endpoints implemented and working
- ✅ Successful integration with jsonplaceholder.typicode.com via proxy
- ✅ Successful integration with rest-external-user via proxy
- ✅ All 40 test cases passing (from config.comprehensive.json)
- ✅ Proper error handling and validation
- ✅ Correct HTTP status codes returned

### Non-Functional Requirements

- ✅ 80%+ test coverage across all packages
- ✅ All tests passing (`go test ./...`)
- ✅ Clean linting (`golangci-lint run`)
- ✅ Well-documented code (godoc for all exported functions)
- ✅ Follows Go best practices and idioms
- ✅ Proper dependency injection with Wire
- ✅ Vertical slice architecture maintained
- ✅ Fast build times (<10 seconds)
- ✅ Low memory footprint

---

## Next Steps

1. **Initialize Project**: Create go.mod, install dependencies
2. **Setup Tooling**: Configure Wire, mockery, golangci-lint, air
3. **Common Package**: Implement HTTP client, errors, response helpers (TDD)
4. **User Domain**: Implement user endpoints with full TDD workflow
5. **Person Domain**: Implement person endpoints with full TDD workflow
6. **Integration**: Wire everything together with DI
7. **E2E Testing**: Run full test suite (40 test cases)
8. **Documentation**: Complete README with usage examples
9. **Optimization**: Profile and optimize if needed

---

**Version**: 1.0
**Last Updated**: 2025-11-14
**Status**: Ready for Implementation
