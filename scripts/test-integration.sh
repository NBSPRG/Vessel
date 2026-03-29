#!/usr/bin/env bash
# scripts/test-integration.sh
#
# Integration test suite for Vessel container runtime.
# Validates build correctness, binary size, cgroup controller availability,
# and kernel-level namespace syscall support on the current Linux host.
#
# Must be run on a Linux system (native or VM) with root privileges for
# cgroup and namespace checks.
#
# Usage:
#   sudo bash scripts/test-integration.sh
#
# Exit codes:
#   0  — all checks passed
#   1  — one or more checks failed

set -euo pipefail

BINARY="./vessel"
PASS=0
FAIL=0
RESULTS=()

# ── helpers ────────────────────────────────────────────────────────────────

ok() {
    local msg="$1"
    PASS=$((PASS + 1))
    RESULTS+=("  [PASS] $msg")
}

fail() {
    local msg="$1"
    FAIL=$((FAIL + 1))
    RESULTS+=("  [FAIL] $msg")
}

section() {
    echo ""
    echo "── $1 ──────────────────────────────────────────"
}

# ── 1. Build ───────────────────────────────────────────────────────────────

section "Build"

if CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$BINARY" . 2>&1; then
    ok "go build succeeded"
else
    fail "go build failed"
    echo "Build failed — cannot continue." >&2
    exit 1
fi

# ── 2. Binary size ─────────────────────────────────────────────────────────

section "Binary size"

BINARY_SIZE=$(stat -c%s "$BINARY" 2>/dev/null || stat -f%z "$BINARY")
BINARY_SIZE_MB=$(echo "scale=2; $BINARY_SIZE / 1048576" | bc)
echo "  Binary: $BINARY  ($BINARY_SIZE bytes / ${BINARY_SIZE_MB} MB)"

# Target: ≤ 10 MB without UPX (stripped Go binary is typically 4-7 MB)
MAX_BYTES=$((10 * 1024 * 1024))
if [ "$BINARY_SIZE" -le "$MAX_BYTES" ]; then
    ok "Binary size ${BINARY_SIZE_MB} MB is within 10 MB limit"
else
    fail "Binary size ${BINARY_SIZE_MB} MB exceeds 10 MB limit"
fi

# Report UPX availability and compressed size
if command -v upx >/dev/null 2>&1; then
    COMPRESSED=$(upx --best --lzma -q "$BINARY" 2>&1 | grep -oP '\d+\.\d+%' | tail -1 || true)
    COMPRESSED_SIZE=$(stat -c%s "$BINARY")
    COMPRESSED_MB=$(echo "scale=2; $COMPRESSED_SIZE / 1048576" | bc)
    ok "UPX compressed binary to ${COMPRESSED_MB} MB (${COMPRESSED})"
    if [ "$COMPRESSED_SIZE" -le $((1024 * 1024)) ]; then
        ok "UPX binary is under 1 MB"
    fi
else
    echo "  (UPX not available — skipping compression check)"
fi

# ── 3. Static analysis ─────────────────────────────────────────────────────

section "Static analysis (go vet)"

if go vet ./... 2>&1; then
    ok "go vet passed with no issues"
else
    fail "go vet reported issues"
fi

# ── 4. Unit tests ──────────────────────────────────────────────────────────

section "Unit tests (go test)"

if go test ./... -v -count=1 2>&1; then
    ok "All unit tests passed"
else
    fail "One or more unit tests failed"
fi

# ── 5. Cgroup controllers (Linux only) ────────────────────────────────────

section "cgroup v1 controllers"

CGROUP_BASE="/sys/fs/cgroup"
if [ -d "$CGROUP_BASE" ]; then
    for controller in memory cpu pids; do
        dir="$CGROUP_BASE/$controller"
        if [ -d "$dir" ]; then
            ok "cgroup controller present: $controller ($dir)"
        else
            fail "cgroup controller missing: $controller"
        fi
    done
else
    echo "  /sys/fs/cgroup not found — skipping cgroup checks (not on Linux?)"
fi

# ── 6. Kernel namespace syscall support ───────────────────────────────────

section "Kernel namespace support"

# Check CONFIG_NAMESPACES via /proc/config.gz or kernel config
KCONFIG=""
if [ -f /proc/config.gz ]; then
    KCONFIG=$(zcat /proc/config.gz 2>/dev/null || true)
elif [ -f /boot/config-"$(uname -r)" ]; then
    KCONFIG=$(cat /boot/config-"$(uname -r)" 2>/dev/null || true)
fi

if [ -n "$KCONFIG" ]; then
    for ns in CONFIG_UTS_NS CONFIG_IPC_NS CONFIG_PID_NS CONFIG_NET_NS CONFIG_USER_NS; do
        if echo "$KCONFIG" | grep -q "^${ns}=y"; then
            ok "Kernel namespace enabled: $ns"
        elif echo "$KCONFIG" | grep -q "^${ns}=m"; then
            ok "Kernel namespace as module: $ns"
        else
            fail "Kernel namespace not found: $ns"
        fi
    done
else
    echo "  Kernel config not accessible — checking /proc/self/ns instead"
    for ns in ipc mnt net pid uts; do
        if [ -e "/proc/self/ns/$ns" ]; then
            ok "Namespace file present: /proc/self/ns/$ns"
        else
            fail "Namespace file missing: /proc/self/ns/$ns"
        fi
    done
fi

# ── 7. /proc/self/exe (required for reexec) ───────────────────────────────

section "reexec prerequisite"

if [ -L /proc/self/exe ]; then
    ok "/proc/self/exe symlink present (reexec supported)"
else
    fail "/proc/self/exe not found — reexec will not work"
fi

# ── Summary ────────────────────────────────────────────────────────────────

echo ""
echo "════════════════════════════════════════════════"
echo " Integration Test Results"
echo "════════════════════════════════════════════════"
for r in "${RESULTS[@]}"; do
    echo "$r"
done
echo ""
echo "  Passed: $PASS  |  Failed: $FAIL"
echo "════════════════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
    echo "RESULT: FAIL"
    exit 1
else
    echo "RESULT: PASS"
    exit 0
fi
