# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new `azuredevops/internal/service/accounts/` package with `data_accounts.go` implementing the `betterado_accounts` framework data source WHEN a Terraform config with `data "betterado_accounts" "test" {}` is evaluated THEN the data source reads the ADO Accounts API (`_apis/accounts`) and populates `accounts` as a computed list of objects with `account_id`, `account_name`, `account_uri`, `account_type`, and `organization_name` in Terraform state; the offline unit test `TestDataAccountsSchema` passes
- [x] AC2: GIVEN the framework provider's `DataSources()` function in `azuredevops/internal/provider/framework_provider.go` WHEN the provider is initialized THEN `betterado_accounts` is present in the framework data source list; it is NOT registered in `azuredevops/provider.go` SDKv2 DataSourcesMap; `provider_test.go`'s `TestProvider_HasChildDataSources` count is NOT incremented (framework sources are not in that map)
- [x] AC3: GIVEN the ADO Accounts and Profile REST APIs WHEN the gap matrix is constructed by inspecting API response shapes THEN `docs/accounts-profile-gap-matrix.md` exists and lists every field from both APIs, marks each as implemented/gap/out-of-scope, and confirms both surfaces are read-only data sources
