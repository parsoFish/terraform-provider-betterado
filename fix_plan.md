# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_user_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_user_entitlement appears in the list and NOT in provider.go ResourcesMap
  - [x] Created resource_user_entitlement_framework.go with full CRUD + ImportState
  - [x] Registered NewUserEntitlementResource in framework_provider.go Resources()
  - [x] Removed ResourceUserEntitlement() from provider.go ResourcesMap (same commit — avoids duplicate resource error)
  - [x] Updated provider_test.go expectedResources (decremented count by 1)
  - [x] Added TestFrameworkProvider_HasUserEntitlementResource to framework_provider_test.go
  - [x] Unit test TestNewUserEntitlementResource_Metadata passes
- [ ] AC2: GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccUserEntitlement_Create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up
  - [x] Acceptance test updated: Providers → ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()
  - [x] checkUserEntitlementExists / checkUserEntitlementDestroyed use getDirectClient() (no Meta() dependency)
  - [x] Idempotency step added (PlanOnly: true, ExpectNonEmptyPlan: false)
  - [x] captureUserEntitlementEvidence calls CaptureLiveEvidence("acceptance-resource-user-entitlement", ...)
  - [ ] Live TF_ACC run needed — awaiting forge gate execution

## Watch out

- licensing_source is returned lower-cased from the API; flattenUserEntitlementFramework normalises to lowercase
- The serviceendpoint package has pre-existing test build failures (assignment mismatch) on main — not introduced here
