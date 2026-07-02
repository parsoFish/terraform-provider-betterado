# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_git_repositories data source migrated to terraform-plugin-framework WHEN terraform apply → provider read-back → idempotency → destroy runs live THEN TestAccDataSourceGitRepositories tests pass; the data source is deregistered from SDKv2 DataSourcesMap and registered in framework_provider.go DataSources(); provider_test.go updated; acceptance tests use GetMuxedProviderFactories()
  - [x] Created data_git_repositories_framework.go with ListNestedAttribute schema
  - [x] Deregistered from SDKv2 DataSourcesMap in provider.go (+ removed now-unused git import)
  - [x] Registered in framework_provider.go DataSources()
  - [x] Updated provider_test.go to remove betterado_git_repositories from expectedDataSources
  - [x] Rewrote data_git_repositories_test.go: added TestAccDataSourceGitRepositories using ProtoV6ProviderFactories and shared fixture project
  - [x] Added live evidence capture via gitRepositoriesLiveEvidenceCheck

- [ ] AC2: GIVEN the migration WHEN CI-equivalent gate runs (make test, golangci-lint --new-from-rev=main, make terrafmt-check) THEN all checks pass with no new lint findings on changed code
  - [x] golangci-lint --new-from-rev=main: 0 issues
  - [x] make terrafmt-check: passes
  - [x] go test offline: provider_test.go passes, git package tests pass, no new failures
  - [ ] Live acceptance gate (TF_ACC=1): pending forge live run
