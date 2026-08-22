#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert gap-registry.md has at least 15 ### sections total (8 from WI-2 + 7 from WI-3)
SECTION_COUNT=$(grep -c '^### ' docs/gap-registry.md 2>/dev/null || echo 0)
[ "$SECTION_COUNT" -ge 15 ] || fail "Expected >=15 ### sections in gap-registry.md, found $SECTION_COUNT"

# Assert no forbidden vocabulary in the 7 Infrastructure matrices
bash .forge/scripts/check-vocab.sh \
  docs/serviceendpoint-gap-matrix.md \
  docs/core-gap-matrix.md \
  docs/build-gap-matrix.md \
  docs/policy-gap-matrix.md \
  docs/git-gap-matrix.md \
  docs/feed-gap-matrix.md \
  docs/wiki-gap-matrix.md

# Assert Infrastructure tier areas are represented in the registry
for area in "Service Endpoint" "Core" "Build" "Policy" "Git" "Feed" "Wiki"; do
  grep -qi "$area" docs/gap-registry.md || fail "Registry missing entry for: $area"
done

echo "WI-3 gate PASSED"
