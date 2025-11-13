# Quick Start Guide - API Comparison Framework

**For developers and AI assistants**

## 🚀 First Time Setup

```bash
cd v1-v2-check-and-fix

# Record v1 baseline behavior (this will take a few minutes)
./record-tests.sh
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
# 1. Start services and run tests (playback mode - uses cached recordings)
./play-tests.sh

# 2. Check what's failing
cat reports/report_*.md | grep "❌"

# 3. Fix code in rest-v2

# 4. Rebuild
cd ../prroxy/rest-v2
go build -o rest-v2 ./cmd/server
cd ../../v1-v2-check-and-fix

# 5. Restart and test again
./remove.sh
./play-tests.sh

# 6. Repeat until all endpoints pass
```

## 🤖 What to Say to AI Assistants

### Initial Setup
```
"Please read the README and action
```

## 📁 Key Files

| File | Purpose |
|------|---------|
| `record-tests.sh` | Record v1 baseline (first time setup) |
| `play-tests.sh` | Test v2 against v1 baseline (daily use) |
| `run-reporter.sh` | Run comparison tests manually |
| `remove.sh` | Stop all services |
| `config.*.json` | Endpoint test configurations |
| `env.linux` / `env.darwin` | OS-specific service ports and paths |
| `FIX-PROCESS.md` | **MANDATORY** TDD fix process |
| `TESTING-WORKFLOW.md` | Detailed workflow documentation |

## 🎯 Common Workflows

### Fresh Capture (V1 Changed)
```bash
./record-tests.sh
```

### Fast Development (Reuse Recordings)
```bash
./play-tests.sh                               # Start + test in playback
# ... fix code ...
./remove.sh && ./play-tests.sh                # Restart + test again
```

### Record Fresh Data (Force)
```bash
PROXY_MODE=record ./play-tests.sh             # Override to record mode
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
│   record-tests.sh   │  ← First time: Capture V1 behavior
└──────────┬──────────┘
           │
           ├─→ RECORD mode: V1 responses saved
           ├─→ Creates recordings/
           ├─→ Runs tests automatically
           └─→ Generates first report

┌─────────────────────┐
│   play-tests.sh     │  ← Development: Fast iteration
└──────────┬──────────┘
           │
           ├─→ PLAYBACK mode: Uses recordings
           ├─→ No external API calls
           ├─→ Runs tests automatically
           └─→ Very fast testing

┌─────────────────────┐
│   run-reporter.sh   │  ← Manual test execution
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
2. **Update OS-specific env file** (`env.linux` or `env.darwin`) with your service paths and ports
3. **Create config files** listing your endpoints
4. **Set custom start commands** in env file if needed (e.g., `REST_V1_START_COMMAND="./gradlew run"`)
5. **Run `./record-tests.sh`** to capture your V1 behavior
6. **Follow the fix process** in FIX-PROCESS.md

See the main README for detailed configuration instructions.

---

**Remember:** The framework is API-agnostic. It works with any REST API as long as you configure the endpoints and service locations!
