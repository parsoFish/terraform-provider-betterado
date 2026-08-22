#!/usr/bin/env bash
set -euo pipefail

# check-vocab.sh <file1> [file2 ...]
# Exits 1 if any forbidden vocabulary token is found in any file.
# Forbidden tokens: mapped, supported, implemented, partial, missing, present, gap-resolved
# and emoji tokens ✅ ⚠️ 🚫

FORBIDDEN_PATTERN='mapped|supported|implemented|partial|missing|present|gap-resolved|✅|⚠️|🚫'

fail() { echo "FAIL: $*" >&2; exit 1; }

for f in "$@"; do
  if grep -qP "$FORBIDDEN_PATTERN" "$f" 2>/dev/null; then
    echo "Forbidden vocabulary found in: $f" >&2
    grep -nP "$FORBIDDEN_PATTERN" "$f" >&2 || true
    fail "Forbidden token in $f"
  fi
done
echo "OK: no forbidden vocabulary tokens found."
