# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (current)

- **AC1** was already complete: `resource_build_definition_framework.go` exists with full schema, `resource_build_definition_framework_test.go` has `TestBuildDefinitionFramework_Schema` which passes.
- **AC2** was missing: `azuredevops/internal/acceptancetests/resource_build_definition_framework_test.go` did not exist.
- Created the acceptance test file with `TestAccBuildDefinition_Framework_basic`:
  - Uses `SharedReleaseFixture` (no project creation — org is at 1000-project cap).
  - `//go:build all || resource_build_definition_framework` tag.
  - Steps: apply with TfsGit repo + non-default path + one variable → read-back assertions → idempotency re-plan (ExpectNonEmptyPlan: false).
  - `captureBuildDefinitionFrameworkEvidence` calls `testutils.CaptureLiveEvidence("acceptance-resource", url, defResp)`.
  - Uses `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()`.
- File compiles cleanly; offline run skips without TF_ACC.

## What worked

- Pattern from `data_build_definition_framework_test.go` — same `SharedReleaseFixture` + `fixture.RepoID` for TfsGit repo, same `captureBuild*Evidence` helper shape.
- The `getDirectClient()` helper is available in the `acceptancetests` package for live evidence capture.
- `captureBuildDefinitionFrameworkEvidence` follows the existing `captureBuildFolderFrameworkEvidence` / `captureBuildDefinitionFrameworkDSEvidence` pattern exactly.

## What didn't work

_(nothing hit dead ends this iteration)_

## Open questions

- AC2 live pass requires `TF_ACC=1` + `AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN` — the orchestrator runs this in the live serve env.

## Notes for reflection

- Two pre-existing SDKv2 mock test failures (`TestBuildDefinition_Create_DoesNotSwallowError`, `TestBuildDefinition_Update_DoesNotSwallowError`) in the build package exist on main and are unrelated to this WI.
- Both ACs are now satisfied; WI-3 is complete pending live gate run.
