# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

Full implementation in one iteration (no prior work). All offline gates green.

**Files created/modified:**
- `azuredevops/internal/service/memberentitlementmanagement/resource_service_principal_entitlement_framework.go` — new framework resource (ServicePrincipalEntitlementResource, NewServicePrincipalEntitlementResource)
- `azuredevops/internal/service/memberentitlementmanagement/resource_service_principal_entitlement_framework_test.go` — unit test (TypeName assertion)
- `azuredevops/internal/provider/framework_provider.go` — added NewServicePrincipalEntitlementResource to Resources()
- `azuredevops/internal/provider/framework_provider_test.go` — added TestFrameworkProvider_HasServicePrincipalEntitlementResource
- `azuredevops/provider.go` — removed betterado_service_principal_entitlement from SDKv2 ResourcesMap; dropped unused memberentitlementmanagement import
- `azuredevops/provider_test.go` — removed betterado_service_principal_entitlement from expectedResources list
- `azuredevops/internal/acceptancetests/resource_service_principal_entitlement_test.go` — switched to ProtoV6ProviderFactories/GetMuxedProviderFactories, getDirectClient(), idempotency re-plan step, CaptureLiveEvidence
- `CHANGELOG.md` — added [Unreleased] FEATURES entry

## What worked

- Followed the exact same pattern as WI-2 (user_entitlement) and WI-3 (group_entitlement) framework migrations.
- Re-used the `geStringDefault` helper from group_entitlement_framework.go (same package, accessible without new type).
- `getDirectClient()` for check helpers (defined in resource_task_group_test.go, same acceptancetests package).
- `CaptureLiveEvidence` from `testutils` package, label `"acceptance-resource-sp-entitlement"`.
- Plan modifiers: `speRequiresReplace()` for `origin_id` and `origin`; `speStateForUnknown()` for computed fields.
- `isServicePrincipalEntitlementDeleted()` checks `accounts.AccountUserStatusValues.Deleted` and `.None`.
- API calls use `r.client.MemberEntitleManagementClient.{Add,Get,Update,Delete}ServicePrincipalEntitlement`.
- Update uses JSON-patch on `/accessLevel` (same pattern as user/group entitlement).
- Build tag on acceptance test: `//go:build (all || resource_service_principal_entitlement) && !exclude_resource_service_principal_entitlement`

## What didn't work

_(nothing — clean first-attempt implementation)_

## Open questions

- AC2 requires live gate with `TF_ACC=1 AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID=<uuid>`. If it fails live, check: (1) does the `origin` field get populated from the API (it's optional+computed with speStateForUnknown — should be fine), (2) does the flatten correctly handle case differences in licensing_source (using strings.ToLower).

## Notes for reflection

- The memberentitlementmanagement import is now fully removed from provider.go (all 3 resources migrated in WI-2, WI-3, WI-4). Worth noting in brain that the migration initiative is complete for memberentitlementmanagement.
