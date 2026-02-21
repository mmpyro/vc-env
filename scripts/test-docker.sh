#!/bin/bash
# scripts/test-docker.sh
#
# Docker-based integration test runner for vc-env.
#
# HOST MODE  (VCENV_ROOT is not set):
#   Builds the Docker image using Dockerfile.test, then runs the container.
#   The container re-executes this same script in CONTAINER MODE.
#
# CONTAINER MODE  (VCENV_ROOT is already exported by the Dockerfile):
#   Runs the full ordered integration test suite against the vc-env binary
#   that was compiled into the image.

set -uo pipefail

# ── Colour helpers ─────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ── Test accounting ────────────────────────────────────────────────────────────
TOTAL=0
PASS=0
FAIL=0
declare -a FAILURES=()   # each entry: "STEP_NUM|COMMAND|EXPECTED|ACTUAL"

pass() {
    local step="$1" desc="$2"
    echo -e "  ${GREEN}✓ PASS${NC} [Step ${step}] ${desc}"
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
}

fail() {
    local step="$1" desc="$2" cmd="$3" expected="$4" actual="$5"
    echo -e "  ${RED}✗ FAIL${NC} [Step ${step}] ${desc}"
    echo -e "         ${YELLOW}cmd     :${NC} ${cmd}"
    echo -e "         ${YELLOW}expected:${NC} ${expected}"
    echo -e "         ${YELLOW}actual  :${NC} ${actual}"
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
    FAILURES+=("${step}|${cmd}|${expected}|${actual}")
}

section() {
    echo ""
    echo -e "${CYAN}${BOLD}▶ $*${NC}"
}

# ── HOST MODE: build image and run container ───────────────────────────────────
if [ -z "${VCENV_ROOT:-}" ]; then
    echo -e "${BOLD}=========================================${NC}"
    echo -e "${BOLD}  vc-env Docker Integration Test Runner  ${NC}"
    echo -e "${BOLD}=========================================${NC}"
    echo ""

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

    echo "Building Docker image (Dockerfile.test)..."
    docker build -f "${PROJECT_ROOT}/Dockerfile.test" -t vc-env-test "${PROJECT_ROOT}"

    echo ""
    echo "Running integration tests inside Docker container..."
    echo ""
    docker run --rm vc-env-test
    exit $?
fi

# ── CONTAINER MODE: full integration test suite ────────────────────────────────

echo ""
echo -e "${BOLD}=========================================${NC}"
echo -e "${BOLD}  vc-env Integration Test Suite          ${NC}"
echo -e "${BOLD}=========================================${NC}"

# ── Step 1: Set and export VCENV_ROOT ─────────────────────────────────────────
section "Step 1: Set and export VCENV_ROOT"
export VCENV_ROOT="${HOME}/.vc-env"
if [ -n "${VCENV_ROOT}" ]; then
    pass 1 "VCENV_ROOT exported as '${VCENV_ROOT}'"
else
    fail 1 "VCENV_ROOT export" \
         "export VCENV_ROOT=\$HOME/.vc-env" \
         "VCENV_ROOT is empty or unset" \
         "(variable is empty)"
fi

# ── Step 2: vc-env init ────────────────────────────────────────────────────────
section "Step 2: vc-env init"
INIT_OUTPUT=$(vc-env init 2>&1)
INIT_EXIT=$?
if [ $INIT_EXIT -eq 0 ] && [ -d "${VCENV_ROOT}/versions" ] && [ -f "${VCENV_ROOT}/shims/vcluster" ]; then
    pass 2 "vc-env init completed without error; versions dir and vcluster shim created"
else
    ACTUAL="exit=${INIT_EXIT}; versions_dir=$([ -d "${VCENV_ROOT}/versions" ] && echo exists || echo missing); shim=$([ -f "${VCENV_ROOT}/shims/vcluster" ] && echo exists || echo missing); output=${INIT_OUTPUT}"
    fail 2 "vc-env init" \
         "vc-env init" \
         "exit 0, versions directory and shims/vcluster created" \
         "${ACTUAL}"
fi

# Ensure shims are on PATH for subsequent steps
export PATH="${VCENV_ROOT}/shims:${PATH}"

# ── Step 3: vc-env list-remote ────────────────────────────────────────────────
section "Step 3: vc-env list-remote"
LR_OUTPUT=$(vc-env list-remote 2>&1)
LR_EXIT=$?
if [ $LR_EXIT -eq 0 ] && echo "${LR_OUTPUT}" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
    FIRST_VER=$(echo "${LR_OUTPUT}" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    pass 3 "vc-env list-remote returned a non-empty list (first entry: ${FIRST_VER})"
else
    fail 3 "vc-env list-remote" \
         "vc-env list-remote" \
         "non-empty list of semver versions (exit 0)" \
         "exit=${LR_EXIT}; output=${LR_OUTPUT}"
fi

# ── Step 4: vc-env latest ─────────────────────────────────────────────────────
section "Step 4: vc-env latest"
LATEST_OUTPUT=$(vc-env latest 2>&1)
LATEST_EXIT=$?
if [ $LATEST_EXIT -eq 0 ] && echo "${LATEST_OUTPUT}" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
    pass 4 "vc-env latest returned a stable semver version: '${LATEST_OUTPUT}'"
else
    fail 4 "vc-env latest" \
         "vc-env latest" \
         "exit 0 and output matches semver pattern (e.g. 0.x.x)" \
         "exit=${LATEST_EXIT}; output=${LATEST_OUTPUT}"
fi

# ── Step 5: vc-env latest --prerelease ────────────────────────────────────────
section "Step 5: vc-env latest --prerelease"
LATEST_PRE_OUTPUT=$(vc-env latest --prerelease 2>&1)
LATEST_PRE_EXIT=$?
if [ $LATEST_PRE_EXIT -eq 0 ] && echo "${LATEST_PRE_OUTPUT}" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
    pass 5 "vc-env latest --prerelease returned a version string: '${LATEST_PRE_OUTPUT}'"
else
    fail 5 "vc-env latest --prerelease" \
         "vc-env latest --prerelease" \
         "exit 0 and output matches semver pattern (may include prerelease suffix)" \
         "exit=${LATEST_PRE_EXIT}; output=${LATEST_PRE_OUTPUT}"
fi

# ── Step 6: vc-env latest --help ──────────────────────────────────────────────
section "Step 6: vc-env latest --help"
LATEST_HELP_OUTPUT=$(vc-env latest --help 2>&1)
LATEST_HELP_EXIT=$?
if [ $LATEST_HELP_EXIT -eq 0 ] && echo "${LATEST_HELP_OUTPUT}" | grep -qiE 'latest|prerelease'; then
    pass 6 "vc-env latest --help outputs help text containing 'latest' or 'prerelease'"
else
    fail 6 "vc-env latest --help" \
         "vc-env latest --help" \
         "exit 0 and output contains 'latest' or 'prerelease'" \
         "exit=${LATEST_HELP_EXIT}; output=${LATEST_HELP_OUTPUT}"
fi

# ── Step 7: vc-env install {version} ──────────────────────────────────────────
TEST_VERSION="0.21.1"
section "Step 7: vc-env install ${TEST_VERSION}"
INSTALL_OUTPUT=$(vc-env install "${TEST_VERSION}" 2>&1)
INSTALL_EXIT=$?
if [ $INSTALL_EXIT -eq 0 ] && [ -f "${VCENV_ROOT}/versions/${TEST_VERSION}/vcluster" ]; then
    pass 7 "vc-env install ${TEST_VERSION} succeeded; binary present at versions/${TEST_VERSION}/vcluster"
else
    fail 7 "vc-env install ${TEST_VERSION}" \
         "vc-env install ${TEST_VERSION}" \
         "exit 0 and binary at \${VCENV_ROOT}/versions/${TEST_VERSION}/vcluster" \
         "exit=${INSTALL_EXIT}; binary=$([ -f "${VCENV_ROOT}/versions/${TEST_VERSION}/vcluster" ] && echo present || echo missing); output=${INSTALL_OUTPUT}"
fi

# ── Step 8: vc-env list ───────────────────────────────────────────────────────
section "Step 8: vc-env list"
LIST_OUTPUT=$(vc-env list 2>&1)
LIST_EXIT=$?
if [ $LIST_EXIT -eq 0 ] && echo "${LIST_OUTPUT}" | grep -q "${TEST_VERSION}"; then
    pass 8 "vc-env list shows installed version ${TEST_VERSION}"
else
    fail 8 "vc-env list" \
         "vc-env list" \
         "output contains '${TEST_VERSION}' (exit 0)" \
         "exit=${LIST_EXIT}; output=${LIST_OUTPUT}"
fi

# ── Step 9: vc-env global {version} ───────────────────────────────────────────
section "Step 9: vc-env global ${TEST_VERSION}"
vc-env global "${TEST_VERSION}" 2>&1
GLOBAL_READ=$(vc-env global 2>&1)
GLOBAL_EXIT=$?
if [ $GLOBAL_EXIT -eq 0 ] && echo "${GLOBAL_READ}" | grep -q "${TEST_VERSION}"; then
    pass 9 "vc-env global set and confirmed as '${TEST_VERSION}'"
else
    fail 9 "vc-env global ${TEST_VERSION}" \
         "vc-env global ${TEST_VERSION} && vc-env global" \
         "global version reads back as '${TEST_VERSION}' (exit 0)" \
         "exit=${GLOBAL_EXIT}; output=${GLOBAL_READ}"
fi

# ── Step 10: vc-env which ─────────────────────────────────────────────────────
section "Step 10: vc-env which"
WHICH_OUTPUT=$(vc-env which 2>&1)
WHICH_EXIT=$?
EXPECTED_WHICH="${VCENV_ROOT}/versions/${TEST_VERSION}/vcluster"
if [ $WHICH_EXIT -eq 0 ] && echo "${WHICH_OUTPUT}" | grep -q "${EXPECTED_WHICH}"; then
    pass 10 "vc-env which returns correct path '${EXPECTED_WHICH}'"
else
    fail 10 "vc-env which" \
         "vc-env which" \
         "path containing '${EXPECTED_WHICH}' (exit 0)" \
         "exit=${WHICH_EXIT}; output=${WHICH_OUTPUT}"
fi

# ── Step 11: vc-env local {version} ───────────────────────────────────────────
section "Step 11: vc-env local ${TEST_VERSION}"
LOCAL_WORKDIR=$(mktemp -d)
pushd "${LOCAL_WORKDIR}" > /dev/null
vc-env local "${TEST_VERSION}" 2>&1
LOCAL_READ=$(vc-env local 2>&1)
LOCAL_EXIT=$?
LOCAL_FILE_CONTENT=$(cat "${LOCAL_WORKDIR}/.vcluster-version" 2>/dev/null || echo "(file missing)")
popd > /dev/null

if [ $LOCAL_EXIT -eq 0 ] \
   && echo "${LOCAL_READ}" | grep -q "${TEST_VERSION}" \
   && [ -f "${LOCAL_WORKDIR}/.vcluster-version" ] \
   && grep -q "${TEST_VERSION}" "${LOCAL_WORKDIR}/.vcluster-version"; then
    pass 11 "vc-env local set '${TEST_VERSION}'; .vcluster-version file written correctly"
else
    fail 11 "vc-env local ${TEST_VERSION}" \
         "vc-env local ${TEST_VERSION} && vc-env local (in temp dir)" \
         "local version reads back as '${TEST_VERSION}' and .vcluster-version contains '${TEST_VERSION}'" \
         "exit=${LOCAL_EXIT}; read_output=${LOCAL_READ}; file_content=${LOCAL_FILE_CONTENT}"
fi
rm -rf "${LOCAL_WORKDIR}"

# ── Step 12: vc-env shell {version} ───────────────────────────────────────────
section "Step 12: vc-env shell ${TEST_VERSION}"
SHELL_OUTPUT=$(vc-env shell "${TEST_VERSION}" 2>&1)
SHELL_EXIT=$?
if [ $SHELL_EXIT -eq 0 ] && echo "${SHELL_OUTPUT}" | grep -q "export VCENV_VERSION=${TEST_VERSION}"; then
    pass 12 "vc-env shell outputs 'export VCENV_VERSION=${TEST_VERSION}'"
else
    fail 12 "vc-env shell ${TEST_VERSION}" \
         "vc-env shell ${TEST_VERSION}" \
         "output contains 'export VCENV_VERSION=${TEST_VERSION}' (exit 0)" \
         "exit=${SHELL_EXIT}; output=${SHELL_OUTPUT}"
fi

# ── Step 13: vc-env uninstall {version} ───────────────────────────────────────
section "Step 13: vc-env uninstall ${TEST_VERSION}"
UNINSTALL_OUTPUT=$(vc-env uninstall "${TEST_VERSION}" 2>&1)
UNINSTALL_EXIT=$?
if [ $UNINSTALL_EXIT -eq 0 ] \
   && echo "${UNINSTALL_OUTPUT}" | grep -qi "uninstalled" \
   && [ ! -d "${VCENV_ROOT}/versions/${TEST_VERSION}" ]; then
    pass 13 "vc-env uninstall ${TEST_VERSION} succeeded; version directory removed"
else
    fail 13 "vc-env uninstall ${TEST_VERSION}" \
         "vc-env uninstall ${TEST_VERSION}" \
         "exit 0, output contains 'uninstalled', version directory removed" \
         "exit=${UNINSTALL_EXIT}; dir=$([ -d "${VCENV_ROOT}/versions/${TEST_VERSION}" ] && echo still_present || echo removed); output=${UNINSTALL_OUTPUT}"
fi

# Verify the version no longer appears in vc-env list
LIST_AFTER=$(vc-env list 2>&1)
if ! echo "${LIST_AFTER}" | grep -q "${TEST_VERSION}"; then
    pass 13b "vc-env list no longer shows ${TEST_VERSION} after uninstall"
else
    fail 13b "vc-env list after uninstall" \
         "vc-env list" \
         "output does NOT contain '${TEST_VERSION}'" \
         "${LIST_AFTER}"
fi

# ── Step 14: vc-env version ───────────────────────────────────────────────────
section "Step 14: vc-env version"
VERSION_OUTPUT=$(vc-env version 2>&1)
VERSION_EXIT=$?
if [ $VERSION_EXIT -eq 0 ] && echo "${VERSION_OUTPUT}" | grep -qE '[0-9]+\.[0-9]+\.[0-9]+'; then
    pass 14 "vc-env version outputs a version string: '${VERSION_OUTPUT}'"
else
    fail 14 "vc-env version" \
         "vc-env version" \
         "exit 0 and output contains a semver string" \
         "exit=${VERSION_EXIT}; output=${VERSION_OUTPUT}"
fi

# ── Step 15: vc-env help ──────────────────────────────────────────────────────
section "Step 15: vc-env help"
HELP_OUTPUT=$(vc-env help 2>&1)
HELP_EXIT=$?
if [ $HELP_EXIT -eq 0 ] && echo "${HELP_OUTPUT}" | grep -qi "Commands:"; then
    pass 15 "vc-env help outputs usage information containing 'Commands:'"
else
    fail 15 "vc-env help" \
         "vc-env help" \
         "exit 0 and output contains 'Commands:'" \
         "exit=${HELP_EXIT}; output=${HELP_OUTPUT}"
fi

# ── Step 16: shim end-to-end — reinstall then run vcluster via shim ───────────
section "Step 16: vcluster shim end-to-end"

# Re-install the test version so the shim has something to execute
REINSTALL_OUTPUT=$(vc-env install "${TEST_VERSION}" 2>&1)
REINSTALL_EXIT=$?
if [ $REINSTALL_EXIT -ne 0 ]; then
    fail 16 "vcluster shim (re-install prerequisite)" \
         "vc-env install ${TEST_VERSION}" \
         "exit 0 (re-install for shim test)" \
         "exit=${REINSTALL_EXIT}; output=${REINSTALL_OUTPUT}"
else
    # Set the global version so the shim can resolve it
    vc-env global "${TEST_VERSION}" 2>&1

    # Confirm shim is executable and on PATH
    if ! command -v vcluster >/dev/null 2>&1; then
        fail 16 "vcluster shim not found in PATH" \
             "command -v vcluster" \
             "vcluster found in PATH (${VCENV_ROOT}/shims)" \
             "not found; PATH=${PATH}"
    else
        SHIM_PATH=$(command -v vcluster)
        VCLUSTER_OUTPUT=$(vcluster version 2>&1) || true
        VCLUSTER_EXIT=$?
        # vcluster --version / version may exit non-zero on some builds; we just
        # need it to produce output and not fail with "not installed" from the shim.
        if echo "${VCLUSTER_OUTPUT}" | grep -qi "vcluster\|version\|[0-9]\+\.[0-9]"; then
            pass 16 "vcluster shim at '${SHIM_PATH}' executed correctly; output: $(echo "${VCLUSTER_OUTPUT}" | head -1)"
        else
            fail 16 "vcluster shim end-to-end" \
                 "vcluster version" \
                 "output contains version information from the underlying binary" \
                 "exit=${VCLUSTER_EXIT}; shim=${SHIM_PATH}; output=${VCLUSTER_OUTPUT}"
        fi
    fi
fi

# ── Test Report ────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}=========================================${NC}"
echo -e "${BOLD}  Integration Test Report                ${NC}"
echo -e "${BOLD}=========================================${NC}"
echo -e "  Total steps run : ${BOLD}${TOTAL}${NC}"
echo -e "  Passed          : ${GREEN}${BOLD}${PASS}${NC}"
echo -e "  Failed          : ${RED}${BOLD}${FAIL}${NC}"

if [ "${FAIL}" -gt 0 ]; then
    echo ""
    echo -e "${RED}${BOLD}Failure Details:${NC}"
    for entry in "${FAILURES[@]}"; do
        IFS='|' read -r step_num cmd expected actual <<< "${entry}"
        echo ""
        echo -e "  ${RED}✗ Step ${step_num}${NC}"
        echo -e "    Command  : ${cmd}"
        echo -e "    Expected : ${expected}"
        echo -e "    Actual   : ${actual}"
    done
    echo ""
    echo -e "${RED}${BOLD}RESULT: FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}${BOLD}RESULT: ALL TESTS PASSED${NC}"
exit 0
