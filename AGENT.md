# Agent Memory — WI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

**Gate failure:** `TestAccWorkitemtrackingprocessInheritedControl_Revert` panicked:
```
panic: interface conversion: interface {} is nil, not *client.AggregatedClient
```
at `resource_workitemtrackingprocess_inherited_control_test.go:236`.

**Root cause:** The test used `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`.
When using the muxed provider factory, the SDKv2 provider singleton (the package-level `provider` var
in `testutils/commons.go`) is NEVER configured. Its `Meta()` returns nil.
`testutils.GetProvider().Meta().(*client.AggregatedClient)` panics on the nil interface cast.

The SAME bug existed in `TestAccWorkitemtrackingprocessSystemControl_Revert`'s `testCheckSystemControlReverted` function.

**Fix applied (committed 70b23a11):**
- Added `getInheritedControlDirectClient()` to `resource_workitemtrackingprocess_inherited_control_test.go`
- Added `getSystemControlDirectClient()` to `resource_workitemtrackingprocess_system_control_test.go`
- Both build `AggregatedClient` from AZDO env vars (same pattern as `CheckProcessDestroyed` and `getDirectClient` in `resource_task_group_test.go`)
- Updated `checkInheritedControlRevertedFunc` and `testCheckSystemControlReverted` to use the direct client
- Added `captureGroupEvidence()` helper and wired into `TestAccWorkitemtrackingprocessGroup_Basic` (AC6)

### Iteration 1

**Gate failure:** `TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes` failed:
```
TF1590010: Extension  is already installed in this organization.
  with betterado_extension.test (on terraform_plugin_test.tf line 26)
```
The extension `ms-devlabs/vsts-extensions-multivalue-control` was left installed from a prior run.
The test created a `betterado_extension` Terraform resource — which fails if extension is already installed.
`TestAccExtension_requireImportError` explicitly tests that TF1590010 IS returned, so we cannot make
`resourceExtensionCreate` idempotent without breaking that test.

**Fix applied (committed b635c9c0):**
- Added `testutils.EnsureExtensionInstalled(t, publisher, extension)` to `testutils/extension.go`:
  idempotent install via direct API call (handles "already installed" gracefully).
- Added `testutils.EnsureExtensionUninstalled(t, publisher, extension)` for cleanup.
- Modified `TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes`:
  - `PreCheck` calls `EnsureExtensionInstalled` instead of creating `betterado_extension` resource.
  - `CheckDestroy` calls `EnsureExtensionUninstalled` then `CheckProcessDestroyed`.
  - Removed `betterado_extension` resource from `groupWithMultipleControlTypes` HCL.
- Same fix for `TestAccWorkitemtrackingprocessControl_Contribution` and `contributionControl` HCL.
- Added `testutils.GetDirectClient()` to `testutils/process.go` (exported version of `getDirectClient`).
- Updated `resource_workitemtrackingprocess_page_test.go` and `_inherited_page_test.go` to call
  `testutils.GetDirectClient()` (fixes pre-existing build failure: `getDirectClient` was in a
  tagged file but referenced from untagged files).
- Added `//nolint:nilerr` to best-effort nil returns in page/state/workitemtype tests.
- Added `//nolint:unused` to `disabledProcess` in shared test file (used only with `all` tag).
- Fixed gofumpt formatting in `resource_{inherited_,}page_framework.go` and
  `resource_{inherited_,}state_framework.go`.

## What worked

- Direct client pattern: build `client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)` from env vars — this is the established pattern for tests using muxed provider factories.
- `//nolint:nilerr` annotation for intentional best-effort nil returns.
- Using `PreCheck` to manage external prerequisites (extension installation) instead of Terraform resources — avoids flakiness from parallel runs and leftover state.

## What didn't work

- Making `resourceExtensionCreate` idempotent to handle TF1590010 — would break `TestAccExtension_requireImportError` which explicitly expects the error.

## Key architecture notes

- `testutils.GetMuxedProviderFactories()`: creates a fresh SDKv2 `azuredevops.Provider()` internally each time the factory is called. The package-level `provider` var in `testutils/commons.go` is NEVER configured when using this factory.
- `testutils.GetProvider().Meta()` is ONLY valid when using `ProviderFactories` (not `ProtoV6ProviderFactories`).
- The SDKv2 resources (group, control, inherited_control, system_control) are still registered in `provider.go`. They ARE served through the mux when using `GetMuxedProviderFactories()`, so acceptance tests for those resources work even before framework migration.
- AC5 requires actual framework migration (new `*_framework.go` files, registration in `framework_provider.go`, deletion of SDKv2 files, provider_test.go count update).
- `TestAccExtension_requireImportError` explicitly tests that TF1590010 IS returned when you define a duplicate extension resource — do NOT change extension Create to be idempotent.

## Open questions

- Are there any other live test failures after the TF1590010 fix? (Need gate result from iteration 1.)
- AC1-AC5 (actual framework migration) may still need to be done — the resources are currently served as SDKv2 through the mux, so acceptance tests may pass, but AC5 requires explicit framework migration.

## Notes for reflection

- Tests that depend on external resources (extensions, etc.) should manage those via `PreCheck`/`CheckDestroy` rather than as Terraform resources to avoid flakiness from parallel execution and leftover state.
- `getDirectClient` should always be defined in untagged files or exported from `testutils` to avoid build failures with selective build tags.
