# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the release_definition schema map WHEN the provider compiles and the schema is inspected THEN the key 'stages' is present; the key 'environment' is absent
- [x] AC2: GIVEN all d.Get / d.Set / d.GetOk calls in resource_release_definition.go WHEN the rename is applied THEN every reference uses 'stages'; no reference uses 'environment' as a schema key
- [x] AC3: GIVEN expandEnvironments / flattenEnvironments helpers WHEN the rename is applied THEN helpers are named expandStages / flattenStages (or similar); old names are gone
- [x] AC4: GIVEN unit tests in resource_release_definition_test.go WHEN the rename is applied THEN all schema.TestResourceDataRaw maps use the 'stages' key; no test references 'environment' as a schema key
