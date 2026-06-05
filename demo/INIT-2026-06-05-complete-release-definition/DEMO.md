# Complete betterado_release_definition: gates+tasks, idempotency, agent_specification, real gate queryId, exhaustive live acceptance test

> _Derived from `demo.json` (ADR 021). Essence:_ Nine work items completed the betterado_release_definition Terraform resource to cover the full ADO REST 7.2 release-pipeline surface. WI-1–5 (prior pass) added acceptance-test refresh, deployment gates schema, triggers, parallel execution, and agentless phases. WI-6 fixed the #1 live-review gap: deployment gates had zero actual gate checks (only timing options) — now each gate block carries workflow task(s), so gates#>0 live. WI-7 fixed three flatten round-trip bugs that caused a perpetual diff (multipliers comma-string, spurious empty parallel_execution block, schedule_trigger branch_filter location). WI-8 added TestAccReleaseDefinition_complete — an exhaustive live acceptance test setting a non-default value for every schema option. WI-9 (final cleanup) removed the stale branch_filter from schedule_trigger schema, set agent_specification in the acceptance test to a real non-default value (ubuntu-22.04), and created a real shared work-item query via betterado_workitemquery for the gate task queryId — so the gate is complete and the live acceptance test is fully idempotent.

## Summary

- 9 WIs completed: acceptance-test refresh, gates schema + tasks, triggers, parallel execution, agentless phases, idempotency fixes, exhaustive live acceptance test, and final cleanup (WI-9)
- 3 flatten round-trip bugs fixed (perpetual diff eliminated): multipliers comma-string, spurious empty parallel_execution block, schedule_trigger branch_filter removed entirely from schema
- TestAccReleaseDefinition_complete (WI-8+WI-9): exhaustive live acceptance test with agent_specification=ubuntu-22.04, real gate queryId via betterado_workitemquery, and full idempotency (ExpectNonEmptyPlan:false)
- 22 top-level offline unit tests pass; live acceptance test passed in-cycle against real ADO org

## Test Evidence

### Five WIs added the full ADO 7.2 schema surface; 20 unit tests pass; initiative quality gate exits 0 after the first pass

- **Before:** betterado_release_definition lacked pre/post_deployment_gates, triggers, parallel_execution, and runOnServer phase support; acceptance tests failed with VS402982/VS402877 against ADO REST 7.2 due to missing retention_policy and pre_deploy_approval
- **After:** Gates (gatesOptions), cd_artifact_trigger + schedule_trigger, multiConfiguration/multiMachine parallel_execution, and runOnServer agentless phase are fully modelled; acceptance test HCL is ADO 7.2-compatible; 20 release-package unit tests pass; quality gate exits 0

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| Release package unit tests | 11 PASS | 22 PASS | +100.0% | within |
| Initiative quality gate (release + taskagent packages) | gate did not cover new test prefixes | exit 0 — all packages ok | — | match |

### deploymentGatesSchema() gains a repeatable gate{} block each with a task{} list; expand populates ReleaseDefinitionGatesStep.Gates; flatten round-trips them; new TestReleaseDefinition_GatesTasks_ExpandFlatten passes

- **Before:** Live ADO API showed gates#=0 — pre/post_deployment_gates persisted gatesOptions (timing) only; no actual gate checks were modelled; gates were present in HCL but had zero effect at deploy time
- **After:** Each gate{} block carries workflow task(s) via workflowTaskSchema(); expandDeploymentGates() populates ReleaseDefinitionGatesStep.Gates; flattenDeploymentGates() round-trips them back; TestReleaseDefinition_GatesTasks_ExpandFlatten exercises a serverGate task and verifies no drops in the round-trip; live acceptance test (WI-8) confirms gates#>0 via checkReleaseDefinitionHasGates()

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_GatesTasks_ExpandFlatten | 0 (prefix did not exist) | 1 PASS | — | incomplete |

### Three flatten bugs caused a perpetual diff after every apply; all three fixed; TestReleaseDefinition_RoundTrip (4 subtests) proves flatten(expand(x))==x for every affected field

- **Before:** A second `terraform plan` reported 1 to change due to three bugs: (1) parallel_execution.multipliers: ADO SDK has Multipliers *string (comma-separated) not []string — flatten got the string back but expand sent a list, so ADO stored 'TargetSlot,Production' while state had ['TargetSlot','Production']; (2) a spurious empty parallel_execution block was emitted for non-parallel/agentless phases causing ADO to want to remove it every plan; (3) schedule_trigger.branch_filter.include was nested inside the schedule sub-object (which has no BranchFilters field in ADO) — it was silently dropped by ADO and re-added by Terraform on every plan
- **After:** Bug 1 fixed: expandParallelExecution joins multipliers with strings.Join so ADO persists the comma string; flattenParallelExecution handles both *string and []interface{} via type-switch. Bug 2 fixed: flattenParallelExecution returns nil for type:none; flattenDeploymentInput only sets the key when non-nil. Bug 3 fixed: branch_filter entirely removed from schedule_trigger schema (WI-9 final cleanup: ADO classic schedule triggers have no branch filter concept; the field was always silently dropped). TestReleaseDefinition_RoundTrip passes with 4 subtests asserting flatten(expand(x))==x for each case.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseDefinition_RoundTrip (4 subtests) | 0 (prefix did not exist) | 1 PASS (4 subtests: multipliers_comma_string, multipliers_array, no_parallel_execution, schedule_trigger_no_branch_filter_no_residual_diff) | — | incomplete |
| Perpetual diff (terraform plan post-apply) | 1 to change (three round-trip bugs) | No changes — infrastructure matches configuration | — | match |

### TF_ACC=1 TestAccReleaseDefinition_complete applies a release definition with non-default values for every schema option; WI-9 added agent_specification=ubuntu-22.04, real gate queryId via betterado_workitemquery, and removed stale branch_filter assertions from unit tests

- **Before:** WI-8 introduced the exhaustive acceptance test but had three remaining gaps found in live review: agent_specification was never set (unset image), gate queryId was empty string (violates WI-9 AC), and schedule_trigger unit tests had stale branch_filter assertions that conflicted with the already-correct schema removal
- **After:** WI-9 closes all three: agent_specification = 'ubuntu-22.04' set in deployment_input; betterado_workitemquery resource creates a real 'All Work Items - Gate Check' query under Shared Queries in the test project, its .id used as queryId for both pre/post gate tasks; unit tests updated to assert schedule_trigger does NOT expose branch_filter in flattened state. checkReleaseDefinitionAgentSpecification() API-level check verifies agentSpecification.identifier persisted. ExpectNonEmptyPlan:false confirms full idempotency.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestAccReleaseDefinition_complete (live ADO) | 1 PASS (WI-8, but agent_specification unset and gate queryId empty) | 1 PASS (WI-9: agent_specification=ubuntu-22.04, real gate queryId, no schedule_trigger branch_filter diff) | — | match |
| agent_specification in live ADO resource | not set (default agent image) | ubuntu-22.04 (asserted by checkReleaseDefinitionAgentSpecification) | — | match |
| Gate queryId on live resource | empty string (gate task incomplete) | real shared query GUID (betterado_workitemquery.gate_query.id) | — | match |
| Idempotency re-plan (ExpectNonEmptyPlan:false) | 1 to change (schedule_trigger branch_filter diff) | No changes (schedule_trigger has no branch_filter — fully idempotent) | — | match |

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0 with all 22 top-level test functions passing (11 pre-initiative + 11 new)

- **Before:** Quality gate covered 11 pre-initiative tests before this initiative started
- **After:** 22 top-level test functions pass: 11 original + TestReleaseDefinition_AccRefresh_RetentionPolicy, _PreDeployApproval (WI-1), _Gates_ExpandFlatten (WI-2), _Triggers_Empty/_ArtifactOnly/_ScheduleOnly/_ExpandFlatten (WI-3), _ParallelExecution_ExpandFlatten (WI-4), _AgentlessPhase_ExpandFlatten (WI-5), _GatesTasks_ExpandFlatten (WI-6), _RoundTrip (WI-7, updated by WI-9 to assert no schedule_trigger branch_filter diff); all taskagent tests pass

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| go test ./service/release/... (top-level test functions) | 11 PASS | 22 PASS (11 new, including subtests for RoundTrip/ParallelExecution/AgentlessPhase) | +100.0% | within |
| go test ./service/taskagent/... (regression) | all PASS | all PASS (exit 0) | 0.0% | match |

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseDefinition_AccRefresh_RetentionPolicy | pass | +1 (WI-1) |
| TestReleaseDefinition_AccRefresh_PreDeployApproval | pass | +1 (WI-1) |
| TestReleaseDefinition_Gates_ExpandFlatten | pass | +1 (WI-2) |
| TestReleaseDefinition_GatesTasks_ExpandFlatten | pass | +1 (WI-6) |
| TestReleaseDefinition_Triggers_Empty | pass | +1 (WI-3) |
| TestReleaseDefinition_Triggers_ArtifactOnly | pass | +1 (WI-3) |
| TestReleaseDefinition_Triggers_ScheduleOnly | pass | +1 (WI-3, updated by WI-9: asserts schedule_trigger has NO branch_filter in flattened state) |
| TestReleaseDefinition_Triggers_ExpandFlatten | pass | +1 (WI-3, updated by WI-9: stale branch_filter assertions removed) |
| TestReleaseDefinition_RoundTrip (4 subtests) | pass | +1 (WI-7 + WI-9: multipliers_comma_string, multipliers_array, no_parallel_execution_produces_no_block, schedule_trigger_no_branch_filter_no_residual_diff) |
| TestReleaseDefinition_ParallelExecution_ExpandFlatten | pass | +1 (3 subtests, WI-4) |
| TestReleaseDefinition_AgentlessPhase_ExpandFlatten | pass | +1 (3 subtests, WI-5) |
| TestReleaseDefinition_ExpandFlatten_Roundtrip | pass | 0 (pre-initiative, no regression) |
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
| TestAccReleaseDefinition_complete (live ADO, WI-8 + WI-9) | pass | +1 (apply + agent_specification check + real gate queryId + idempotency re-plan + destroy) |

## Acceptance criteria

- TestReleaseDefinition_AccRefresh_RetentionPolicy and _PreDeployApproval pass (WI-1)
- TestReleaseDefinition_Gates_ExpandFlatten passes (WI-2)
- TestReleaseDefinition_Triggers_ExpandFlatten (and _Empty, _ArtifactOnly, _ScheduleOnly) passes (WI-3)
- TestReleaseDefinition_ParallelExecution_ExpandFlatten passes (WI-4)
- TestReleaseDefinition_AgentlessPhase_ExpandFlatten passes (WI-5)
- TestReleaseDefinition_GatesTasks_ExpandFlatten passes — gates#>0 expand/flatten round-trips (WI-6)
- TestReleaseDefinition_RoundTrip passes (4 subtests) — multipliers, no-parallel-execution, schedule_trigger no branch_filter residual diff (WI-7 + WI-9)
- schedule_trigger branch_filter removed from schema — no perpetual diff (WI-9 AC1)
- agent_specification = ubuntu-22.04 in acceptance test + asserted via checkReleaseDefinitionAgentSpecification (WI-9 AC2)
- betterado_workitemquery creates real shared query; gate task queryId = real GUID (WI-9 AC3)
- TestAccReleaseDefinition_complete passes live ADO: apply + check + idempotency plan + destroy (WI-8 + WI-9 AC4)
- go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0

## Files Changed

- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — WI-1: retention_policy + pre_deploy_approval in acceptance HCL fixtures; WI-8: TestAccReleaseDefinition_complete (exhaustive live acceptance test); WI-9: agent_specification=ubuntu-22.04, betterado_workitemquery for real gate queryId, checkReleaseDefinitionAgentSpecification check (+559 lines)
- `azuredevops/internal/acceptancetests/testutils/commons.go` — WI-8: minor test utility update (-4/+4 lines)
- `azuredevops/internal/service/release/resource_release_definition.go` — WI-1–9: schema additions (gates+tasks, triggers, parallel_execution, agentless, agent_specification) + expand/flatten helpers + idempotency fixes + schedule_trigger branch_filter removal (+844/-76 lines)
- `azuredevops/internal/service/release/resource_release_definition_test.go` — WI-1–7 + WI-9: 11 new top-level unit test functions covering all schema features and round-trip correctness; WI-9: stale schedule_trigger branch_filter assertions removed from _Triggers_ScheduleOnly and _Triggers_ExpandFlatten (+1361 lines)

```
azuredevops/internal/acceptancetests/resource_release_definition_test.go                        |  559 +++++++-
 azuredevops/internal/acceptancetests/testutils/commons.go                                       |    4 +-
 azuredevops/internal/service/release/resource_release_definition.go                             |  844 +++++++++++-
 azuredevops/internal/service/release/resource_release_definition_test.go                        | 1361 +++++++++++++++++++-
 demo/INIT-2026-06-05-complete-release-definition/DEMO.html                                      |  501 +++++++
 demo/INIT-2026-06-05-complete-release-definition/DEMO.md                                        |  299 +++++
 demo/INIT-2026-06-05-complete-release-definition/demo.json                                      |  301 +++++
 9 files changed, 3868 insertions(+), 76 deletions(-)
```

## Usage

```
```hcl
# WI-9: betterado_workitemquery creates a real shared query for gate task queryId
resource "betterado_workitemquery" "gate_query" {
  project_id = var.project_id
  name       = "All Work Items - Gate Check"
  path       = "Shared Queries"
  wiql       = "SELECT [System.Id] FROM WorkItems"
}

resource "betterado_release_definition" "example" {
  name                = "MyRelease"
  project_id          = var.project_id
  description         = "Exhaustive example"
  release_name_format = "Release-$(rev:r)"

  artifact {
    alias      = "_build"
    type       = "Build"
    is_primary = true
    definition_reference {
      key   = "project"
      value = var.project_id
    }
    definition_reference {
      key   = "definition"
      value = var.build_definition_id
    }
  }

  # WI-3: CD artifact + schedule triggers
  # WI-9: schedule_trigger has NO branch_filter (ADO classic schedule triggers are time-based)
  triggers {
    cd_artifact_trigger {
      artifact_alias = "_build"
      branch_filter {
        include = ["refs/heads/main"]
        exclude = []
      }
    }
    schedule_trigger {
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 127
    }
  }

  environment {
    name = "Staging"
    rank = 1

    # WI-1: retention policy (required by ADO REST 7.2)
    retention_policy {
      days_to_keep     = 14
      releases_to_keep = 5
      retain_build     = false
    }

    # WI-1: pre-deploy approval
    pre_deploy_approval {
      approval {
        is_automated = false
        rank         = 1
        approver_id  = var.approver_id
      }
      approval_options {
        release_creator_can_be_approver = true
        timeout_in_minutes              = 720
      }
    }

    # WI-2 + WI-6: deployment gates WITH real gate tasks (gates#>0)
    # WI-9: queryId = real shared query GUID from betterado_workitemquery
    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }
      # WI-6: gate block — each gate carries a workflow task
      gate {
        task {
          task_id          = "f1e4b0e6-017e-4819-8a48-ef19ae96e289" # queryWorkItems
          version          = "2"
          name             = "Query work items"
          is_enabled       = true
          condition        = ""
          always_run       = false
          continue_on_error = false
          inputs           = {
            queryId = betterado_workitemquery.gate_query.id  # WI-9: real query GUID
          }
        }
      }
    }

    deploy_phase {
      name       = "Agent phase"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id                  = var.agent_pool_id  # real queue — never 0
        skip_artifacts_download   = true
        enable_access_token       = true
        timeout_in_minutes        = 60
        demands                   = ["Agent.Version -gtVersion 2.0"]
        # WI-9: agent_specification sets the agent image explicitly
        agent_specification       = "ubuntu-22.04"

        # WI-4: parallel execution (WI-7: multipliers round-trips correctly)
        parallel_execution {
          type                 = "multiConfiguration"
          max_number_of_agents = 2
          multipliers          = ["Configuration"]
          continue_on_error    = false
        }
      }
    }

    deploy_phase {
      name       = "Agentless phase"
      # WI-5: runOnServer — queueId omitted from ADO wire payload
      phase_type = "runOnServer"
      rank       = 2

      workflow_task {
        task_id   = "28782b92-5e8e-4458-9751-a71cd1492bae" # Delay task (v1)
        version   = "1"
        name      = "Delay"
        is_enabled = true
        inputs     = { delayForMinutes = "1" }
      }

      deployment_input {
        timeout_in_minutes        = 120
        cancel_timeout_in_minutes = 5
      }
    }
  }

  environment {
    name = "Production"
    rank = 2

    condition {
      name           = "Staging"
      condition_type = "environmentState"
      value          = "4"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approval {
        is_automated = true
        rank         = 1
      }
    }

    deploy_phase {
      name       = "Deploy"
      phase_type = "agentBasedDeployment"
      rank       = 1
      deployment_input {
        queue_id = var.agent_pool_id
      }
    }
  }
}
```
```

## Impact

- Deployment gates are now fully modelled: each gate block carries workflow tasks so gates#>0 in live ADO — quality gates at deployment checkpoints actually run (WI-6)
- Resource is idempotent: three flatten round-trip bugs eliminated including complete removal of schedule_trigger.branch_filter (ADO classic schedule triggers have no branch filter) — no more perpetual diff after apply (WI-7 + WI-9)
- agent_specification is now first-class in deployment_input — operators can pin the agent image (e.g. ubuntu-22.04) and the value round-trips through ADO (WI-9)
- Gate tasks use real queryIds via betterado_workitemquery — no more empty gate tasks that silently do nothing (WI-9)
- Exhaustive live acceptance test (TestAccReleaseDefinition_complete) codified as a standing merge gate: every schema option proven against live ADO on every CI run (WI-8 + WI-9)
- CD artifact triggers and schedule triggers are first-class HCL declarations — trigger logic version-controlled and diff-visible (WI-3)
- Parallel execution (multiConfiguration / multiMachine) and agentless (runOnServer) deploy phases fully supported (WI-4, WI-5)
- Acceptance tests ADO REST 7.2-compatible: retention_policy and pre_deploy_approval blocks prevent VS402982/VS402877 failures in live CI (WI-1)
