#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $1" >&2; exit 1; }

REGISTRY="docs/gap-registry.md"
grep -q "^## Vocabulary" "$REGISTRY" || fail "Vocabulary section missing from registry"

# Check 8 registry rows present
for area in "identity" "graph" "security" "permissions" "securityroles" "memberentitlementmanagement" "notification" "servicehook"; do
  grep -q "$area" "$REGISTRY" || fail "Registry missing row for area: $area"
done

# Check no forbidden tokens in matrix table rows
MATRICES=(
  "docs/identity-gap-matrix.md"
  "docs/graph-gap-matrix.md"
  "docs/security-gap-matrix.md"
  "docs/permissions-gap-matrix.md"
  "docs/securityroles-gap-matrix.md"
  "docs/memberentitlementmanagement-gap-matrix.md"
  "docs/notification-gap-matrix.md"
  "docs/servicehook-gap-matrix.md"
)
FORBIDDEN='mapped|partial|missing|implemented|✅|⚠️|🚫|read-only|present|missing-writable|missing-computed|breaking-deferral'
for f in "${MATRICES[@]}"; do
  [[ -f "$f" ]] || fail "Matrix file not found: $f"
  if grep -qE "\| .*($FORBIDDEN)" "$f"; then
    grep -E "\| .*($FORBIDDEN)" "$f" >&2
    fail "Forbidden token found in $f"
  fi
done

echo "OK: gap-registry-wi4a checks passed"
