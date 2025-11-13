# Prox - Proxy Testing Framework for REST API Comparison

A comprehensive testing framework for comparing REST API implementations using proxy recording and playback capabilities.

---

## ⚡ For AI Assistants: Quick Start

**🚨 CRITICAL - READ THESE DOCS IN ORDER:**

1. **First:** [`v1-v2-check-and-fix/TESTING-WORKFLOW.md`](v1-v2-check-and-fix/TESTING-WORKFLOW.md) - Understanding the workflow
2. **Second:** [`v1-v2-check-and-fix/FIX-PROCESS.md`](v1-v2-check-and-fix/FIX-PROCESS.md) - **MANDATORY** step-by-step fix process

**⚠️ DO NOT start fixing endpoints until you have read FIX-PROCESS.md!**

The FIX-PROCESS.md contains:
- **Branching strategy** - Create feature branches for each fix
- **TDD workflow** - Write failing tests FIRST, then fix
- **Build and verification** - Rebuild and test properly
- **AI communication requirements** - What to report at each step

**🎯 When you see: "please read the readme we want to follow the process and fix"**

This is your signal to:
1. **READ FIX-PROCESS.md** (if you haven't already)
2. Start the testing workflow (`./initialize-workflow.sh`)
3. Run the reporter to see what's failing
4. **For EACH failing endpoint**, follow the FIX-PROCESS.md exactly
5. Re-run the reporter after each fix to verify progress
6. Continue until all endpoints pass

**TL;DR:** Use these scripts in this order:
```bash
cd v1-v2-check-and-fix
./initialize-workflow.sh              # First time setup
./run-reporter.sh config.comprehensive.json  # See what's failing
# → Read FIX-PROCESS.md and follow it for EACH endpoint fix
```

---

## 🎯 Overview

Prox helps you test and compare two versions of REST APIs by:
- Recording external API calls through a proxy server
- Playing back cached responses for deterministic testing
- Comparing responses between API v1 and v2
- Generating detailed reports of matches and mismatches

## 🚀 Quick Start

```bash
# Navigate to the test orchestration directory
cd v1-v2-check-and-fix

# Initialize workflow (first time setup)
./initialize-workflow.sh

# Run comparison tests
./run-reporter.sh config.comprehensive.json

# View results
cat reports/report_*.md | head -50
```

## 📁 Repository Structure

```
.
├── v1-v2-check-and-fix/    # Test orchestration and configs
│   ├── config.*.json       # Test configurations
│   ├── initialize-workflow.sh # First-time setup (auto-detects mode)
│   ├── start.sh            # Start services (record/playback mode)
│   ├── run-reporter.sh     # Run comparison tests
│   ├── remove.sh           # Cleanup script
│   ├── README.md           # Usage guide
│   └── TESTING-WORKFLOW.md # Detailed workflow documentation
│
├── prroxy/                 # Main proxy and API implementations
│   ├── proxy/              # Go proxy server (record/playback)
│   ├── rest-v1/            # TypeScript REST API (v1)
│   ├── rest-v2/            # Go REST API (v2, hexagonal)
│   └── rest-external-user/ # Mock external API service (port 3006)
│
├── reporter/               # Go comparison tool
│   └── cmd/reporter/       # CLI for comparing responses
│
└── utils/                  # Helper scripts
```

## 🔧 Components

### 1. Proxy Server (prroxy/proxy/)
- Records HTTP interactions to disk
- Replays cached responses in playback mode
- Supports multiple target services
- Port: 8099

### 2. REST APIs
- **prroxy/rest-v1** (TypeScript/Express) - Port 3002
- **prroxy/rest-v2** (Go/Hexagonal) - Port 3004
- **prroxy/rest-external-user** (Go/Gin) - Port 3006 - Mock external service

### 3. Reporter Tool
- Compares API responses
- Generates markdown reports
- Shows detailed mismatches
- Configurable endpoint testing

## 📝 Available Test Configs

- **config.person-lookup.json** - Full person search (25 endpoints)
- **config.person-by-surname.json** - Surname-only search (5 endpoints)
- **config.person-by-dob.json** - DOB-only search (5 endpoints)
- **config.user-endpoints.json** - External API tests (10 endpoints)

## 🎓 Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm/yarn

### Installation

```bash
# Clone the repository
git clone https://github.com/jonathanleahy/prox.git
cd prox

# Install dependencies
cd prroxy/proxy && go mod download && cd ../..
cd prroxy/rest-v1 && npm install && cd ../..
cd reporter && go mod download && cd ..
```

### Running Tests

**Quick Test (Record & Compare):**
```bash
cd compare-v1-v2
./test-record.sh config.person-lookup.json
```

**Fast Test (Playback):**
```bash
cd compare-v1-v2
./test-playback.sh config.person-lookup.json
```

**Manual Control:**
```bash
cd compare-v1-v2
PROXY_MODE=record ./start.sh
./run-reporter.sh config.person-lookup.json
./remove.sh  # Cleanup
```

## 📊 Test Modes

### Record Mode
- Captures all external API calls
- Stores responses in `recordings/`
- Use for initial data capture or refresh

### Playback Mode
- Uses cached responses
- No external API calls
- Deterministic, faster testing

## 🧪 Creating Custom Tests

Create a new config file in `compare-v1-v2/`:

```json
{
  "base_url_v1": "http://0.0.0.0:3002",
  "base_url_v2": "http://0.0.0.0:3004",
  "iterations": 1,
  "endpoints": [
    {
      "path": "/api/person?surname=Smith&dob=1990-01-01",
      "method": "GET"
    }
  ]
}
```

Run your test:
```bash
./test-record.sh config.custom.json
```

## 📚 Documentation

- [v1-v2-check-and-fix/README.md](v1-v2-check-and-fix/README.md) - Detailed usage guide
- [v1-v2-check-and-fix/TESTING-WORKFLOW.md](v1-v2-check-and-fix/TESTING-WORKFLOW.md) - Comprehensive workflow documentation
- [prroxy/README.md](prroxy/README.md) - Proxy implementation details
- [reporter/README.md](reporter/README.md) - Reporter tool details

## 🏗️ Architecture

```
┌─────────────┐    ┌─────────────┐    ┌──────────────────┐
│  prroxy/    │───▶│   prroxy/   │───▶│   prroxy/        │
│  rest-v1    │    │   proxy     │    │   rest-external- │
│  (TS/Node)  │    │  (Record/   │    │   user (Go/Gin)  │
│  Port 3002  │    │   Playback) │    │   Port 3006      │
└─────────────┘    │  Port 8099  │    └──────────────────┘
                   └─────────────┘
┌─────────────┐           │
│  prroxy/    │           │
│  rest-v2    │           │
│  (Go/Hex)   │           ▼
│  Port 3004  │    ┌─────────────┐
└─────────────┘    │  Reporter   │
                   │  (Compare)  │
                   └─────────────┘
```

## 🤝 Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT

## 🔗 Links

- [GitHub Repository](https://github.com/jonathanleahy/prox)
- [Issue Tracker](https://github.com/jonathanleahy/prox/issues)
