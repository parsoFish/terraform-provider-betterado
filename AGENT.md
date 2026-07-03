# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

**Root cause of last gate failure:** The acceptance tests were creating `betterado_project` inline in HCL but the ADO org is at its 1000-project cap → `Failed to add a project as this organization already has 1000 projects`.

**What was done:**

1. Created `azuredevops/internal/service/wiki/resource_wiki_framework.go` — a full `resource.Resource` implementation (`WikiResource`) with:
   - Metadata/Schema/Configure/Create/Read/Update/Delete/ImportState
   - Inline plan modifiers `wikiRequiresReplace()` + `wikiUseStateForUnknown()` (NOT using `stringplanmodifier` package — it's not in vendor)
   - Inline `oneOfStringValidator` (NOT using `terraform-plugin-framework-validators` — not in vendor)
   - Import via `resource.ImportStatePassthroughID`
   - `Configure()` type-asserts `req.ProviderData.(*client.AggregatedClient)`

2. Registered `wiki.NewWikiResource` in `framework_provider.go` Resources() slice.

3. Removed `betterado_wiki: wiki.ResourceWiki()` from `provider.go` ResourcesMap.

4. Deleted `resource_wiki.go` and `resource_wiki_test.go` (SDKv2 files).

5. Rewrote `azuredevops/internal/acceptancetests/resource_wiki_test.go`:
   - Uses `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` (not `Providers: testutils.GetProviders()`)
   - Uses `SharedFixtureProjectName` (betterado-standing-demo) instead of creating new projects
   - Has `ExpectNonEmptyPlan: false` idempotency step
   - Has `captureWikiEvidence()` for forge demo live-evidence
   - `checkWikiDestroyedFramework()` uses `getWikiDirectClient()` pattern (same as task group tests)
   - Tests: `TestAccWikiResource_projectWiki` and `TestAccWikiResource_codeWiki`

6. Decremented `provider_test.go` SDKv2 resource count (removed `betterado_wiki` from expectedResources).

7. Confirmed: `go build -mod=vendor .` passes, `go test ./azuredevops/ -run TestProvider` passes.

## What worked

- Inline plan modifiers following the pattern in `taskagent/resource_task_group_framework.go`
- Inline `oneOfStringValidator` since `terraform-plugin-framework-validators` is not in vendor
- Using `SharedFixtureProjectName` / `data "betterado_project"` data source for tests instead of creating new projects
- `getWikiDirectClient()` pattern for CheckDestroy (same as task group test)

## What didn't work

- `stringplanmodifier` package is not in vendor — must use inline plan modifier structs
- `terraform-plugin-framework-validators` is not in vendor — must use inline validator struct
- Creating projects inline in HCL — org is at 1000-project cap

## Key patterns for future iterations

- `strings.TrimRight(orgURL, "/")` needed before building API URLs
- The wiki API URL for live evidence: `{orgURL}/{projectID}/_apis/wiki/wikis/{wikiID}?api-version=7.1`
- `wikiRequiresReplace()` is like SDKv2 `ForceNew: true`
- `wikiUseStateForUnknown()` prevents perpetual diffs for Computed attributes

## Open questions

- Will the project wiki delete logic work? Project wikis use `DeleteRepository` on the backing git repo.
  - This is the same logic as the SDKv2 version, so should be correct.
- The standing project may already have a project wiki — the test creates a NEW uniquely-named one with `wikiName := testutils.GenerateResourceName()` so it shouldn't conflict.

## Notes for reflection

- The 1000-project cap is a consistent pattern across this initiative — all tests should use SharedFixtureProjectName.
- The `terraform-plugin-framework-validators` package is not vendored — keep using inline validators.
