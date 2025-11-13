# AI Assistant Instructions for API Migration Compatibility

## Mission
Achieve functional compatibility between two API implementations through systematic comparison and fixes. The implementations may be in different languages, frameworks, or architectures.

## System Overview
- **Source-of-Truth API**: The reference implementation (existing/legacy system)
- **New Implementation API**: The replacement system (must match source-of-truth behavior)
- **Proxy**: Records and replays external API calls for consistent testing
- **Comparison System**: Automated tools to identify differences

## Workflow Process

### 1. Assess Current State
```bash
# Check what services exist
ls -la /home/jon/personal/prroxy/

# Review any existing documentation
find . -name "*.md" -o -name "README*" | head -10

# Understand the service architecture (ports, dependencies, etc.)
grep -r "PORT\|port" --include="*.env" --include="*.json" --include="*.yml" 2>/dev/null | head -20
```

### 2. Verify Services Can Build/Run
```bash
# The comparison script will tell you if services aren't running
# Each service has its own build/run mechanism
# Check for common patterns:
ls -la */Makefile */package.json */go.mod */pom.xml */build.* 2>/dev/null
```

### 3. Start Required Services
Three services must be running:
1. **Proxy server** - Records/replays external API calls
2. **Source-of-truth API** - The reference implementation (v1)
3. **New implementation API** - The version being migrated (v2)

Each service should have a `start.sh` script:
```bash
# Check for startup scripts
ls -la rest-*/start.sh proxy/start.sh

# If scripts exist, run them:
./proxy/start.sh           # Usually port 8080
./rest-v1/start.sh          # Usually port 3002 or 3000
./rest-v2/start.sh          # Usually port 3004 or 8082

# Or check README files for manual startup commands
```

Each service should have a stop.sh script:
```bash
# Check for stop scripts
ls -la rest-*/stop.sh proxy/stop.sh

# If scripts don't exist, create them to stop and kill the app running on the port
./proxy/stop.sh
./rest-v1/stop.sh
./rest-v2/stop.sh
```

**Note:** The comparison script expects:
- Source-of-truth API on port 3002 (configurable via REST_V1_URL env var)
- New implementation on port 3004 (configurable via REST_V2_URL env var)

### 4. Run Initial Comparison to Generate Baseline
**IMPORTANT:** Always run the comparison first to establish the current state and identify what needs fixing.

```bash
# Default configuration (ports 3002 and 3004)
/home/jon/personal/prroxy/scripts/compare-v1-v2-with-reports.sh

# Or with custom ports/URLs:
REST_V1_URL=http://0.0.0.0:3000 \
REST_V2_URL=http://0.0.0.0:8082 \
./scripts/compare-v1-v2-with-reports.sh

# This creates:
# - comparison-report.json (machine-readable data)
# - comparison-report.md (human summary)
# - detailed-reports/ (individual test analyses)
```

This initial run establishes:
- Current pass rate (baseline)
- Which endpoints are failing
- What types of failures exist
- Data needed for next-issue.md generation

### 5. Generate First Issue Report
```bash
# Automatically identify the highest-priority issue
/home/jon/personal/prroxy/scripts/generate-next-issue.sh

# Review the generated report
cat /home/jon/personal/prroxy/test-data/next-issue.md
```

### 6. Understand the Issue
The next-issue.md report provides:
- **Problem Description**: What's failing and why
- **Root Cause**: Technical analysis of the difference
- **Fix Location**: Which files/components need changes
- **Solution**: Specific changes required
- **Verification Steps**: How to test the fix

### 7. Apply the Fix
Follow the instructions in next-issue.md:
- Navigate to the specified file(s)
- Apply the recommended changes
- The report includes exact code snippets when possible

### 8. Rebuild Modified Service
```bash
# Find and run the appropriate build command
# Common patterns:
make build          # Makefile
npm run build       # Node.js
go build           # Go
mvn package        # Java/Maven
./gradlew build    # Java/Gradle
cargo build        # Rust
```

### 9. Test the Fix
```bash
# Restart the modified service (usually the new implementation)
# Use the service's restart script or kill and restart manually
pkill -f "rest-v2" && ./rest-v2/start.sh  # Example

# Test the specific failing endpoint
# The next-issue.md provides exact test commands

# Verify improvement
/home/jon/personal/prroxy/scripts/compare-v1-v2-with-reports.sh
```

### 10. Check Progress
Review the new comparison report:
- Did the pass rate improve?
- Did the specific issue get resolved?
- Were any regressions introduced?

### 11. Create Pull Request

When fixes are ready to merge:

```bash
# Create PR with summary of changes
gh pr create --title "Fix [issue description]" --body "$(cat <<'EOF'
## Summary
[Brief description of what was fixed]

## Changes
- File: [path]
- Issue: [what was wrong]
- Fix: [what was changed]
- Result: [expected improvement]

## Test Results
Before: X/Y tests passing
After: Expected improvement in [specific tests]

## Workflow Followed
1. Ran comparison to generate baseline
2. Generated next-issue.md
3. Applied fix per recommendations
4. Rebuilt service
5. Committed changes

🤖 Generated following AI-README workflow
EOF
)"
```

The PR link will be displayed - save it for review.

### 12. Iterate
Repeat steps 5-11 until reaching acceptable compatibility:
- Target may be 100% compatibility
- Or a defined subset (e.g., 80% of critical endpoints)
- Check project requirements for success criteria
- Each iteration creates a new PR for review

## Key Files and Their Purposes

### Reports (Generated)
- `test-data/comparison-report.md` - Overall compatibility status
- `test-data/next-issue.md` - Current issue to fix with instructions
- `test-data/detailed-reports/` - Individual endpoint analyses

### Scripts (Tools)
- `scripts/compare-v1-v2-with-reports.sh` - Main comparison tool
- `scripts/generate-next-issue.sh` - Issue prioritization tool

### Test Data
- `test-data/test-cases-with-responses.json` - Expected behaviors (if exists)
- `proxy/recordings/` - Captured external API interactions

## Decision Framework

### Issue Priority (Automated)
The system prioritizes issues by impact:
1. **Response mismatches** (same status code, different content)
2. **Status code differences** (e.g., 200 vs 500)
3. **Missing functionality** (endpoint not implemented)

### Complexity Assessment
- **Simple**: Response formatting (minutes to fix)
- **Medium**: Business logic differences (hours to fix)
- **Complex**: Architectural mismatches (may need redesign)

## Important Principles

1. **Source-of-Truth is Sacred**: New implementation must match source-of-truth's behavior exactly
2. **Data-Driven Fixes**: Let the reports guide your actions
3. **Test After Each Change**: Verify fixes don't break other endpoints
4. **Document Patterns**: Similar issues often have similar solutions
5. **Incremental Progress**: Each fix should improve the pass rate

## Common Patterns to Watch For

1. **Response Structure**: Field names, nesting, data types
2. **Status Codes**: Success/error code mismatches
3. **Data Completeness**: Missing or truncated data
4. **Format Differences**: Dates, numbers, strings
5. **Null Handling**: null vs undefined vs missing fields
6. **Array Ordering**: Sorted vs unsorted results
7. **Pagination**: Different default limits
8. **Error Messages**: Different error response formats

## Success Metrics

Track progress through:
- **Pass Rate**: Percentage of endpoints returning identical responses
- **Status Matches**: Percentage with correct HTTP status codes
- **Response Matches**: Percentage with identical response bodies
- **Regression Count**: Previously passing tests that now fail

## Troubleshooting

If comparison fails:
1. Check all services are running (proxy, source-of-truth, new implementation)
2. Verify services are on correct ports (default: 8080, 3002, 3004)
3. Check proxy has recordings for the endpoints
4. Review service logs for errors
5. Try with explicit environment variables:
   ```bash
   REST_V1_URL=http://0.0.0.0:YOUR_PORT \
   REST_V2_URL=http://0.0.0.0:YOUR_PORT \
   ./scripts/compare-v1-v2-with-reports.sh
   ```

If fixes don't work:
1. Verify the change was saved
2. Confirm service was rebuilt
3. Ensure service was restarted
4. Check for caching issues

## Final Verification

When you believe compatibility is achieved:
```bash
# Run final comparison
/home/jon/personal/prroxy/scripts/compare-v1-v2-with-reports.sh

# Success indicators:
# - Pass rate meets target (e.g., 80%, 100%)
# - No critical endpoints failing
# - No regressions from baseline
```

## Notes

- The system is language-agnostic (v1 and v2 can be any technology)
- External dependencies are simulated via proxy recordings
- The comparison is behavioral, not code-based
- Success is measured by API compatibility, not implementation details
- The next-issue.md report is self-contained with all fix details

---

**Remember**: Trust the reports. The system automatically identifies issues, prioritizes them, and provides specific fixes. Your role is to apply these fixes systematically until the desired compatibility level is achieved.