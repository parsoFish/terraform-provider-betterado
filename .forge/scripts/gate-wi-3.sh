#!/bin/bash
set -e
for f in docs/serviceendpoint-gap-matrix.md docs/core-gap-matrix.md \
  docs/build-gap-matrix.md docs/policy-gap-matrix.md docs/git-gap-matrix.md \
  docs/feed-gap-matrix.md docs/wiki-gap-matrix.md; do
  test -f "$f"
  BAD=$(grep '^|' "$f" | grep -cE '\bmapped\b|\bsupported\b|\bimplemented\b|\bpartial\b|\bmissing\b|\bpresent\b|gap-resolved' || true)
  test "$BAD" -eq 0
done
test -f docs/gap-registry.md
test $(grep -c 'v7.2 delta' docs/gap-registry.md) -ge 15
test -f .forge/scripts/gate-wi-3.sh
echo WI-3 PASSED
