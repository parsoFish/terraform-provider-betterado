# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (structural): `azuredevops/internal/acceptancetests/resource_state_upgrade_smoke_test.go`
      created with `TestAccTaskGroupStateUpgradeSmoke`. The file compiles (`go build -tags all`
      passes), the test is discovered by `-list`, and it exercises the framework provider's
      UpgradeState wiring. Evidence label: `task-group-state-upgrade-live` → writes
      `.forge/live-evidence/task-group-state-upgrade-live.json` on a live run.
      **COMPLETE**: live run passes. Evidence written.

- [x] AC2: live TF_ACC run completing without error.
      **COMPLETE** (iteration 1): `TestAccTaskGroupStateUpgradeSmoke` passes — apply, read-back
      assertions, idempotency plan (No changes), destroy all pass. Evidence captured to
      `.forge/live-evidence/task-group-state-upgrade-live.json` with real ADO API response.

## How AC2 was unblocked

The test was rewritten to use `smokeResolveProject()` which auto-discovers an existing
wellFormed ADO project via `GetProjects` API (no project creation needed). This bypasses the
org's 1000-project limit that blocked iteration 0. Overridable via `AZDO_TEST_EXISTING_PROJECT`.
