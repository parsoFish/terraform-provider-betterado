## Why

ADO Marketplace extension lifecycle (install/uninstall/read-back) was only possible through the legacy SDKv2 `betterado_extension` resource. That resource pre-dates the framework migration and is not mux-free-ready. Operators who want to manage extension installations as idiomatic Terraform resources (with full plan/apply/destroy lifecycle) had no framework-native option. This initiative adds `betterado_extension_install` — a terraform-plugin-framework resource — so new configurations can adopt the framework path without depending on the SDKv2 mux, and existing users have a clear migration target documented in the gap matrix.

## What

- **`docs/gallery-extensionmanagement-gap-matrix.md`** — endpoint-by-endpoint mapping of ADO Gallery (`_apis/gallery`) and ExtensionManagement (`_apis/extensionmanagement`) v7.1 against the provider's current coverage; explicitly resolves the boundary between `betterado_extension` (existing SDKv2) and `betterado_extension_install` (new framework); triages `betterado_extension_settings` and `betterado_marketplace_extension` with rationale.
- **`azuredevops/internal/service/extensionmanagement/resource_extension_install_framework.go`** — `terraform-plugin-framework` `resource.Resource` with Create/Read/Update/Delete using `ExtensionManagementClient`; schema: `publisher_id` (Required, RequiresReplace), `extension_id` (Required, RequiresReplace), `version` (Computed), `disabled` (Optional); 404-in-Read removes resource from state gracefully.
- **`azuredevops/internal/service/extensionmanagement/resource_extension_install_framework_test.go`** — unit tests (`TestExtensionInstallResource_ExpandFlatten`) proving expand→flatten round-trip data integrity for all schema fields.
- **`azuredevops/internal/acceptancetests/resource_extension_install_test.go`** — live TF_ACC acceptance test (`TestAccExtensionInstall_basic`): apply → read-back (`publisher_id`, `extension_id`, `version` set) → idempotency re-plan (`ExpectNonEmptyPlan: false`) → destroy; calls `CaptureLiveEvidence` with real ExtensionManagement REST GET URL.
- **`azuredevops/internal/provider/framework_provider.go`** — registers `extensionmanagement.NewExtensionInstallResource` in `Resources()`; **zero** changes to `azuredevops/provider.go` (SDKv2 maps untouched — AC-4 framework-only registration).
- **`azuredevops/internal/provider/framework_provider_test.go`** — `TestFrameworkProvider_HasExtensionInstallResource` confirms `betterado_extension_install` appears in `Resources()`.
- **`docs/resources/extension_install.md`** — tfplugindocs-generated registry docs for all schema attributes.
- **`examples/resources/betterado_extension_install/resource.tf`** — canonical HCL example embedded in registry docs.
- **`CHANGELOG.md`** — draft entry under `## [Unreleased]` for `betterado_extension_install`.
- **`PROVIDER_VERSION.txt`** — bumped to next semver.

## How

1. **WI-1** wrote the gap matrix and a stub `gap_matrix_test.go` (build tag `all || gallery`) that asserts the file exists and contains all three candidate resource names — ensures the gate fails on a clean tree and passes after the doc is present.
2. **WI-2** implemented the framework resource following the `resource_task_group_framework.go` pattern: `ExtensionInstallResource` struct, `NewExtensionInstallResource()` factory, CRUD methods via `clients.ExtensionManagementClient`, 404-in-Read guard (`resp.State.RemoveResource`), and expand/flatten helpers tested by `TestExtensionInstallResource_ExpandFlatten`. No `framework_provider.go` changes in this WI (shape discipline).
3. **WI-3** added the live acceptance test mirroring `resource_task_group_test.go`: uses `ProtoV6ProviderFactories`, `getDirectClient()` for `CheckDestroy` and evidence capture, and calls `testutils.CaptureLiveEvidence("acceptance-resource-extension-install", ...)` with the ExtensionManagement REST URL so `forge demo render` back-fills the real API response into `demo.json`.
4. **WI-3 gate fix** (commit `7b928209`) registered `extensionmanagement.NewExtensionInstallResource` in `framework_provider.go` — an honest deviation from the planned WI-4-only shape rule, required to unblock the WI-3 live gate. **WI-4** added the provider-level registration test (`TestFrameworkProvider_HasExtensionInstallResource`), ran `make docs` + `git checkout -- docs/guides/`, and updated `CHANGELOG.md` and `PROVIDER_VERSION.txt`. Confirmed with `grep -n 'extension_install\|marketplace_extension' azuredevops/provider.go` → 0 matches.

Changed files (from `git diff --name-only main...HEAD`):
- `CHANGELOG.md`
- `PROVIDER_VERSION.txt`
- `azuredevops/internal/acceptancetests/resource_extension_install_test.go`
- `azuredevops/internal/provider/framework_provider.go`
- `azuredevops/internal/provider/framework_provider_test.go`
- `azuredevops/internal/service/extensionmanagement/resource_extension_install_framework.go`
- `azuredevops/internal/service/extensionmanagement/resource_extension_install_framework_test.go`
- `azuredevops/internal/service/gallery/gap_matrix_test.go`
- `docs/gallery-extensionmanagement-gap-matrix.md`
- `docs/resources/extension_install.md`
- `examples/resources/betterado_extension_install/resource.tf`
- `forge/history/INIT-2026-07-01-new-api-gallery-extensionmanagement/demo/demo.json`
- `forge/history/INIT-2026-07-01-new-api-gallery-extensionmanagement/demo/DEMO.md`
