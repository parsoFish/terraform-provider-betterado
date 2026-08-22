#!/usr/bin/env bash
# check-vocab.sh — asserts no forbidden vocabulary appears in the supplied files.
# Usage: bash .forge/scripts/check-vocab.sh <file1> [file2 ...]
set -euo pipefail

fail() { echo "FAIL: $*" >&2; exit 1; }

FORBIDDEN_PATTERN='(\bmapped\b|\bsupported\b|\bimplemented\b|\bpartial\b|\bmissing\b|\bpresent\b|gap-resolved|✅|⚠️|🚫)'

if [ "$#" -eq 0 ]; then
  echo "Usage: $0 <file> [file ...]" >&2
  exit 1
fi

any_fail=0
for f in "$@"; do
  if [ ! -f "$f" ]; then
    echo "FAIL: file not found: $f" >&2
    any_fail=1
    continue
  fi
  # Use grep -P for Perl-compatible regex (supports \b word boundaries and Unicode)
  if grep -Pn "$FORBIDDEN_PATTERN" "$f" 2>/dev/null; then
    echo "FAIL: forbidden vocabulary found in $f (lines above)" >&2
    any_fail=1
  fi
done

if [ "$any_fail" -ne 0 ]; then
  fail "Forbidden vocabulary found in one or more files — normalize them first"
fi

echo "check-vocab: PASSED — no forbidden vocabulary in checked files"
