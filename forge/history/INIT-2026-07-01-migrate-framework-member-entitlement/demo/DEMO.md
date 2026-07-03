# Migrate betterado_user_entitlement, betterado_group_entitlement, and betterado_service_principal_entitlement to terraform-plugin-framework

> _Derived from `demo.json` (ADR 021). Essence:_ Three SDKv2 resources — betterado_user_entitlement, betterado_group_entitlement, and betterado_service_principal_entitlement — are now served via the mux provider using terraform-plugin-framework. Prior: each was registered in provider.go ResourcesMap using the SDKv2 helper/schema path. After: each is deregistered from ResourcesMap, added to framework_provider.go Resources(), and backed by a framework resource type with Configure() wired to *client.AggregatedClient. ConfigValidators enforce attribute constraints; schema defaults eliminate known-after-apply drift; a case-insensitive plan modifier suppresses case-only API diffs. Live acceptance evidence captured for group_entitlement (TestAccGroupEntitlement_Create ran with TF_ACC=1); user_entitlement and service_principal_entitlement offline-only in this cycle (no TF_ACC environment). Gap matrix produced for the full REST API v7.1 surface.

## Summary

- betterado_user_entitlement, betterado_group_entitlement, betterado_service_principal_entitlement migrated from SDKv2 to terraform-plugin-framework — no schema change visible to users.
- All three resources deregistered from provider.go ResourcesMap and wired into framework_provider.go Resources(); mux provider serves both SDK paths without conflict.
- UWI-2 review fixes: 6 orphaned SDKv2 files deleted; ConfigValidators re-add mutual-exclusivity constraints; user_entitlement gets schema defaults + case-insensitive plan modifier for account_license_type/licensing_source.
- docs/memberentitlementmanagement-gap-matrix.md produced: every REST API v7.1 field listed with gap status (covered / gap-deferred).
- CHANGELOG.md updated with DRAFT Unreleased entries; registry docs regenerated; PROVIDER_VERSION.txt bumped to 1.3.0.
- Live acceptance evidence available for group_entitlement; user_entitlement and service_principal_entitlement acceptance claims downgraded to not-run (no TF_ACC=1 environment in this cycle).
- Branch: `forge/INIT-2026-07-01-migrate-framework-member-entitlement`
- Commit: `9fcb639b`

## Intent & Outcome

> _Assessed intent:_ Three SDKv2 resources — betterado_user_entitlement, betterado_group_entitlement, and betterado_service_principal_entitlement — are now served via the mux provider using terraform-plugin-framework. Prior: each was registered in provider.go ResourcesMap using the SDKv2 helper/schema path. After: each is deregistered from ResourcesMap, added to framework_provider.go Resources(), and backed by a framework resource type with Configure() wired to *client.AggregatedClient. ConfigValidators enforce attribute constraints; schema defaults eliminate known-after-apply drift; a case-insensitive plan modifier suppresses case-only API diffs. Live acceptance evidence captured for group_entitlement (TestAccGroupEntitlement_Create ran with TF_ACC=1); user_entitlement and service_principal_entitlement offline-only in this cycle (no TF_ACC environment). Gap matrix produced for the full REST API v7.1 surface.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Member Entitlement Management REST API v7.1 schema WHEN compared against the SDKv2 schemas for user_entitlement, group_entitlement, and service_principal_entitlement THEN docs/memberentitlementmanagement-gap-matrix.md exists and lists every field with writable gaps marked resolved or deferred | ✓ met | docs/memberentitlementmanagement-gap-matrix.md exists on branch (added commit 3cc814b3); tables for all three resources list ADO REST fields, terraform attributes, writability, and gap status (covered/gap-deferred). Gate test TestGapMatrixExists → pass (go test -tags all -run TestGapMatrixExists ./azuredevops/internal/service/memberentitlementmanagement/). |
| 2 | GIVEN betterado_user_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_user_entitlement appears in the list and NOT in provider.go ResourcesMap | ✓ met | framework_provider.go Resources() includes memberentitlementmanagement.NewUserEntitlementResource(); provider.go ResourcesMap no longer contains 'betterado_user_entitlement'. TestFrameworkProvider_HasUserEntitlementResource → pass (go test -tags all -run TestFrameworkProvider_HasUserEntitlementResource ./azuredevops/internal/provider/). |
| 3 | GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccUserEntitlement_Create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up | ~ partial | Framework implementation is code-complete (ConfigValidators, schema defaults, case-insensitive modifier); test wiring updated to GetMuxedProviderFactories(). TestAccUserEntitlement_Create NOT run live — no TF_ACC=1 environment available in this cycle; no .forge/live-evidence/acceptance-resource-user-entitlement.json present. Offline unit test TestNewUserEntitlementResource_Metadata → pass. |
| 4 | GIVEN betterado_group_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_group_entitlement appears in the list and NOT in provider.go ResourcesMap | ✓ met | framework_provider.go Resources() includes memberentitlementmanagement.NewGroupEntitlementResource(); provider.go ResourcesMap no longer contains 'betterado_group_entitlement'. TestFrameworkProvider_HasGroupEntitlementResource → pass. |
| 5 | GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccGroupEntitlement_Create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up | ✓ met | TestAccGroupEntitlement_Create ran live; REST GET captured at https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1 (2026-07-03T03:18:03Z) — response: id=7bece247, displayName=group-038c153d, licenseRule.status=active, status=applied. ExpectNonEmptyPlan: false → PASS. |
| 6 | GIVEN betterado_service_principal_entitlement migrated to terraform-plugin-framework WHEN the framework provider's Resources() is inspected THEN betterado_service_principal_entitlement appears in the list and NOT in provider.go ResourcesMap | ✓ met | framework_provider.go Resources() includes memberentitlementmanagement.NewServicePrincipalEntitlementResource(); provider.go ResourcesMap no longer contains 'betterado_service_principal_entitlement'. TestFrameworkProvider_HasServicePrincipalEntitlementResource → pass. |
| 7 | GIVEN the framework resource is wired with a real AggregatedClient WHEN TestAccServicePrincipalEntitlement_create runs with TF_ACC=1 via GetMuxedProviderFactories() THEN apply succeeds, read-back returns correct values, idempotency re-plan shows no diff, destroy cleans up | ~ partial | Framework implementation is code-complete; test wiring updated to GetMuxedProviderFactories(). TestAccServicePrincipalEntitlement_create NOT run live — TF_ACC=1 environment not available and AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID not set; no .forge/live-evidence/acceptance-resource-sp-entitlement.json present. Offline unit test TestNewServicePrincipalEntitlementResource_Metadata → pass. |
| 8 | GIVEN all three entitlement resources migrated to framework WHEN make docs runs and docs are committed THEN docs/resources/user_entitlement.md, docs/resources/group_entitlement.md, docs/resources/service_principal_entitlement.md reflect current framework schemas; docs/guides/ is restored | ✓ met | docs/resources/user_entitlement.md, docs/resources/group_entitlement.md, docs/resources/service_principal_entitlement.md present in branch diff (committed 742dfa66). make docs ran + git checkout -- docs/guides/ executed; docs/guides/ restored in same commit. |
| 9 | GIVEN live evidence captured in WI-2, WI-3, WI-4 acceptance tests WHEN demo.json is inspected THEN a checkpoint exists with liveEvidence.url pointing to a real REST GET URL | ✓ met | checkpoint 'acceptance-resource' carries liveEvidence.url = https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1 (capturedAt 2026-07-03T03:18:03Z; real ADO REST GET response present in response field). |
| 10 | GIVEN all three resources migrated WHEN CHANGELOG.md is read THEN an Unreleased entry notes the framework migration of the three entitlement resources | ✓ met | CHANGELOG.md ## [Unreleased] section contains FEATURES bullets for betterado_user_entitlement, betterado_group_entitlement, betterado_service_principal_entitlement framework migrations (committed 742dfa66). |
| 11 | GIVEN a user-visible schema change has shipped WHEN PROVIDER_VERSION.txt is read THEN the semver has been bumped (patch or minor) relative to its pre-cycle value | ✓ met | PROVIDER_VERSION.txt = 1.3.0 (prior release was 1.2.0 per CHANGELOG.md ## [1.2.0] heading; bumped minor for three new framework resource implementations, committed 742dfa66). |

## Visual Changes

### Initiative CI gate (release + taskagent packages) passes on branch HEAD

- **Before:** Gate packages (release, taskagent) were unaffected by the SDKv2 entitlement registrations — these pass on main and remain green after migration
- **After:** ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.009s; ok .../taskagent 0.006s; ok .../taskagent/validate 0.004s — all 3 packages green on branch HEAD (9fcb639b)

### Framework provider Resources() lists all three entitlement types; SDKv2 ResourcesMap entries removed

- **Before:** Prior: betterado_user_entitlement, betterado_group_entitlement, betterado_service_principal_entitlement registered only in provider.go ResourcesMap (SDKv2 path); framework provider did not list them
- **After:** After: TestFrameworkProvider_HasUserEntitlementResource, TestFrameworkProvider_HasGroupEntitlementResource, TestFrameworkProvider_HasServicePrincipalEntitlementResource all PASS; ResourcesMap entries removed

### Live group entitlement created via framework resource path; ADO REST GET confirms entitlement exists

- **Before:** Prior: betterado_group_entitlement used SDKv2 helper/schema path; no framework Configure() wiring existed
- **After:** After: TestAccGroupEntitlement_Create applied via mux→framework path; GET /groupentitlements/{id} returned id=7bece247, displayName=group-038c153d, licenseRule.status=active, status=applied. ExpectNonEmptyPlan: false → PASS; destroy cleaned up
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1` _(captured 2026-07-03T03:18:03Z)_

```json
{"id":"7bece247-5904-4d44-b7e3-29f96502d9ae","group":{"displayName":"group-038c153d-c86e-443c-b6f6-3d97378025d0","origin":"vsts","originId":"7bece247-5904-4d44-b7e3-29f96502d9ae","principalName":"[davidgparsonson]\\group-038c153d-c86e-443c-b6f6-3d97378025d0"},"licenseRule":{"accountLicenseType":"express","licensingSource":"account","status":"active"},"status":"applied"}
```

### Live evidence — acceptance-resource-group-entitlement

- **After:** Real API GET against the live system: https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1` _(captured 2026-07-03T03:18:03Z)_

```json
{
  "extensionRules": [],
  "group": {
    "_links": {
      "membershipState": {
        "href": "https://vssps.dev.azure.com/davidgparsonson/_apis/Graph/MembershipStates/vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA"
      },
      "memberships": {
        "href": "https://vssps.dev.azure.com/davidgparsonson/_apis/Graph/Memberships/vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA"
      },
      "self": {
        "href": "https://vssps.dev.azure.com/davidgparsonson/_apis/Graph/Groups/vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA"
      },
      "storageKey": {
        "href": "https://vssps.dev.azure.com/davidgparsonson/_apis/Graph/StorageKeys/vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA"
      }
    },
    "descriptor": "vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA",
    "displayName": "group-038c153d-c86e-443c-b6f6-3d97378025d0",
    "url": "https://vssps.dev.azure.com/davidgparsonson/_apis/Graph/Groups/vssgp.Uy0xLTktMTU1MTM3NDI0NS0xMDkyNDk4Mzc1LTMxODAyNjQyNjYtMjgwODQ0NTIyNS0zOTEzMDQ5NjE1LTEtMzY2NjM1MjYwMC0zOTQ2OTQyMjc4LTMxMjU3MDgwMzktMzcyNDg3MjE2MA",
    "origin": "vsts",
    "originId": "7bece247-5904-4d44-b7e3-29f96502d9ae",
    "subjectKind": "group",
    "domain": "vstfs:///Framework/IdentityDomain/c7331e41-8ebd-4afb-a765-7929e93c660f",
    "principalName": "[davidgparsonson]\\group-038c153d-c86e-443c-b6f6-3d97378025d0",
    "description": ""
  },
  "id": "7bece247-5904-4d44-b7e3-29f96502d9ae",
  "lastExecuted": "2026-07-03T03:17:59.9689895Z",
  "licenseRule": {
    "accountLicenseType": "express",
    "assignmentSource": "unknown",
    "licenseDisplayName": "Basic",
    "licensingSource": "account",
    "msdnLicenseType": "none",
    "status": "active",
    "statusMessage": ""
  },
  "projectEntitlements": [],
  "status": "applied"
}
```

## API / Behaviour Diff

### provider.go ResourcesMap — removed betterado_user_entitlement (removed)

**Before:**
```
"betterado_user_entitlement": memberentitlementmanagement.ResourceUserEntitlement()
```
**After:**
```
(moved to framework_provider.go Resources() as memberentitlementmanagement.NewUserEntitlementResource())
```

### provider.go ResourcesMap — removed betterado_group_entitlement (removed)

**Before:**
```
"betterado_group_entitlement": memberentitlementmanagement.ResourceGroupEntitlement()
```
**After:**
```
(moved to framework_provider.go Resources() as memberentitlementmanagement.NewGroupEntitlementResource())
```

### provider.go ResourcesMap — removed betterado_service_principal_entitlement (removed)

**Before:**
```
"betterado_service_principal_entitlement": memberentitlementmanagement.ResourceServicePrincipalEntitlement()
```
**After:**
```
(moved to framework_provider.go Resources() as memberentitlementmanagement.NewServicePrincipalEntitlementResource())
```

### framework_provider.go Resources() — added three entitlement resources (added)

**Before:**
```
// betterado_user/group/service_principal_entitlement not in framework provider
```
**After:**
```
memberentitlementmanagement.NewUserEntitlementResource(),
memberentitlementmanagement.NewGroupEntitlementResource(),
memberentitlementmanagement.NewServicePrincipalEntitlementResource(),
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... (initiative CI gate — offline) | pass | 3/3 packages green on branch HEAD |
| TestGapMatrixExists (go test -tags all -run TestGapMatrixExists ./azuredevops/internal/service/memberentitlementmanagement/) | pass | +1 new gate test (WI-1) |
| TestNewUserEntitlementResource_Metadata (offline unit, WI-2) | pass | +1 new test |
| TestFrameworkProvider_HasUserEntitlementResource (offline unit, WI-2) | pass | +1 new test |
| TestAccUserEntitlement_Create (TF_ACC=1, live ADO, WI-2) | skip | Skipped — no TF_ACC=1 environment available in this cycle; offline unit tests pass |
| TestNewGroupEntitlementResource_Metadata (offline unit, WI-3) | pass | +1 new test |
| TestFrameworkProvider_HasGroupEntitlementResource (offline unit, WI-3) | pass | +1 new test |
| TestAccGroupEntitlement_Create (TF_ACC=1, live ADO, WI-3) | pass | Updated to GetMuxedProviderFactories(); live REST GET evidence captured; idempotency PASS |
| TestNewServicePrincipalEntitlementResource_Metadata (offline unit, WI-4) | pass | +1 new test |
| TestFrameworkProvider_HasServicePrincipalEntitlementResource (offline unit, WI-4) | pass | +1 new test |
| TestAccServicePrincipalEntitlement_create (TF_ACC=1, live ADO, WI-4) | skip | Skipped — TF_ACC=1 not available; AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID not set; offline unit tests pass |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/memberentitlementmanagement/resource_user_entitlement_framework.go` — new file — framework resource implementation (WI-2)
- `azuredevops/internal/service/memberentitlementmanagement/resource_user_entitlement_framework_test.go` — new file — metadata unit test (WI-2)
- `azuredevops/internal/service/memberentitlementmanagement/resource_group_entitlement_framework.go` — new file — framework resource implementation (WI-3)
- `azuredevops/internal/service/memberentitlementmanagement/resource_group_entitlement_framework_test.go` — new file — metadata unit test (WI-3)
- `azuredevops/internal/service/memberentitlementmanagement/resource_service_principal_entitlement_framework.go` — new file — framework resource implementation (WI-4)
- `azuredevops/internal/service/memberentitlementmanagement/resource_service_principal_entitlement_framework_test.go` — new file — metadata unit test (WI-4)
- `azuredevops/internal/service/memberentitlementmanagement/gap_matrix_test.go` — new file — gate test for gap matrix existence (WI-1)
- `azuredevops/internal/provider/framework_provider.go` — changed — added three NewXxxEntitlementResource() to Resources()
- `azuredevops/internal/provider/framework_provider_test.go` — changed — added HasXxxEntitlementResource tests
- `azuredevops/provider.go` — changed — removed three entitlement entries from ResourcesMap
- `azuredevops/provider_test.go` — changed — SDKv2 resource count decremented by 3
- `azuredevops/internal/acceptancetests/resource_user_entitlement_test.go` — changed — switched to GetMuxedProviderFactories()
- `azuredevops/internal/acceptancetests/resource_group_entitlement_test.go` — changed — switched to GetMuxedProviderFactories()
- `azuredevops/internal/acceptancetests/resource_service_principal_entitlement_test.go` — changed — switched to GetMuxedProviderFactories()
- `docs/memberentitlementmanagement-gap-matrix.md` — new file — REST API v7.1 field gap matrix (WI-1)
- `docs/resources/user_entitlement.md` — changed — regenerated by make docs (WI-5)
- `docs/resources/group_entitlement.md` — changed — regenerated by make docs (WI-5)
- `docs/resources/service_principal_entitlement.md` — changed — regenerated by make docs (WI-5)
- `examples/resources/betterado_user_entitlement/resource.tf` — new file — HCL example for docs embed (WI-5)
- `examples/resources/betterado_group_entitlement/resource.tf` — new file — HCL example for docs embed (WI-5)
- `examples/resources/betterado_service_principal_entitlement/resource.tf` — new file — HCL example for docs embed (WI-5)
- `CHANGELOG.md` — changed — DRAFT Unreleased entries for all three migrations (WI-5)
- `PROVIDER_VERSION.txt` — changed — bumped to 1.3.0 (WI-5)
- `forge/history/INIT-2026-07-01-migrate-framework-member-entitlement/demo/demo.json` — this file — unifier demo artefact

```
74 files changed, 3629 insertions(+), 3289 deletions(-)
```

## Usage

```
# betterado_user_entitlement (now framework-backed)
resource "betterado_user_entitlement" "example" {
  principal_name       = "user@example.com"
  account_license_type = "express"
}

# betterado_group_entitlement (now framework-backed)
resource "betterado_group_entitlement" "example" {
  display_name         = "my-group"
  account_license_type = "express"
}

# betterado_service_principal_entitlement (now framework-backed)
resource "betterado_service_principal_entitlement" "example" {
  origin_id            = "<aad-service-principal-object-id>"
  origin               = "aad"
  account_license_type = "express"
}
```

## Impact

- All three entitlement resources now use the terraform-plugin-framework code path — better type-safety, cleaner plan/apply diagnostics, aligned with provider long-term architecture.
- No schema change: existing Terraform configurations continue to work without modification.
- Mux provider eliminates any 'Duplicate resource type' risk; both SDK paths coexist with a single authoritative registration per resource.
- docs/memberentitlementmanagement-gap-matrix.md provides a permanent reference for future API surface expansion across these three resources.
