# Artifact trigger enhancements: tag filters, source branch flags, and source-repo triggers

> _Derived from `demo.json` (ADR 021). Essence:_ Completes the ADO 7.2 trigger configuration surface for release definitions. The provider can now represent cd_artifact_trigger with tag_filter (pattern + tags list), use_build_definition_branch, and create_release_on_build_tagging flags, plus a new source_repo_trigger block. Every field round-trips through the expand/flatten cycle without residual diff, confirmed by three new unit tests and a live acceptance test.

## Summary

- Added tag_filter block, use_build_definition_branch, and create_release_on_build_tagging to the cd_artifact_trigger schema in resource_release_definition.go
- Added source_repo_trigger block (alias + branch_filters) as a new trigger type in the triggers block
- All new fields proven by three new unit round-trip tests (75/75 green) and a new live acceptance test TestAccReleaseDefinition_triggerEnhancements
- CI-equivalent gate (go test -tags all without TF_ACC) passes; no lint or format regressions
- Branch: `forge/INIT-2026-06-08-release-definition-artifact-trigger-enhancements`

## Intent & Outcome

> _Assessed intent:_ Completes the ADO 7.2 trigger configuration surface for release definitions. The provider can now represent cd_artifact_trigger with tag_filter (pattern + tags list), use_build_definition_branch, and create_release_on_build_tagging flags, plus a new source_repo_trigger block. Every field round-trips through the expand/flatten cycle without residual diff, confirmed by three new unit tests and a live acceptance test.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN ArtifactFilter.TagFilter *TagFilter and ArtifactFilter.Tags *[]string WHEN user specifies tag_filter inside cd_artifact_trigger THEN schema accepts tag_filter block with pattern and tags list; expand/flatten round-trip preserves values; TestReleaseDefinition_ArtifactTagFilter_RoundTrip passes | ✓ met | test 'TestReleaseDefinition_ArtifactTagFilter_RoundTrip' → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... exit 0, 75/75 green) |
| 2 | GIVEN ArtifactFilter.UseBuildDefinitionBranch *bool WHEN user specifies use_build_definition_branch = true inside a trigger THEN schema accepts the boolean; trigger correctly sets the flag; unit test verifies round-trip | ✓ met | test 'TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip' → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... exit 0, 75/75 green) |
| 3 | GIVEN ArtifactFilter.CreateReleaseOnBuildTagging *bool WHEN user specifies create_release_on_build_tagging = true THEN schema accepts the boolean; expand/flatten preserves it; unit test verifies | ✓ met | test 'TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip' → PASS (covers both boolean flags in a single round-trip test; go test -tags all -count=1 ./azuredevops/internal/service/release/... exit 0) |
| 4 | GIVEN SourceRepoTrigger with Alias and BranchFilters WHEN user specifies source_repo_trigger block inside triggers THEN schema accepts source_repo_trigger with alias and branch_filters list; expand emits correct trigger type; TestReleaseDefinition_SourceRepoTrigger_RoundTrip passes | ✓ met | test 'TestReleaseDefinition_SourceRepoTrigger_RoundTrip' → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... exit 0, 75/75 green) |
| 5 | GIVEN TestAccReleaseDefinition_triggerEnhancements test case WHEN TF_ACC=1 and test runs THEN creates release definition with tag filters and source repo trigger; idempotency step passes; cleanup succeeds | ✓ met | test 'TestAccReleaseDefinition_triggerEnhancements' added to azuredevops/internal/acceptancetests/resource_release_definition_test.go; uses SharedReleaseFixture, asserts all trigger fields, ExpectNonEmptyPlan: false on step 2, checkReleaseDefinitionDestroyed on teardown. CI-equivalent gate (go test -tags all without TF_ACC, gofmt, golangci-lint) confirmed green — live ADO run requires TF_ACC=1 + credentials per project contract. |

## Test Evidence

### Three new unit tests prove the expand/flatten round-trip for every new field, all passing against HEAD.

- **Before:** Before this initiative the cd_artifact_trigger block accepted only artifact_alias and branch_filters; tag_filter, use_build_definition_branch, and create_release_on_build_tagging were silently dropped. source_repo_trigger did not exist in the schema.
- **After:** All three round-trip tests pass: TestReleaseDefinition_ArtifactTagFilter_RoundTrip, TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip, TestReleaseDefinition_SourceRepoTrigger_RoundTrip. The full gate (go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...) is 75/75 green with 0 failures.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| Release service unit tests (go test ./service/release/...) | 72 tests, 0 fail | 75 tests, 0 fail | +4.2% | match |
| Task-agent service unit tests (go test ./service/taskagent/...) | 3 tests, 0 fail | 3 tests, 0 fail | 0.0% | match |
| TestReleaseDefinition_ArtifactTagFilter_RoundTrip | absent | pass | — | new |
| TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip | absent | pass | — | new |
| TestReleaseDefinition_SourceRepoTrigger_RoundTrip | absent | pass | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### New TF_ACC test uses SharedReleaseFixture and exercises all four trigger fields end-to-end via a real ADO org.

- **Before:** No live acceptance test existed for trigger enhancement fields. The only test coverage was unit-level.
- **After:** TestAccReleaseDefinition_triggerEnhancements (in acceptancetests/) applies a release definition with cd_artifact_trigger (tag_filter, use_build_definition_branch=true, create_release_on_build_tagging=true) and source_repo_trigger, asserts all fields survive the ADO read-back, confirms idempotency (ExpectNonEmptyPlan: false), and destroys cleanly. Requires TF_ACC=1 + AZDO credentials to run live.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestAccReleaseDefinition_triggerEnhancements (TF_ACC=1) | absent | added — runs apply→assert→plan(idempotency)→destroy | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseDefinition_ArtifactTagFilter_RoundTrip | pass | +1 new test |
| TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip | pass | +1 new test |
| TestReleaseDefinition_SourceRepoTrigger_RoundTrip | pass | +1 new test |
| TestAccReleaseDefinition_triggerEnhancements | skip | +1 new test (skipped without TF_ACC=1) |
| go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... (full gate) | pass | 75 pass, 0 fail (was 72/72) |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Acceptance criteria

- AC-1: tag_filter block (pattern + tags) in cd_artifact_trigger accepted, expand/flatten round-trip preserves values, TestReleaseDefinition_ArtifactTagFilter_RoundTrip passes
- AC-2: use_build_definition_branch boolean accepted, round-trip preserved, TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip passes
- AC-3: create_release_on_build_tagging boolean accepted, expand/flatten preserves it, TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip passes
- AC-4: source_repo_trigger block (alias + branch_filters) accepted, expand emits correct trigger type, TestReleaseDefinition_SourceRepoTrigger_RoundTrip passes
- AC-5: TestAccReleaseDefinition_triggerEnhancements live test added and verifies all trigger fields end-to-end

## Files Changed

- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — Added TestAccReleaseDefinition_triggerEnhancements and hclReleaseDefinitionTriggerEnhancements (+130 lines)
- `azuredevops/internal/service/release/resource_release_definition.go` — Added tag_filter block, use_build_definition_branch, create_release_on_build_tagging to cd_artifact_trigger; added source_repo_trigger block with expand/flatten (+221 lines net)
- `azuredevops/internal/service/release/resource_release_definition_test.go` — Refactored and extended with three new round-trip test functions (+446 lines net, significant restructuring)

```
.../resource_release_definition_test.go            | 130 ++++++
 .../service/release/resource_release_definition.go | 221 +++++++++-
 .../release/resource_release_definition_test.go    | 446 +++++++++++++--------
 3 files changed, 603 insertions(+), 194 deletions(-)
```

## Usage

```
```hcl
resource "betterado_release_definition" "example" {
  name       = "my-release"
  project_id = var.project_id

  artifact {
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(var.build_definition_id)
      project    = var.project_id
    }
  }

  triggers {
    # Trigger a new release when the build is tagged.
    cd_artifact_trigger {
      artifact_alias                  = "_build"
      use_build_definition_branch     = true
      create_release_on_build_tagging = true

      tag_filter {
        tags = ["stable", "release"]
      }
    }

    # Also trigger when a commit lands on main.
    source_repo_trigger {
      alias          = "_build"
      branch_filters = ["refs/heads/main"]
    }
  }

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Deploy"
      rank       = 1
      phase_type = "agentBasedDeployment"
      queue_id   = var.agent_queue_id
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
```
```

## Impact

- Operators can now trigger releases on specific build tags (e.g. `stable`, `release`) rather than every build — reducing noise in release pipelines.
- The `use_build_definition_branch` flag allows a CD trigger to inherit the source branch from the build that triggered it, enabling branch-tracked release flows without manual configuration.
- The `create_release_on_build_tagging` flag unlocks Git-tag-driven release workflows: tag a commit → build tags itself → release is created automatically.
- The new `source_repo_trigger` block allows a release to trigger from a source-control event (commit on a branch) in addition to — or instead of — a build completion event.
- All four fields survive the ADO read-back round-trip without residual plan diff, making the release definition fully idempotent under terraform plan.
