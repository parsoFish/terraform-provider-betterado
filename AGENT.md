# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

**Iteration 1 (2026-07-02) — full implementation committed:**
- Created `resource_dashboard_framework.go` — full framework resource (Create/Read/Update/Delete/ImportState).
  - Used locally-defined plan modifier types (not `stringplanmodifier.*` — NOT vendored).
  - `readIntoModel()` takes no context.
- Removed `betterado_dashboard` from SDKv2 ResourcesMap in `provider.go`; dropped unused import.
- Added `dashboard.NewDashboardResource` to `framework_provider.go` Resources().
- Updated `provider_test.go`: removed `"betterado_dashboard"` with comment.
- Rewrote `resource_dashboard_test.go`: ProtoV6ProviderFactories, getDirectDashboardClient(), live evidence, idempotency steps.
- Created `docs/dashboard-gap-matrix.md` (14 Dashboard struct fields documented).
- Created `examples/resources/betterado_dashboard/resource.tf`.
- Ran `make docs` + `git checkout -- docs/guides/`.
- All local gates pass: `make test` ✓, `golangci-lint --new-from-rev=main` (0 issues) ✓, `make terrafmt-check` ✓.
- Committed as `feat: migrate betterado_dashboard to terraform-plugin-framework`.

## What worked

- Locally-defined plan modifier structs (see taskagent and release packages as templates).
- `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` for mux tests.
- `getDirectDashboardClient()` pattern for CheckDestroy/check funcs under mux (no Meta()).
- gofumpt separate from gofmt — run `gofumpt -w` to satisfy lint.
- Split evidence into `tryCaptureDashboardEvidence()` + `_ = try...() //nolint:errcheck` to avoid nilerr and errcheck findings.

## What didn't work

- Importing `stringplanmodifier`/`int64planmodifier` sub-packages — NOT vendored.
- `_ = func()` without `//nolint:errcheck` triggers errcheck.
- `if err != nil { return nil }` triggers nilerr.

**Iteration 2 (2026-07-02) — ADO project limit fix committed:**
- Switched all dashboard HCL helpers to `data "betterado_project" "test"` using `SharedFixtureProjectName`
  ("betterado-standing-demo") — the persistent shared project (same as resource_task_group_test.go).
- Removed `projectName` parameter from all `hclDashboard*` helper functions.
- Removed `projectName := testutils.GenerateResourceName()` from all `TestAccDashboard_*` functions.
- Teams are still created per-test (don't count toward project cap).
- All local gates pass: `make test` ✓, `golangci-lint --new-from-rev=main` (0 issues) ✓, `make terrafmt-check` ✓.
- Committed as `fix(acc): use SharedFixtureProjectName data source to avoid ADO 1000-project cap`.

## What worked

- Locally-defined plan modifier structs (see taskagent and release packages as templates).
- `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` for mux tests.
- `getDirectDashboardClient()` pattern for CheckDestroy/check funcs under mux (no Meta()).
- gofumpt separate from gofmt — run `gofumpt -w` to satisfy lint.
- Split evidence into `tryCaptureDashboardEvidence()` + `_ = try...() //nolint:errcheck` to avoid nilerr and errcheck findings.
- `data "betterado_project" "test" { name = SharedFixtureProjectName }` pattern for acceptance tests that need a project (avoids the org 1000-project cap permanently).

## What didn't work

- Importing `stringplanmodifier`/`int64planmodifier` sub-packages — NOT vendored.
- `_ = func()` without `//nolint:errcheck` triggers errcheck.
- `if err != nil { return nil }` triggers nilerr.
- Creating a new `resource "betterado_project"` in acceptance tests — org is permanently at 1000-project cap.

## Open questions

None — all known issues addressed. Awaiting live gate re-run from forge.

## Notes for reflection

- ADO org project limit (1000 projects) is a PERMANENT constraint. ALL acceptance tests that need a project
  must use `SharedFixtureProjectName` via a data source. Never create a new project in acceptance tests
  unless the WI specifically tests project CRUD.
- `SharedFixtureProjectName` defined in `azuredevops/internal/acceptancetests/shared_fixtures.go` line ~314.
- This pattern is used by resource_task_group_test.go and now resource_dashboard_test.go.
