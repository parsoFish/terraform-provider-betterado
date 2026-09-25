#!/bin/bash
set -e
test -f docs/gap-registry.md
test $(grep -c '## Vocabulary\|## Classification\|## Area index\|## Priority backlog' docs/gap-registry.md) -ge 4
grep -q 'gap-open' docs/gap-registry.md
grep -q 'gap-deferred' docs/gap-registry.md
grep -q 'covered' docs/gap-registry.md
grep -q 'out-of-scope' docs/gap-registry.md
test -f .forge/scripts/gate-wi-1.sh
echo WI-1 PASSED
