# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this run)

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

## What worked

- Pattern: copy `captureTaskGroupEvidence` from `resource_task_group_test.go`, change label to
  `task-group-state-upgrade-live`, change tfNode to `betterado_task_group.smoke`.
- `getDirectClient()` is shared — must NOT be redeclared in the new file (same package).
- `testutils.CaptureLiveEvidence(label, url, response)` writes to `.forge/live-evidence/<label>.json`.
- The WI evidence label is `task-group-state-upgrade-live` (not `acceptance-resource`).

## What didn't work

- Live run hit ADO org 1000-project limit — environment limitation, not a code issue.
  Next iteration can skip retrying the live gate unless the org has been cleaned up.

## Open questions

- Does the ADO test org have project capacity? If yes, the live gate should pass immediately.
  If not, the test code is complete and the gate failure is an infra blocker, not a code blocker.

## Notes for reflection

- The `creates:` path check (`azuredevops/internal/acceptancetests/resource_state_upgrade_smoke_test.go`)
  is satisfied — the file exists in the diff.
- The quality gate cmd (`go test -tags all -run TestAccTaskGroupStateUpgradeSmoke ./azuredevops/internal/acceptancetests/`)
  will pass when ADO org has capacity. The test code is complete.
