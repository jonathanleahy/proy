# Quick Start Guide - API Comparison Framework

**For developers and AI assistants**

## 🚀 First Time Setup

```bash
cd v1-v2-check-and-fix

# Initialize and capture V1 behavior (this will take a few minutes)
./initialize-workflow.sh
```

**What this does:**
- Deletes old recordings, reports, and temporary files
- Starts services in RECORD mode
- Captures V1 API behavior as "ground truth"
- Runs full comparison and generates report
- Creates recordings for fast playback mode

## 📊 Check Results

```bash
# View the report
cat reports/report_*.md | head -50

# Or check for failures
cat reports/report_*.md | grep "❌"
```

**Report shows:**
- ✅ Passing endpoints (V1 and V2 match)
- ❌ Failing endpoints (differences found)
- Detailed comparison of response differences

## 🔧 Development Loop (Fixing Endpoints)

Once initialized, use this fast iteration cycle:

```bash
# 1. Start services in playback mode (uses cached recordings)
./start.sh

# 2. Run comparison
./run-reporter.sh config.comprehensive.json

# 3. Check what's failing
cat reports/report_*.md | grep "❌"

# 4. Fix code in rest-v2

# 5. Rebuild
cd ../prroxy/rest-v2
go build -o rest-v2 ./cmd/server
cd ../../v1-v2-check-and-fix

# 6. Restart and test
./remove.sh
./start.sh
./run-reporter.sh config.comprehensive.json

# 7. Repeat until all endpoints pass
```

## 🤖 What to Say to AI Assistants

### Initial Setup
```
"Please read the README and action
```

## 📁 Key Files

| File | Purpose |
|------|---------|
| `initialize-workflow.sh` | Full reset + record V1 behavior |
| `start.sh` | Start services (playback mode by default) |
| `run-reporter.sh` | Run comparison tests |
| `remove.sh` | Stop all services |
| `config.*.json` | Endpoint test configurations |
| `env` | Service ports and paths |
| `FIX-PROCESS.md` | **MANDATORY** TDD fix process |
| `TESTING-WORKFLOW.md` | Detailed workflow documentation |

## 🎯 Common Workflows

### Fresh Capture (V1 Changed)
```bash
./initialize-workflow.sh
```

### Fast Development (Reuse Recordings)
```bash
./start.sh                                    # Start in playback
./run-reporter.sh config.comprehensive.json   # Run tests
# ... fix code ...
./remove.sh && ./start.sh                     # Restart
./run-reporter.sh config.comprehensive.json   # Test again
```

### Record Fresh Data (Force)
```bash
PROXY_MODE=record ./start.sh                  # Override to record mode
./run-reporter.sh config.comprehensive.json
```

### Test Specific Endpoints
```bash
# Create a custom config
cp config.comprehensive.json config.my-test.json
# Edit config.my-test.json to include only your endpoints
./run-reporter.sh config.my-test.json
```

## 🎓 Understanding the Process

**Workflow:**
```
┌─────────────────────┐
│ initialize-workflow │  ← First time: Capture V1 behavior
└──────────┬──────────┘
           │
           ├─→ RECORD mode: V1 responses saved
           ├─→ Creates recordings/
           └─→ Generates first report

┌─────────────────────┐
│      start.sh       │  ← Development: Fast iteration
└──────────┬──────────┘
           │
           ├─→ PLAYBACK mode: Uses recordings
           ├─→ No external API calls
           └─→ Very fast testing

┌─────────────────────┐
│   run-reporter.sh   │  ← Compare V1 vs V2
└──────────┬──────────┘
           │
           ├─→ Calls both APIs
           ├─→ Compares responses
           └─→ Generates report

┌─────────────────────┐
│   FIX-PROCESS.md    │  ← Fix each endpoint
└──────────┬──────────┘
           │
           ├─→ Write test (TDD)
           ├─→ Fix code
           ├─→ Rebuild & verify
           └─→ Commit when passing
```

## ✅ Success Criteria

Your V2 API is ready when:
- ✅ All endpoints return 200 status (or expected status)
- ✅ Response data matches V1 exactly
- ✅ Report shows: "Passing: 100%, Failing: 0%"

## 🔄 For New Projects

To adapt this framework for your APIs:

1. **Copy the framework** to your project
2. **Update `env` file** with your service paths and ports
3. **Create config files** listing your endpoints
4. **Modify `start.sh`** if your services have different startup commands
5. **Run `initialize-workflow.sh`** to capture your V1 behavior
6. **Follow the fix process** in FIX-PROCESS.md

See the main README for detailed configuration instructions.

---

**Remember:** The framework is API-agnostic. It works with any REST API as long as you configure the endpoints and service locations!
