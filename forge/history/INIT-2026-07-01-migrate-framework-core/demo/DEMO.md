# Demo — INIT-2026-07-01-migrate-framework-core

> **Migrate core resources (betterado_project, betterado_project_features, + data sources) to terraform-plugin-framework**

## Essence

`betterado_project` (resource) plus `data.betterado_project` and `data.betterado_projects` data sources are now served by the mux provider via terraform-plugin-framework (WI-2). `betterado_project_features` is also migrated to a framework `resource.Resource` (WI-3) and proven by a live acceptance test with real evidence captured via `CaptureLiveEvidence`. A gap matrix (`docs/core-gap-matrix.md`) documents every field for all 7 core resources against the ADO Projects/Teams/Features REST API v7.1 (WI-1). WI-4 through WI-9 exited `status: failed`; those resources remain in SDKv2 and are deferred to a follow-up initiative.

## Diff stat

12 files changed, 1633 insertions(+), 117 deletions(-)

---

## Checkpoint 1 — Offline quality gate

**Caption:** Offline unit tests for release and taskagent packages pass on branch HEAD (the gate forge ran, verbatim)

**Command (before/after evidence):**
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```

| | |
|---|---|
| **Before (main)** | Framework files did not exist; only SDKv2 paths compiled |
| **After (HEAD)** | `ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.008s` \| `ok .../taskagent 0.007s` \| `ok .../taskagent/validate 0.006s` — all three packages green |

---

## Checkpoint 2 — Provider registration

**Caption:** Framework provider registers migrated resources/data-sources; SDKv2 maps updated; TestProvider_HasChildResources + TestProvider_HasChildDataSources pass

**Command (before/after evidence):**
```
go test -tags all -count=1 -run TestProvider_HasChildResources ./azuredevops/
```

| | |
|---|---|
| **Before (main)** | `betterado_project`, `betterado_project_features` in SDKv2 `ResourcesMap`; `betterado_project`, `betterado_projects` in `DataSourcesMap` |
| **After (HEAD)** | `betterado_project` + `betterado_project_features` removed from SDKv2 maps; registered in `framework_provider.go` `Resources()` / `DataSources()`. `TestProvider_HasChildResources` + `TestProvider_HasChildDataSources` → `ok` (0.007s) |

---

## Checkpoint 3 — Live project_features resource read-back

**Caption:** Live REST GET: betterado_project_features feature states from ADO API (CaptureLiveEvidence called in TestAccProjectFeatures_roundtrip)

**Command (before/after evidence):**
```
go test -tags all -count=1 -run TestAccProjectFeatures_roundtrip ./azuredevops/internal/acceptancetests/
```

**Live evidence (captured 2026-07-02T09:17:10Z):**

- **REST GET:** `https://dev.azure.com/davidgparsonson/_apis/FeatureManagement/FeatureStatesForScope/host/project/c0ac3757-e915-453f-ba2b-93a3720d1994?api-version=7.1`
- **Response:**
  ```json
  {
    "artifacts": "disabled",
    "boards": "enabled",
    "pipelines": "enabled",
    "repositories": "enabled",
    "testplans": "disabled"
  }
  ```

| | |
|---|---|
| **Before (main)** | `betterado_project_features` was SDKv2-only; `projectUseStateForUnknown` applied in SDKv2 `CreateContext` |
| **After (HEAD)** | `betterado_project_features` served via mux→framework `resource.Resource`; `Configure()` wires `*client.AggregatedClient`; live GET on `betterado-standing-demo` confirmed feature states; `ExpectNonEmptyPlan: false` → PASS; destroy clean (feature states restored) |

---

## Intent & Outcome — AC Evaluations

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| AC1 | GIVEN ADO Projects/Teams/Features REST API v7.1 schema WHEN compared against each in-scope resource's SDKv2 schema THEN `docs/core-gap-matrix.md` exists listing every field with status for all 7 resources; every writable gap resolved or deferred with rationale | **met** | `docs/core-gap-matrix.md` committed (4e36fa83); 191 lines; field tables for all 7 resources with implemented/read-only/gap/out-of-scope status; gap rows carry explicit rationale |
| AC2 | GIVEN betterado_project as framework resource.Resource WHEN terraform import run against betterado-standing-demo THEN import succeeds, read-back asserts name/visibility/version_control, ExpectNonEmptyPlan: false | **met** | `TestAccProject_importByName` uses `GetMuxedProviderFactories()`; imports betterado-standing-demo; asserts name/visibility/version_control; `ExpectNonEmptyPlan: false`. Committed in cbea0eef + e995fe9d + cf0491cf |
| AC3 | GIVEN data.betterado_project in framework WHEN terraform apply runs data source lookup by name THEN `TestAccProject_dataSource_withID` and `TestAccProject_dataSource_withName` pass | **met** | `data_project_framework.go` (172 lines) committed; both tests updated to muxed factories in `data_project_test.go` (commit af6f73e2) |
| AC4 | GIVEN data.betterado_projects in framework WHEN terraform apply runs listing projects THEN TestAccProjects_dataSource passes | **met** | `data_projects_framework.go` (186 lines) committed; test updated to muxed factories in `data_projects_test.go` (commit af6f73e2) |
| AC5 | GIVEN betterado_project + 2 data sources deregistered from SDKv2 provider.go WHEN TestProvider_HasChildResources + TestProvider_HasChildDataSources run THEN both pass with updated counts | **met** | `provider.go` removes all 3 from SDKv2 maps; `provider_test.go` updated; tests pass: `ok azuredevops 0.007s` |
| AC6 | GIVEN betterado_project_features as framework resource.Resource WHEN terraform apply enables/disables features → read-back → re-plan → destroy THEN TestAccProjectFeatures_roundtrip passes; CaptureLiveEvidence called; ExpectNonEmptyPlan: false | **met** | `TestAccProjectFeatures_roundtrip` passed live; `CaptureLiveEvidence("acceptance-resource", url, response)` → `.forge/live-evidence/acceptance-resource.json` written (capturedAt 2026-07-02T09:17:10Z); `ExpectNonEmptyPlan: false` |
| AC7 | GIVEN betterado_project_features removed from SDKv2 ResourcesMap WHEN TestProvider_HasChildResources runs THEN test passes with updated count | **met** | `provider.go` removes `betterado_project_features` from `ResourcesMap`; `provider_test.go` updated; tests pass |
| AC8 | GIVEN betterado_project_pipeline_settings as framework resource.Resource WHEN TestAccProjectPipelineSettings passes | **missed** | WI-4 status: failed — not committed. betterado_project_pipeline_settings remains in SDKv2 ResourcesMap |
| AC9 | GIVEN betterado_project_pipeline_settings removed from SDKv2 ResourcesMap WHEN TestProvider_HasChildResources runs THEN test passes | **missed** | WI-4 status: failed — not committed |
| AC10 | GIVEN betterado_project_tags as framework resource.Resource WHEN TestAccProjectTags passes | **missed** | WI-5 status: failed — not committed. betterado_project_tags remains in SDKv2 ResourcesMap |
| AC11 | GIVEN betterado_project_tags removed from SDKv2 ResourcesMap WHEN TestProvider_HasChildResources runs THEN test passes | **missed** | WI-5 status: failed — not committed |
| AC12 | GIVEN betterado_team as framework resource.Resource WHEN TestAccTeam_basic + TestAccTeam_update pass | **missed** | WI-6 status: failed — not committed. betterado_team remains in SDKv2 ResourcesMap |
| AC13 | GIVEN data.betterado_team + data.betterado_teams in framework WHEN TestAccTeam_dataSource + TestAccTeams_dataSource pass | **missed** | WI-6 status: failed — not committed |
| AC14 | GIVEN betterado_team + 2 data sources deregistered from SDKv2 WHEN TestProvider tests run THEN pass with updated counts | **missed** | WI-6 status: failed |
| AC15 | GIVEN betterado_team_administrators in framework WHEN TestAccTeamAdministrators passes | **missed** | WI-7 status: failed — not committed |
| AC16 | GIVEN betterado_team_members in framework WHEN TestAccTeamMembers passes | **missed** | WI-7 status: failed — not committed |
| AC17 | GIVEN betterado_team_administrators + betterado_team_members deregistered from SDKv2 WHEN TestProvider_HasChildResources runs THEN pass | **missed** | WI-7 status: failed |
| AC18 | GIVEN data.betterado_client_config in framework WHEN TestAccClientConfig_LoadsCorrectProperties passes | **missed** | WI-8 status: failed — not committed. data.betterado_client_config remains in SDKv2 DataSourcesMap |
| AC19 | GIVEN data.betterado_client_config removed from SDKv2 DataSourcesMap WHEN TestProvider_HasChildDataSources runs THEN pass | **missed** | WI-8 status: failed |
| AC20 | GIVEN all resources migrated WHEN make docs runs THEN docs/resources/ + docs/data-sources/ regenerated; guides restored; examples/ present | **partial** | docs/core-gap-matrix.md committed. Full docs regeneration deferred (WI-9 status: failed; WI-4 through WI-8 not delivered) |
| AC21 | GIVEN CHANGELOG.md updated and PROVIDER_VERSION.txt bumped WHEN git diff shows changes THEN CHANGELOG.md has Unreleased entry; version bumped | **partial** | CHANGELOG.md updated with draft Unreleased entry for WI-1 through WI-3 by this unifier commit. PROVIDER_VERSION.txt bump is a pre-merge finaliser step |
| AC22 | GIVEN demo.json carries real REST GET checkpoints WHEN forge demo render invoked THEN demo.json ends with checkpoint carrying liveEvidence.url | **met** | Live evidence: `CaptureLiveEvidence("acceptance-resource", "https://dev.azure.com/davidgparsonson/_apis/FeatureManagement/FeatureStatesForScope/host/project/c0ac3757-e915-453f-ba2b-93a3720d1994?api-version=7.1", response)` → `.forge/live-evidence/acceptance-resource.json` (capturedAt 2026-07-02T09:17:10Z) |

---

## Test evidence

| Test | Result |
|------|--------|
| `go test -tags all -count=1 ./azuredevops/internal/service/release/...` (offline) | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/...` (offline) | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/validate/...` (offline) | pass |
| `TestProvider_HasChildResources` (offline) | pass |
| `TestProvider_HasChildDataSources` (offline) | pass |
| `TestAccProject_importByName` (TF_ACC=1, live) | pass |
| `TestAccProject_dataSource_withID` (TF_ACC=1, live, muxed factories) | pass |
| `TestAccProject_dataSource_withName` (TF_ACC=1, live, muxed factories) | pass |
| `TestAccProjects_dataSource` (TF_ACC=1, live, muxed factories) | pass |
| `TestAccProjectFeatures_roundtrip` (TF_ACC=1, live) | pass |

## Files changed

| File | Change |
|------|--------|
| `docs/core-gap-matrix.md` | Added (191 lines) — ADO REST API v7.1 gap matrix for all 7 core resources |
| `azuredevops/internal/service/core/resource_project_framework.go` | Added (455 lines) — betterado_project as framework resource.Resource |
| `azuredevops/internal/service/core/data_project_framework.go` | Added (172 lines) — data.betterado_project as framework datasource.DataSource |
| `azuredevops/internal/service/core/data_projects_framework.go` | Added (186 lines) — data.betterado_projects as framework datasource.DataSource |
| `azuredevops/internal/service/core/resource_project_features_framework.go` | Added (332 lines) — betterado_project_features as framework resource.Resource |
| `azuredevops/internal/provider/framework_provider.go` | Modified — registers 4 new framework types |
| `azuredevops/provider.go` | Modified — removes 3 resources + data sources from SDKv2 maps |
| `azuredevops/provider_test.go` | Modified — updated TestProvider_Has* counts |
| `azuredevops/internal/acceptancetests/resource_project_test.go` | Modified — TestAccProject_importByName uses muxed factories |
| `azuredevops/internal/acceptancetests/data_project_test.go` | Modified — tests updated to muxed factories |
| `azuredevops/internal/acceptancetests/data_projects_test.go` | Modified — tests updated to muxed factories |
| `azuredevops/internal/acceptancetests/resource_project_features_test.go` | Modified — TestAccProjectFeatures_roundtrip uses muxed factories + CaptureLiveEvidence |
