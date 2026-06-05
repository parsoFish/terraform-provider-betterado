# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a `betterado_release_definition` resource includes a `triggers` block containing a `cd_artifact_trigger` (artifact alias + branch filter) and a `schedule_trigger` (cron expression + branch filter + time zone) WHEN `expandReleaseDefinition` processes the resource data THEN the resulting `ReleaseDefinition.Triggers` slice contains one `ArtifactFilter`-type trigger and one `ScheduleTrigger`-type trigger with all fields correctly populated
- [x] AC2: GIVEN a `ReleaseDefinition` from ADO whose `Triggers` slice contains both artifact and schedule triggers WHEN `flattenReleaseDefinition` processes it THEN the Terraform state contains a `triggers` block with matching `cd_artifact_trigger` and `schedule_trigger` sub-blocks
- [x] AC3: GIVEN new unit tests `TestReleaseDefinition_Triggers_ExpandFlatten` exist in the release package WHEN `go test -tags all -count=1 -run TestReleaseDefinition_Triggers ./azuredevops/internal/service/release/` is executed THEN the tests pass and exit 0
