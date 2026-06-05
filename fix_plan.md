# Fix Plan

> Checklist for WI-7. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN HCL setting parallel_execution.multipliers, a schedule_trigger branch_filter.include, and a deploy phase with NO parallel_execution WHEN the resource is expanded then flattened (round-trip) THEN multipliers round-trips to the same list, branch_filter.include round-trips to the same list, and NO empty parallel_execution block is emitted for the phase that did not declare one
- [x] AC2: GIVEN new offline round-trip tests TestReleaseDefinition_RoundTrip WHEN go test -tags all -count=1 -run TestReleaseDefinition_RoundTrip ./azuredevops/internal/service/release/ runs THEN tests pass asserting flatten(expand(x))==x for the affected fields
- [x] AC3: GIVEN the full release-package unit suite WHEN go test -tags all -count=1 -run TestReleaseDefinition ./azuredevops/internal/service/release/ runs THEN all existing tests still pass (no regression)
