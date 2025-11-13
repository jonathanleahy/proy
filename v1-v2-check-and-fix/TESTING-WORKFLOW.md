# Testing Workflow Guide

> Think of this as your friendly companion for the sometimes-tricky journey of API validation. We'll help you systematically prove that your new v2 API behaves just like your trusted v1 API – without losing your sanity in the process.

---

## 🎯 The Big Picture

Imagine you're renovating a house. You want the new kitchen to work exactly like the old one – same light switches, same water pressure, same everything – but with a fresh look underneath. That's essentially what we're doing here with your APIs.

This guide walks you through a proven process to:
- Capture how your v1 API actually behaves in the real world
- Compare your v2 implementation against that behavior
- Find and fix the differences systematically
- Build confidence that nothing breaks when you switch over

**The beauty of this approach:** You'll catch subtle differences that might otherwise only show up when real users start hitting your API. Better to find them now, right?

---

## 🏃‍♂️ The "I Just Want to Get Started" Version

### ⚡ FOR AI ASSISTANT: Use These Scripts (Do NOT do things manually)

**Available Scripts:**

| Script | Purpose | When to Run |
|--------|---------|------------|
| `./record-tests.sh` | Records v1 baseline behavior in RECORD mode, runs tests automatically | First time setup, or when v1 changes |
| `./play-tests.sh` | Starts services in PLAYBACK mode, runs tests automatically (fast iteration) | Development loop - fixing endpoints |
| `./run-reporter.sh config.comprehensive.json` | Manually runs specific test suite and generates a report showing which endpoints pass/fail | When you want to test without restarting services |
| `./remove.sh` | Stops all running services | When you need to clean up |

**AI Workflow (Automatic - USE THESE):**
```bash
# Step 1: First time - record v1 baseline
cd v1-v2-check-and-fix
./record-tests.sh

# Step 2: Development loop - test v2 against v1 baseline
./play-tests.sh

# Step 3: When fixing (repeat this cycle):
# - Make code changes to rest-v2
# - Rebuild: cd ../prroxy/rest-v2 && go build -o rest-v2 ./cmd/server
# - Test again
./remove.sh
./play-tests.sh
```

**Manual Version (if you prefer to do it step-by-step):**
```bash
# Step 1: First time - record v1 behavior
./record-tests.sh

# Step 2: Development loop - fix one endpoint at a time
./play-tests.sh
# ... make code changes ...
./remove.sh
./play-tests.sh
```

**⚠️ IMPORTANT:** AI should use the automatic workflow with scripts. Do not go off on your own and implement manually - the scripts handle all the complexity.

---

## 📋 What You'll Need

- Your services installed and ready to go
- A terminal window and basic command-line comfort
- About 30 minutes for your first run-through
- Patience for the "one endpoint at a time" approach (trust us on this)

---

## 🚀 Let's Walk Through This Together

### Step 1: Getting Everything Running

First things first – we need to get all the pieces talking to each other. Think of this as setting up the stage before the play begins.

**For your very first time**, we'll record what "normal" looks like:

```bash
cd v1-v2-check-and-fix
./record-tests.sh
```

**What this does:**
- 🧹 Always cleans up old runtime files (tmp/, reports/, recordings/) for a fresh start
- 📹 Starts in RECORD mode and captures fresh v1 behavior
- 🧪 Runs tests automatically to generate baseline
- 🎯 Ensures you start with a clean slate every time

**Alternatively, you can do it manually:**
```bash
# Start in record mode (captures v1 behavior)
PROXY_MODE=record ./play-tests.sh

# Tests run automatically, wait for completion
```

You'll see some logging output as services start up. This is good – it means things are happening. The system is now:
- Watching your v1 API like a careful observer
- Ready to capture every detail of how it responds
- Setting up your v2 API for comparison
- Cleaning up old test results (but keeping any recordings you might reuse)

**Quick check:** Make sure everything's awake and listening:
```bash
lsof -i:3002 -i:3004 -i:3006 -i:8099
```

If you see processes listed, you're golden. If not, check the log files in the `tmp/` folder – they'll tell you what went wrong.

---

### Step 2: Capturing the "Ground Truth"

Now we let your v1 API show us how it's supposed to work. This is like taking a photograph of the current behavior – we want to capture every detail.

**If you used the initialize script, this step is already done!** Skip ahead to Step 3.

**If you're doing it manually:**

```bash
./run-reporter.sh config.comprehensive.json
```

**What you'll see:** The system will start calling your v1 endpoints, and the proxy will quietly record everything – the requests, the responses, the timing, even the headers. It's building a complete picture of "this is how things work today."

**Don't panic if you see failures** – those are actually good news! They're the roadmap of what needs fixing in v2.

---

### Step 3: Reading the Tea Leaves

After the recording finishes, you'll get a report that tells the story. Let's look at what it's saying:

```bash
cat reports/report_*.md
```

**Green checkmarks (✅) mean:** "All good! v2 matches v1 perfectly."  
**Red X's (❌) mean:** "Houston, we have a problem. These don't match."

**The report will show you things like:**
- Which specific endpoints are different
- Exactly how the responses differ
- Whether it's a small detail or a completely missing endpoint

**Real example from the field:**
```
GET /api/user/1/summary
Status: ❌ Mismatch
V1 says: { "userId": 1, "summary": "Active user" }
V2 says: { "error": "Not found" }
```

This tells us v2 is missing the user summary endpoint entirely. Clear, actionable information.

---

### Step 4: Picking Your First Fight

Here's where discipline pays off. You'll be tempted to fix everything at once. Resist that urge.

**Pick ONE failure – just one – and fix that.**

Why? Because:
- Small changes are easier to get right
- If something breaks, you know exactly what caused it
- Your future self (and your teammates) will thank you for clear, focused commits
- You'll build momentum with quick wins

**Start with something simple** – maybe an endpoint that's just returning the wrong format, rather than one that's completely missing. Build confidence first.

**🚨 MANDATORY: Read [FIX-PROCESS.md](FIX-PROCESS.md) NOW before fixing any endpoints! 🚨**

This document contains the exact step-by-step process you MUST follow for each fix:
- Creating feature branches
- Writing tests FIRST (TDD)
- Building and verification steps
- Communication requirements

---

### Step 5: Writing the "Before" Test

Before you change anything, write a test that proves the problem exists. This might seem backwards, but it's actually brilliant:

1. **It forces you to understand exactly what's broken**
2. **It gives you a clear definition of "fixed"**
3. **It prevents the problem from coming back later**

```bash
cd ../prroxy/rest-v2/tests/integration
```

Create a test that expects the correct behavior. When you run it, it should fail – that's how you know you're testing the right thing.

**The test that fails today is the same test that will pass tomorrow** when you've fixed the issue.

---

### Step 6: Making It Work

Now – and only now – do you actually fix the code. You know exactly what's wrong, you have a test that proves it, and you have a clear target to hit.

**This is usually the easiest part** because all the investigation work is done. You're just implementing what you now understand needs to exist.

**The key insight:** You're not guessing what the endpoint should do – you have the recorded v1 behavior as your specification.

---

### Step 7: Proving It Works

Run your test again. This time it should pass.

```bash
cd prroxy/rest-v2
go test ./tests/integration/... -v -run YourTestName
```

**Green is good!** But don't stop there...

---

### Step 8: The Full System Check

Now we make sure your fix works in the complete environment and doesn't break anything else:

```bash
cd v1-v2-check-and-fix
./remove.sh
./play-tests.sh
```

**What to look for:**
- Your previously failing endpoint should now show a happy green checkmark
- The total failure count should go down by one
- No new failures should appear (you didn't break anything else)

---

### Step 9: Sharing Your Victory

Time to make it official:

```bash
git add .
git commit -m "Implement GET /api/user/:id/summary endpoint

- Add user summary handler with proper response format
- Add comprehensive tests
- Matches v1 behavior exactly"

git push origin your-branch-name
```

**Pro tip:** Write commit messages that explain WHY you made the change, not just WHAT you changed. Your future self will thank you.

---

### Step 10: Rinse and Repeat

One down, however many left to go. But here's the beautiful thing: **you now have a proven process that works.**

Look at your latest report, pick the next failure, and repeat the cycle. Each one gets easier because you're building expertise and momentum.

---

## 🎯 Making This Your Own

### The Development Modes

**Record Mode** (for capturing behavior):
```bash
./record-tests.sh
# Or manually: PROXY_MODE=record ./play-tests.sh
```
- Makes real API calls to capture actual behavior
- Slower but necessary for accuracy
- Use this when you need fresh data

**Playback Mode** (for development):
```bash
./play-tests.sh
```
- Uses cached responses – much faster
- Perfect for iterative development
- No external dependencies
- Tests run automatically

### Starting Fresh

The `record-tests.sh` script **always** cleans up and starts fresh every time:

```bash
./record-tests.sh
```

This ensures you always have clean recordings captured from v1. There's no need for special flags - it's the default behavior.

---

## 🛠️ The Essential Commands

**The ones you'll use constantly:**
```bash
# Record v1 baseline (first time or fresh start)
./record-tests.sh

# Test v2 against v1 baseline (most common - daily development)
./play-tests.sh

# Run comparison tests manually
./run-reporter.sh config.comprehensive.json

# Stop everything
./remove.sh

# Check what's failing
cat reports/report_*.md | grep "❌"

# Run specific test in v2
cd ../prroxy/rest-v2 && go test ./tests/integration/... -v -run TestName
```

**Everything else is just decoration.**

---

## 🚨 When Things Don't Go According to Plan

**Services being cranky?**
```bash
./remove.sh  # Kill everything and start over
```

**Not sure if things are running?**
```bash
lsof -i:3002 -i:3004 -i:3006 -i:8099  # See what's listening
```

**Tests failing mysteriously?**
```bash
cat tmp/rest-v2.log | grep -i error  # Check what v2 is complaining about
```

**Recordings not happening?**
```bash
echo $PROXY_MODE  # Should be "record"
cat tmp/proxy.log | head -10  # Check proxy mood
```

---

## 🎓 The Bigger Picture

**What you've built:** A systematic way to ensure API compatibility that you can reuse whenever you make significant changes. 

**The real magic:** Once you've gone through this process once, you have:
- Confidence that your v2 really does match your v1
- A comprehensive test suite that prevents regression
- A repeatable process for future API evolution
- Documentation of what each endpoint actually does (your recordings)

**The mindset shift:** Instead of hoping your APIs match, you now have proof they do.

---

## 🎉 Your Next Steps

1. **Pick your first failure** – start with something manageable
2. **Write that failing test** – resist the urge to skip this step
3. **Implement your fix** – you know exactly what needs to be done
4. **Verify everything still works** – run the full comparison
5. **Commit and share** – let the world see your beautiful, compatible code

**Remember:** Every expert was once a beginner who didn't quit. You've got this!

---

*Happy coding, and may all your endpoints match perfectly!* 🚀