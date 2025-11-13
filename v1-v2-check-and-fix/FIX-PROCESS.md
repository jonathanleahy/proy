# Endpoint Fix Process

**🚨 FOR AI ASSISTANTS: READ THIS ENTIRE DOCUMENT BEFORE FIXING ANY ENDPOINT! 🚨**

**This is the MANDATORY step-by-step process you MUST follow for every endpoint fix.**

---

This document outlines the systematic process for fixing each failing endpoint, one at a time.

## ⚠️ CRITICAL: AI Communication Requirements

**BEFORE STARTING EACH FIX, THE AI MUST:**

1. **Clearly state what endpoint is being fixed**
   - Example: "I'm fixing GET /api/user/:id/summary"

2. **Explain the issue**
   - What is currently broken?
   - What does the test expect?
   - Why is v2 failing?
   - Example: "V2 currently returns 404 but should return user data with post count"

3. **Summarize the implementation plan**
   - What service method will be created/modified?
   - What data transformation is needed?
   - How will v1 behavior be replicated?

4. **During implementation, provide progress updates**
   - "Writing the test that proves the problem..."
   - "Implementing the service method..."
   - "Rebuilding and testing..."
   - "Running full system comparison..."

5. **After the fix, report the results**
   - Did the test pass?
   - Did the full system comparison succeed?
   - What's the new failure count?
   - Is the user ready to move to the next endpoint?

6. **Use the TodoWrite tool** to track progress so the user can see:
   - What endpoint is currently being worked on (in_progress)
   - What's pending
   - What's been completed

---



## The Fix Workflow

For each failing endpoint, follow this process exactly:

### 1. Create a Feature Branch
```bash
git checkout -b fix/endpoint-name
```
Use a descriptive name like `fix/user-summary-endpoint` or `fix/person-lookup-response`

### 2. Write a Failing Test
Before making any code changes, write a test that:
- Proves the problem exists
- Clearly documents what the endpoint should do
- Will pass once the fix is complete

```bash
cd prroxy/rest-v2
# Create or modify test in tests/integration/
# Test should FAIL when you first run it
go test ./tests/integration/... -v -run TestName
```

**Why test first?**
- Forces you to understand exactly what's broken
- Gives you a clear definition of "fixed"
- Prevents the problem from coming back later

### 3. Fix the Code
Implement the fix in the rest-v2 service. Reference the v1 implementation and the recorded responses from the proxy to understand what the endpoint should do.

### 4. Rebuild the System
```bash
cd prroxy/rest-v2
go build -o rest-v2 ./cmd/server
```

### 5. Run Tests
Verify your test now passes:
```bash
cd prroxy/rest-v2
go test ./tests/integration/... -v -run TestName
```

All tests should pass, and specifically your new test should now pass.

### 6. Run Full System Comparison
Verify the fix works in the complete environment:
```bash
cd v1-v2-check-and-fix
./remove.sh                              # Clean up
./play-tests.sh                          # Start in playback + run tests
```

**Check the report:**
- Your previously failing endpoint should now show ✅
- No new failures should appear
- Overall failure count should decrease by 1

### 7. Commit and Push
```bash
git add .
git commit -m "Fix endpoint name

- Implement [description of what was fixed]
- Add comprehensive tests
- Matches v1 behavior exactly"

git push origin fix/endpoint-name
```

## Key Rules

1. **One endpoint per branch** - Keep changes focused and atomic
2. **Always test first** - Write the failing test before fixing code
3. **Always verify in full system** - Use run-reporter.sh to confirm the fix works end-to-end
4. **No broken tests** - All tests must pass before moving to the next endpoint
5. **Document your findings** - If you discover anything about the endpoint, add comments to the test

## Identifying Which Endpoint to Fix

After running the reporter, look at the report:
```bash
cat reports/report_*.md | grep "❌"
```

This shows all failing endpoints. Pick one and follow the process above.

## Example Fix Workflow

```bash
# 1. See what's failing
cd v1-v2-check-and-fix
./play-tests.sh                          # Test v2 against v1 baseline
cat reports/report_*.md | grep "❌"

# 2. Start fixing (example: user summary endpoint)
git checkout -b fix/user-summary-endpoint

# 3. Write failing test
cd ../prroxy/rest-v2/tests/integration
# Create test_user_summary.go
go test ./tests/integration -v -run TestUserSummary  # Should FAIL

# 4. Fix the code
# Edit cmd/server/handlers.go or services/userService.go
# Implement the endpoint based on v1 behavior

# 5. Rebuild
go build -o rest-v2 ./cmd/server

# 6. Test passes
go test ./tests/integration -v -run TestUserSummary  # Should PASS

# 7. Full system check
cd ../../..  # Back to v1-v2-check-and-fix
./remove.sh
./play-tests.sh                          # Test again
./run-reporter.sh config.comprehensive.json
# Check report - your endpoint should now be ✅

# 8. Commit
git add .
git commit -m "Fix user summary endpoint

- Implement GET /api/user/:id/summary
- Add integration tests
- Matches v1 response format exactly"
git push origin fix/user-summary-endpoint
```

## Troubleshooting

**Test won't compile?**
- Check Go syntax in your test file
- Make sure imports are correct
- Run `go fmt` to format the file

**Endpoint fix doesn't work in reporter?**
- Check tmp/rest-v2.log for errors
- Make sure rebuild succeeded (`go build`)
- Verify the endpoint path matches exactly (case-sensitive)

**Services won't start?**
- Run `./remove.sh` to clean up
- Check log files in `tmp/`
- Ensure ports 3002, 3004, 3006, 8099 are available

**Reporter shows different failure?**
- Make sure you rebuilt: `go build -o rest-v2 ./cmd/server`
- Services may need time to start, try waiting 5 seconds before running reporter
- Check that you're using playback mode: `PROXY_MODE=playback`
