# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-03)

Completed WI-2 in full in a single iteration.

**Approach:**
1. The `testplan` SDK package is in `third_party/azure-devops-go-api/azuredevops/v7/testplan/` but was NOT in `vendor/` or `vendor/modules.txt`. Copied both `client.go` and `models.go` to `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan/` and added the entry to `vendor/modules.txt`.
2. Generated `azdosdkmocks/testplan_sdk_mock.go` manually (no mockgen available at runtime) by following the same pattern as `taskagent_sdk_mock.go` — covering all methods in the testplan Client interface.
3. Added `TestPlanClient testplan.Client` field to `AggregatedClient` and called `testplan.NewClient(ctx, connection)` (returns `Client` with no error, unlike most other SDK constructors) in `GetAzdoClient()`.
4. Created `azuredevops/internal/service/testplan/` with:
   - `resource_test_plan_framework.go` — full CRUD; inline plan modifiers (no `stringplanmodifier` subpackage, which is not vendored); expand/flatten for all fields including RFC3339 dates via `azuredevops.Time`.
   - `data_test_plan_framework.go` — datasource with project_id + plan_id inputs; same flatten logic.
   - `resource_test_plan_framework_test.go` — `TestUnitTestPlan_expandFlatten`, `TestUnitTestPlan_Read404`, `TestUnitTestPlan_Schema`, `TestUnitTestPlan_Create_Error`.
   - `data_test_plan_framework_test.go` — `TestUnitTestPlanDataSource_Schema`.
5. Registered `testplan.NewTestPlanResource` in `Resources()` and `testplan.NewTestPlanDataSource` in `DataSources()` in `framework_provider.go`.

## What worked

- `testplan.NewClient()` returns `Client` (no error) — do NOT use `err :=` pattern.
- Use inline plan modifier types inside the same file — `stringplanmodifier` subpackage is not vendored.
- `azuredevops.WrappedError` has a `StatusCode *int` field to detect 404.
- `converter.Int()` exists for `*int` values.
- Build tags: `//go:build all || resource_test_plan` (single tag covers both resource and datasource tests in same package).

## What didn't work

- `stringplanmodifier.RequiresReplace()` / `stringplanmodifier.UseStateForUnknown()` — package not vendored; use inline structs.
- Using `err :=` pattern with `testplan.NewClient()` — it returns no error.

## Open questions

_(none — all ACs satisfied)_

## Notes for reflection

- The `testplan` package being in `third_party` but not `vendor` is a pattern worth noting in the brain: when adding new SDK packages from third_party, always check vendor/modules.txt and vendor/ directory.
- Inline plan modifiers are necessary when framework subpackages (stringplanmodifier, etc.) are not vendored.
