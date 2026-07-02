# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

**Root cause of prior gate failure:** Gate runs `-run TestAccGitRepositoryBranch` but test functions were named `TestAccGitRepoBranch_*` (missing "itory"). Zero tests matched → forge rejected with "[no tests to run]".

**What was done:**

1. Created `azuredevops/internal/service/git/resource_git_repository_branch_framework.go`:
   - Full framework implementation (Create/Read/Delete/ImportState; Update returns error — all fields ForceNew).
   - Uses `gitStateString()` and `gitRequiresReplace()` from same package — `stringplanmodifier` is NOT in vendor.
   - `resolveRefToCommitID()` uses `switch {}` (gocritic flags if-else chains on 3 conditions).
   - `updateRefsFramework()` is distinct name to avoid duplicate with SDKv2 `updateRefs`.
2. Deregistered from SDKv2 `provider.go` ResourcesMap (commented with migration note).
3. Registered `git.NewGitRepositoryBranchResource` in `framework_provider.go` Resources().
4. Removed from `provider_test.go` TestProvider_HasChildResources expectedResources.
5. Renamed 4 acceptance test functions: `TestAccGitRepoBranch_*` → `TestAccGitRepositoryBranch_*`.
6. Switched tests to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`.

**Verification (offline):**
- `go build -tags all ./...` → clean
- `golangci-lint run --new-from-rev=main ./azuredevops/...` → 0 issues
- `go test -tags all -run TestAccGitRepositoryBranch ./azuredevops/internal/acceptancetests/ -v` → 4 tests SKIP (no TF_ACC), exit 0

## What worked

- Package-local `gitRequiresReplace()` and `gitStateString()` (not `stringplanmodifier` — not vendored).
- The gate `-run TestAccGitRepositoryBranch` requires "Repository" in full — "Repo" won't match.

## What didn't work

- `stringplanmodifier` package not in vendor → compilation failure if imported.
- if-else chain on 3 conditions → gocritic violation → use `switch {}`.

## Open questions

_(none blocking)_

## Notes for reflection

- Test function name vs gate `-run` pattern mismatch is an easy mistake. Always grep actual function names against gate pattern before committing.
