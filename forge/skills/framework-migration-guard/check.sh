#!/usr/bin/env bash
# forge/skills/framework-migration-guard/check.sh
#
# Guards the ongoing SDK v2 → Plugin Framework migration against leaving
# orphaned dead code. Run from the repo root.
#
# Exit codes:
#   0 — all resource constructors are registered; no orphans
#   1 — one or more orphan constructors found (dead code)
#
# Dependencies: grep, sed, sort, comm — all standard POSIX tools.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
SERVICE_DIR="$REPO_ROOT/azuredevops/internal/service"
SDK_PROVIDER="$REPO_ROOT/azuredevops/provider.go"
FW_PROVIDER="$REPO_ROOT/azuredevops/internal/provider/framework_provider.go"

# ── 1. Enumerate all resource constructor names found in service/**/*.go ──────
#
# SDK v2 pattern:   func ResourceXxx() *schema.Resource
# Framework pattern: func NewXxxResource() resource.Resource
#                    func NewXxxResource() func() resource.Resource  (also matches)
#
# We collect just the bare function name (e.g. ResourceFoo, NewFooResource).

sdk_constructors=$(
  grep -rh --include="*.go" \
    -E "^func (Resource[A-Z][A-Za-z0-9_]*)\(\)" \
    "$SERVICE_DIR" \
  | sed -E 's/^func ([A-Za-z0-9_]+)\(\).*/\1/' \
  | sort -u
)

fw_constructors=$(
  grep -rh --include="*.go" \
    -E "^func (New[A-Z][A-Za-z0-9_]*Resource)\(\)" \
    "$SERVICE_DIR" \
  | sed -E 's/^func ([A-Za-z0-9_]+)\(\).*/\1/' \
  | sort -u
)

all_constructors=$(printf '%s\n%s\n' "$sdk_constructors" "$fw_constructors" | sort -u)

# ── 2. Extract names referenced in each registration site ─────────────────────
#
# provider.go  ResourcesMap: values are calls like   taskagent.ResourceFoo()
# framework_provider.go Resources(): values are      taskagent.NewFooResource,
#                                                     (bare reference, no parens)

sdk_registered=$(
  grep -Eoh '\b(Resource[A-Z][A-Za-z0-9_]*)\(' "$SDK_PROVIDER" \
  | sed 's/($//' \
  | sort -u
)

fw_registered=$(
  grep -Eoh '\b(New[A-Z][A-Za-z0-9_]*Resource)\b' "$FW_PROVIDER" \
  | sort -u
)

registered=$(printf '%s\n%s\n' "$sdk_registered" "$fw_registered" | sort -u)

# ── 3. Find constructors present in service/ but absent from both providers ───

orphans=$(comm -23 \
  <(echo "$all_constructors") \
  <(echo "$registered") \
)

# ── 4. Report ─────────────────────────────────────────────────────────────────

if [ -z "$orphans" ]; then
  echo "framework-migration-guard: OK — no orphaned resource constructors found."
  exit 0
fi

echo "framework-migration-guard: FAIL — orphaned resource constructors detected:"
echo ""
while IFS= read -r name; do
  # Find the file that declares this constructor for a helpful error message
  file=$(grep -rhl --include="*.go" "^func ${name}()" "$SERVICE_DIR" 2>/dev/null | head -1 || true)
  if [ -n "$file" ]; then
    rel="${file#$REPO_ROOT/}"
    echo "  ORPHAN: $name  ($rel)"
  else
    echo "  ORPHAN: $name"
  fi
done <<< "$orphans"

echo ""
echo "These constructors are defined in service/ but registered in NEITHER"
echo "  $SDK_PROVIDER (ResourcesMap)"
echo "  $FW_PROVIDER (Resources())"
echo ""
echo "See forge/skills/framework-migration-guard/SKILL.md for remediation steps."
exit 1
