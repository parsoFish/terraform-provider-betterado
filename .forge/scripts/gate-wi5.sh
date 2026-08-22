#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert ## Priority backlog section exists
grep -q '^## Priority backlog' docs/gap-registry.md || fail "Missing ## Priority backlog section in gap-registry.md"

# Assert priority backlog has Tier 1, Tier 2, Tier 3 subsections
grep -q 'Tier 1' docs/gap-registry.md || fail "Missing Tier 1 in priority backlog"
grep -q 'Tier 2' docs/gap-registry.md || fail "Missing Tier 2 in priority backlog"
grep -q 'Tier 3' docs/gap-registry.md || fail "Missing Tier 3 in priority backlog"

# Assert profile.md net-new table has more than 5 rows (original had 5)
NETNEWTABLE_ROWS=$(grep -c 'betterado_' brain/projects/terraform-provider-betterado/profile.md 2>/dev/null || echo 0)
[ "$NETNEWTABLE_ROWS" -gt 5 ] || fail "profile.md net-new table has <=5 betterado_ rows; expected >5"

# Assert api-coverage-roadmap.md no longer contains "parked plan"
grep -qi 'parked plan' docs/api-coverage-roadmap.md && fail "api-coverage-roadmap.md still contains 'parked plan'"

# Assert gap-registry.md has all 31 areas (### sections)
SECTION_COUNT=$(grep -c '^### ' docs/gap-registry.md 2>/dev/null || echo 0)
[ "$SECTION_COUNT" -ge 31 ] || fail "Expected >=31 ### sections in gap-registry.md, found $SECTION_COUNT"

# Assert all 31 matrices are clean (run check-vocab.sh on all)
bash .forge/scripts/check-vocab.sh \
  docs/release-definition-gap-matrix.md \
  docs/release-folder-gap-matrix.md \
  docs/release-definition-permissions-gap-matrix.md \
  docs/task-group-gap-matrix.md \
  docs/taskagent-gap-matrix.md \
  docs/approvalsandchecks-gap-matrix.md \
  docs/pipelinesapproval-gap-matrix.md \
  docs/pipelines-v2-gap-matrix.md \
  docs/serviceendpoint-gap-matrix.md \
  docs/core-gap-matrix.md \
  docs/build-gap-matrix.md \
  docs/policy-gap-matrix.md \
  docs/git-gap-matrix.md \
  docs/feed-gap-matrix.md \
  docs/wiki-gap-matrix.md \
  docs/identity-gap-matrix.md \
  docs/graph-gap-matrix.md \
  docs/security-gap-matrix.md \
  docs/permissions-gap-matrix.md \
  docs/securityroles-gap-matrix.md \
  docs/memberentitlementmanagement-gap-matrix.md \
  docs/notification-gap-matrix.md \
  docs/servicehook-gap-matrix.md \
  docs/dashboard-gap-matrix.md \
  docs/extension-gap-matrix.md \
  docs/gallery-extensionmanagement-gap-matrix.md \
  docs/featuremanagement-gap-matrix.md \
  docs/workitemtracking-gap-matrix.md \
  docs/workitemtrackingprocess-gap-matrix.md \
  docs/accounts-profile-gap-matrix.md \
  docs/test-gap-matrix.md

echo "WI-5 gate PASSED"
