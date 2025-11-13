# 🔄 HTTP Testing Proxy

A simple, elegant HTTP proxy designed specifically for testing teams to record and replay HTTP interactions. Built with Go using TDD and BDD methodologies, following professional design patterns for maintainability and ease of use.

## 🌟 Features

- **🔴 Record Mode**: Capture all HTTP requests and responses
- **▶️ Playback Mode**: Replay recorded interactions for consistent testing
- **🎯 Full Request Matching**: Ensures exact match of URL, method, headers, and body
- **📁 Organized Storage**: Recordings organized by service in JSON format
- **🎮 Web Dashboard**: User-friendly UI for managing recordings
- **📊 Statistics**: Track hits, misses, and recording counts
- **🐳 Docker Support**: Easy deployment with container support
- **🚀 Zero Config**: Works out of the box with sensible defaults

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Build and run with Docker
make docker
make docker-run

# Access the proxy (using a real API endpoint)
curl "http://0.0.0.0:8080/proxy?target=jsonplaceholder.typicode.com/users"

# View dashboard
open http://0.0.0.0:8080/admin/ui
```

### Local Installation

```bash
# Install dependencies
make deps

# Build the proxy
make build

# Run the proxy
make run
```

## 📖 Usage Guide

### Basic Proxy Usage

The proxy works by intercepting requests sent to it and either forwarding them (record mode) or returning saved responses (playback mode).

> **Note**: The examples below use real, publicly available APIs. Replace these URLs with your actual API endpoints when testing your applications.

#### Record Mode
```bash
# Switch to record mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"record"}'

# Make a request through the proxy (using a real API)
curl "http://0.0.0.0:8080/proxy?target=jsonplaceholder.typicode.com/users"
```

#### Playback Mode
```bash
# Switch to playback mode
curl -X POST http://0.0.0.0:8080/admin/mode \
  -H "Content-Type: application/json" \
  -d '{"mode":"playback"}'

# Same request will return the recorded response
curl "http://0.0.0.0:8080/proxy?target=jsonplaceholder.typicode.com/users"
```

#### Real-World Examples
```bash
# JSONPlaceholder (Testing API)
curl "http://0.0.0.0:8080/proxy?target=jsonplaceholder.typicode.com/posts/1"

# GitHub API
curl "http://0.0.0.0:8080/proxy?target=api.github.com/users/github"

# HTTPBin (Testing Service)
curl "http://0.0.0.0:8080/proxy?target=httpbin.org/json"

# Your own API
curl "http://0.0.0.0:8080/proxy?target=your-api.com/v1/users"
```

### Management API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/status` | GET | View current status and statistics |
| `/admin/mode` | GET/POST | Get or set current mode (record/playback) |
| `/admin/recordings` | GET | List all recordings |
| `/admin/recordings` | DELETE | Clear all recordings |
| `/admin/ui` | GET | Web dashboard interface |
| `/health` | GET | Health check endpoint |

### Web Dashboard

Access the dashboard at `http://0.0.0.0:8080/admin/ui` for a visual interface to:
- Switch between record and playback modes
- View all recorded interactions
- Monitor statistics
- Clear recordings

## 🔧 Configuration

### Command Line Flags

```bash
./proxy --port=8080 --recordings-dir=./recordings --mode=record
```

### Environment Variables

```bash
export PROXY_PORT=8080
export PROXY_HOST=0.0.0.0
export PROXY_RECORDINGS_DIR=./recordings
export PROXY_MODE=playback
export PROXY_TLS_SKIP_VERIFY=true
```

### Configuration File

Create `proxy.yaml`:

```yaml
server:
  port: 8080
  host: 0.0.0.0
storage:
  type: filesystem
  path: ./recordings
mode:
  default: playback
tls:
  skip_verify: true
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Generate coverage report
make coverage

# Run specific test
go test -v ./internal/storage/...
```

### BDD Scenarios

The proxy includes comprehensive BDD test scenarios in `tests/bdd/proxy.feature`:
- Recording new interactions
- Replaying recorded interactions
- Full request matching
- Mode switching
- Recording management

## 🏗️ Architecture

### Design Patterns Used

- **Repository Pattern**: Storage abstraction for recordings
- **Strategy Pattern**: Record/Playback mode implementations
- **Middleware Pattern**: Request processing pipeline
- **Factory Pattern**: Handler creation
- **Singleton Pattern**: Configuration management

### Project Structure

```
proxy/
├── cmd/proxy/           # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # Request middleware
│   ├── mode/          # Record/Playback implementations
│   ├── models/        # Data models
│   └── storage/       # Storage repository
├── web/               # Dashboard UI
├── tests/             # Test files
├── Dockerfile         # Container configuration
├── Makefile          # Build commands
└── README.md         # Documentation
```

## 🔄 Integration with CRM API

### Update docker-compose.yml

Add the proxy service to your existing `docker-compose.yml`:

```yaml
services:
  proxy:
    build: ./proxy
    ports:
      - "8080:8080"
    volumes:
      - ./proxy/recordings:/app/recordings
    environment:
      - PROXY_MODE=playback
```

### Update CRM API Environment

Modify `scripts/local/crm-api.env` to route through proxy:

```bash
# Instead of direct URLs
EVENTS_API_URL=http://api-events-dev.pismolabs.io

# Use proxy URLs
EVENTS_API_URL=http://proxy:8080/proxy?target=api-events-dev.pismolabs.io
```

## 📝 Make Commands

```bash
make help              # Show all available commands
make build            # Build the proxy binary
make test             # Run tests
make run              # Start the proxy
make docker           # Build Docker image
make docker-run       # Run in Docker
make clean            # Clean build artifacts
make clean-recordings # Clear all recordings
make stats            # Show recording statistics
```

## 🎯 Use Cases

### 1. Testing Without External Dependencies
Record interactions once, then run tests in playback mode without requiring access to external services.

### 2. Consistent Test Data
Ensure tests always receive the same responses, eliminating flakiness from external service variability.

### 3. Offline Development
Continue development and testing even when external services are unavailable.

### 4. Performance Testing
Eliminate network latency to focus on application performance.

### 5. Debugging
Inspect exact requests and responses for troubleshooting.

## 📊 Recording Storage

Recordings are stored as JSON files organized by service:

```
recordings/
├── jsonplaceholder_typicode_com/
│   ├── <hash1>.json
│   └── <hash2>.json
├── api_github_com/
│   └── <hash3>.json
├── your_api_com/
│   └── <hash4>.json
```

Each recording contains:
- Request details (method, URL, headers, body)
- Response details (status, headers, body)
- Metadata (target service, duration)

## 🔒 Security Notes

- The proxy accepts self-signed certificates by default (configurable)
- No authentication is implemented (add as needed for production)
- Recordings may contain sensitive data - secure appropriately

## 🤝 Contributing

This proxy was built using Test-Driven Development (TDD) and Behavior-Driven Development (BDD) practices. When contributing:

1. Write tests first
2. Implement functionality
3. Ensure all tests pass
4. Update documentation

## 📄 License

Internal use only. Developed for testing purposes.

## 🆘 Support

For issues or questions:
1. Check the dashboard at `http://0.0.0.0:8080/admin/ui`
2. View logs with `docker logs testing-proxy`
3. Check recordings in the `./recordings` directory

## 🎓 Tips for Testers

1. **Start in Record Mode**: Capture all interactions first
2. **Switch to Playback**: Run tests with consistent data
3. **Use the Dashboard**: Visual management is easier
4. **Clear Periodically**: Keep recordings organized
5. **Check Statistics**: Monitor cache hits/misses

## 🚦 Status Codes

- `200`: Successful proxy operation
- `404`: No recording found (playback mode)
- `400`: Invalid request (missing target)
- `500`: Proxy internal error

---

Built with ❤️ for the testing team using Go, TDD, and clean architecture principles.