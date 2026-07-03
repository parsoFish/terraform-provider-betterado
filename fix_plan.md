# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_feature_flag resource with feature_id, scope_name, scope_value, and state configured WHEN the Create CRUD method runs THEN SetFeatureStateForScope is called with the correct featureId, userScope (scope_name), scopeName (scope_name), scopeValue (scope_value), and patch state; the method reads back state via GetFeatureStateForScope and stores computed fields (overridden, reason)
- [x] AC2: GIVEN a betterado_feature_flag resource already created WHEN the Read CRUD method runs (e.g. during refresh) THEN GetFeatureStateForScope is called; if the API returns a 404 the resource is removed from state (RemoveResource); if the state field is 'undefined' the resource is treated as deleted
- [x] AC3: GIVEN a betterado_feature_flag resource with state changed in config WHEN the Update CRUD method runs THEN SetFeatureStateForScope is called with the new state; state is re-read and stored
- [x] AC4: GIVEN a betterado_feature_flag resource being destroyed WHEN the Delete CRUD method runs THEN SetFeatureStateForScope is called with state 'undefined' to remove management (restore to default); state is removed from Terraform
- [x] AC5: GIVEN unit tests using the gomock mock from WI-2 WHEN go test -tags all -run TestFeatureFlagCRUD ./azuredevops/internal/service/featuremanagement/ is run THEN all four CRUD method unit tests pass: TestFeatureFlagCreate, TestFeatureFlagRead, TestFeatureFlagUpdate, TestFeatureFlagDelete

## Status: ALL ACs COMPLETE

All four CRUD tests pass under the gate command:
`go test -tags all -run TestFeatureFlagCRUD ./azuredevops/internal/service/featuremanagement/`
