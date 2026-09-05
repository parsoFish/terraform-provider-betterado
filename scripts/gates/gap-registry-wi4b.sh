#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $1" >&2; exit 1; }

REGISTRY="docs/gap-registry.md"
grep -q "^## Vocabulary" "$REGISTRY" || fail "Vocabulary section missing from registry"

# Check 8 registry rows present for Collaboration/Long-tail areas
for area in "dashboard" "extension" "gallery-extensionmanagement" "featuremanagement" "workitemtracking" "workitemtrackingprocess" "accounts-profile" "test"; do
  grep -q "$area" "$REGISTRY" || fail "Registry missing row for area: $area"
done

# Check no forbidden tokens in matrix table rows or legend blocks
MATRICES=(
  "docs/dashboard-gap-matrix.md"
  "docs/extension-gap-matrix.md"
  "docs/gallery-extensionmanagement-gap-matrix.md"
  "docs/featuremanagement-gap-matrix.md"
  "docs/workitemtracking-gap-matrix.md"
  "docs/workitemtrackingprocess-gap-matrix.md"
  "docs/accounts-profile-gap-matrix.md"
  "docs/test-gap-matrix.md"
)
FORBIDDEN='mapped|partial|missing|implemented|✅|⚠️|🚫|read-only|present|missing-writable|missing-computed|breaking-deferral'
for f in "${MATRICES[@]}"; do
  [[ -f "$f" ]] || fail "Matrix file not found: $f"
  if grep -qE "\| .*($FORBIDDEN)" "$f"; then
    grep -E "\| .*($FORBIDDEN)" "$f" >&2
    fail "Forbidden token found in $f"
  fi
done

echo "OK: gap-registry-wi4b checks passed"
