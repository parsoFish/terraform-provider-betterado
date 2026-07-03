# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

**Gate failure analyzed:** `.forge/last-gate-failure.md` showed `TestAccVariableGroup*` and `TestAccVariableGroupVariable*` tests failing with "Failed to add a project as this organization already has 1000 projects".

**Root cause:** `resource_variable_group_variable_test.go` tests were:
1. Using `resource "betterado_project" "test"` in HCL fixtures — creating a new project each run, hitting the 1000-project org cap.
2. Using `Providers: testutils.GetProviders()` (SDKv2 singleton) instead of `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`.
3. Using `testutils.GetProvider().Meta().(*client.AggregatedClient)` in check helpers — incompatible with mux provider (Meta() is nil in mux path).

`resource_variable_group_test.go` and `data_variable_group_test.go` already used fixture project — variable_group_variable tests were the stragglers.

**Fix applied (commit 6778cfcb):**
- Rewrote `resource_variable_group_variable_test.go` to use `data "betterado_project" "fixture" { name = SharedFixtureProjectName }` in all HCL templates.
- Removed `projectName` parameter from fixture functions.
- Switched all 3 tests to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`.
- Replaced `testutils.GetProvider().Meta().(*client.AggregatedClient)` with `testutils.GetDirectClient()` in `checkVariableGroupVariableFromState`.
- Fixed pre-existing gofmt drift in `resource_variable_group_framework.go` and `provider.go` (unblocked `make test` fmtcheck).

**Verification:** `make test` passes. `go test -tags all -count=1 ./azuredevops/internal/service/...` passes.

## What worked

- Fixture project pattern: `data "betterado_project" "fixture" { name = SharedFixtureProjectName }` — already used in variable_group, environment, task_group tests.
- `GetDirectClient()` works for check helpers with mux provider; `GetProvider().Meta()` is nil in mux path.
- Run `gofmt -w <file>` to fix drift; `make test` fmtcheck catches it.

## What didn't work

_(none)_

## Open questions

_(none)_

## Notes for reflection

- The 1000-project org cap has been a recurring pain point. Every new acceptance test must use the standing fixture project. The pattern is now consistently applied across all taskagent acceptance tests.
- Pre-existing gofmt drift should be fixed opportunistically when touching files.
