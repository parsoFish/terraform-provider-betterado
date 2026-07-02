# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 — Full framework migration

**Gate context:** .forge/last-gate-failure.md showed all failures were "organization already has 1000 projects" — live ADO environment capacity issue, NOT code defects.

**Work done:**

1. Created `azuredevops/internal/service/git/resource_git_repository_framework.go`:
   - Implements `resource.Resource`, `ResourceWithConfigure`, `ResourceWithImportState`
   - Inline plan modifiers (NOT sub-packages — those don't exist in vendor)
   - `initialization` uses `schema.ListNestedBlock` + `types.List` model field
   - `expandInitialization()` uses `initList.ElementsAs(ctx, &blocks, false)`
   - `flattenGitRepositoryFramework()` preserves `Initialization`/`ParentRepositoryID` from prior state
   - `gitRepositoryRead()` reuses the same-package SDKv2 helper (unexported but same package = accessible)
   - `Size` in ADO API is `uint64` — must cast: `int64(*repo.Size)`

2. Created `azuredevops/internal/service/git/data_git_repository_framework.go`:
   - Reuses `getGitRepositoriesByNameAndProject()` from `data_git_repositories.go`

3. Registered both in `framework_provider.go`; deregistered from `provider.go`

4. Updated `provider_test.go` to remove git_repository from both expectedResources and expectedDataSources

5. Updated all acceptance tests to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`:
   - `resource_git_repository_test.go` — all 12 tests
   - `data_git_repository_test.go` — both tests
   - `checkGitRepoExists`/`checkGitRepoDestroyed` now use `getDirectClient()`
   - Added `captureGitRepositoryEvidence()` for forge evidence

6. **Key refactor:** moved `getDirectClient()` from `resource_task_group_test.go` (build-tagged)
   to new `direct_client_test.go` (no build tag) so files without build constraints can use it.

## What worked

- Inline plan modifiers following `resource_task_group_framework.go` pattern
- `gofumpt -w` (stricter than `gofmt`) satisfies the gofumpt linter
- `go build -tags all ./...` catches vendor/import issues early
- `errcheck` requires named variable assignment, not `_ =` with multi-return

## What didn't work

- Sub-packages like `booldefault`, `stringplanmodifier` NOT in vendor — must inline
- `_ = someCall()` does NOT suppress errcheck — must `err := ...; if err != nil { ... }`

## Open questions

- Live gate needs ADO org with <1000 projects. No code fix available.
- `password` field: `WriteOnly: true` in SDKv2 becomes `Sensitive: true` in framework (no WriteOnly concept).
- If idempotency fails on `initialization` block (ADO never echoes back), may need to explicitly handle empty list vs null list.

### Iteration 2 — Fix missing os import (build error)

**Gate context:** `.forge/last-gate-failure.md` showed `resource_task_group_test.go:226:31: undefined: os`.

**Root cause:** Iteration 1 removed `"os"` from `resource_task_group_test.go`'s imports when moving
`getDirectClient()` to `direct_client_test.go`. But `os.Getenv("AZDO_ORG_SERVICE_URL")` is still
called at line 226 in the evidence capture closure.

**Fix:** Added `"os"` back to the import block in `resource_task_group_test.go`.

**Result:** `go build -tags all ./...` passes. `golangci-lint --new-from-rev=main` → 0 issues.
Gate test now builds (exits 0.008s with no-TF_ACC skip).

## Notes for reflection

- `getDirectClient()` should be in a shared no-build-tag file from the start in all future migration WIs.
- The "1000 projects" gate blocker is a live environment issue, not a code issue.
- When moving code to a shared file, carefully audit ALL usages of removed imports — not just the moved function, but ALL callers in the original file.
