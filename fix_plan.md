# Fix Plan

> Checklist for UWI-2. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN the initiative scope of 7 resources + 5 data sources and the 8 undelivered types (project_pipeline_settings, project_tags, team, team_administrators, team_members resources; team, teams, client_config data sources) WHEN the remaining work items are delivered THEN all 8 remaining types are migrated to the framework with SDKv2 deregistration, framework registration, provider_test count updates, live TestAcc per type against the restored fixture (team resources may create live inside the fixture project), CaptureLiveEvidence under distinct per-type labels, and docs/CHANGELOG updated
- [ ] AC2: GIVEN SDKv2 betterado_project validated name (StringIsNotWhiteSpace) and visibility/version_control (StringInSlice) WHEN the framework resource Schema() is finalized THEN equivalent stringvalidator validators are attached to each attribute
- [ ] AC3: GIVEN SDKv2 betterado_project_features validated project_id (IsUUID) and feature keys/values (validateProjectFeatures) WHEN the framework resource Schema() is finalized THEN equivalent validators are attached (UUID pattern + map-value validator with the same accepted feature names/states)
- [ ] AC4: GIVEN SDKv2 betterado_project carried a wired features TypeMap attribute and the WI-1 gap matrix classifies it implemented WHEN the framework betterado_project ships THEN features is present in the schema and wired through create/read/update — OR the gap matrix and CHANGELOG explicitly reclassify it as a deliberate breaking deferral with rationale
- [ ] AC5: GIVEN the features attribute regression went undetected by every gate WHEN it is fixed THEN an offline schema/roundtrip test asserts the attribute exists and is wired so a future migration cannot silently drop it

## Completed sub-tasks

- [x] **Iteration 6 gate fix:** `TestAccProjectFeatures_roundtrip` was failing with "Provider produced
  inconsistent result after apply" on `testplans`. Root cause: ADO org lacks TestPlans license;
  `SetFeatureStateForScope` returns HTTP 200 but leaves testplans state unchanged. Provider read-back
  sees "disabled" while plan expected "enabled".
  - Fix 1: Changed acceptance test to use `artifacts`+`boards` (license-free) instead of `testplans`+`artifacts`.
  - Fix 2: `applyFeatureStates` now checks the returned `ContributedFeatureState` and surfaces a clear error
    if the API doesn't apply the requested state (prevents misleading "inconsistent result after apply" panics).
  - Committed: `b154c6de`
