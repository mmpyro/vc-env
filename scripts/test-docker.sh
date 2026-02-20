#!/bin/bash
# Docker-based integration test script for vc-env
# This script can be run either:
# 1. Inside a Docker container (when VCENV_ROOT is already set)
# 2. On the host to build and run the Docker container

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    PASS=$((PASS + 1))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    FAIL=$((FAIL + 1))
}

info() {
    echo -e "${YELLOW}→${NC} $1"
}

# If we're not inside the container, build and run it
if [ -z "${VCENV_ROOT:-}" ]; then
    echo "Building vc-env for Linux..."
    
    # Detect host architecture for correct binary
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) GOARCH="amd64" ;;
        aarch64|arm64) GOARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    GOOS=linux GOARCH=$GOARCH go build -ldflags "-X main.Version=0.1.0" \
        -o build/vc-env-linux-$GOARCH ./cmd/vc-env
    
    echo "Building Docker image..."
    # Copy the correct binary for the Dockerfile
    cp build/vc-env-linux-$GOARCH build/vc-env-linux-test
    docker build -f Dockerfile.test -t vc-env-test .
    rm -f build/vc-env-linux-test
    
    echo "Running integration tests in Docker..."
    docker run --rm vc-env-test
    exit $?
fi

# ============================================================
# Integration tests (run inside Docker container)
# ============================================================

echo "========================================="
echo "  vc-env Integration Tests"
echo "========================================="
echo ""

# Test 1: vc-env version
info "Test: vc-env version"
OUTPUT=$(vc-env version 2>&1)
if echo "$OUTPUT" | grep -q "0.1.0"; then
    pass "vc-env version prints 0.1.0"
else
    fail "vc-env version: expected '0.1.0', got '$OUTPUT'"
fi

# Test 2: vc-env help
info "Test: vc-env help"
OUTPUT=$(vc-env help 2>&1)
if echo "$OUTPUT" | grep -q "Commands:"; then
    pass "vc-env help shows commands"
else
    fail "vc-env help: expected 'Commands:', got '$OUTPUT'"
fi

# Test 3: vc-env init
info "Test: vc-env init"
OUTPUT=$(vc-env init 2>&1)
if [ -d "$VCENV_ROOT/versions" ]; then
    pass "vc-env init creates versions directory"
else
    fail "vc-env init: versions directory not created"
fi

if [ -f "$VCENV_ROOT/shims/vcluster" ]; then
    pass "vc-env init creates vcluster shim"
else
    fail "vc-env init: vcluster shim not created"
fi

# Test 4: vc-env list (empty)
info "Test: vc-env list (empty)"
OUTPUT=$(vc-env list 2>&1)
if [ -z "$OUTPUT" ]; then
    pass "vc-env list shows nothing when empty"
else
    fail "vc-env list: expected empty output, got '$OUTPUT'"
fi

# Test 5: vc-env list-remote
info "Test: vc-env list-remote"
OUTPUT=$(vc-env list-remote 2>&1)
if echo "$OUTPUT" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
    pass "vc-env list-remote shows versions"
else
    fail "vc-env list-remote: no versions found in output"
fi

# Get a version to install (use a known stable version)
TEST_VERSION="0.21.1"

# Test 6: vc-env install
info "Test: vc-env install $TEST_VERSION"
OUTPUT=$(vc-env install "$TEST_VERSION" 2>&1)
if [ -f "$VCENV_ROOT/versions/$TEST_VERSION/vcluster" ]; then
    pass "vc-env install downloads and stores binary"
else
    fail "vc-env install: binary not found at $VCENV_ROOT/versions/$TEST_VERSION/vcluster"
    echo "Output: $OUTPUT"
fi

# Test 7: vc-env install (already installed)
info "Test: vc-env install $TEST_VERSION (already installed)"
OUTPUT=$(vc-env install "$TEST_VERSION" 2>&1)
if echo "$OUTPUT" | grep -q "already installed"; then
    pass "vc-env install skips already installed version"
else
    fail "vc-env install: expected 'already installed', got '$OUTPUT'"
fi

# Test 8: vc-env list (with installed version)
info "Test: vc-env list"
OUTPUT=$(vc-env list 2>&1)
if echo "$OUTPUT" | grep -q "$TEST_VERSION"; then
    pass "vc-env list shows installed version"
else
    fail "vc-env list: expected '$TEST_VERSION', got '$OUTPUT'"
fi

# Test 9: vc-env global
info "Test: vc-env global $TEST_VERSION"
vc-env global "$TEST_VERSION" 2>&1
OUTPUT=$(vc-env global 2>&1)
if echo "$OUTPUT" | grep -q "$TEST_VERSION"; then
    pass "vc-env global sets and reads version"
else
    fail "vc-env global: expected '$TEST_VERSION', got '$OUTPUT'"
fi

# Test 10: vc-env which (with global version)
info "Test: vc-env which"
OUTPUT=$(vc-env which 2>&1)
if echo "$OUTPUT" | grep -q "$VCENV_ROOT/versions/$TEST_VERSION/vcluster"; then
    pass "vc-env which shows correct path"
else
    fail "vc-env which: expected path containing '$TEST_VERSION', got '$OUTPUT'"
fi

# Test 11: vc-env local
info "Test: vc-env local $TEST_VERSION"
cd /tmp
vc-env local "$TEST_VERSION" 2>&1
OUTPUT=$(vc-env local 2>&1)
if echo "$OUTPUT" | grep -q "$TEST_VERSION"; then
    pass "vc-env local sets and reads version"
else
    fail "vc-env local: expected '$TEST_VERSION', got '$OUTPUT'"
fi

# Verify .vcluster-version file was created
if [ -f "/tmp/.vcluster-version" ]; then
    pass "vc-env local creates .vcluster-version file"
else
    fail "vc-env local: .vcluster-version file not created"
fi

# Test 12: vc-env shell
info "Test: vc-env shell $TEST_VERSION"
OUTPUT=$(vc-env shell "$TEST_VERSION" 2>&1)
if echo "$OUTPUT" | grep -q "export VCENV_VERSION=$TEST_VERSION"; then
    pass "vc-env shell outputs export command"
else
    fail "vc-env shell: expected export command, got '$OUTPUT'"
fi

# Test 13: vc-env shell (no version set)
info "Test: vc-env shell (no version set)"
OUTPUT=$(vc-env shell 2>&1) || true
if echo "$OUTPUT" | grep -q "no shell version"; then
    pass "vc-env shell reports no version set"
else
    fail "vc-env shell: expected 'no shell version', got '$OUTPUT'"
fi

# Test 14: Verify shim works
info "Test: vcluster shim"
export PATH="$VCENV_ROOT/shims:$PATH"
export VCENV_VERSION="$TEST_VERSION"
if command -v vcluster >/dev/null 2>&1; then
    OUTPUT=$(vcluster version 2>&1) || true
    pass "vcluster shim is executable and in PATH"
else
    fail "vcluster shim not found in PATH"
fi
unset VCENV_VERSION

# Test 15: vc-env uninstall
info "Test: vc-env uninstall $TEST_VERSION"
OUTPUT=$(vc-env uninstall "$TEST_VERSION" 2>&1)
if echo "$OUTPUT" | grep -q "uninstalled"; then
    pass "vc-env uninstall removes version"
else
    fail "vc-env uninstall: expected 'uninstalled', got '$OUTPUT'"
fi

# Verify directory was removed
if [ ! -d "$VCENV_ROOT/versions/$TEST_VERSION" ]; then
    pass "vc-env uninstall removes version directory"
else
    fail "vc-env uninstall: version directory still exists"
fi

# Test 16: vc-env uninstall (not installed)
info "Test: vc-env uninstall $TEST_VERSION (not installed)"
OUTPUT=$(vc-env uninstall "$TEST_VERSION" 2>&1) || true
if echo "$OUTPUT" | grep -q "not installed"; then
    pass "vc-env uninstall fails for non-installed version"
else
    fail "vc-env uninstall: expected 'not installed', got '$OUTPUT'"
fi

# Test 17: Error cases
info "Test: vc-env global (no version configured)"
# Remove global version file
rm -f "$VCENV_ROOT/version"
OUTPUT=$(vc-env global 2>&1) || true
if echo "$OUTPUT" | grep -q "no global version"; then
    pass "vc-env global reports no version configured"
else
    fail "vc-env global: expected 'no global version', got '$OUTPUT'"
fi

# Summary
echo ""
echo "========================================="
echo "  Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "========================================="

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
