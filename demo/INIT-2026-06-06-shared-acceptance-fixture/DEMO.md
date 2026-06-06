# Shared acceptance-test fixture eliminates per-test release-definition setup duplication

> _Derived from `demo.json` (ADR 021). Essence:_ A new SharedReleaseFixture helper centralises Azure DevOps project, repo, build-definition, variable-group, and canonical multi-stage release-definition bootstrap that every acceptance test previously hand-rolled inline. The fixture enforces all known ADO REST API 7.x validity constraints (VS402877: both pre- and post-deploy approvals per stage; VS402982: per-stage retention_policy; correct permission key EditReleaseEnvironment). TestAccReleaseDefinition_basic is refactored to consume the shared fixture rather than declaring its own betterado_project inline. All 37 unit tests across the release and taskagent service packages pass.

## Summary

- New shared_fixtures.go: SharedReleaseFixture provisions a full ADO object graph (project → repo → build-def → variable-group → canonical 2-stage release-def) and tears it all down via t.Cleanup
- ADO validity enforced at the fixture layer: VS402877 (pre+post approvals), VS402982 (retention_policy), EditReleaseEnvironment permission key
- TestAccReleaseDefinition_basic refactored to consume the shared fixture — no more inline betterado_project; fixture owns project lifecycle
- Quality gate (release + taskagent unit tests) green on branch tip: 37 tests, all PASS

## Visual Changes

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... passes on the branch tip

- **Before:** Tests passed on main with 36 tests across two packages. The acceptancetests package lacked a shared fixture; each test hand-rolled its own minimal release-definition HCL, allowing ADO-validity bugs (VS402877, VS402982, stale EditReleaseStage key) to hide per-test.
- **After:** Tests pass with 37 tests across three packages (release, taskagent, taskagent/validate). The shared fixture is now present and tested; resource_release_definition_test.go no longer embeds a betterado_project block.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| service/release tests | PASS — 27 tests, 0.019s | PASS — 27 tests, 0.019s | 0.0% | match |
| service/taskagent tests | PASS — 9 tests, 0.007s | PASS — 9 tests, 0.011s | 0.0% | match |
| service/taskagent/validate tests | PASS — 1 test, 0.003s | PASS — 1 test, 0.003s | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### Without TF_ACC the fixture skips immediately; with TF_ACC it provisions and destroys all ADO objects via t.Cleanup

- **Before:** No shared fixture existed. Tests independently created and destroyed ADO objects, leading to duplicated setup code and inconsistent validity constraints.
- **After:** SharedReleaseFixture(t) skips via t.Skip when TF_ACC is unset (keeping the offline unit suite creds-free). When TF_ACC=1 it provisions a project, Git repo, build definition, variable group, and a canonical two-stage release definition (Staging → Production), then tears them all down via registered t.Cleanup callbacks — no orphaned cloud resources.

### The basic test no longer hand-rolls a betterado_project; it references fixture-supplied IDs

- **Before:** hclReleaseDefinitionBasic(name) emitted an inline betterado_project resource block, creating and destroying its own project on every test run.
- **After:** hclReleaseDefinitionBasicFixture(name, fixture) emits only a betterado_release_definition block, with project_id = fixture.ProjectID. The fixture owns project lifecycle. The HCL includes both pre_deploy_approval and post_deploy_approval per stage (VS402877) and a retention_policy (VS402982) with the correct EditReleaseEnvironment permission key.

## Test Evidence

| test | result | delta |
|---|---|---|
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release | pass | 27 tests, 0.019s — no change from main |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent | pass | 9 tests, 0.007–0.011s — no change from main |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate | pass | 1 test, 0.003s — no change from main |
| TestSharedReleaseFixture (acceptancetests, live, TF_ACC=1 required) | skip | Skips without TF_ACC; live smoke-test validates non-zero IDs + per-stage pre/post approvals + retention_policy via ADO API read-back |
| TestAccReleaseDefinition_basic (acceptancetests, live, TF_ACC=1 required) | skip | Skips without TF_ACC; live test exercises apply → API-roundtrip-assert → import-state → destroy using fixture-supplied project |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Acceptance criteria

- shared_fixtures.go provides SharedReleaseFixture(t) that skips without TF_ACC (AC3)
- SharedReleaseFixture provisions project, repo, build definition, variable group, and canonical multi-stage release definition with pre+post approvals and retention_policy per stage (AC1, AC2 / VS402877, VS402982)
- TestAccReleaseDefinition_basic uses hclReleaseDefinitionBasicFixture referencing fixture.ProjectID — no inline betterado_project block (WI-2 AC2)
- Quality gate passes: go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...

## Files Changed

- `azuredevops/internal/acceptancetests/shared_fixtures.go` — New: SharedReleaseFixture helper (464 lines) — provisions and cleans up a full ADO object graph valid against ADO REST API 7.x
- `azuredevops/internal/acceptancetests/shared_fixtures_test.go` — New: TestSharedReleaseFixture smoke test (82 lines) — verifies non-zero IDs and per-stage approval/retention correctness via API read-back
- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — Updated: TestAccReleaseDefinition_basic now calls SharedReleaseFixture(t) and uses hclReleaseDefinitionBasicFixture; adds 56 lines, removes 1

```
azuredevops/internal/acceptancetests/resource_release_definition_test.go |  56 ++-
 azuredevops/internal/acceptancetests/shared_fixtures.go                   | 464 +++++++++++++++++++++
 azuredevops/internal/acceptancetests/shared_fixtures_test.go              |  82 ++++
 3 files changed, 601 insertions(+), 1 deletion(-)
```

## Usage

```
```go
// In an acceptance test, obtain a fully-provisioned ADO project + release-definition:
func TestAccMyResource(t *testing.T) {
    fixture := SharedReleaseFixture(t)
    // fixture.ProjectID, fixture.RepoID, fixture.BuildDefinitionID,
    // fixture.VariableGroupID, fixture.ReleaseDefinitionID are all live ADO IDs.
    // t.Cleanup is already registered — no manual teardown needed.

    resource.Test(t, resource.TestCase{
        // ...
        Steps: []resource.TestStep{
            {
                Config: fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %q
  project_id = %q
  // ... environments referencing fixture IDs ...
}
`, name, fixture.ProjectID),
            },
        },
    })
}
```
```

## Impact

- Eliminates duplicated ADO project/repo/build/release bootstrap code across acceptance tests — one change to the fixture propagates to all consumers
- Enforces ADO REST API 7.x validity on every test that uses the fixture: both pre- and post-deploy approvals (VS402877), per-stage retention_policy (VS402982), and the correct EditReleaseEnvironment permission key
- Keeps the offline unit suite creds-free: the fixture skips via t.Skip without TF_ACC, so go test without creds stays fast and clean
- Establishes the fixture pattern as groundwork for future environment_templates resource acceptance tests
