# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new `azuredevops/internal/service/profile/` package with `data_profile.go` implementing the `betterado_profile` framework data source WHEN a Terraform config with `data "betterado_profile" "me" { id = "me" }` is evaluated THEN the data source reads the ADO Profile API (`_apis/profile/profiles/{id}`) and populates `display_name`, `email_address`, `public_alias`, `id`, and `avatar_url` in Terraform state; the offline unit test `TestDataProfileSchema` passes
- [x] AC2: GIVEN the framework provider's `DataSources()` function in `azuredevops/internal/provider/framework_provider.go` WHEN the provider is initialized THEN `betterado_profile` is present in the framework data source list; it is NOT registered in `azuredevops/provider.go` SDKv2 DataSourcesMap
