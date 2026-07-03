# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_group_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_group_entitlement appears in the list and NOT in provider.go ResourcesMap
- [ ] AC2: GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccGroupEntitlement_Create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up

## AC1 — DONE (iter 0)

All changes committed in ce56fcb4:
- `resource_group_entitlement_framework.go` — full CRUD + ImportState + inline plan modifiers + geStringDefault
- `framework_provider.go` — added `NewGroupEntitlementResource`
- `provider.go` — removed SDKv2 registration (same commit to avoid "Duplicate resource type")
- `provider_test.go` — removed from expected resources list
- `framework_provider_test.go` — added `TestFrameworkProvider_HasGroupEntitlementResource`
- `resource_group_entitlement_framework_test.go` — unit test `TestNewGroupEntitlementResource_Metadata` passes
- Acceptance test updated: `ProtoV6ProviderFactories`, `getDirectClient()`, idempotency step, `CaptureLiveEvidence("acceptance-resource-group-entitlement", ...)`
- Fixed gofumpt on `resource_user_entitlement_framework.go` (pre-existing lint debt)

## AC2 — PENDING (live gate)

AC2 requires TF_ACC=1 against real ADO. Cannot be verified offline. The quality_gate_cmd is:

  go test -tags all -run TestAccGroupEntitlement_Create ./azuredevops/internal/acceptancetests/

This will be verified by the orchestrator's live gate run.
