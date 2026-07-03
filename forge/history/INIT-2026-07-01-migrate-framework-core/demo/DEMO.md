# Migrate all core resources and data sources to terraform-plugin-framework

> _Derived from `demo.json` (ADR 021). Essence:_ Migrates all 7 core resources (betterado_project, betterado_project_features, betterado_project_pipeline_settings, betterado_project_tags, betterado_team, betterado_team_administrators, betterado_team_members) and 5 data sources (data.betterado_project, data.betterado_projects, data.betterado_team, data.betterado_teams, data.betterado_client_config) to terraform-plugin-framework, served through the mux provider. Produces a gap matrix (docs/core-gap-matrix.md) comparing all core resource schemas against the ADO REST API v7.1. Bug fix: applyFeatureStates surfaces silent ADO feature-state rejections (testplans license restriction). Fixture safety hardened. Missing validators and features attribute in betterado_project schema are documented as deferred regressions. provider_test.go updated to reflect all SDKv2 deregistrations.

## Intent & Outcome

> _Assessed intent:_ Migrates all 7 core resources and 5 data sources to terraform-plugin-framework, served through the mux provider.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Projects/Teams/Features REST API v7.1 schema WHEN compared against each in-scope resource's current SDKv2 schema THEN docs/core-gap-matrix.md exists and lists every field for all 7 resources with status and rationale | ✓ met | docs/core-gap-matrix.md committed in 4e36fa83; contains field-by-field tables for all 7 resources with implemented/read-only/gap/out-of-scope status; every gap row carries explicit rationale. |
| 2 | GIVEN betterado_project resource implemented as resource.Resource WHEN terraform import is run against betterado-standing-demo THEN import succeeds, read-back asserts attributes, idempotency re-plan shows no diff | ✓ met | TestAccProject_importByName uses GetMuxedProviderFactories(); imports betterado-standing-demo; asserts name/visibility/version_control; ExpectNonEmptyPlan: false. Committed in cbea0eef + e995fe9d + cf0491cf. |
| 3 | GIVEN data.betterado_project data source implemented WHEN terraform apply runs with a data source lookup THEN data source returns correct project fields | ✓ met | data_project_framework.go (172 lines) committed; TestAccProject_dataSource_withID and TestAccProject_dataSource_withName updated to use GetMuxedProviderFactories(). Committed in af6f73e2. |
| 4 | GIVEN data.betterado_projects data source implemented WHEN terraform apply runs listing projects THEN TestAccProjects_dataSource passes | ✓ met | data_projects_framework.go (186 lines) committed; TestAccProjects_dataSource updated to use GetMuxedProviderFactories(). Committed in af6f73e2. |
| 5 | GIVEN betterado_project, data.betterado_project, data.betterado_projects deregistered from SDKv2 WHEN TestProvider_HasChildResources and TestProvider_HasChildDataSources run THEN both pass | ✓ met | provider.go removes betterado_project from ResourcesMap; data sources from DataSourcesMap. TestProvider_HasChildResources + TestProvider_HasChildDataSources pass → ok 0.007s. |
| 6 | GIVEN betterado_project_features implemented WHEN terraform apply enables/disables features THEN TestAccProjectFeatures_roundtrip passes; CaptureLiveEvidence called; ExpectNonEmptyPlan: false | ✓ met | TestAccProjectFeatures_roundtrip passes (uses artifacts+boards; applyFeatureStates checks ContributedFeatureState return). CaptureLiveEvidence wrote live-evidence/acceptance-resource.json (capturedAt: 2026-07-02T09:17:10Z). |
| 7 | GIVEN betterado_project_features removed from SDKv2 ResourcesMap WHEN TestProvider_HasChildResources runs THEN test passes | ✓ met | provider.go removes betterado_project_features. TestProvider_HasChildResources passes → ok 0.007s. |
| 8 | GIVEN SDKv2 betterado_project validators WHEN framework resource Schema() is finalized THEN equivalent validators are attached | ✗ missed | resource_project_framework.go Schema() carries no stringvalidator for name or visibility/version_control. Deferred to follow-up initiative. |
| 9 | GIVEN SDKv2 betterado_project_features validators WHEN framework resource Schema() is finalized THEN equivalent validators are attached | ✗ missed | resource_project_features_framework.go Schema() carries no UUID validator on project_id and no map-value validator for feature keys/values. Deferred to follow-up initiative. |
| 10 | GIVEN SDKv2 betterado_project carried features TypeMap WHEN framework betterado_project ships THEN features is present OR gap matrix and CHANGELOG explicitly reclassify | ✓ met | CHANGELOG.md ## [Unreleased] BREAKING CHANGES explicitly documents the deliberate removal of features inline TypeMap with migration guidance and rationale. |
| 11 | GIVEN features attribute regression WHEN it is fixed THEN offline test asserts attribute exists | ~ partial | resource_project_features_absence_test.go committed (101 lines) — verifies betterado_project Schema() does NOT include features (documents deliberate breaking removal). No positive regression test added. |
| 12 | GIVEN betterado_project_pipeline_settings implemented WHEN terraform apply sets pipeline settings THEN TestAccProjectPipelineSettings passes | ~ partial | resource_project_pipeline_settings_framework.go (309 lines) committed; registered in framework_provider.go; removed from SDKv2. TestProvider_HasChildResources passes. Live acceptance test not confirmed run in CI. |
| 13 | GIVEN betterado_project_pipeline_settings removed from SDKv2 WHEN TestProvider_HasChildResources runs THEN passes | ✓ met | provider.go removes betterado_project_pipeline_settings; provider_test.go updated. TestProvider_HasChildResources passes → ok 0.007s. |
| 14 | GIVEN betterado_project_tags implemented WHEN terraform apply adds tags THEN TestAccProjectTags passes | ~ partial | resource_project_tags_framework.go (332 lines) committed; registered; removed from SDKv2. TestProvider_HasChildResources passes. Live acceptance test not confirmed run in CI. |
| 15 | GIVEN betterado_project_tags removed from SDKv2 WHEN TestProvider_HasChildResources runs THEN passes | ✓ met | provider.go removes betterado_project_tags; provider_test.go updated. TestProvider_HasChildResources passes → ok 0.007s. |
| 16 | GIVEN betterado_team implemented WHEN terraform apply creates a team THEN TestAccTeam_basic and TestAccTeam_update pass | ~ partial | resource_team_framework.go (414 lines) committed; registered; removed from SDKv2. TestProvider_HasChildResources passes. Live acceptance test not confirmed run in CI. |
| 17 | GIVEN data.betterado_team and data.betterado_teams implemented WHEN terraform apply reads teams THEN tests pass | ~ partial | data_team_framework.go (189 lines) and data_teams_framework.go (255 lines) committed; registered; removed from SDKv2. TestProvider_HasChildDataSources passes. Live acceptance tests not confirmed run in CI. |
| 18 | GIVEN betterado_team removed from SDKv2 and data sources deregistered WHEN provider tests run THEN both pass with updated counts | ✓ met | provider.go removes betterado_team from ResourcesMap and DataSourcesMap; provider_test.go updated. Both provider tests pass → ok 0.007s. |
| 19 | GIVEN betterado_team_administrators implemented WHEN terraform apply sets administrators THEN TestAccTeamAdministrators passes | ~ partial | resource_team_administrators_framework.go (350 lines) committed; registered; removed from SDKv2. TestProvider_HasChildResources passes. Live acceptance test not confirmed run in CI. |
| 20 | GIVEN betterado_team_members implemented WHEN terraform apply sets members THEN TestAccTeamMembers passes | ~ partial | resource_team_members_framework.go (327 lines) committed; registered; removed from SDKv2. TestProvider_HasChildResources passes. Live acceptance test not confirmed run in CI. |
| 21 | GIVEN betterado_team_administrators and betterado_team_members removed from SDKv2 WHEN TestProvider_HasChildResources runs THEN passes | ✓ met | provider.go removes both; provider_test.go updated. TestProvider_HasChildResources passes → ok 0.007s. |
| 22 | GIVEN data.betterado_client_config implemented WHEN terraform apply reads provider config metadata THEN TestAccClientConfig_LoadsCorrectProperties passes | ~ partial | data_client_config_framework.go (138 lines) committed; registered; removed from SDKv2. TestProvider_HasChildDataSources passes. Live acceptance test not confirmed run in CI. |
| 23 | GIVEN data.betterado_client_config removed from SDKv2 WHEN TestProvider_HasChildDataSources runs THEN passes | ✓ met | provider.go removes betterado_client_config; provider_test.go updated. TestProvider_HasChildDataSources passes → ok 0.007s. |
| 24 | GIVEN all core resources and data sources migrated WHEN make docs runs THEN docs regenerated for every migrated resource | ~ partial | WI-9 docs regeneration not run for the full set. The gap matrix is committed. tfplugindocs regeneration deferred to post-review finalisation. |
| 25 | GIVEN CHANGELOG.md updated and PROVIDER_VERSION.txt bumped WHEN git diff HEAD shows changes THEN CHANGELOG.md has new entry under ## Unreleased | ~ partial | CHANGELOG.md has a DRAFT ## [Unreleased] entry (BREAKING CHANGES + FEATURES + BUG FIXES) covering all WI-1 through WI-8 deliverables. PROVIDER_VERSION.txt bump is a pre-merge finaliser step. |
| 26 | GIVEN demo.json carries real REST GET checkpoints WHEN forge demo render is invoked THEN demo.json ends with checkpoint carrying liveEvidence.url | ✓ met | Live evidence captured by TestAccProjectFeatures_roundtrip → CaptureLiveEvidence('acceptance-resource') → live-evidence/acceptance-resource.json: url=https://dev.azure.com/davidgparsonson/_apis/FeatureManagement/FeatureStatesForScope/host/project/c0ac3757-e915-453f-ba2b-93a3720d1994?api-version=7.1; capturedAt=2026-07-02T09:17:10Z. |

## Visual Changes

### Offline unit gate: release + taskagent packages green — verbatim gate forge ran

- **Before:** Gate runs against main branch (pre-migration)
- **After:** Gate passes on branch HEAD: ok release 0.007s; ok taskagent 0.006s; ok taskagent/validate 0.004s

### Framework provider registers all migrated resources/data-sources; SDKv2 maps fully updated

- **Before:** betterado_project, betterado_project_features, betterado_project_pipeline_settings, betterado_project_tags, betterado_team, betterado_team_administrators, betterado_team_members in SDKv2 ResourcesMap; betterado_team, betterado_teams, betterado_client_config in DataSourcesMap
- **After:** All 7 core resources and 5 data sources removed from SDKv2 maps; all registered in framework_provider.go Resources()/DataSources(). TestProvider_HasChildResources + TestProvider_HasChildDataSources pass.

### Live REST GET: betterado_project_features feature states from ADO API (CaptureLiveEvidence, capturedAt 2026-07-02T09:17:10Z)

- **Before:** betterado_project_features served via SDKv2 schema helper; testplans feature toggle silently failed (license restriction) causing 'inconsistent result after apply' panic
- **After:** betterado_project_features served via framework resource.Resource; applyFeatureStates checks ContributedFeatureState return — surfaces license errors; test uses artifacts+boards (license-free); live GET: artifacts=disabled boards=enabled pipelines=enabled repositories=enabled testplans=disabled

## Files Changed

```
166 files changed, 12493 insertions(+), 466 deletions(-)
```
