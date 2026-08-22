#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert gap-registry.md has at least 8 new ## sections (one per Release/Pipeline area)
SECTION_COUNT=$(grep -c '^### ' docs/gap-registry.md 2>/dev/null || echo 0)
[ "$SECTION_COUNT" -ge 8 ] || fail "Expected >=8 ### sections in gap-registry.md, found $SECTION_COUNT"

# Assert no forbidden vocabulary in the 8 Release/Pipeline matrices
bash .forge/scripts/check-vocab.sh \
  docs/release-definition-gap-matrix.md \
  docs/release-folder-gap-matrix.md \
  docs/release-definition-permissions-gap-matrix.md \
  docs/task-group-gap-matrix.md \
  docs/taskagent-gap-matrix.md \
  docs/approvalsandchecks-gap-matrix.md \
  docs/pipelinesapproval-gap-matrix.md \
  docs/pipelines-v2-gap-matrix.md

# Assert each area has a v7.1→v7.2 delta line in the registry
grep -q 'v7\.1.*v7\.2' docs/gap-registry.md || fail "No v7.1→v7.2 delta lines found in gap-registry.md"

# Assert gap-open count lines are present in the registry
grep -q 'gap-open count' docs/gap-registry.md || fail "No gap-open count lines found in gap-registry.md"

echo "WI-2 gate PASSED"
