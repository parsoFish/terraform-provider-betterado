## Why

Both `betterado_dashboard` and `betterado_extension` were implemented as SDKv2 `*schema.Resource` resources. The project is migrating all resources to terraform-plugin-framework (served via the mux provider), which enables protocol v6, better plan modifiers, stronger type safety, and enables future schema evolution without the SDKv2 `GetRawConfig` workarounds. This initiative completes the migration for the dashboard and extension packages, and produces field-by-field gap matrices against the ADO REST API v7.1 for both resources.

## What

Files changed on this branch (`git diff --name-only main...HEAD`):

- `azuredevops/internal/service/dashboard/resource_dashboard_framework.go` — new terraform-plugin-framework implementation of `betterado_dashboard` (463 lines); Configure() wires `*client.AggregatedClient` from ProviderData; plan modifiers preserve ForceNew on `project_id`/`team_id`
- `azuredevops/internal/service/dashboard/resource_dashboard_framework_test.go` — unit tests for framework resource schema and plan modifiers (206 lines)
- `azuredevops/internal/service/extension/resource_extension_framework.go` — new terraform-plugin-framework implementation of `betterado_extension` (476 lines); `disabled` attribute uses `types.Bool` with `IsNull()` check to preserve SDKv2 `GetRawConfig` semantics; Configure() wires `*client.AggregatedClient`
- `azuredevops/internal/provider/framework_provider.go` — `Resources()` extended with `NewDashboardResource` and `NewExtensionResource` (+4 lines)
- `azuredevops/provider.go` — `betterado_dashboard` and `betterado_extension` removed from SDKv2 `ResourcesMap`; corresponding SDKv2 imports dropped (54 line delta)
- `azuredevops/provider_test.go` — `TestProvider_HasChildResources` expected list updated; `betterado_dashboard` and `betterado_extension` removed with comments that they are now framework resources (-6 lines)
- `azuredevops/internal/acceptancetests/resource_dashboard_test.go` — acceptance tests updated to use `GetMuxedProviderFactories()`; `CaptureLiveEvidence` call added before destroy (300 line delta)
- `azuredevops/internal/acceptancetests/resource_extension_test.go` — acceptance tests updated to use `GetMuxedProviderFactories()`; `CaptureLiveEvidence` call added before destroy (113 line delta)
- `docs/dashboard-gap-matrix.md` — new gap matrix: all 14 Dashboard struct fields mapped with mapped/missing/writable status and deferral rationale (+47 lines)
- `docs/extension-gap-matrix.md` — new gap matrix: all InstalledExtension struct fields of interest mapped (+49 lines)
- `docs/resources/dashboard.md` — regenerated via `make docs`; all attributes documented (+59 line delta)
- `docs/resources/extension.md` — regenerated via `make docs`; all attributes documented (+35 line delta)
- `examples/resources/betterado_dashboard/resource.tf` — new example with non-default values for `description` and `refresh_interval` (+25 lines)
- `examples/resources/betterado_extension/resource.tf` — new example with non-default values (+5 lines)
- `CHANGELOG.md` — draft `## [Unreleased]` entries for both resource migrations and gap matrices (+27 lines)
- `PROVIDER_VERSION.txt` — patch version bumped 1.2.0 → 1.2.1

## How

**WI-1 (Dashboard):** Implemented `resource_dashboard_framework.go` following the `resource_task_group_framework.go` pattern. `Configure()` asserts `req.ProviderData.(*client.AggregatedClient)` non-nil. Both project-scoped and team-scoped dashboard CRUD operations preserved. `betterado_dashboard` deregistered from SDKv2 `ResourcesMap` in the same commit to avoid `Duplicate resource type` at apply. Acceptance tests upgraded to `GetMuxedProviderFactories()`.

**WI-2 (Extension):** Implemented `resource_extension_framework.go`. The `disabled` field uses `types.Bool` with an `IsNull()` check rather than `GetRawConfig().AsValueMap()["disabled"]` to preserve the SDKv2 semantics that distinguish an explicitly-set `false` from an unset attribute. `betterado_extension` deregistered from SDKv2 `ResourcesMap` in the same commit.

**WI-3 (Release artefacts):** CHANGELOG.md updated under `## [Unreleased]`; PROVIDER_VERSION.txt incremented 1.2.0 → 1.2.1.

**Gap matrices:** Both `docs/dashboard-gap-matrix.md` and `docs/extension-gap-matrix.md` enumerate every ADO REST API v7.1 struct field with a mapped/missing/writable assessment and explicit deferral rationale for server-computed fields (ETag, ModifiedDate, ModifiedBy, LastAccessedDate, Widgets, GroupId).
