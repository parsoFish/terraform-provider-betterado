## Why

The `workitemtracking` package contained six resources/data-sources still implemented in terraform-plugin-sdk/v2: `betterado_workitem`, `betterado_workitemtracking_field`, `betterado_workitemquery`, `betterado_workitemquery_folder`, `betterado_area`, and `betterado_iteration`. Keeping SDKv2 implementations alongside the framework mux increases maintenance surface, prevents using framework-only features (typed state, plan modifiers, write-only attributes), and blocks eventual removal of the SDKv2 dependency. This initiative migrates all six types so the workitemtracking surface is 100% framework.

## What

- **`betterado_workitem`** — new `resource_workitem_framework.go`; SDKv2 `resource_workitem.go` deleted; `TestAccWorkItem_basic` updated to `GetMuxedProviderFactories` + `CaptureLiveEvidence("acceptance-resource-workitem", ...)`.
- **`betterado_workitemtracking_field`** — new `resource_field_framework.go`; SDKv2 `resource_field.go` deleted; `TestAccWorkItemTrackingField_Basic` updated; `CaptureLiveEvidence("acceptance-resource-workitemtracking-field", ...)` called.
- **`betterado_workitemquery`** — new `resource_workitemquery_framework.go`; SDKv2 `resource_workitemquery.go` and `resource_workitemquery_test.go` deleted; `TestAccWorkItemQuery_UnderArea` updated; `CaptureLiveEvidence("acceptance-resource-workitemquery", ...)`.
- **`betterado_workitemquery_folder`** — new `resource_workitemquery_folder_framework.go`; SDKv2 `resource_workitemquery_folder.go` and `resource_workitemquery_folder_test.go` deleted; `TestAccWorkItemQueryFolder_UnderArea` updated; `CaptureLiveEvidence("acceptance-resource-workitemquery-folder", ...)`.
- **`betterado_area`** — new `data_area_framework.go`; SDKv2 `data_area.go` deleted; `TestAccAreaDataSource_Read` written with `GetMuxedProviderFactories`; `CaptureLiveEvidence("acceptance-resource-area", ...)`.
- **`betterado_iteration`** — new `data_iteration_framework.go`; SDKv2 `data_iteration.go` deleted; `TestAccIterationDataSource_Read` written with `GetMuxedProviderFactories`; `CaptureLiveEvidence("acceptance-resource-iteration", ...)`.
- **Provider wiring** — all six types deregistered from `provider.go` ResourcesMap/DataSourcesMap; added to `framework_provider.go` Resources()/DataSources(); `provider_test.go` counts updated.
- **Docs** — `docs/workitemtracking-gap-matrix.md` added; six resource/data-source docs regenerated via `make docs`; `docs/guides/` restored; six example `.tf` files created under `examples/`.
- **Release** — `CHANGELOG.md` `## [Unreleased]` section updated; `PROVIDER_VERSION.txt` bumped to `1.9.1` (from `1.9.0` on main).

Files changed (from `git diff --name-only main...HEAD`):

- `.forge/project.json`
- `CHANGELOG.md`
- `PROVIDER_VERSION.txt`
- `azuredevops/internal/acceptancetests/data_area_test.go`
- `azuredevops/internal/acceptancetests/data_iteration_test.go`
- `azuredevops/internal/acceptancetests/resource_workitem_test.go`
- `azuredevops/internal/acceptancetests/resource_workitemquery_folder_test.go`
- `azuredevops/internal/acceptancetests/resource_workitemquery_test.go`
- `azuredevops/internal/acceptancetests/resource_workitemtracking_field_test.go`
- `azuredevops/internal/provider/framework_provider.go`
- `azuredevops/internal/service/workitemtracking/data_area.go` (deleted)
- `azuredevops/internal/service/workitemtracking/data_area_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/data_iteration.go` (deleted)
- `azuredevops/internal/service/workitemtracking/data_iteration_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/resource_field.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_field_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/resource_workitem.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_workitem_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/resource_workitem_test.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery_folder.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery_folder_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery_folder_test.go` (deleted)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery_framework.go` (new)
- `azuredevops/internal/service/workitemtracking/resource_workitemquery_test.go` (deleted)
- `azuredevops/provider.go`
- `azuredevops/provider_test.go`
- `docs/data-sources/area.md`
- `docs/data-sources/iteration.md`
- `docs/resources/workitem.md`
- `docs/resources/workitemquery.md`
- `docs/resources/workitemquery_folder.md`
- `docs/resources/workitemtracking_field.md`
- `docs/workitemtracking-gap-matrix.md`
- `examples/data-sources/betterado_area/data-source.tf`
- `examples/data-sources/betterado_iteration/data-source.tf`
- `examples/resources/betterado_workitem/resource.tf`
- `examples/resources/betterado_workitemquery/resource.tf`
- `examples/resources/betterado_workitemquery_folder/resource.tf`
- `examples/resources/betterado_workitemtracking_field/resource.tf`
- `forge/history/INIT-2026-07-01-migrate-framework-workitemtracking/demo/demo.json`
- `forge/history/INIT-2026-07-01-migrate-framework-workitemtracking/demo/DEMO.md`

## How

Each resource/data-source was migrated in its own WI (WI-2 through WI-5), serialised by dependency edges to prevent hidden-coupling violations on shared files (`framework_provider.go`, `provider.go`, `provider_test.go`):

1. **WI-1** — produced `docs/workitemtracking-gap-matrix.md`: field-by-field API vs schema comparison for all 6 types; writable gaps resolved or explicitly deferred.
2. **WI-2** — migrated `betterado_workitem`: new framework resource with typed state, `RequiresReplace()` plan modifiers, UUID/non-whitespace validators; SDKv2 file deleted; acceptance test updated to mux factory + `CaptureLiveEvidence`.
3. **WI-3** — migrated `betterado_workitemtracking_field`: framework resource preserving `WriteOnly: true` (`restore` field), computed attributes (`can_sort_by`, `is_queryable`, `is_identity`, `is_picklist`, `supported_operations`), identity-type override in `Read`; SDKv2 file deleted.
4. **WI-4** — migrated `betterado_workitemquery` and `betterado_workitemquery_folder` together (shared `ExactlyOneOf(parent_id, area)` constraint via `resource.ConfigValidator`; both share the same provider files); SDKv2 files and unit test files deleted.
5. **WI-5** — migrated `betterado_area` and `betterado_iteration` data sources together (shared `utils/classification.go` helper; framework-native `classification_framework.go` written); SDKv2 files deleted; new acceptance tests written.
6. **WI-6** — `make docs` regenerated six docs pages; `docs/guides/` restored; `CHANGELOG.md` updated; `PROVIDER_VERSION.txt` bumped to `1.9.1`.

All migrations followed the project's framework migration checklist: mux registration, `Configure()` wiring `*client.AggregatedClient` from `ProviderData`, `GetMuxedProviderFactories()` in acceptance tests, `SharedFixtureProjectName` (no ad-hoc project creation), `CaptureLiveEvidence` during live read-back. The CI-equivalent gate (`go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`) is green on branch HEAD.
