#!/bin/bash
set -e
for f in docs/dashboard-gap-matrix.md docs/extension-gap-matrix.md \
  docs/gallery-extensionmanagement-gap-matrix.md docs/featuremanagement-gap-matrix.md \
  docs/workitemtracking-gap-matrix.md docs/workitemtrackingprocess-gap-matrix.md \
  docs/accounts-profile-gap-matrix.md docs/test-gap-matrix.md; do
  test -f "$f"
  BAD=$(grep '^|' "$f" | grep -cE '\bmapped\b|\bsupported\b|\bimplemented\b|\bpartial\b|\bmissing\b|\bpresent\b|gap-resolved' || true)
  test "$BAD" -eq 0
done
test -f docs/gap-registry.md
test $(grep -c 'v7.2 delta' docs/gap-registry.md) -ge 31
test -f .forge/scripts/gate-wi-4b.sh
echo WI-4b PASSED
