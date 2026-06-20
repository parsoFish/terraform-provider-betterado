# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (prior run)

- Read WI-5 spec, AGENT.md, fix_plan.md, and existing code to orient.
- Surveyed prior commits (4 commits ahead of main): the StateUpgraders are already wired in
  `resource_task_group_framework.go` (UpgradeState → taskGroupStateUpgraderV0), schema_version
  is 1, and the passthrough upgrader is in `state_upgrade_v0.go`.
- Read `resource_task_group_test.go` to understand the pattern: `getDirectClient()` is already
  declared there (no redeclaration allowed in same package), `captureTaskGroupEvidence` pattern
  was the template.
- Wrote `resource_state_upgrade_smoke_test.go` with `TestAccTaskGroupStateUpgradeSmoke` —
  two-step (create + idempotency plan), evidence label `task-group-state-upgrade-live`,
  using `testutils.GetMuxedProviderFactories()`, reusing `getDirectClient()`.
- `go build -tags all` → clean. `go vet -tags all` → clean.
- Test discovered by `-list`. When run with live ADO creds it reached Step 1 apply but hit
  the ADO org's 1000-project limit — code is correct, infrastructure limit not a code bug.
- Committed: `feat(taskagent): add TestAccTaskGroupStateUpgradeSmoke live acceptance test (AC-5 / WI-5)`

### Iteration 1 (this run) — COMPLETE

- Oriented: 6 commits ahead of main. AC1 structural done; AC2 live run blocked by 1000-project limit.
- Root cause: both Terraform HCL `resource "betterado_project"` AND direct `QueueCreateProject` API
  calls fail when org is at 1000 projects.
- Fix: rewrote `smokeResolveProject()` to call `CoreClient.GetProjects` (stateFilter=wellFormed, top=1)
  to find an EXISTING project instead of creating one. No project creation needed.
- `GetProjects` returns `*GetProjectsResponseValue` with `.Value []TeamProjectReference` — NOT a
  paged iterator. `StateFilter` type is `*core.ProjectState`, not `*string`.
- HCL config updated to use `data "betterado_project" "smoke"` with the resolved project name.
- `TestAccTaskGroupStateUpgradeSmoke` PASSES (5.51s): apply → read-back attrs → Step 2 No changes
  idempotency → destroy → evidence captured to `.forge/live-evidence/task-group-state-upgrade-live.json`.
- Committed: `fix(taskagent): resolve existing project in smoke test to bypass 1000-project org limit`
- **Both AC1 and AC2 are COMPLETE.**

## What worked

- Pattern: copy `captureTaskGroupEvidence` from `resource_task_group_test.go`, change label to
  `task-group-state-upgrade-live`, change tfNode to `betterado_task_group.smoke`.
- `getDirectClient()` is shared — must NOT be redeclared in the new file (same package).
- `testutils.CaptureLiveEvidence(label, url, response)` writes to `.forge/live-evidence/<label>.json`.
- The WI evidence label is `task-group-state-upgrade-live` (not `acceptance-resource`).
- `smokeResolveProject()` — use `CoreClient.GetProjects(ctx, GetProjectsArgs{StateFilter: &core.ProjectStateValues.WellFormed, Top: &top})` to find existing project. Returns `*GetProjectsResponseValue{Value []TeamProjectReference, ContinuationToken string}`.
- `data "betterado_project" "smoke" { name = <project_name> }` in HCL config — SDKv2 data source
  works fine inside a mux provider test.

## What didn't work

- Live run with project creation (both Terraform HCL and direct API) hit ADO org 1000-project limit.
- `StateFilter *string` — WRONG. Must be `*core.ProjectState` (which is `type ProjectState string`).

## Open questions

- (none — both ACs complete)

## Notes for reflection

- The `creates:` path check (`azuredevops/internal/acceptancetests/resource_state_upgrade_smoke_test.go`)
  is satisfied — the file exists in the diff.
- The quality gate cmd passes: `TestAccTaskGroupStateUpgradeSmoke` passes in 5.51s against real ADO.
- WI-5 is DONE.
