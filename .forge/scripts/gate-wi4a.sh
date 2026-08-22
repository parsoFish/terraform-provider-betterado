#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert no forbidden vocabulary in the 8 Identity/Security tier A matrices
bash .forge/scripts/check-vocab.sh \
  docs/identity-gap-matrix.md \
  docs/graph-gap-matrix.md \
  docs/security-gap-matrix.md \
  docs/permissions-gap-matrix.md \
  docs/securityroles-gap-matrix.md \
  docs/memberentitlementmanagement-gap-matrix.md \
  docs/notification-gap-matrix.md \
  docs/servicehook-gap-matrix.md

# Assert Identity/Security tier A areas are represented in the registry
for area in "Identity" "Graph" "Security" "Permissions" "Security Roles" "Member Entitlement" "Notification" "Service Hook"; do
  grep -qi "$area" docs/gap-registry.md || fail "Registry missing entry for: $area"
done

# Assert gap-open count lines are present for these areas
grep -c 'gap-open count' docs/gap-registry.md | awk -v n=8 '$1>=n' \
  || fail "Fewer than 8 gap-open count lines in registry (WI-4a may be missing entries)"

echo "WI-4a gate PASSED"
