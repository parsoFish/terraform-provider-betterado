#!/bin/bash
set -e
test -f docs/gap-registry.md
grep -q '## Priority backlog' docs/gap-registry.md
grep -qE 'Tier 1|Tier 2|Tier 3' docs/gap-registry.md
test -f docs/profile-corrections.md
test $(grep -c '```' docs/profile-corrections.md) -ge 4
test -f docs/api-coverage-roadmap.md
! grep -q 'parked plan' docs/api-coverage-roadmap.md
for feat in FEAT-1 FEAT-2 FEAT-3 FEAT-4; do
  grep -qE "${feat}.*(complete|partial|not-started|shipped)" docs/api-coverage-roadmap.md
done
test -f .forge/scripts/gate-wi-5.sh
echo WI-5 PASSED
