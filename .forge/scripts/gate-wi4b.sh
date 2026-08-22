#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert no forbidden vocabulary in the 8 Collaboration/Long-tail tier B matrices
bash .forge/scripts/check-vocab.sh \
  docs/dashboard-gap-matrix.md \
  docs/extension-gap-matrix.md \
  docs/gallery-extensionmanagement-gap-matrix.md \
  docs/featuremanagement-gap-matrix.md \
  docs/workitemtracking-gap-matrix.md \
  docs/workitemtrackingprocess-gap-matrix.md \
  docs/accounts-profile-gap-matrix.md \
  docs/test-gap-matrix.md

# Assert Collaboration/Long-tail tier B areas are represented in the registry
for area in "Dashboard" "Extension" "Gallery" "Feature Management" "Work Item Tracking" "Work Item Tracking Process" "Accounts" "Test"; do
  grep -qi "$area" docs/gap-registry.md || fail "Registry missing entry for: $area"
done

# Assert total ### sections is now at least 31 (all areas covered across all tier WIs)
# WI-2: 8, WI-3: 7, WI-4a: 8, WI-4b: 8 = 31 total
SECTION_COUNT=$(grep -c '^### ' docs/gap-registry.md 2>/dev/null || echo 0)
[ "$SECTION_COUNT" -ge 31 ] || fail "Expected >=31 ### sections in gap-registry.md (all 31 areas), found $SECTION_COUNT"

echo "WI-4b gate PASSED"
