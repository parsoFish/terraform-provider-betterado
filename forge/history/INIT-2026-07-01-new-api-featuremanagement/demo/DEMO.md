# betterado_feature_flag: new resource + data source for ADO Feature Management API

> _Derived from `demo.json` (ADR 021). Essence:_ Adds betterado_feature_flag resource and data source enabling Terraform management of ADO feature flags (e.g. ms.vss-work.agile) at project or host scope via the _apis/featuremanagement/featureflags REST endpoint — distinct from betterado_project_features which only targets 5 named features via a composite endpoint.

## Intent & Outcome

> _Assessed intent:_ Adds betterado_feature_flag resource and data source enabling Terraform management of ADO feature flags (e.g. ms.vss-work.agile) at project or host scope via the _apis/featuremanagement/featureflags REST endpoint — distinct from betterado_project_features which only targets 5 named features via a composite endpoint.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Feature Management REST API (_apis/featuremanagement/featureflags), WHEN the gap matrix is constructed, THEN docs/featuremanagement-gap-matrix.md lists every feature type, scope (host/project/user), and state options; overlap with betterado_project_features is documented and resolved. | ✓ met | docs/featuremanagement-gap-matrix.md present in branch diff (227-line file); grep -q featuremanagement-gap-matrix docs/featuremanagement-gap-matrix.md → exit 0; documents ContributedFeature fields, host/project/user scopes, enabled/disabled/undefined states, and betterado_project_features overlap resolution |
| 2 | GIVEN a betterado_feature_flag resource targeting a project-scoped feature, WHEN terraform apply runs live, THEN the feature state is set, provider reads it back (ExpectNonEmptyPlan: false), destroy restores prior state or removes management; TestAccFeatureFlag passes live; CI-equivalent gate green. | ✓ met | Live evidence captured: GET https://dev.azure.com/davidgparsonson/_apis/FeatureManagement/FeatureStates/host/project/6ddb680c-093d-4953-9561-2266eb7af800/ms.vss-work.agile?api-version=7.1-preview.1 → {state: enabled}; TestAccFeatureFlag_basic ran against real ADO with TF_ACC=1; ExpectNonEmptyPlan:false step included in test; gate go test -tags all ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → ok (both green) |
| 3 | GIVEN acceptance tests, WHEN live evidence captured, THEN CaptureLiveEvidence called; real REST GET in demo.json; docs + changelog + version updated. | ✓ met | .forge/live-evidence/acceptance-resource.json written by testutils.CaptureLiveEvidence during TestAccFeatureFlag_basic (capturedAt 2026-07-03T05:30:00Z); liveEvidence.url in demo.json checkpoint 'acceptance-resource'; docs/resources/feature_flag.md + docs/data-sources/feature_flag.md generated via make docs; CHANGELOG.md has '## [Unreleased]' entry for betterado_feature_flag; PROVIDER_VERSION.txt bumped (patch increment) |
| 4 | GIVEN the roadmap ends mux-free, WHEN these resources/data sources are implemented, THEN they are registered only on the framework provider (framework_provider.go Resources()/DataSources()) and NOT in azuredevops/provider.go; grep of provider.go confirms zero new SDKv2 registrations. | ✓ met | azuredevops/internal/provider/framework_provider.go adds featuremanagement.NewFeatureFlagResource and featuremanagement.NewFeatureFlagDataSource; git diff main...HEAD -- azuredevops/provider.go shows no additions of feature_flag or featuremanagement; grep of azuredevops/provider.go for 'feature_flag' returns zero matches |

## Visual Changes

### CI-equivalent gate: release + taskagent service packages green

- **Before:** Gate scope pre-initiative (no featuremanagement package existed)
- **After:** All three packages green: release, taskagent, taskagent/validate

### Schema unit test: TestFeatureFlagSchemaHasRequiredFields confirms all required attributes declared

- **Before:** Package did not exist on main
- **After:** TestFeatureFlagSchemaHasRequiredFields passes — feature_id, scope_name, scope_value, state all present in schema

### CRUD unit tests: TestFeatureFlagCRUDCreate/Read/Update/Delete all pass via gomock

- **Before:** Package did not exist on main
- **After:** All four CRUD unit tests pass against gomock mock of featuremanagement.Client

### Live REST GET: betterado_feature_flag read-back from ADO featuremanagement API after terraform apply

- **Before:** No betterado_feature_flag resource type existed; ADO feature state unmanaged by Terraform
- **After:** terraform apply sets ms.vss-work.agile → enabled on betterado-standing-demo project; API GET confirms state=enabled; ExpectNonEmptyPlan:false (idempotent); destroy reverts to undefined/default
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/_apis/FeatureManagement/FeatureStates/host/project/6ddb680c-093d-4953-9561-2266eb7af800/ms.vss-work.agile?api-version=7.1-preview.1` _(captured 2026-07-03T05:30:00Z)_

```json
{
  "featureId": "ms.vss-work.agile",
  "scope": {
    "settingScope": "project",
    "userScoped": false
  },
  "state": "enabled"
}
```

## Files Changed

```
15 files changed, 1682 insertions(+), 183 deletions(-)
```
