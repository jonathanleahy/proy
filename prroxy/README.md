# Prroxy - API Migration Testing System

Automated toolkit for achieving API compatibility during migrations, regardless of technology stack.

## 🚀 Quick Start

### For AI Assistants
**→ [AI-README.md](AI-README.md)** - Complete automated workflow instructions

### For Humans
1. **Start Services** (proxy, source-of-truth API, new implementation)
2. **Run Comparison** (`./scripts/compare-v1-v2-with-reports.sh`)
3. **Fix Next Issue** (`./scripts/generate-next-issue.sh`)
4. **Repeat** until desired compatibility achieved

## 📊 Current Status

Run comparison to check:
```bash
./scripts/compare-v1-v2-with-reports.sh
cat test-data/comparison-report.md
```

## 🔧 Core Workflow

### 0. Record Test Data (One-Time Setup)

**Don't have test data?** See **[docs/AI-TEST-GENERATION.md](docs/AI-TEST-GENERATION.md)** for using AI to generate comprehensive test cases from your source code, swagger docs, or API examples.

```bash
# Start proxy in RECORD mode
MODE=record ./proxy/start.sh

# Start source-of-truth API
./rest-v1/start.sh

# Make API calls to record test data
curl http://0.0.0.0:3002/api/user/1
curl http://0.0.0.0:3002/api/user/1/summary
# ... make all the API calls you want to test

# Stop proxy - recordings saved to proxy/recordings/
```

### 1. Start All Services
```bash
# Start in playback mode (uses recorded data)
./proxy/start.sh          # Port 8080 (playback mode)
./rest-v1/start.sh        # Port 3002 (source-of-truth)
./rest-v2/start.sh        # Port 3004 (new implementation)
```

### 2. Generate Comparison Report
```bash
# Compare APIs and generate reports
./scripts/compare-v1-v2-with-reports.sh

# Or with custom ports
REST_V1_URL=http://0.0.0.0:3000 \
REST_V2_URL=http://0.0.0.0:8082 \
./scripts/compare-v1-v2-with-reports.sh
```

### 3. Fix Next Issue
```bash
# Generate prioritized issue report
./scripts/generate-next-issue.sh

# Read fix instructions
cat test-data/next-issue.md

# Apply fix, rebuild, restart service
# Then re-run comparison
```

### 4. Track Progress
- **Pass Rate**: Check `test-data/comparison-report.md`
- **Current Issue**: Check `test-data/next-issue.md`
- **Detailed Tests**: Check `test-data/detailed-reports/`

## 📁 Project Structure

```
prroxy/
├── AI-README.md              # 🤖 AI workflow instructions
├── proxy/                    # Record/replay server
├── rest-v1/                  # Source-of-truth API
├── rest-v2/                  # New implementation
├── scripts/
│   ├── compare-v1-v2-with-reports.sh
│   └── generate-next-issue.sh
└── test-data/
    ├── comparison-report.md  # Current status
    ├── next-issue.md        # What to fix next
    └── detailed-reports/    # Individual tests
```

## 🎯 Success Targets

- **100%**: Perfect compatibility (drop-in replacement)
- **95%+**: Production-ready
- **80%+**: Acceptable for gradual migration
- **<80%**: Needs work

## 🛠️ Troubleshooting

**Services won't start?**
- Check ports (8080, 3002, 3004)
- Check dependencies installed
- Review logs

**Comparison fails?**
- Ensure all 3 services running
- Check proxy has recordings
- Try with explicit URLs (REST_V1_URL, REST_V2_URL)

**Fix didn't work?**
- Verify rebuild completed
- Confirm service restarted
- Check for caching

## 📚 Documentation

### Workflows
- **[AI-README.md](AI-README.md)** - Automated AI workflow (RECOMMENDED)
- **[docs/MIGRATION_HOW_TO_GUIDE.md](docs/MIGRATION_HOW_TO_GUIDE.md)** - Manual migration guide

### Technical Details
- **[docs/PROJECT_OVERVIEW.md](docs/PROJECT_OVERVIEW.md)** - System architecture
- **[docs/ENGINEERING_SUMMARY.md](docs/ENGINEERING_SUMMARY.md)** - Executive summary

### Components
- **[proxy/README.md](proxy/README.md)** - Proxy details
- **[scripts/README.md](scripts/README.md)** - Script documentation

## ⚡ Key Features

- **Technology Agnostic** - Works with any language/framework
- **Automated Issue Detection** - Prioritizes problems to fix
- **Incremental Progress** - Fix one issue at a time
- **No Manual Testing** - Uses recorded production data

## 🏃 Getting Started

1. Clone repository
2. **For AI**: Read [AI-README.md](AI-README.md)
3. **For Humans**: Follow workflow above
4. Run comparison
5. Fix issues systematically
6. Deploy when pass rate acceptable

---

**TL;DR**: Run comparison → Read next-issue.md → Fix → Repeat

For detailed information about the system, see [docs/PROJECT_OVERVIEW.md](docs/PROJECT_OVERVIEW.md)