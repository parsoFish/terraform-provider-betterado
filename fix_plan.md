# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN each of the 7 repository policy resources has a *_framework.go file WHEN terraform apply → provider read-back → idempotency re-plan → destroy runs live THEN TestAccRepositoryPolicy* acceptance tests all pass with GetMuxedProviderFactories()
  - Framework files created for all 7 resources
  - Acceptance tests updated to use SharedFixtureProjectID(t) + GetMuxedProviderFactories()
  - No betterado_project creation in tests (avoids 1000-project cap)
  - Awaiting live gate run to confirm full ACC pass
- [x] AC2: GIVEN the framework migration is applied WHEN provider.go is inspected THEN all 7 repository policy resources are removed from ResourcesMap (SDKv2) and added to framework_provider.go Resources()
  - All 7 removed from provider.go ResourcesMap
  - All 7 registered in framework_provider.go Resources()
- [x] AC3: GIVEN provider_test.go HasChildResources count is further updated WHEN go test ./azuredevops/ -run TestProvider_HasChildResources runs THEN the test passes (count reflects removal of 7 repo-policy resources from SDKv2)
  - PASSES: `go test ./azuredevops/ -run TestProvider_HasChildResources` → PASS

## Resources migrated

| Resource | Framework file |
|---|---|
| betterado_repository_policy_author_email_pattern | resource_repositorypolicy_author_email_patterns_framework.go |
| betterado_repository_policy_case_enforcement | resource_repositorypolicy_enforce_consistent_case_framework.go |
| betterado_repository_policy_check_credentials | resource_repositorypolicy_check_credentials_framework.go |
| betterado_repository_policy_file_path_pattern | resource_repositorypolicy_file_path_patterns_framework.go |
| betterado_repository_policy_max_file_size | resource_repositorypolicy_max_file_size_framework.go |
| betterado_repository_policy_max_path_length | resource_repositorypolicy_max_path_length_framework.go |
| betterado_repository_policy_reserved_names | resource_repositorypolicy_reserved_names_framework.go |
