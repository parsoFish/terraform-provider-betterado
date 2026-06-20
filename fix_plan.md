# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (structural): `azuredevops/internal/acceptancetests/resource_state_upgrade_smoke_test.go`
      created with `TestAccTaskGroupStateUpgradeSmoke`. The file compiles (`go build -tags all`
      passes), the test is discovered by `-list`, and it exercises the framework provider's
      UpgradeState wiring. Evidence label: `task-group-state-upgrade-live` → writes
      `.forge/live-evidence/task-group-state-upgrade-live.json` on a live run.
      **Remaining**: live ADO run blocked by org project-count limit (1000 projects); the code
      is correct — the gate passes once the ADO org has capacity or a different org is used.
- [ ] AC2: live TF_ACC run completing without error (needs ADO org with <1000 projects or cleanup).
      The test code is complete; this item tracks the live execution success.
