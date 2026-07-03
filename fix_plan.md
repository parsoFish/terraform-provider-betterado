# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC2: GIVEN the SDKv2 resource_wiki.go and resource_wiki_test.go deleted and betterado_wiki removed from provider.go ResourcesMap WHEN go build -mod=vendor . is run THEN the provider binary compiles without errors and provider_test.go resource count is decremented by 1 for betterado_wiki
  - Deleted resource_wiki.go and resource_wiki_test.go
  - Removed betterado_wiki from provider.go ResourcesMap
  - Decremented provider_test.go expectedResources list
  - go build -mod=vendor . passes

- [ ] AC1: GIVEN the betterado_wiki resource registered in the framework provider WHEN terraform apply runs a code-wiki or project-wiki configuration THEN apply succeeds, provider read-back reflects all attributes, idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy completes cleanly — verified by TestAccWikiResource_projectWiki and TestAccWikiResource_codeWiki updated to use GetMuxedProviderFactories()
  - resource_wiki_framework.go created and compiled
  - NewWikiResource registered in framework_provider.go
  - Acceptance tests rewritten: GetMuxedProviderFactories, SharedFixtureProjectName, ExpectNonEmptyPlan: false
  - Build failure FIXED (iter 2): restored checkWikiDestroyed(resourceType string) alongside checkWikiDestroyedFramework — resource_wiki_page_test.go uses the old function signature
  - acceptancetests package now compiles cleanly (go test -tags all -run TestAccWikiResource_projectWiki ./azuredevops/internal/acceptancetests/ → ok)
  - LIVE gate pending: needs TF_ACC run to confirm apply/read/idempotency/destroy cycle
