# Surface environment_trigger, schedule, process_parameters, properties on betterado_release_definition

> _Derived from `demo.json` (ADR 021). Essence:_ Four previously un-surfaced blocks/fields in the release definition environment schema are now readable and writable via Terraform: environment_trigger (rollback/redeploy trigger config), schedule (cron-style scheduled deployment), process_parameters (task input overrides), and properties (arbitrary string key-value pairs). The expand/flatten paths are tested end-to-end with four new round-trip unit tests, and a live acceptance test guards idempotency against a real ADO org.

## Summary

- Added four new optional sub-blocks/fields to the `environment` block: `environment_trigger`, `schedule`, `process_parameters`, `properties`
- All expand/flatten paths are covered by four new round-trip unit tests (44→48 tests in the release package)
- Custom JSON unmarshaler for `ReleaseSchedule.DaysToRelease` fixes the ADO integer-vs-enum mismatch in the vendored SDK
- New acceptance test `TestAccReleaseDefinition_environmentConfig` guards live idempotency when TF_ACC credentials are available
- Branch: `INIT-2026-06-08-release-definition-environment-config-surface`
- Commit: `a8dcd5b887277e1bffd38802054d4bcbb8b6132a`

## Intent & Outcome

> _Assessed intent:_ Four previously un-surfaced blocks/fields in the release definition environment schema are now readable and writable via Terraform: environment_trigger (rollback/redeploy trigger config), schedule (cron-style scheduled deployment), process_parameters (task input overrides), and properties (arbitrary string key-value pairs). The expand/flatten paths are tested end-to-end with four new round-trip unit tests, and a live acceptance test guards idempotency against a real ADO org.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN an environment block with an environment_trigger sub-block containing definition_environment_id, trigger_type, trigger_content WHEN expandEnvironments processes the Terraform state THEN env.EnvironmentTriggers is populated with the correct EnvironmentTrigger SDK struct values | ✓ met | TestReleaseDefinition_EnvironmentTriggers_RoundTrip → PASS (go test -tags all -run TestReleaseDefinition_EnvironmentTriggers_RoundTrip ./azuredevops/internal/service/release/) |
| 2 | GIVEN an ADO ReleaseDefinitionEnvironment with EnvironmentTriggers populated WHEN flattenEnvironments processes the API response THEN the environment_trigger block in state reflects the correct definition_environment_id, trigger_type, trigger_content values | ✓ met | TestReleaseDefinition_EnvironmentTriggers_RoundTrip (flatten half) → PASS; same test asserts state matches original HCL after flatten |
| 3 | GIVEN an environment block with a schedule sub-block containing days_to_release, start_hours, start_minutes, time_zone_id, job_id WHEN expandEnvironments processes the Terraform state THEN env.Schedules is populated with the correct ReleaseSchedule SDK struct values | ✓ met | TestReleaseDefinition_EnvironmentSchedules_RoundTrip → PASS (go test -tags all -run TestReleaseDefinition_EnvironmentSchedules_RoundTrip ./azuredevops/internal/service/release/) |
| 4 | GIVEN an ADO ReleaseDefinitionEnvironment with Schedules populated WHEN flattenEnvironments processes the API response THEN the schedule block in state reflects days_to_release, start_hours, start_minutes, time_zone_id, job_id | ✓ met | TestReleaseDefinition_EnvironmentSchedules_RoundTrip (flatten half) → PASS; asserts days_to_release=62, start_hours=3, start_minutes=30, time_zone_id=UTC |
| 5 | GIVEN an environment block with a process_parameters sub-block containing an inputs list (name, default_value, parameter_type) WHEN expandEnvironments processes the Terraform state THEN env.ProcessParameters is populated with a distributedtaskcommon.ProcessParameters having Inputs entries matching the HCL | ✓ met | TestReleaseDefinition_EnvironmentProcessParameters_RoundTrip → PASS (go test -tags all -run TestReleaseDefinition_EnvironmentProcessParameters_RoundTrip ./azuredevops/internal/service/release/) |
| 6 | GIVEN an ADO ReleaseDefinitionEnvironment with ProcessParameters populated WHEN flattenEnvironments processes the API response THEN the process_parameters block in state reflects all inputs name/default_value/parameter_type values | ✓ met | TestReleaseDefinition_EnvironmentProcessParameters_RoundTrip (flatten half) → PASS; asserts name=myParam, default_value=default, parameter_type=string |
| 7 | GIVEN an environment block with a properties map containing string key-value pairs WHEN expandEnvironments processes the Terraform state THEN env.Properties is set to the map[string]interface{} derived from the HCL map[string]string | ✓ met | TestReleaseDefinition_EnvironmentProperties_RoundTrip → PASS (go test -tags all -run TestReleaseDefinition_EnvironmentProperties_RoundTrip ./azuredevops/internal/service/release/) |
| 8 | GIVEN an ADO ReleaseDefinitionEnvironment with Properties set WHEN flattenEnvironments processes the API response THEN the properties TypeMap in state contains the correct string keys and values | ✓ met | TestReleaseDefinition_EnvironmentProperties_RoundTrip (flatten half) → PASS; asserts properties[env]=prod after round-trip |
| 9 | GIVEN unit tests TestReleaseDefinition_EnvironmentTriggers_RoundTrip, TestReleaseDefinition_EnvironmentSchedules_RoundTrip, TestReleaseDefinition_ProcessParameters_RoundTrip, TestReleaseDefinition_EnvironmentProperties_RoundTrip WHEN go test -tags all -run TestReleaseDefinition_Environment.*RoundTrip is executed THEN all four tests pass | ✓ met | go test -tags all -run 'TestReleaseDefinition_Environment.*RoundTrip' ./azuredevops/internal/service/release/ → 4/4 PASS (0.004s) |
| 10 | GIVEN a new TestAccReleaseDefinition_environmentConfig acceptance test that creates a release definition with environment_trigger, schedule, and properties blocks configured WHEN TF_ACC=1 is set and the test runs against live ADO THEN the test creates the release definition successfully, a PlanOnly step with ExpectNonEmptyPlan:false confirms idempotency, and destroy completes without error | ~ partial | TestAccReleaseDefinition_environmentConfig is present and compiles (go test -tags all -run TestAccReleaseDefinition_environmentConfig ./azuredevops/internal/acceptancetests/ exits 0). Live ADO run requires TF_ACC=1 + credentials; not available in the offline harness. The test structure (2-step: apply+check → plan-only idempotency) matches the AC contract. |
| 11 | GIVEN the acceptance test file azuredevops/internal/acceptancetests/resource_release_definition_test.go WHEN go test -tags all -run TestAccReleaseDefinition_environmentConfig ./azuredevops/internal/acceptancetests/ is executed with TF_ACC=1 THEN the test function is found and runs | ~ partial | go test -tags all -run TestAccReleaseDefinition_environmentConfig ./azuredevops/internal/acceptancetests/ exits 0 (test found, skipped without TF_ACC). Function presence and compilation confirmed. Live run blocked by absent credentials. |

## Test Evidence

### Four new unit round-trip tests confirm each environment sub-block expands to the correct SDK struct and flattens back to identical Terraform state.

- **Before:** environment_trigger, schedule, process_parameters, and properties fields were silently dropped: expand produced zero-value structs, flatten produced empty state. No test coverage existed for these paths.
- **After:** All four round-trip tests pass: EnvironmentTriggers, Schedules, ProcessParameters, and Properties each survive a full expand→flatten cycle with value fidelity. Gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/...` → PASS (48/48 tests).

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| release package tests (total) | 44 | 48 | +9.1% | within |
| TestReleaseDefinition_EnvironmentTriggers_RoundTrip | not present | PASS | — | new |
| TestReleaseDefinition_EnvironmentSchedules_RoundTrip | not present | PASS | — | new |
| TestReleaseDefinition_EnvironmentProcessParameters_RoundTrip | not present | PASS | — | new |
| TestReleaseDefinition_EnvironmentProperties_RoundTrip | not present | PASS | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### A new acceptance test creates a release definition with environment_trigger, schedule, and properties blocks, verifies attribute values, confirms idempotency (ExpectNonEmptyPlan: false), then destroys. Runs live when TF_ACC=1; skips cleanly in offline CI.

- **Before:** No acceptance test exercised the three new environment sub-blocks. Drift between schema and ADO API would be invisible to the test suite.
- **After:** TestAccReleaseDefinition_environmentConfig is present and compiles. With TF_ACC=1 it creates the definition, asserts environment_trigger.0.trigger_type=rollbackRedeploy + schedule.0.start_hours=3 + schedule.0.time_zone_id=UTC + properties.env=staging, runs a plan-only idempotency check, then destroys. Without TF_ACC it skips (offline CI stays green).

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseDefinition_EnvironmentTriggers_RoundTrip | pass | new |
| TestReleaseDefinition_EnvironmentSchedules_RoundTrip | pass | new |
| TestReleaseDefinition_EnvironmentProcessParameters_RoundTrip | pass | new |
| TestReleaseDefinition_EnvironmentProperties_RoundTrip | pass | new |
| TestAccReleaseDefinition_environmentConfig | skip | new (skips without TF_ACC; live run requires credentials) |
| service/release full suite (48 tests) | pass | +4 vs main (44→48) |
| service/taskagent full suite | pass | unchanged |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/release/resource_release_definition.go` — Schema additions (environment_trigger, schedule, process_parameters, properties blocks) + expand/flatten helpers + vendor schedule_unmarshal integration
- `azuredevops/internal/service/release/resource_release_definition_test.go` — Four new round-trip unit tests for the four new environment sub-blocks
- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — New TestAccReleaseDefinition_environmentConfig acceptance test + hclReleaseDefinitionEnvironmentConfig HCL helper
- `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/release/schedule_unmarshal.go` — Custom JSON unmarshaler for ReleaseSchedule.DaysToRelease which ADO returns as an integer bitmask rather than the SDK's string enum

```
azuredevops/internal/acceptancetests/resource_release_definition_test.go    | 101 +++++++
 azuredevops/internal/service/release/resource_release_definition.go        | 332 +++++++++++++++++++++
 azuredevops/internal/service/release/resource_release_definition_test.go   | 204 +++++++++++++
 vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/release/schedule_unmarshal.go |  89 ++++++
 4 files changed, 726 insertions(+)
```

## Usage

```
```hcl
resource "betterado_release_definition" "example" {
  name       = "My Release"
  project_id = data.betterado_project.example.id
  path       = "\\"

  environment {
    name = "Staging"
    rank = 1

    # Trigger a re-deploy when a rollback occurs in another environment
    environment_trigger {
      definition_environment_id = 2
      trigger_type              = "rollbackRedeploy"
      trigger_content           = "{}"
    }

    # Schedule a deployment every weekday at 03:00 UTC
    schedule {
      days_to_release = 62   # Mon–Fri bitmask
      start_hours     = 3
      start_minutes   = 0
      time_zone_id    = "UTC"
    }

    # Override a task input across the phase
    process_parameters {
      input {
        name           = "BuildConfiguration"
        default_value  = "Release"
        parameter_type = "string"
      }
    }

    # Arbitrary key-value metadata ADO stores on the environment
    properties = {
      env  = "staging"
      team = "platform"
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }
  }
}
```
```

## Impact

- Operators can now configure deployment triggers between environments (rollback redeploy, deployment group redeploy) entirely in HCL, previously requiring a manual portal click.
- Scheduled release deployments (cron-style, per-environment) are now expressible in Terraform without drift — the bitmask DaysToRelease field round-trips correctly via the new custom JSON unmarshaler.
- Process-parameter overrides (task input defaults scoped to an environment) are surfaced so template-driven pipeline configurations can be fully managed as code.
- Arbitrary environment metadata (the ADO `properties` bag) is now readable and writable, enabling tagging/labelling workflows and third-party integrations that consume those fields.
