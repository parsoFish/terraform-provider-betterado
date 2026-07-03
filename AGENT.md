# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration)

**Root cause:** Gate failure was `TestAccWikiPageResource_basic` crashing with "Invalid resource type betterado_wiki" because:
1. The test file was using `Providers: testutils.GetProviders()` (SDKv2 only)
2. `betterado_wiki` had already been migrated to the framework provider in WI-2
3. `betterado_wiki_page` was still in the SDKv2 `provider.go` ResourcesMap, not yet in the framework provider

**What was done:**
- Created `azuredevops/internal/service/wiki/resource_wiki_page_framework.go` — full framework `resource.Resource` implementation of `betterado_wiki_page` with:
  - Inline UUID validator (`wikiPageUUIDValidator`) and not-empty validator (`wikiPageNotEmptyValidator`)
  - Reuses `wikiRequiresReplace()` and `wikiUseStateForUnknown()` from the same `wiki` package (defined in `resource_wiki_framework.go`)
  - `pageLock = sync.Mutex{}` mutex preserved for concurrency safety
  - CRUD: Create/Read/Update/Delete using the ADO wiki API
  - `flattenPage()` helper for ETag parsing
- Registered `wiki.NewWikiPageResource` in `framework_provider.go` Resources() slice
- Removed `"betterado_wiki_page": wiki.ResourceWikiPage()` from `provider.go` ResourcesMap and removed the `wiki` import
- Deleted `resource_wiki_page.go` (SDKv2 implementation)
- Rewrote `resource_wiki_page_test.go` with:
  - Build tag `//go:build (all || resource_wiki_page) && !exclude_resource_wiki_page`
  - `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
  - `checkWikiPageDestroyedByAttrs` — direct ADO client CheckDestroy
  - `captureWikiPageEvidence` — live evidence capture
  - `TestAccWikiPageResource_basic` with idempotency step (ExpectNonEmptyPlan: false)
  - `TestAccWikiPageResource_update` with update + idempotency step
  - Both tests use `betterado_project` inline + `betterado_wiki` (projectWiki type)
- Removed `"betterado_wiki_page"` from `provider_test.go` expectedResources

## What worked

- **Reusing inline plan modifiers from the same package:** The `wiki` package's `resource_wiki_framework.go` already defines `wikiRequiresReplace()` and `wikiUseStateForUnknown()` — these work directly in `resource_wiki_page_framework.go` without importing the unavailable `stringplanmodifier` package.
- `stringplanmodifier` is NOT in vendor — always use inline plan modifiers as the wiki package already does.
- `go build -mod=vendor ./...` passes after migration.
- `golangci-lint run --new-from-rev=main` passes — only lint issue was `else { if }` → fixed to `else if`.
- `provider_test.go` test `TestProvider_HasChildResources` passes after removing `betterado_wiki_page` from expected list.

## What didn't work

- Importing `github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier` — this package is NOT vendored. Use inline plan modifier structs instead.

## Open questions

_(things that aren't blocking but would be useful to clarify)_

## Notes for reflection

- The `stringplanmodifier` package from terraform-plugin-framework is not in vendor — projects that want `UseStateForUnknown()` or `RequiresReplace()` must implement inline struct plan modifiers. This is a common pattern in this codebase.
