# Agent Memory — UWI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (this iteration) — COMPLETE

**Problem:** Two ACs to satisfy:
1. AC1: demo.json needed BOTH liveEvidence checkpoints (already done by prior WIs)
2. AC2: PR was CONFLICTING with main — needed merge + conflict resolution

**What happened:**
- AC1 was already satisfied: forge/history/.../demo/demo.json had both dashboard-acceptance-resource and acceptance-resource checkpoints with real ADO URLs
- AC2 required merging origin/main into the branch
  - `git merge origin/main --no-commit --no-ff` revealed 5 conflicts
  - Conflicts: .forge/project.json, CHANGELOG.md, PROVIDER_VERSION.txt, framework_provider.go, provider.go
  - Root cause: main had merged build + memberentitlement framework migrations (#49, #53) since our branch diverged

**Conflict resolution strategy:**
- `framework_provider.go`: imports → keep BOTH sets (dashboard+extension from our branch, build+memberentitlementmanagement from main); Resources() was auto-merged correctly (all resources present)
- `provider.go`: two conflict regions — take main's approach (remove build/entitlement/pipeline_auth from SDKv2) PLUS our approach (remove dashboard/extension from SDKv2). Result: all 6 resource types removed from SDKv2 ResourcesMap, only comments remain
- `CHANGELOG.md`: keep our [Unreleased] dashboard+extension entries, prepend main's [1.3.0] block between [Unreleased] and [1.2.0]
- `PROVIDER_VERSION.txt`: bump to 1.3.1 (main=1.3.0, our branch had 1.2.1 from a prior WI)
- `.forge/project.json`: take main's version (had gofumpt addition to ci_fix_cmd)

**Post-merge issue:** gofmt found one alignment issue in provider.go (betterado_build_folder_permissions tab stop). Fixed with `make fmt`.

## What worked

- `git merge origin/main --no-commit --no-ff` to preview conflicts before committing
- Taking main's side for provider.go SDKv2 removals (build/entitlement) and adding our dashboard/extension removals on top
- `make fmt` after manual conflict resolution to catch formatting mismatches
- Full build `go build ./...` to verify before committing

## What didn't work

- Initial conflict resolution for provider.go had a formatting issue (inconsistent tab alignment on betterado_build_folder_permissions line) that caused gofmt to fail in CI

## Final state

- All 4 GitHub CI checks GREEN (depscheck, go-lint, terrafmt, test)
- PR 45: mergeable=MERGEABLE, mergeStateStatus=CLEAN
- Both ACs satisfied; no outstanding issues

## Open questions

_(none)_

## Notes for reflection

- When merging branches with fan-in files (provider.go, framework_provider.go), the correct approach is to take the union of both sets of changes (SDKv2 removals from all migrations)
- Always run `make fmt` after manual conflict resolution in Go files — golangci-lint catches gofmt issues with any tab misalignment
