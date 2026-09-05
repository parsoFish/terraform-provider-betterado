#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $1" >&2; exit 1; }

REGISTRY="docs/gap-registry.md"
grep -q "^## Vocabulary" "$REGISTRY" || fail "Vocabulary section missing from registry"

# Check 8 registry rows present
for area in "release-definition" "release-folder" "release-definition-permissions" "task-group" "taskagent" "approvalsandchecks" "pipelinesapproval" "pipelines-v2"; do
  grep -q "$area" "$REGISTRY" || fail "Registry missing row for area: $area"
done

# Check no forbidden tokens in matrix table rows
MATRICES=(
  "docs/release-definition-gap-matrix.md"
  "docs/release-folder-gap-matrix.md"
  "docs/release-definition-permissions-gap-matrix.md"
  "docs/task-group-gap-matrix.md"
  "docs/taskagent-gap-matrix.md"
  "docs/approvalsandchecks-gap-matrix.md"
  "docs/pipelinesapproval-gap-matrix.md"
  "docs/pipelines-v2-gap-matrix.md"
)
FORBIDDEN='mapped|partial|missing|implemented|✅|⚠️|🚫|read-only|present|missing-writable|missing-computed|breaking-deferral'
for f in "${MATRICES[@]}"; do
  [[ -f "$f" ]] || fail "Matrix file not found: $f"
  if grep -qE "\| .*($FORBIDDEN)" "$f"; then
    grep -E "\| .*($FORBIDDEN)" "$f" >&2
    fail "Forbidden token found in $f"
  fi
done

echo "OK: gap-registry-wi2 checks passed"
