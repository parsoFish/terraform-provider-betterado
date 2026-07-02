# Agent Memory — UWI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (2026-07-02) — commit c60f0970

**Context:** No prior iterations; starting fresh on iteration 0. No `.forge/last-gate-failure.md` existed.

**Problem:** Commit 28580b96 introduced a `t.Skipf` in `preCheckDashboard` that silently skipped all dashboard tests when the fixture project was missing. This means the gate passed with zero test runs (a false pass). The WI requires reverting this to fail-loud behavior.

**Changes made:**
1. `azuredevops/internal/acceptancetests/resource_dashboard_test.go`:
   - Replaced inline `t.Skipf` block in `preCheckDashboard` with a call to `resolveOrCreateFixtureProject(t, clients)` — the shared fail-loud resolver that calls `t.Fatalf` if the fixture project is missing
   - Removed unused `"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"` import
   - Changed `CaptureLiveEvidence` label from `"acceptance-resource"` to `"dashboard-acceptance-resource"` so dashboard and extension evidence use distinct labels and both files survive in `.forge/live-evidence/`

2. `CHANGELOG.md`:
   - Updated dashboard entry to accurately describe live verification via betterado-standing-demo fixture project, noting the `dashboard-acceptance-resource` label

3. `forge/history/INIT-2026-07-01-migrate-framework-dashboard-extension/demo/demo.json`:
   - Added `dashboard-acceptance-resource` checkpoint entry alongside the existing `acceptance-resource` extension checkpoint — proves both labels survive together once the live gate runs

**Quality gate:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → all green.

## What worked

- `resolveOrCreateFixtureProject(t, clients)` is the correct fail-loud helper in `shared_fixtures.go` (line 324). It already calls `t.Fatalf` with a clear message instructing to restore from the ADO recycle bin.
- The `core` import was ONLY used in the old `t.Skipf` block's `GetProject` call — safe to remove once that block is replaced.
- The dashboard evidence label collision was straightforward to fix: change the label string from `"acceptance-resource"` to `"dashboard-acceptance-resource"` in `tryCaptureDashboardEvidence`.

## What didn't work

_(none to record for iteration 0)_

## Open questions

- AC2 requires the live gate to actually pass (TF_ACC=1). The code is correctly set up; the live gate will determine if the tests pass or if there are dashboard resource runtime issues.
- If the live gate still fails after this iteration, the failure output will indicate whether it's a fixture lookup error (project still missing) or a test assertion error (resource behavior problem).

## Notes for reflection

- The pattern: `t.Skipf` = silent false pass, `t.Fatalf` = loud failure that blocks promotion. Always prefer `t.Fatalf` for infrastructure preconditions.
- Evidence label uniqueness: when multiple resources share `CaptureLiveEvidence`, they must use distinct labels or writes overwrite each other.
