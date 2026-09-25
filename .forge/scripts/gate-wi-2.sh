#!/bin/bash
set -e
for f in docs/release-definition-gap-matrix.md docs/release-folder-gap-matrix.md \
  docs/release-definition-permissions-gap-matrix.md docs/task-group-gap-matrix.md \
  docs/taskagent-gap-matrix.md docs/approvalsandchecks-gap-matrix.md \
  docs/pipelinesapproval-gap-matrix.md docs/pipelines-v2-gap-matrix.md; do
  test -f "$f"
  BAD=$(grep '^|' "$f" | grep -cE '\bmapped\b|\bsupported\b|\bimplemented\b|\bpartial\b|\bmissing\b|\bpresent\b|gap-resolved' || true)
  test "$BAD" -eq 0
done
test -f docs/gap-registry.md
test $(grep -c 'v7.2 delta' docs/gap-registry.md) -ge 8
test -f .forge/scripts/gate-wi-2.sh
echo WI-2 PASSED
