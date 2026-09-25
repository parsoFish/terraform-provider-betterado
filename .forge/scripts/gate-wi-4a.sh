#!/bin/bash
set -e
for f in docs/identity-gap-matrix.md docs/graph-gap-matrix.md \
  docs/security-gap-matrix.md docs/permissions-gap-matrix.md \
  docs/securityroles-gap-matrix.md docs/memberentitlementmanagement-gap-matrix.md \
  docs/notification-gap-matrix.md docs/servicehook-gap-matrix.md; do
  test -f "$f"
  BAD=$(grep '^|' "$f" | grep -cE '\bmapped\b|\bsupported\b|\bimplemented\b|\bpartial\b|\bmissing\b|\bpresent\b|gap-resolved' || true)
  test "$BAD" -eq 0
done
test -f docs/gap-registry.md
test $(grep -c 'v7.2 delta' docs/gap-registry.md) -ge 23
test -f .forge/scripts/gate-wi-4a.sh
echo WI-4a PASSED
