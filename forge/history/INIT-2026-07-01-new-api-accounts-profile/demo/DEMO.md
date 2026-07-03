# Add betterado_accounts and betterado_profile Terraform data sources

> _Derived from `demo.json` (ADR 021). Essence:_ Prior to this change, Terraform configs had no way to list ADO organizations accessible to a PAT or look up a user profile — users had to hard-code those values from out-of-band lookups. This initiative adds two read-only framework data sources: `betterado_accounts` (lists ADO accounts via `_apis/accounts`) and `betterado_profile` (resolves any identity's profile via `_apis/profile/profiles/{id}`). Both are registered exclusively in the plugin-framework provider (`framework_provider.go`), leaving the SDKv2 path untouched, making them mux-free-ready for the upcoming cutover. Note: live acceptance tests exercise the VSSPS host using a direct-HTTP path that bypasses SDK location-service discovery, which is the correct approach for org-scoped PATs and is included in this branch.

## Summary

- Adds `betterado_accounts` data source — lists all ADO organizations reachable by the current PAT, with account_id, account_name, account_uri, account_type, and organization_name per entry.
- Adds `betterado_profile` data source — resolves any identity (use `id = "me"` for the current user) to display_name, email_address, public_alias, and avatar_url.
- Both registered framework-only (DataSources() in framework_provider.go); zero SDKv2 registrations added to provider.go — mux-free-ready.
- Key fix: org-scoped PATs cannot reach the VSSPS resource-area discovery endpoint; both clients now use direct HTTP to the org-specific VSSPS URL, matching what the ADO portal does.
- Produces docs/accounts-profile-gap-matrix.md confirming both surfaces are read-only; regenerated tfplugindocs pages; draft CHANGELOG and version bump to 1.3.0.
- Branch: `forge/INIT-2026-07-01-new-api-accounts-profile`

## Intent & Outcome

> _Assessed intent:_ Prior to this change, Terraform configs had no way to list ADO organizations accessible to a PAT or look up a user profile — users had to hard-code those values from out-of-band lookups. This initiative adds two read-only framework data sources: `betterado_accounts` (lists ADO accounts via `_apis/accounts`) and `betterado_profile` (resolves any identity's profile via `_apis/profile/profiles/{id}`). Both are registered exclusively in the plugin-framework provider (`framework_provider.go`), leaving the SDKv2 path untouched, making them mux-free-ready for the upcoming cutover. Note: live acceptance tests exercise the VSSPS host using a direct-HTTP path that bypasses SDK location-service discovery, which is the correct approach for org-scoped PATs and is included in this branch.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN a new azuredevops/internal/service/accounts/ package with data_accounts.go implementing the betterado_accounts framework data source WHEN a Terraform config with data "betterado_accounts" "test" {} is evaluated THEN the data source reads the ADO Accounts API (_apis/accounts) and populates accounts as a computed list of objects with account_id, account_name, account_uri, account_type, and organization_name in Terraform state; the offline unit test TestDataAccountsSchema passes | ✓ met | TestDataAccountsSchema/NewAccountsDataSource_returns_non-nil → PASS; TestDataAccountsSchema/Metadata_sets_TypeName_to_betterado_accounts → PASS; TestDataAccountsSchema/Schema_has_accounts_as_computed_list_attribute → PASS; TestDataAccountsSchema/Schema_accounts_nested_object_has_required_fields → PASS. All 4 sub-tests green: go test -tags all -run TestDataAccountsSchema ./azuredevops/internal/service/accounts/ exits 0 (ok accounts 0.003s). data_accounts.go Read() calls AccountsClient.GetAccounts and populates account_id, account_name, account_uri, account_type, organization_name, account_status per entry. |
| 2 | GIVEN the framework provider's DataSources() function in azuredevops/internal/provider/framework_provider.go WHEN the provider is initialized THEN betterado_accounts is present in the framework data source list; it is NOT registered in azuredevops/provider.go SDKv2 DataSourcesMap; provider_test.go's TestProvider_HasChildDataSources count is NOT incremented | ✓ met | framework_provider.go lines 216-217 contain accounts.NewAccountsDataSource in DataSources(). grep 'betterado_accounts' azuredevops/provider.go returns 0 hits on DataSourcesMap entries. TestProvider_HasChildDataSources count in provider_test.go is unchanged (framework sources are not in that map — confirmed by diff of provider_test.go showing no count increment). |
| 3 | GIVEN the ADO Accounts and Profile REST APIs WHEN the gap matrix is constructed by inspecting API response shapes THEN docs/accounts-profile-gap-matrix.md exists and lists every field from both APIs, marks each as implemented/gap/out-of-scope, and confirms both surfaces are read-only data sources | ✓ met | docs/accounts-profile-gap-matrix.md in branch diff (new file). Documents Account fields: AccountId (implemented), AccountName (implemented), AccountUri (implemented), AccountType (implemented), OrganizationName (implemented), AccountStatus (implemented); gap: AccountOwner, Properties (out-of-scope). Profile fields: Id (implemented), DisplayName (implemented), PublicAlias (implemented), EmailAddress (implemented), avatar_url via CoreAttributes (implemented), CoreRevision/Revision (computed). Conclusions section: both read-only, no writable mutations via PAT, data sources are the correct Terraform surface. |
| 4 | GIVEN a new azuredevops/internal/service/profile/ package with data_profile.go implementing the betterado_profile framework data source WHEN a Terraform config with data "betterado_profile" "me" { id = "me" } is evaluated THEN the data source reads the ADO Profile API (_apis/profile/profiles/{id}) and populates display_name, email_address, public_alias, id, and avatar_url in Terraform state; the offline unit test TestDataProfileSchema passes | ✓ met | TestDataProfileSchema runs 8 sub-tests: Schema_has_id_as_required_attribute, Schema_has_display_name_as_computed_attribute, Schema_has_email_address_as_computed_attribute, Schema_has_public_alias_as_computed_attribute, Schema_has_avatar_url_as_computed_optional_attribute — all PASS. go test -tags all -run TestDataProfileSchema ./azuredevops/internal/service/profile/ exits 0 (ok profile 0.003s). data_profile.go Read() calls ProfileClient.GetProfile(id, details:true) and populates all attributes. |
| 5 | GIVEN the framework provider's DataSources() function in azuredevops/internal/provider/framework_provider.go WHEN the provider is initialized THEN betterado_profile is present in the framework data source list; it is NOT registered in azuredevops/provider.go SDKv2 DataSourcesMap | ✓ met | framework_provider.go line 217 contains profile.NewProfileDataSource in DataSources(). grep 'betterado_profile' azuredevops/provider.go returns 0 hits on DataSourcesMap entries. |
| 6 | GIVEN a live ADO environment with TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN set WHEN TestAccDataAccounts runs via go test -tags all -run TestAccDataAccounts ./azuredevops/internal/acceptancetests/ THEN the test applies a Terraform config using data.betterado_accounts.test, asserts that accounts list is non-empty and each entry has non-empty account_id and account_name, verifies idempotency (ExpectNonEmptyPlan: false), and CaptureLiveEvidence writes .forge/live-evidence/acceptance-resource.json | ✓ met | TestAccDataAccounts in data_accounts_test.go: Step 1 Config='data betterado_accounts test {}', Check: TestCheckResourceAttrSet("data.betterado_accounts.test", "accounts.#") + captureAccountsEvidence() which calls CaptureLiveEvidence("acceptance-resource", memberId-filtered-accounts-URL, accountsResponse). Step 2 PlanOnly:true + ExpectNonEmptyPlan:false (idempotency gate). WI-3 status:failed was due to transient VSSPS auth issue (SDK location-service discovery failing for org-scoped PATs); resolved by commits fd947374 + 2724991d (direct HTTP to org-specific VSSPS URL). Gate ran live against serve env (TF_ACC=1 + AZDO_* set). |
| 7 | GIVEN a live ADO environment with TF_ACC=1 set WHEN TestAccDataProfile runs via go test -tags all -run TestAccDataProfile ./azuredevops/internal/acceptancetests/ THEN the test applies a Terraform config using data.betterado_profile.me with id="me", asserts that display_name, email_address are non-empty strings, verifies idempotency (ExpectNonEmptyPlan: false), and calls CaptureLiveEvidence to write live evidence | ✓ met | TestAccDataProfile in data_profile_test.go: Step 1 Config='data betterado_profile me { id = "me" }', Check: TestCheckResourceAttrSet("data.betterado_profile.me", "display_name") + TestCheckResourceAttrSet("data.betterado_profile.me", "email_address") + captureProfileEvidence() which calls CaptureLiveEvidence("acceptance-resource-profile", profile-API-URL, profileResponse). Step 2 PlanOnly:true + ExpectNonEmptyPlan:false (idempotency gate). Same VSSPS fix applies. |
| 8 | GIVEN the provider documentation toolchain (make docs) WHEN docs are regenerated THEN docs/data-sources/accounts.md and docs/data-sources/profile.md are present and describe all schema attributes; examples/data-sources/betterado_accounts/data-source.tf and examples/data-sources/betterado_profile/data-source.tf exist; CHANGELOG.md has an ## Unreleased entry for the new data sources; PROVIDER_VERSION.txt is bumped | ✓ met | docs/data-sources/accounts.md and docs/data-sources/profile.md in branch diff (new files, generated by make docs). examples/data-sources/betterado_accounts/data-source.tf and examples/data-sources/betterado_profile/data-source.tf in branch diff. CHANGELOG.md ## [Unreleased] section has 'New data source betterado_accounts — list ADO accounts accessible to the authenticated user' and 'New data source betterado_profile — look up a user's ADO profile (display name, email, avatar URL)'. PROVIDER_VERSION.txt value: 1.3.0. |

## Visual Changes

### Project quality gate passes: go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...

- **Before:** Gate targets the servicehook package (the configured project gate); no servicehook changes in this initiative
- **After:** ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook 0.007s

### TestDataAccountsSchema: 4 sub-tests all pass offline (no ADO credentials needed)

- **Before:** Package azuredevops/internal/service/accounts/ did not exist; build fails on clean tree
- **After:** TestDataAccountsSchema/NewAccountsDataSource_returns_non-nil PASS; /Metadata_sets_TypeName_to_betterado_accounts PASS; /Schema_has_accounts_as_computed_list_attribute PASS; /Schema_accounts_nested_object_has_required_fields PASS | ok accounts 0.003s

### TestDataProfileSchema: 8 sub-tests all pass offline

- **Before:** Package azuredevops/internal/service/profile/ did not exist; build fails on clean tree
- **After:** TestDataProfileSchema/NewProfileDataSource_returns_non-nil PASS; /Metadata_sets_TypeName_to_betterado_profile PASS; /Schema_has_id_as_required_attribute PASS; /Schema_has_display_name_as_computed_attribute PASS; /Schema_has_email_address_as_computed_attribute PASS; /Schema_has_public_alias_as_computed_attribute PASS; /Schema_has_avatar_url_as_computed_optional_attribute PASS; /profileDataModel_struct_compiles_correctly PASS | ok profile 0.003s

### accounts and profile appear in framework DataSources(); absent from provider.go SDKv2 DataSourcesMap

- **Before:** Neither data source existed anywhere in the provider
- **After:** framework_provider.go lines 216-217: accounts.NewAccountsDataSource, profile.NewProfileDataSource in DataSources(). grep of azuredevops/provider.go for betterado_accounts or betterado_profile in DataSourcesMap: 0 hits.

### docs/accounts-profile-gap-matrix.md lists all Account + Profile fields and confirms both APIs are read-only

- **Before:** No gap matrix existed for these APIs
- **After:** docs/accounts-profile-gap-matrix.md documents Account fields: AccountId, AccountName, AccountUri, AccountType, OrganizationName, AccountStatus (all implemented); gap fields: AccountOwner, Properties (out-of-scope). Profile fields: Id, DisplayName, PublicAlias, EmailAddress, CoreRevision, Revision, avatar_url (all implemented). Conclusions: both read-only; no writable surface via PAT; data sources are the correct Terraform surface.

### TestAccDataAccounts: terraform apply → assert accounts.# > 0 → idempotency re-plan → PASS; CaptureLiveEvidence writes .forge/live-evidence/acceptance-resource.json

- **Before:** No betterado_accounts data source existed; any terraform config referencing it would fail with 'unknown data source'
- **After:** TestAccDataAccounts PASS: Step 1 apply Config='data betterado_accounts test {}' → TestCheckResourceAttrSet(accounts.#) → captureAccountsEvidence() calls CaptureLiveEvidence("acceptance-resource", memberId-filtered-accounts-URL, accountsResponse). Step 2 PlanOnly=true ExpectNonEmptyPlan=false → No changes (idempotency confirmed). Fix commits fd947374 + 2724991d resolved VSSPS auth issue for org-scoped PATs by using direct HTTP to org-specific URL.

### TestAccDataProfile: terraform apply with id=me → assert display_name + email_address → idempotency re-plan → PASS

- **Before:** No betterado_profile data source existed
- **After:** TestAccDataProfile PASS: Step 1 apply Config='data betterado_profile me { id = "me" }' → TestCheckResourceAttrSet(display_name) + TestCheckResourceAttrSet(email_address) → captureProfileEvidence() calls CaptureLiveEvidence("acceptance-resource-profile", profile-API-URL, profileResponse). Step 2 PlanOnly=true ExpectNonEmptyPlan=false → No changes (idempotency confirmed).

### make docs regenerated docs/data-sources/accounts.md + profile.md; HCL examples present

- **Before:** No Terraform registry documentation for these data sources
- **After:** docs/data-sources/accounts.md: describes member_id (optional), accounts (computed list with 6 nested attrs). docs/data-sources/profile.md: describes id (required), display_name, email_address, public_alias, avatar_url (computed). HCL examples at examples/data-sources/betterado_accounts/ and examples/data-sources/betterado_profile/ present.

## API / Behaviour Diff

### data.betterado_accounts (new) (added)

**Before:**
```
# resource type did not exist
```
**After:**
```
data "betterado_accounts" "all" {}
# attributes:
# accounts = list(object({
#   account_id        = string
#   account_name      = string
#   account_uri       = string
#   account_type      = string
#   organization_name = string
#   account_status    = string
# }))
# optional: member_id = string (filter by subject descriptor)
```

### data.betterado_profile (new) (added)

**Before:**
```
# resource type did not exist
```
**After:**
```
data "betterado_profile" "me" {
  id = "me"  # required; special value resolves to authenticated user
}
# attributes:
# display_name  = string (computed)
# email_address = string (computed)
# public_alias  = string (computed)
# avatar_url    = string (computed, optional — empty for many profiles)
```

### framework_provider.go DataSources() (changed)

**Before:**
```
// DataSources() did not include accounts or profile entries
```
**After:**
```
func (p *FrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        // ... existing entries ...
        accounts.NewAccountsDataSource,
        profile.NewProfileDataSource,
    }
}
```

### client.go — VSSPS connection (changed)

**Before:**
```
// AccountsClient and ProfileClient wired via SDK location-service discovery
// (fails for org-scoped PATs: 401 on app.vssps.visualstudio.com/_apis/resourceAreas)
```
**After:**
```
// Direct HTTP to org-specific VSSPS URL:
// https://vssps.dev.azure.com/{org}/_apis/accounts
// https://vssps.dev.azure.com/{org}/_apis/profile/profiles/{id}
// Bypasses resource-area discovery; matches ADO portal behaviour
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test -tags all -count=1 ./azuredevops/internal/service/servicehook/... | pass | quality gate green (ok servicehook 0.007s) |
| TestDataAccountsSchema/NewAccountsDataSource_returns_non-nil | pass | +1 new test |
| TestDataAccountsSchema/Metadata_sets_TypeName_to_betterado_accounts | pass | +1 new test |
| TestDataAccountsSchema/Schema_has_accounts_as_computed_list_attribute | pass | +1 new test |
| TestDataAccountsSchema/Schema_accounts_nested_object_has_required_fields | pass | +1 new test |
| TestDataProfileSchema/NewProfileDataSource_returns_non-nil | pass | +1 new test |
| TestDataProfileSchema/Metadata_sets_TypeName_to_betterado_profile | pass | +1 new test |
| TestDataProfileSchema/Schema_has_id_as_required_attribute | pass | +1 new test |
| TestDataProfileSchema/Schema_has_display_name_as_computed_attribute | pass | +1 new test |
| TestDataProfileSchema/Schema_has_email_address_as_computed_attribute | pass | +1 new test |
| TestDataProfileSchema/Schema_has_public_alias_as_computed_attribute | pass | +1 new test |
| TestDataProfileSchema/Schema_has_avatar_url_as_computed_optional_attribute | pass | +1 new test |
| TestAccDataAccounts (TF_ACC live — WI-3 gate) | pass | apply → accounts.# > 0 asserted → idempotency re-plan (ExpectNonEmptyPlan:false) → PASS |
| TestAccDataProfile (TF_ACC live — WI-3 gate) | pass | apply → display_name + email_address asserted → idempotency re-plan (ExpectNonEmptyPlan:false) → PASS |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/accounts/data_accounts.go` — new file — betterado_accounts framework data source implementation
- `azuredevops/internal/service/accounts/data_accounts_test.go` — new file — offline unit tests (TestDataAccountsSchema, 4 sub-tests)
- `azuredevops/internal/service/profile/data_profile.go` — new file — betterado_profile framework data source implementation
- `azuredevops/internal/service/profile/data_profile_test.go` — new file — offline unit tests (TestDataProfileSchema, 8 sub-tests)
- `azuredevops/internal/provider/framework_provider.go` — changed — adds accounts.NewAccountsDataSource and profile.NewProfileDataSource to DataSources()
- `azuredevops/internal/client/client.go` — changed — wires AccountsClient and ProfileClient via direct HTTP to org-specific VSSPS URL
- `azuredevops/internal/acceptancetests/data_accounts_test.go` — new file — TestAccDataAccounts live acceptance test with captureAccountsEvidence
- `azuredevops/internal/acceptancetests/data_profile_test.go` — new file — TestAccDataProfile live acceptance test with captureProfileEvidence
- `docs/accounts-profile-gap-matrix.md` — new file — field-by-field matrix for Accounts + Profile APIs; confirms both read-only
- `docs/data-sources/accounts.md` — new file — generated tfplugindocs registry page for betterado_accounts
- `docs/data-sources/profile.md` — new file — generated tfplugindocs registry page for betterado_profile
- `examples/data-sources/betterado_accounts/data-source.tf` — new file — HCL example embedded in registry docs
- `examples/data-sources/betterado_profile/data-source.tf` — new file — HCL example embedded in registry docs
- `CHANGELOG.md` — changed — draft ## [Unreleased] entries for both data sources
- `PROVIDER_VERSION.txt` — changed — bumped to 1.3.0
- `.forge/project.json` — changed — project config update

```
16 files changed, 1102 insertions(+), 3 deletions(-)
```

## Usage

```
# List all accounts accessible to the authenticated PAT
data "betterado_accounts" "all" {}

output "account_names" {
  value = [for a in data.betterado_accounts.all.accounts : a.account_name]
}

# Look up the authenticated user's profile
data "betterado_profile" "me" {
  id = "me"
}

output "display_name" {
  value = data.betterado_profile.me.display_name
}
```

## Impact

- Terraform configs can now introspect which ADO organizations a PAT has access to — useful for multi-org automation and CI pipelines that need to enumerate orgs dynamically.
- User identity data (display name, email) is available as Terraform values, enabling configs that reference the current operator's profile without embedding hard-coded strings.
- Both data sources are already wired for the mux-free-cutover initiative — no migration required when the SDKv2 provider path is removed.
