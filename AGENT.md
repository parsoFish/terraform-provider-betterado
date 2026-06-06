# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (current)

**Goal:** Refactor `TestAccReleaseDefinition_basic` to consume `SharedReleaseFixture` from WI-1.

**What was done:**
1. Read `WI-2.md`, `resource_release_definition_test.go`, `shared_fixtures.go`, `shared_fixtures_test.go`.
2. `shared_fixtures.go` (WI-1) already exists with `SharedReleaseFixture(t *testing.T) SharedFixtureResult` — provides `ProjectID`, `RepoID`, `BuildDefinitionID`, `VariableGroupID`, `ReleaseDefinitionID`.
3. Refactored `TestAccReleaseDefinition_basic` to call `SharedReleaseFixture(t)` at the top of the test.
4. Added `hclReleaseDefinitionBasicFixture(name string, fixture SharedFixtureResult)` — emits only a `betterado_release_definition` block; `project_id` is set to `fixture.ProjectID` (hard-coded UUID string, no `betterado_project` resource).
5. Left `hclReleaseDefinitionBasic(name string)` unchanged for `TestAccReleaseDefinition_update` and other tests (per WI-2 constraint: scope is exactly `TestAccReleaseDefinition_basic`).
6. Ran `go build -tags all ./...` (clean) and `go vet -tags all ./azuredevops/internal/acceptancetests/` (clean).
7. Ran `go test -tags all -count=1 -run TestAccReleaseDefinition_basic ./azuredevops/internal/acceptancetests/` — exits 0 (skips correctly without TF_ACC).
8. Committed as `feat: refactor TestAccReleaseDefinition_basic to consume SharedReleaseFixture` (8c97ba95).

## What worked

- **Fixture pattern:** `SharedReleaseFixture(t)` called before `resource.ParallelTest` — the fixture sets up via ADO SDK; `t.Cleanup` handles teardown; the Terraform test lifecycle only manages the release definition itself.
- **HCL approach:** New HCL function takes `(name string, fixture SharedFixtureResult)` and uses `%[2]q` for the project UUID — cleanest approach avoiding any `betterado_project` reference in the HCL.
- **Preserving old function:** Keeping `hclReleaseDefinitionBasic(name)` unchanged avoids touching `_update`, `_withDeploymentInput`, `_withApprovalOptions`, `_withEnvironmentOptions` tests (WI-2 constraint: don't refactor all tests).

## What didn't work

_(none in this iteration)_

## Open questions

- AC3 (live TF_ACC run) verifies no ADO errors / no orphans — this requires the orchestrator to run the gate with `TF_ACC=1`. The structural changes satisfy VS402877 (pre+post approvals) and VS402982 (retention_policy) in the HCL. Live verification is the orchestrator's job.

## Notes for reflection

- The `SharedFixtureResult.ProjectID` is a UUID string — using it directly as `project_id` in HCL (string-interpolated) is correct; the provider accepts UUID string for project_id.
- `TestAccReleaseDefinition_basic` now no longer creates/destroys an ADO project itself — the fixture owns the project lifecycle. This is the "clean destroy" pattern.
- `hclReleaseDefinitionBasicFixture` is intentionally kept separate from `hclReleaseDefinitionBasic` to avoid breaking other tests. A future cleanup could unify them, but that's out of scope for WI-2.
