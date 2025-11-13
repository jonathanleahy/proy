#!/bin/bash

# Build script for Prroxy project
# Builds all applications: proxy, reporter, rest-v1, rest-v2

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
}

# Get script directory (project root)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"

print_header "Prroxy Build Script"
print_info "Project root: ${PROJECT_ROOT}"
print_info "Build directory: ${BUILD_DIR}"

# Create build directory
mkdir -p "${BUILD_DIR}"

# Track build status
BUILDS_SUCCEEDED=0
BUILDS_FAILED=0

# ============================================================
# Build Proxy (Go)
# ============================================================
print_header "Building Proxy"
if [ -d "${PROJECT_ROOT}/proxy" ]; then
    cd "${PROJECT_ROOT}/proxy"
    print_info "Building proxy..."

    if go build -o "${BUILD_DIR}/proxy" cmd/proxy/main.go; then
        print_success "Proxy built successfully: ${BUILD_DIR}/proxy"
        BUILDS_SUCCEEDED=$((BUILDS_SUCCEEDED + 1))
    else
        print_error "Proxy build failed"
        BUILDS_FAILED=$((BUILDS_FAILED + 1))
    fi
else
    print_warning "Proxy directory not found, skipping"
fi

# ============================================================
# Build Reporter (Go)
# ============================================================
print_header "Building Reporter"
if [ -d "${PROJECT_ROOT}/reporter" ]; then
    cd "${PROJECT_ROOT}/reporter"
    print_info "Building reporter..."

    if go build -o "${BUILD_DIR}/reporter" cmd/reporter/main.go; then
        print_success "Reporter built successfully: ${BUILD_DIR}/reporter"
        BUILDS_SUCCEEDED=$((BUILDS_SUCCEEDED + 1))
    else
        print_error "Reporter build failed"
        BUILDS_FAILED=$((BUILDS_FAILED + 1))
    fi
else
    print_warning "Reporter directory not found, skipping"
fi

# ============================================================
# Build REST v1 (Node.js/TypeScript)
# ============================================================
print_header "Building REST v1"
if [ -d "${PROJECT_ROOT}/rest-v1" ]; then
    cd "${PROJECT_ROOT}/rest-v1"

    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        if npm install; then
            print_success "Dependencies installed"
        else
            print_error "npm install failed"
            BUILDS_FAILED=$((BUILDS_FAILED + 1))
            cd "${PROJECT_ROOT}"
            continue
        fi
    fi

    print_info "Building REST v1..."
    if npm run build; then
        print_success "REST v1 built successfully: ${PROJECT_ROOT}/rest-v1/dist"

        # Type-check test files
        print_info "Type-checking test files..."
        if npx tsc --project tsconfig.test.json; then
            print_success "Test files type-check passed"
            BUILDS_SUCCEEDED=$((BUILDS_SUCCEEDED + 1))
        else
            print_error "Test files have TypeScript errors"
            BUILDS_FAILED=$((BUILDS_FAILED + 1))
        fi
    else
        print_error "REST v1 build failed"
        BUILDS_FAILED=$((BUILDS_FAILED + 1))
    fi
else
    print_warning "REST v1 directory not found, skipping"
fi

# ============================================================
# Build REST v2 (Go)
# ============================================================
print_header "Building REST v2"
if [ -d "${PROJECT_ROOT}/rest-v2" ]; then
    cd "${PROJECT_ROOT}/rest-v2"
    print_info "Building REST v2..."

    if go build -o "${BUILD_DIR}/rest-v2" cmd/server/main.go; then
        print_success "REST v2 built successfully: ${BUILD_DIR}/rest-v2"
        BUILDS_SUCCEEDED=$((BUILDS_SUCCEEDED + 1))
    else
        print_error "REST v2 build failed"
        BUILDS_FAILED=$((BUILDS_FAILED + 1))
    fi
else
    print_warning "REST v2 directory not found, skipping"
fi

# ============================================================
# Build Summary
# ============================================================
cd "${PROJECT_ROOT}"

print_header "Build Summary"
echo ""
echo "Build directory: ${BUILD_DIR}"
echo ""

if [ -f "${BUILD_DIR}/proxy" ]; then
    print_success "proxy        $(stat -c%s "${BUILD_DIR}/proxy" 2>/dev/null | numfmt --to=iec-i --suffix=B || echo "$(stat -f%z "${BUILD_DIR}/proxy" 2>/dev/null) bytes")"
fi

if [ -f "${BUILD_DIR}/reporter" ]; then
    print_success "reporter     $(stat -c%s "${BUILD_DIR}/reporter" 2>/dev/null | numfmt --to=iec-i --suffix=B || echo "$(stat -f%z "${BUILD_DIR}/reporter" 2>/dev/null) bytes")"
fi

if [ -d "${PROJECT_ROOT}/rest-v1/dist" ]; then
    print_success "rest-v1      ${PROJECT_ROOT}/rest-v1/dist/"
fi

if [ -f "${BUILD_DIR}/rest-v2" ]; then
    print_success "rest-v2      $(stat -c%s "${BUILD_DIR}/rest-v2" 2>/dev/null | numfmt --to=iec-i --suffix=B || echo "$(stat -f%z "${BUILD_DIR}/rest-v2" 2>/dev/null) bytes")"
fi

echo ""
print_info "Total builds succeeded: ${BUILDS_SUCCEEDED}"
if [ ${BUILDS_FAILED} -gt 0 ]; then
    print_error "Total builds failed: ${BUILDS_FAILED}"
fi

echo ""
print_header "Usage"
echo ""
echo "Proxy:         ${BUILD_DIR}/proxy"
echo "Reporter:      ${BUILD_DIR}/reporter --config config.json"
echo "REST v1:       cd rest-v1 && npm start"
echo "REST v2:       ${BUILD_DIR}/rest-v2"
echo ""

# Exit with error if any builds failed
if [ ${BUILDS_FAILED} -gt 0 ]; then
    print_error "Build completed with errors"
    exit 1
else
    print_success "All builds completed successfully!"
    exit 0
fi
