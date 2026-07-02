# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_git_repository_branch resource migrated to terraform-plugin-framework WHEN terraform apply → provider read-back → idempotency re-plan (ExpectNonEmptyPlan:false) → destroy runs live THEN TestAccGitRepositoryBranch tests pass; the resource is deregistered from SDKv2 provider.go ResourcesMap and registered in framework_provider.go Resources(); provider_test.go updated; acceptance tests use GetMuxedProviderFactories()
  - [x] Created `resource_git_repository_branch_framework.go` (Create/Read/Update/Delete/ImportState)
  - [x] Deregistered from SDKv2 `provider.go` ResourcesMap
  - [x] Registered `NewGitRepositoryBranchResource` in `framework_provider.go` Resources()
  - [x] Removed from `provider_test.go` expectedResources (TestProvider_HasChildResources)
  - [x] Renamed test functions to `TestAccGitRepositoryBranch_*` (gate pattern match)
  - [x] Switched tests to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
- [x] AC2: GIVEN the migration WHEN CI-equivalent gate runs (make test, golangci-lint --new-from-rev=main, make terrafmt-check) THEN all checks pass with no new lint findings on changed code
  - [x] `gofmt` clean (all changed files pass `gofmt -l`)
  - [x] `go build -tags all ./...` succeeds
  - [x] `golangci-lint run --new-from-rev=main ./azuredevops/...` → 0 issues (fixed gocritic if-else chain)
  - [x] Live acceptance test: TestAccGitRepositoryBranch_* — tests now use shared fixture project (no project create); gate failure was 1000-project cap → fixed
