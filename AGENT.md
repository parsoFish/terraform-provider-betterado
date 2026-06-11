# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

**Primary task**: Add `TestAccReleaseDefinition_environmentConfig` + `hclReleaseDefinitionEnvironmentConfig` to `azuredevops/internal/acceptancetests/resource_release_definition_test.go`.

The test uses `SharedReleaseFixture(t)` (avoids creating an inline project). The HCL exercises `environment_trigger`, `schedule`, and `properties` blocks with required `retention_policy`, `pre_deploy_approval`, and `post_deploy_approval` blocks (needed by ADO VS402877/VS402982).

**Three bugs discovered and fixed during live test runs:**

1. **`schedule.daysToRelease` type mismatch** — ADO REST returns `"daysToRelease": 62` (JSON integer) but the Go SDK's `ReleaseSchedule.DaysToRelease` is `*ScheduleDays` (a Go string type). Standard JSON unmarshal fails. **Fix**: Added `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/release/schedule_unmarshal.go` with custom `MarshalJSON` (emits integer for numeric strings) and `UnmarshalJSON` (accepts JSON number via `json.Number`) for `ReleaseSchedule`.

2. **`properties` typed wrapper** — ADO returns environment properties as `{"key": {"$type": "System.String", "$value": "actual-value"}}` but the flatten path just `fmt.Sprintf`-d the map, producing `map[$type:System.String $value:staging]`. **Fix**: Updated `flattenEnvironments` in `resource_release_definition.go` to unwrap the `$value` key when the property value is a `map[string]interface{}`.

3. **`environment_trigger.definition_environment_id` perpetual drift** — ADO auto-fills `definition_environment_id` with the environment's ID when the user sends `0`. On read, ADO returns the env's ID (e.g., `4`), causing a perpetual diff against the user's `0`. **Fix**: Changed schema from `Optional: true, Default: 0` to `Optional: true, Computed: true` so Terraform uses ADO's returned value in state without diffing against the unset user config.

## What worked

- Using `SharedReleaseFixture(t)` (existing fixture) with just `fixture.ProjectID` — no inline project needed in the HCL.
- Adding custom `MarshalJSON`/`UnmarshalJSON` directly to the vendor package (`vendor/.../release/schedule_unmarshal.go`) — Go allows adding methods to types within the same package; adding a file to vendor is allowed and doesn't break `go mod vendor`.
- `Optional: true, Computed: true` (no Default) for auto-filled ID fields — Terraform treats unset as "use whatever the provider returns."

## What didn't work

- `Optional: true, Default: 0` for `definition_environment_id` causes a perpetual diff because ADO replaces the 0 with the actual environment ID.
- Relying on `fmt.Sprintf("%v", v)` for properties values — ADO wraps string values in a typed object.

## Open questions

_(none blocking)_

## Notes for reflection

- The ADO Go SDK has several type mismatches with the REST API (string enums vs integer bitmasks). The `ReleaseSchedule.DaysToRelease` issue may not be the only one — worth scanning for similar patterns if environment-level schedules appear in other resources.
- The `$type`/`$value` wrapper pattern for properties is an ADO serialisation convention for typed objects; any resource that reads `env.Properties` or similar needs the same unwrap logic.
