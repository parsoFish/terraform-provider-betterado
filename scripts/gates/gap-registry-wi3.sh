#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $1" >&2; exit 1; }

REGISTRY="docs/gap-registry.md"
grep -q "^## Vocabulary" "$REGISTRY" || fail "Vocabulary section missing from registry"

# Check 7 registry rows present
for area in "serviceendpoint" "core" "build" "policy" "git" "feed" "wiki"; do
  grep -q "$area" "$REGISTRY" || fail "Registry missing row for area: $area"
done

# Check no forbidden tokens in matrix table rows
MATRICES=(
  "docs/serviceendpoint-gap-matrix.md"
  "docs/core-gap-matrix.md"
  "docs/build-gap-matrix.md"
  "docs/policy-gap-matrix.md"
  "docs/git-gap-matrix.md"
  "docs/feed-gap-matrix.md"
  "docs/wiki-gap-matrix.md"
)
FORBIDDEN='mapped|partial|missing|implemented|✅|⚠️|🚫|read-only|present|missing-writable|missing-computed|breaking-deferral'
for f in "${MATRICES[@]}"; do
  [[ -f "$f" ]] || fail "Matrix file not found: $f"
  if grep -qE "\| .*($FORBIDDEN)" "$f"; then
    grep -E "\| .*($FORBIDDEN)" "$f" >&2
    fail "Forbidden token found in $f"
  fi
done

echo "OK: gap-registry-wi3 checks passed"
