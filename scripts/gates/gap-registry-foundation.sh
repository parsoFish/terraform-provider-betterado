#!/usr/bin/env bash
set -euo pipefail

REGISTRY="docs/gap-registry.md"

fail() { echo "FAIL: $1" >&2; exit 1; }

[[ -f "$REGISTRY" ]] || fail "$REGISTRY does not exist"

grep -q "^## Vocabulary" "$REGISTRY" || fail "$REGISTRY missing ## Vocabulary section"

grep -q '`covered`' "$REGISTRY"         || fail "$REGISTRY missing 'covered' token definition"
grep -q '`gap-open`' "$REGISTRY"        || fail "$REGISTRY missing 'gap-open' token definition"
grep -q '`gap-deferred`' "$REGISTRY"    || fail "$REGISTRY missing 'gap-deferred' token definition"
grep -q '`out-of-scope`' "$REGISTRY"    || fail "$REGISTRY missing 'out-of-scope' token definition"

echo "OK: gap-registry-foundation checks passed"
