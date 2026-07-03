# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_service_principal_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_service_principal_entitlement appears in the list and NOT in provider.go ResourcesMap
- [ ] AC2: GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccServicePrincipalEntitlement_create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up

## Completed (iter 0)

- [x] Created `resource_service_principal_entitlement_framework.go` with full CRUD + ImportState
- [x] Created `resource_service_principal_entitlement_framework_test.go` (unit test: TypeName assertion)
- [x] Added `NewServicePrincipalEntitlementResource` to `framework_provider.go` Resources()
- [x] Added `TestFrameworkProvider_HasServicePrincipalEntitlementResource` to `framework_provider_test.go`
- [x] Removed `betterado_service_principal_entitlement` from `provider.go` SDKv2 ResourcesMap
- [x] Dropped unused `memberentitlementmanagement` import from `provider.go`
- [x] Updated `provider_test.go` expectedResources list
- [x] Updated acceptance test: ProtoV6ProviderFactories, getDirectClient(), idempotency step, CaptureLiveEvidence
- [x] Added CHANGELOG [Unreleased] entry
- [x] All offline tests pass: TestNewServicePrincipalEntitlementResource_Metadata, TestFrameworkProvider_HasServicePrincipalEntitlementResource, TestProvider_HasChildResources
- [x] golangci-lint: 0 issues on changed code
- [x] go build: clean

## Remaining

- AC2 requires live gate: `TF_ACC=1 AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID=<uuid>` to run full acceptance test
