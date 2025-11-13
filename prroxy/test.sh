#!/bin/bash

# Test script for Prroxy project
# Runs all tests for: proxy, reporter, rest-v1, rest-v2

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_coverage() {
    echo -e "${CYAN}📊${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
}

# Get script directory (project root)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

print_header "Prroxy Test Suite"
print_info "Project root: ${PROJECT_ROOT}"

# Track test status
TESTS_PASSED=0
TESTS_FAILED=0

# Option flags
COVERAGE=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -c, --coverage    Run tests with coverage reports"
            echo "  -v, --verbose     Run tests with verbose output"
            echo "  -h, --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                # Run all tests"
            echo "  $0 --coverage     # Run all tests with coverage"
            echo "  $0 -v             # Run all tests with verbose output"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# ============================================================
# Test Proxy (Go)
# ============================================================
print_header "Testing Proxy"
if [ -d "${PROJECT_ROOT}/proxy" ]; then
    cd "${PROJECT_ROOT}/proxy"
    print_info "Running proxy tests..."

    if [ "$COVERAGE" = true ]; then
        if [ "$VERBOSE" = true ]; then
            if go test -v -cover ./...; then
                print_success "Proxy tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Proxy tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            if go test -cover ./... 2>&1; then
                print_success "Proxy tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Proxy tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
    else
        if [ "$VERBOSE" = true ]; then
            if go test -v ./...; then
                print_success "Proxy tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Proxy tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            if go test ./...; then
                print_success "Proxy tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Proxy tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
    fi
else
    print_warning "Proxy directory not found, skipping"
fi

# ============================================================
# Test Reporter (Go)
# ============================================================
print_header "Testing Reporter"
if [ -d "${PROJECT_ROOT}/reporter" ]; then
    cd "${PROJECT_ROOT}/reporter"
    print_info "Running reporter tests..."

    if [ "$COVERAGE" = true ]; then
        if [ "$VERBOSE" = true ]; then
            if go test -v -cover ./...; then
                print_success "Reporter tests passed"
                echo ""
                print_coverage "Coverage breakdown:"
                go test -cover ./... 2>&1 | grep -E "coverage:|ok" | grep -v "no test files"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Reporter tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            TEST_OUTPUT=$(go test -cover ./... 2>&1)
            if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
                print_error "Reporter tests failed"
                echo "$TEST_OUTPUT"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            else
                print_success "Reporter tests passed"
                echo ""
                print_coverage "Coverage breakdown:"
                echo "$TEST_OUTPUT" | grep -E "coverage:|ok" | grep -v "no test files"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            fi
        fi
    else
        if [ "$VERBOSE" = true ]; then
            if go test -v ./...; then
                print_success "Reporter tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Reporter tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            if go test ./...; then
                print_success "Reporter tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "Reporter tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
    fi
else
    print_warning "Reporter directory not found, skipping"
fi

# ============================================================
# Test REST v1 (Node.js/TypeScript)
# ============================================================
print_header "Testing REST v1"
if [ -d "${PROJECT_ROOT}/rest-v1" ]; then
    cd "${PROJECT_ROOT}/rest-v1"

    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        if npm install > /dev/null 2>&1; then
            print_success "Dependencies installed"
        else
            print_error "npm install failed"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            cd "${PROJECT_ROOT}"
            continue
        fi
    fi

    print_info "Running REST v1 tests..."

    if [ "$COVERAGE" = true ]; then
        if npm run test:coverage; then
            print_success "REST v1 tests passed"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            print_error "REST v1 tests failed"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        if [ "$VERBOSE" = true ]; then
            if npm test -- --verbose; then
                print_success "REST v1 tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "REST v1 tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            if npm test; then
                print_success "REST v1 tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "REST v1 tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
    fi
else
    print_warning "REST v1 directory not found, skipping"
fi

# ============================================================
# Test REST v2 (Go)
# ============================================================
print_header "Testing REST v2"
if [ -d "${PROJECT_ROOT}/rest-v2" ]; then
    cd "${PROJECT_ROOT}/rest-v2"
    print_info "Running REST v2 tests..."

    if [ "$COVERAGE" = true ]; then
        if [ "$VERBOSE" = true ]; then
            if go test -v -cover ./...; then
                print_success "REST v2 tests passed"
                echo ""
                print_coverage "Coverage breakdown:"
                go test -cover ./... 2>&1 | grep -E "coverage:|ok" | grep -v "no test files"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "REST v2 tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            TEST_OUTPUT=$(go test -cover ./... 2>&1)
            if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
                print_error "REST v2 tests failed"
                echo "$TEST_OUTPUT"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            else
                print_success "REST v2 tests passed"
                echo ""
                print_coverage "Coverage breakdown:"
                echo "$TEST_OUTPUT" | grep -E "coverage:|ok" | grep -v "no test files"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            fi
        fi
    else
        if [ "$VERBOSE" = true ]; then
            if go test -v ./...; then
                print_success "REST v2 tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "REST v2 tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            if go test ./...; then
                print_success "REST v2 tests passed"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                print_error "REST v2 tests failed"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
    fi
else
    print_warning "REST v2 directory not found, skipping"
fi

# ============================================================
# Test Summary
# ============================================================
cd "${PROJECT_ROOT}"

print_header "Test Summary"
echo ""

if [ ${TESTS_PASSED} -gt 0 ]; then
    print_success "Test suites passed: ${TESTS_PASSED}"
fi

if [ ${TESTS_FAILED} -gt 0 ]; then
    print_error "Test suites failed: ${TESTS_FAILED}"
fi

echo ""

if [ "$COVERAGE" = true ]; then
    print_header "Overall Coverage Summary"
    echo ""

    print_coverage "Proxy:"
    cd "${PROJECT_ROOT}/proxy" && go test -cover ./... 2>&1 | grep coverage | head -5 || true
    echo ""

    print_coverage "Reporter:"
    cd "${PROJECT_ROOT}/reporter" && go test -cover ./... 2>&1 | grep coverage | head -5 || true
    echo ""

    if [ -d "${PROJECT_ROOT}/rest-v1" ]; then
        print_coverage "REST v1: See above for detailed coverage report"
        echo ""
    fi

    print_coverage "REST v2:"
    cd "${PROJECT_ROOT}/rest-v2" && go test -cover ./... 2>&1 | grep coverage | head -5 || true
    echo ""
fi

cd "${PROJECT_ROOT}"

print_header "Commands"
echo ""
echo "Run individual test suites:"
echo "  Proxy:      cd proxy && go test ./..."
echo "  Reporter:   cd reporter && go test ./..."
echo "  REST v1:    cd rest-v1 && npm test"
echo "  REST v2:    cd rest-v2 && go test ./..."
echo ""
echo "Run with coverage:"
echo "  ./test.sh --coverage"
echo ""
echo "Run with verbose output:"
echo "  ./test.sh --verbose"
echo ""

# Exit with error if any tests failed
if [ ${TESTS_FAILED} -gt 0 ]; then
    print_error "Test suite completed with failures"
    exit 1
else
    print_success "All test suites passed!"
    exit 0
fi
