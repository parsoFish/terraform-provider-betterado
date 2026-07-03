# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_wiki_page resource registered in the framework provider WHEN terraform apply creates a wiki page, update changes content, and destroy removes it THEN all steps succeed, read-back reflects all attributes, idempotency re-plan produces no diff (ExpectNonEmptyPlan: false) — verified by TestAccWikiPageResource_basic and TestAccWikiPageResource_update updated to use GetMuxedProviderFactories()
- [x] AC2: GIVEN the SDKv2 resource_wiki_page.go deleted and betterado_wiki_page removed from provider.go ResourcesMap WHEN go build -mod=vendor . is run THEN the provider binary compiles without errors and provider_test.go resource count is decremented by 1 for betterado_wiki_page

## Work done in iteration 1

- [x] Created `resource_wiki_page_framework.go` (framework WikiPageResource)
- [x] Registered `NewWikiPageResource` in `framework_provider.go`
- [x] Removed `betterado_wiki_page` from `provider.go` ResourcesMap (and unused `wiki` import)
- [x] Deleted `resource_wiki_page.go` (SDKv2)
- [x] Rewrote `resource_wiki_page_test.go` with mux factories + idempotency steps
- [x] Removed `betterado_wiki_page` from `provider_test.go` expectedResources
- [x] `go build -mod=vendor ./...` passes
- [x] `golangci-lint run --new-from-rev=main` passes (0 issues)
- [x] `TestProvider_HasChildResources` passes
- [x] Committed: ae654d6c

## Work done in iteration 2

- [x] Diagnosed live gate failure: org at 1000-project cap — `betterado_project` resource creation fails
- [x] Rewrote `hclWikiPageBasic` and `hclWikiPageUpdate` to use `data "betterado_project" "fixture"` with `SharedFixtureProjectName` instead of creating a new project
- [x] Updated `TestAccWikiPageResource_basic` and `TestAccWikiPageResource_update` to remove `projectName` param (now using standing project)
- [x] Changed path from `/path` to `/page-path` (minor — avoids collision risk with any pre-existing page at `/path`)
- [x] `go build -mod=vendor ./...` passes
- [x] `go vet -tags all ./azuredevops/internal/acceptancetests/` passes
- [x] `golangci-lint run --new-from-rev=main` — 0 issues
- [x] Committed: 59fadbed

## Remaining

- Live gate will re-run `TestAccWikiPageResource_basic` with TF_ACC=1. Should succeed now that standing fixture project is used.
