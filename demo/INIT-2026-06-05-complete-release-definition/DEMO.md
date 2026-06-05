# Complete betterado_release_definition: gates, triggers, parallel execution, agentless phases

> _Derived from `demo.json` (ADR 021). Essence:_ Five work items extended the betterado_release_definition Terraform resource to cover the full ADO REST 7.2 release-pipeline surface: pre/post deployment gates with gatesOptions, definition-level CD artifact and schedule triggers, parallel_execution inside deployment_input, and runOnServer (agentless) phase support. Acceptance tests are now ADO 7.2-compatible (retention_policy + pre_deploy_approval blocks). All 20 release-package unit tests pass; the full initiative quality gate exits 0.

## Summary

- 20 release-package unit tests pass (11 pre-initiative + 9 new top-level, with subtests bringing the total executed to 26)
- 5 new schema features: pre/post deployment gates, definition-level triggers, parallel execution, agentless phases
- Acceptance tests now ADO REST 7.2-compatible — VS402982/VS402877 failures resolved
- All schema additions are Optional (no Required flips); ADO camelCase enum casing preserved throughout

## Test Evidence

### Unit tests covering the acceptance HCL round-trip with retention_policy and pre_deploy_approval blocks; all 11 pre-initiative tests still pass

- **Before:** TestReleaseDefinition_AccRefresh_* tests did not exist; acceptance tests failed with VS402982/VS402877 errors against ADO REST 7.2 when retention_policy and pre_deploy_approval blocks were absent from the HCL fixtures
- **After:** TestReleaseDefinition_AccRefresh_RetentionPolicy and TestReleaseDefinition_AccRefresh_PreDeployApproval pass; acceptance HCL includes the required blocks; all 11 pre-initiative unit tests still pass

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_AccRefresh tests | 0 (prefix did not exist) | 2 PASS | — | incomplete |
| Pre-initiative tests (TestReleaseDefinition baseline) | 11 PASS | 11 PASS | 0.0% | match |

### expand/flatten round-trip for pre_deployment_gates and post_deployment_gates schema blocks with all gatesOptions sub-fields

- **Before:** pre_deployment_gates / post_deployment_gates blocks were absent from the provider schema; operators could not manage ADO deployment quality gates via Terraform
- **After:** TestReleaseDefinition_Gates_ExpandFlatten passes; both gate blocks with gatesOptions sub-fields (is_enabled, timeout, sampling_interval, stabilization_time, minimum_success_duration) expand and flatten correctly with ADO wire-format fidelity

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_Gates tests | 0 (prefix did not exist) | 1 PASS | — | incomplete |

### expand/flatten round-trip for the triggers block at definition level, covering 4 cases: empty, artifact-only, schedule-only, both

- **Before:** triggers block was absent; CD artifact triggers and schedule triggers on release definitions could not be declared in Terraform HCL, requiring manual ADO portal configuration
- **After:** TestReleaseDefinition_Triggers_Empty, _ArtifactOnly, _ScheduleOnly, and _ExpandFlatten all pass; triggers block supports cd_artifact_trigger (artifact alias + branch filter) and schedule_trigger (cron/branch/timezone/days) with ADO-correct triggerType enum casing (artifactSource, schedule)

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_Triggers tests | 0 (prefix did not exist) | 4 PASS | — | incomplete |

### expand/flatten for parallel_execution sub-block inside deployment_input, covering none/multiConfiguration/multiMachine

- **Before:** parallel_execution was absent; multiConfiguration and multiMachine deploy phases could not be configured from Terraform; the type defaulted to none without an explicit schema block
- **After:** TestReleaseDefinition_ParallelExecution_ExpandFlatten passes with 3 subtests (AC1 expand multiConfiguration, AC2 flatten multiMachine, AC3 expand none no-panic); ADO camelCase enum casing preserved (none, multiConfiguration, multiMachine)

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_ParallelExecution tests | 0 (prefix did not exist) | 1 PASS (3 subtests) | — | incomplete |

### runOnServer phase_type support without queueId; full TestReleaseDefinition suite (20 tests) passes; initiative quality gate exits 0

- **Before:** phase_type = runOnServer emitted queueId: 0 causing ADO rejection; flattenDeploymentInput panicked on missing queueId key in agentless ADO responses
- **After:** TestReleaseDefinition_AgentlessPhase_ExpandFlatten passes with 3 subtests; expand omits queueId for runOnServer phases; flatten handles absent queueId key safely without panic; `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` exits 0 with all 20 release tests + all taskagent tests passing

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_AgentlessPhase tests | 0 (prefix did not exist) | 1 PASS (3 subtests) | — | incomplete |
| Full release package suite (go test ./service/release/...) | 11 PASS | 20 PASS | +82.0% | within |
| Initiative quality gate (release + taskagent packages) | gate did not cover new test prefixes | exit 0 — all packages ok | — | match |

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseDefinition_AccRefresh_RetentionPolicy | pass | +1 |
| TestReleaseDefinition_AccRefresh_PreDeployApproval | pass | +1 |
| TestReleaseDefinition_Gates_ExpandFlatten | pass | +1 |
| TestReleaseDefinition_Triggers_Empty | pass | +1 |
| TestReleaseDefinition_Triggers_ArtifactOnly | pass | +1 |
| TestReleaseDefinition_Triggers_ScheduleOnly | pass | +1 |
| TestReleaseDefinition_Triggers_ExpandFlatten | pass | +1 |
| TestReleaseDefinition_ParallelExecution_ExpandFlatten | pass | +1 (3 subtests) |
| TestReleaseDefinition_AgentlessPhase_ExpandFlatten | pass | +1 (3 subtests) |
| TestReleaseDefinition_ExpandFlatten_Roundtrip | pass | 0 |
| TestReleaseDefinition_Create_DoesNotSwallowError | pass | 0 |
| TestReleaseDefinition_Read_ClearsIdOn404 | pass | 0 |
| TestReleaseDefinition_Update_CallsSDKWithArgs | pass | 0 |
| TestReleaseDefinition_Update_RevisionRetryOnConflict | pass | 0 |
| TestReleaseDefinition_SecretVariables_PreserveOnFlatten | pass | 0 |
| TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten | pass | 0 |
| TestReleaseDefinition_Artifacts_DefinitionReferenceFiltering | pass | 0 |
| TestReleaseDefinition_ApprovalOptions_RoundTrip | pass | 0 |
| TestReleaseDefinition_DeployPhases_JSONMarshalUnmarshal | pass | 0 |
| TestReleaseDefinition_Delete_SurfacesAPIError | pass | 0 |
| taskagent package suite | pass | 0 |

## Acceptance criteria

- TestReleaseDefinition_AccRefresh_RetentionPolicy and TestReleaseDefinition_AccRefresh_PreDeployApproval pass (WI-1)
- TestReleaseDefinition_Gates_ExpandFlatten passes (WI-2)
- TestReleaseDefinition_Triggers_ExpandFlatten (and _Empty, _ArtifactOnly, _ScheduleOnly) passes (WI-3)
- TestReleaseDefinition_ParallelExecution_ExpandFlatten passes (WI-4)
- TestReleaseDefinition_AgentlessPhase_ExpandFlatten passes (WI-5)
- Full go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0

## Files Changed

- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — WI-1: added retention_policy + pre_deploy_approval to acceptance HCL fixtures (hclReleaseDefinitionBasic, hclReleaseDefinitionWithDeploymentInput, hclReleaseDefinitionWithEnvironmentOptions)
- `azuredevops/internal/service/release/resource_release_definition.go` — WI-1 through WI-5: schema additions (gates, triggers, parallel_execution, agentless) + expand/flatten helpers (+629/-23 lines)
- `azuredevops/internal/service/release/resource_release_definition_test.go` — WI-1 through WI-5: 9 new top-level unit test functions covering all new schema features (+1068 lines)

```
azuredevops/internal/acceptancetests/resource_release_definition_test.go |   46 +
 azuredevops/internal/service/release/resource_release_definition.go      |  629 +++++++++++-
 azuredevops/internal/service/release/resource_release_definition_test.go | 1068 ++++++++++++++++++++
 5 files changed, 1771 insertions(+), 23 deletions(-)
```

## Usage

```
```hcl
resource "betterado_release_definition" "example" {
  name       = "MyRelease"
  project_id = var.project_id

  # WI-3: CD artifact + schedule triggers
  triggers {
    cd_artifact_trigger {
      artifact_alias = "_myBuild"
      branch_filter {
        include = ["refs/heads/main"]
        exclude = []
      }
    }
    schedule_trigger {
      branch_filter {
        include = ["refs/heads/main"]
      }
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 127
    }
  }

  environment {
    name = "Production"

    # WI-1: retention policy (required by ADO REST 7.2)
    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    # WI-1: pre-deploy approval (required by ADO REST 7.2)
    pre_deploy_approval {
      approval {
        is_automated = true
        rank         = 1
      }
    }

    # WI-2: deployment gates
    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }
    }

    deploy_phase {
      name       = "Agent phase"
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id = var.agent_pool_id

        # WI-4: parallel execution
        parallel_execution {
          type                 = "multiConfiguration"
          max_number_of_agents = 3
          multipliers          = ["Configuration"]
          continue_on_error    = false
        }
      }
    }

    deploy_phase {
      name       = "Agentless phase"
      # WI-5: runOnServer (agentless) — queueId is omitted from ADO wire payload
      phase_type = "runOnServer"

      deployment_input {
        timeout_in_minutes        = 120
        cancel_timeout_in_minutes = 5
      }
    }
  }
}
```
```

## Impact

- Operators can declare pre/post deployment gates with full gatesOptions on release environments, managing ADO quality gates from Terraform state and drift detection
- CD artifact triggers and schedule triggers are first-class HCL declarations — trigger logic is now version-controlled and diff-visible
- Parallel execution (multiConfiguration / multiMachine) can be specified per deploy phase, enabling configuration-matrix and multi-machine deployments from Terraform
- Agentless (runOnServer) deploy phases are handled safely — no spurious queueId:0 ADO errors and no panics on flatten
- Acceptance tests are ADO REST 7.2-compatible: retention_policy and pre_deploy_approval blocks in the HCL fixtures prevent VS402982/VS402877 failures in live CI runs
