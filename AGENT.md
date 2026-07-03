# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (first iteration — completed full implementation)

Implemented the complete framework migration for `betterado_group_entitlement`:

**Files written/modified:**
- `azuredevops/internal/service/memberentitlementmanagement/resource_group_entitlement_framework.go` — NEW
  Full CRUD + ImportState. Key design:
  - `GroupEntitlementResource` struct with `*client.AggregatedClient` configured via `Configure()` reading from `req.ProviderData.(*client.AggregatedClient)`.
  - Inline plan modifiers `gePlanRequiresReplace` / `geUseStateForUnknown` (same pattern as `resource_user_entitlement_framework.go` — no `stringplanmodifier` package in vendor).
  - Inline `geStaticStringDefault` implementing `defaults.String` (same pattern as `release/framework_defaults.go` `staticStringDefault` — no `stringdefault` package in vendor).
  - `geStringDefault(string(licensing.AccountLicenseTypeValues.Express))` for `account_license_type`.
  - `geStringDefault(string(licensing.LicensingSourceValues.Account))` for `licensing_source`.
  - Delete: calls `DeleteGroupEntitlement`, then if `origin == "vsts"` also calls `GraphClient.DeleteGroup` (matches SDKv2 logic).
  - ImportState: UUID only (matches SDKv2).

- `azuredevops/internal/service/memberentitlementmanagement/resource_group_entitlement_framework_test.go` — NEW
  Unit test `TestNewGroupEntitlementResource_Metadata` asserting TypeName == "betterado_group_entitlement".

- `azuredevops/internal/provider/framework_provider.go` — MODIFIED
  Added `memberentitlementmanagement.NewGroupEntitlementResource` to `Resources()`.

- `azuredevops/internal/provider/framework_provider_test.go` — MODIFIED
  Added `TestFrameworkProvider_HasGroupEntitlementResource`.

- `azuredevops/provider.go` — MODIFIED
  Removed `"betterado_group_entitlement": memberentitlementmanagement.ResourceGroupEntitlement()` from SDKv2 ResourcesMap. Added comment. SAME COMMIT as framework registration to avoid "Duplicate resource type betterado_group_entitlement" at apply.

- `azuredevops/provider_test.go` — MODIFIED
  Removed `"betterado_group_entitlement"` from the `expectedResources` list. Added comment following the release-resource pattern.

- `azuredevops/internal/acceptancetests/resource_group_entitlement_test.go` — MODIFIED
  - Added build tag `//go:build (all || resource_group_entitlement) && !exclude_resource_group_entitlement`.
  - `TestAccGroupEntitlement_Create`: changed `Providers: testutils.GetProviders()` to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`. Added idempotency step (`PlanOnly: true, ExpectNonEmptyPlan: false`). Added `captureGroupEntitlementEvidence(tfNode)` check.
  - `checkGroupEntitlementExists()` / `checkGroupEntitlementDestroyed()`: changed from `testutils.GetProvider().Meta().(*client.AggregatedClient)` to `getDirectClient()` (required for ProtoV6ProviderFactories — no SDKv2 Meta singleton).
  - Added `captureGroupEntitlementEvidence(tfNode)` function using `testutils.CaptureLiveEvidence("acceptance-resource-group-entitlement", ...)`.

- `azuredevops/internal/service/memberentitlementmanagement/resource_user_entitlement_framework.go` — gofumpt fix only (pre-existing lint debt, surfaced by `--new-from-rev=main`).

**Build status:** `go build ./...` — clean.
**Unit tests:** all pass (`TestNewGroupEntitlementResource_Metadata`, `TestFrameworkProvider_HasGroupEntitlementResource`, existing memberentitlementmanagement tests).
**golangci-lint --new-from-rev=main:** 0 issues.
**gofmt / gofumpt:** clean.
**Offline gate:** The acceptance test `TestAccGroupEntitlement_Create` without TF_ACC silently skips — this is expected. Live gate (with TF_ACC=1) is pending orchestrator.

## What worked

- Vendor has `github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults` and `resource/schema/planmodifier` as interface packages — but NOT `stringdefault` or `stringplanmodifier` as top-level packages. Must implement inline.
- Pattern: implement `geStaticStringDefault` inline (same as `release/framework_defaults.go` `staticStringDefault`).
- Pattern: implement `gePlanRequiresReplace`/`geUseStateForUnknown` inline (same as user entitlement resource).
- Remove from SDKv2 AND add to framework in the SAME commit (per WI-3 mandatory checklist item 1). Otherwise: "Duplicate resource type betterado_group_entitlement" at apply.
- Acceptance test: `getDirectClient()` (defined in `resource_task_group_test.go`) must be used instead of `testutils.GetProvider().Meta()` when using `ProtoV6ProviderFactories`.

## What didn't work

- `stringplanmodifier` and `stringdefault` packages are NOT in vendor. Using them fails `go build` with "import lookup disabled by -mod=vendor". Must implement inline.

## Open questions

_(none)_

## Notes for reflection

- The `getDirectClient()` pattern (build ADO client from env vars, not Meta singleton) is now the standard for all framework-migrated acceptance tests.
- The inline `geStaticStringDefault` / `geRequiresReplace` / `geStateForUnknown` helpers are duplicated from the user entitlement resource. A shared `memberentitlementmanagement` internal helper package could DRY these up, but out of scope here.
