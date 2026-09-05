#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $1" >&2; exit 1; }

REGISTRY="docs/gap-registry.md"
PROFILE="brain/projects/terraform-provider-betterado/profile.md"
ROADMAP="docs/api-coverage-roadmap.md"

# Priority backlog section exists
grep -q "^## Priority backlog" "$REGISTRY" || fail "Registry missing ## Priority backlog section"

# All 31 areas present in Coverage index
AREAS=(
  "release-definition" "release-folder" "release-definition-permissions"
  "task-group" "taskagent" "approvalsandchecks" "pipelinesapproval" "pipelines-v2"
  "serviceendpoint" "core" "build" "policy" "git" "feed" "wiki"
  "identity" "graph" "security" "permissions" "securityroles"
  "memberentitlementmanagement" "notification" "servicehook"
  "dashboard" "extension" "gallery-extensionmanagement" "featuremanagement"
  "workitemtracking" "workitemtrackingprocess" "accounts-profile" "test"
)
for area in "${AREAS[@]}"; do
  grep -q "$area" "$REGISTRY" || fail "Registry missing Coverage index row for: $area"
done

# Profile net-new table has >5 rows (count lines with | betterado_ pattern)
NET_NEW_COUNT=$(grep -c '| `betterado_' "$PROFILE" 2>/dev/null || true)
if [[ -z "$NET_NEW_COUNT" ]]; then
  NET_NEW_COUNT=0
fi
[[ "$NET_NEW_COUNT" -gt 5 ]] || fail "profile.md net-new table has $NET_NEW_COUNT rows (need >5)"

# api-coverage-roadmap.md no longer contains "parked plan"
if grep -qi "parked plan" "$ROADMAP"; then
  fail "api-coverage-roadmap.md still contains 'parked plan'"
fi

echo "OK: gap-registry-wi5 checks passed"
