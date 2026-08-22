#!/usr/bin/env bash
set -euo pipefail
fail() { echo "FAIL: $*" >&2; exit 1; }

# Assert gap-registry.md exists
test -f docs/gap-registry.md || fail "docs/gap-registry.md missing"

# Assert vocabulary section with all four canonical tokens
grep -q 'covered' docs/gap-registry.md || fail "missing 'covered' token"
grep -q 'gap-open' docs/gap-registry.md || fail "missing 'gap-open' token"
grep -q 'gap-deferred' docs/gap-registry.md || fail "missing 'gap-deferred' token"
grep -q 'out-of-scope' docs/gap-registry.md || fail "missing 'out-of-scope' token"

# Assert ## Vocabulary section header exists
grep -q '## Vocabulary' docs/gap-registry.md || fail "missing ## Vocabulary section"

# Assert check-vocab.sh helper exists and is executable
test -f .forge/scripts/check-vocab.sh || fail ".forge/scripts/check-vocab.sh missing"
test -x .forge/scripts/check-vocab.sh || fail ".forge/scripts/check-vocab.sh not executable"

# Assert gate-wi1.sh itself exists (self-referential but confirms creates worked)
test -f .forge/scripts/gate-wi1.sh || fail ".forge/scripts/gate-wi1.sh missing"

echo "WI-1 gate PASSED"
