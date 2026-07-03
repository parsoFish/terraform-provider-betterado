# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN betterado_workitemtrackingprocess_group resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessGroup_Basic, TestAccWorkitemtrackingprocessGroup_Update, TestAccWorkitemtrackingprocessGroup_Move, and TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes all pass
- [ ] AC2: GIVEN betterado_workitemtrackingprocess_control resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessControl_Basic, TestAccWorkitemtrackingprocessControl_Update, TestAccWorkitemtrackingprocessControl_Move, and TestAccWorkitemtrackingprocessControl_Contribution all pass
- [ ] AC3: GIVEN betterado_workitemtrackingprocess_inherited_control resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessInheritedControl_Basic, TestAccWorkitemtrackingprocessInheritedControl_Update, and TestAccWorkitemtrackingprocessInheritedControl_Revert all pass
- [ ] AC4: GIVEN betterado_workitemtrackingprocess_system_control resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessSystemControl_Basic, TestAccWorkitemtrackingprocessSystemControl_Update, and TestAccWorkitemtrackingprocessSystemControl_Revert all pass
- [ ] AC5: GIVEN all four SDKv2 resource files and their unit test files removed and deregistered WHEN provider.go ResourcesMap inspected THEN group, control, inherited_control, system_control registered ONLY in framework_provider.go; orphaned SDKv2 files deleted; provider_test.go resource count updated
- [x] AC6: GIVEN live acceptance test running WHEN group resource is read back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-group", url, apiResponse) is called in TestAccWorkitemtrackingprocessGroup_Basic

## Iteration 0 status

**Blocker fixed (iteration 0):**
- Fixed nil panic in `TestAccWorkitemtrackingprocessInheritedControl_Revert` and `TestAccWorkitemtrackingprocessSystemControl_Revert`:
  - Both used `ProtoV6ProviderFactories` (muxed) but called `GetProvider().Meta()` which is nil in mux context
  - Fix: use `getInheritedControlDirectClient()` / `getSystemControlDirectClient()` to build client from AZDO env vars
- Added `captureGroupEvidence()` helper to `TestAccWorkitemtrackingprocessGroup_Basic` (AC6)

## Iteration 1 status

**Blocker fixed (iteration 1):**
- Fixed `TF1590010: Extension is already installed` flakiness in `TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes`:
  - Extension `ms-devlabs/vsts-extensions-multivalue-control` left installed from prior run
  - Cannot make `resourceExtensionCreate` idempotent (would break `TestAccExtension_requireImportError`)
  - Fix: use `testutils.EnsureExtensionInstalled/Uninstalled` in PreCheck/CheckDestroy instead of Terraform resource
  - Same fix applied to `TestAccWorkitemtrackingprocessControl_Contribution`
- Fixed pre-existing build failure: `getDirectClient` referenced from untagged test files but defined in tagged file
  - Added `testutils.GetDirectClient()` exported helper
  - Updated page test files to use it
- Fixed nilerr, unused, gofumpt lint issues in changed/new files

**Still pending:**
- AC1-AC5: The resources (group, control, inherited_control, system_control) are still SDKv2 in provider.go.
  The acceptance tests use `GetMuxedProviderFactories()`, which serves SDKv2 resources through the mux,
  so they MAY pass already without framework migration. But AC5 requires actual migration.
- Framework migration: create `resource_group_framework.go`, `resource_control_framework.go`,
  `resource_inherited_control_framework.go`, `resource_system_control_framework.go`;
  register in framework_provider.go; deregister from provider.go; delete SDKv2 files.
