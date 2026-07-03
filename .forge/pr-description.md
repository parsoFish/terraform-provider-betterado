## Why

The `workitemtrackingprocess` package contained 13 resources and 4 data sources that still used the deprecated Terraform Plugin SDK v2 (SDKv2) machinery. Keeping them in SDKv2 blocked the provider from moving fully to Plugin Framework, which is the strategic target for all future Terraform provider development at betterado. This migration closes the largest remaining SDKv2 surface in the provider (process management: processes, work item types, states, pages, groups, controls, lists, fields, and rules).

## What

- **13 resources migrated** to terraform-plugin-framework with new `*_framework.go` implementations registered in `framework_provider.go` (and deregistered from `provider.go`):
  - `betterado_workitemtrackingprocess_process`
  - `betterado_workitemtrackingprocess_workitemtype`
  - `betterado_workitemtrackingprocess_state`
  - `betterado_workitemtrackingprocess_inherited_state`
  - `betterado_workitemtrackingprocess_page`
  - `betterado_workitemtrackingprocess_inherited_page`
  - `betterado_workitemtrackingprocess_list`
  - `betterado_workitemtrackingprocess_field`
  - `betterado_workitemtrackingprocess_rule`
  - `betterado_workitemtrackingprocess_control`
  - `betterado_workitemtrackingprocess_group`
  - `betterado_workitemtrackingprocess_inherited_control`
  - `betterado_workitemtrackingprocess_system_control`
- **4 data sources migrated** to terraform-plugin-framework:
  - `betterado_workitemtrackingprocess_process`
  - `betterado_workitemtrackingprocess_processes`
  - `betterado_workitemtrackingprocess_workitemtype`
  - `betterado_workitemtrackingprocess_workitemtypes`
- **`docs/workitemtrackingprocess-gap-matrix.md`** produced — 413-line table comparing all 13 resources and 4 data sources against the ADO Work Item Tracking Process REST API v7.1; every field marked `resolved`, `deferred`, or `n/a-computed`.
- **Acceptance tests** updated to `testutils.GetMuxedProviderFactories()` for all 17 resource/data-source test files; `CaptureLiveEvidence` calls added per resource for live REST evidence.
- **Registry docs regenerated** (`make docs` → `git checkout -- docs/guides/`) and HCL examples added under `examples/resources/workitemtrackingprocess_*/` and `examples/data-sources/betterado_workitemtrackingprocess_*/`.
- **CHANGELOG.md** updated under `## [Unreleased]` with ENHANCEMENTS entry; `PROVIDER_VERSION.txt` bumped.
- **Provider counts** in `provider_test.go` updated to reflect all 17 types moving from SDKv2 to framework.

Key files changed:
- `azuredevops/internal/service/workitemtrackingprocess/resource_*_framework.go` (13 new framework resource files)
- `azuredevops/internal/service/workitemtrackingprocess/data_*_framework.go` (4 new framework data source files)
- `azuredevops/internal/provider/framework_provider.go` (17 new registrations)
- `azuredevops/provider.go` (17 deregistrations from SDKv2 maps)
- `azuredevops/provider_test.go` (counts updated)
- `azuredevops/internal/acceptancetests/resource_workitemtrackingprocess_*_test.go` (GetMuxedProviderFactories + CaptureLiveEvidence)
- `docs/workitemtrackingprocess-gap-matrix.md` (new, 413 lines)
- `docs/resources/workitemtrackingprocess_*.md`, `docs/data-sources/workitemtrackingprocess_*.md` (regenerated)
- `examples/resources/betterado_workitemtrackingprocess_*/resource.tf` (13 new examples)
- `examples/data-sources/betterado_workitemtrackingprocess_*/main.tf` (4 new examples)
- `CHANGELOG.md`, `PROVIDER_VERSION.txt`

## How

Each resource was migrated following the project's framework migration checklist:

1. **New `*_framework.go` file** in `azuredevops/internal/service/workitemtrackingprocess/` implementing `resource.Resource` or `datasource.DataSource`. SDKv2 validators were reproduced as framework `Validators:` entries; `ForceNew` fields became `RequiresReplace()` plan modifiers.
2. **Registered** in `azuredevops/internal/provider/framework_provider.go` `Resources()` / `DataSources()` slices.
3. **Deregistered** from `azuredevops/provider.go` `ResourcesMap` / `DataSourcesMap` (with explanatory comments).
4. **Acceptance tests** updated to use `testutils.GetMuxedProviderFactories()` and include a `CaptureLiveEvidence` call that writes `.forge/live-evidence/<label>.json` during the live read-back step before destroy.
5. Resources were delivered in dependency order: process (WI-2) → workitemtype (WI-3) → state/inherited_state (WI-4) → page/inherited_page (WI-5) → group/control/inherited_control/system_control (WI-6) → list/field (WI-7) → rule (WI-8).
6. The `rule` resource required an extra fix (`SetNestedBlock` for condition/action nested blocks) to support HCL block syntax at apply time.
7. WI-9 closed out with `make docs` regeneration, example creation, changelog entry, and version bump.

The offline gate (`go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`) passes green on HEAD. The provider compiles cleanly (`go build -mod=vendor . → exit 0`). Live TF_ACC gates ran per-WI against real ADO; every named `TestAcc*` test passed.
