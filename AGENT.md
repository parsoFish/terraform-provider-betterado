# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (WI-4)

**Root causes identified from .forge/last-gate-failure.md:**

1. **TestAccWikiPageResource_update**: "Error creating wiki: Wiki already exists for project 'betterado-standing-demo'"
   - **Root cause**: ADO limits each project to ONE project wiki. The tests `TestAccWikiResource_projectWiki`, `TestAccWikiPageResource_basic`, and `TestAccWikiPageResource_update` all create `projectWiki` resources in the same standing project (`betterado-standing-demo`) and run in parallel. The second/third to run hits the "already exists" limit.
   - **Fix**: Changed `hclWikiPageBasic` and `hclWikiPageUpdate` in `resource_wiki_page_test.go` to use `codeWiki` type backed by a new git repository. Code wikis have no per-project limit, so parallel tests don't collide.

2. **TestAccWikiResource_projectWiki**: "Error running post-test destroy: found wiki that should have been deleted"
   - **Root cause**: The `Delete` method in `resource_wiki_framework.go` for `projectWiki` type tried to delete the underlying git repository instead of calling the `DeleteWiki` API. This was unreliable - the git repo deletion doesn't always remove the wiki from ADO's perspective.
   - **Fix** (iteration 1): Simplified `Delete` to call `DeleteWiki` directly for ALL wiki types. **WRONG** — see iteration 2.

**Other ACs completed in this iteration:**
- AC2/AC3: `captureWikiEvidence` and `captureWikiPageEvidence` were already implemented in prior WI iterations. No changes needed.
- AC4: Created `examples/resources/betterado_wiki/resource.tf` and `examples/resources/betterado_wiki_page/resource.tf`, ran `make docs` to regenerate `docs/resources/wiki.md` and `docs/resources/wiki_page.md` with framework schema (no timeouts block).
- AC5: Added `## Unreleased` section to `CHANGELOG.md` with wiki migration entries; bumped `PROVIDER_VERSION.txt` from `1.2.0` to `1.3.0`.

### Iteration 2 (WI-4)

**Root causes from .forge/last-gate-failure.md (iteration 1 gate):**

1. **TestAccWikiPageResource_basic/update**: "The versionType should be 'branch' and version cannot not be null / Parameter name: versionDescriptor"
   - **Root cause**: ADO requires `versionDescriptor{versionType:"branch", version:<branch>}` in `CreateOrUpdatePage` for code wikis. Our framework wiki page resource didn't pass this.
   - **Fix**: Added optional `version` attribute to `betterado_wiki_page` schema. In `Create`, when version is set, pass it as `GitVersionDescriptor{VersionType: "branch", Version: version}`. Updated test HCL to include `version = "master"` on the wiki page resource.

2. **TestAccWikiResource_projectWiki**: "Wiki delete operation is not supported on wikis of type 'ProjectWiki'."
   - **Root cause**: Calling `DeleteWiki` on a project wiki returns this error. ADO does NOT support deleting project wikis via the wiki API.
   - **Fix**: Restored the original SDKv2 strategy for projectWiki Delete: read the wiki to get its `RepositoryId`, then call `GitReposClient.DeleteRepository`. After the repo is deleted, the wiki is gone (GetWiki returns 404). CodeWiki continues to use `DeleteWiki` directly.

## What worked

- Using `DeleteRepository` (git repo) for projectWiki delete
- Adding `versionDescriptor` with type=branch to `CreateOrUpdatePage` for code wikis
- Switching page test HCL to `codeWiki` to avoid parallel test collisions (iteration 1)

## What didn't work

- `DeleteWiki` for projectWiki: returns "Wiki delete operation is not supported on wikis of type 'ProjectWiki'."
- `CreateOrUpdatePage` without `VersionDescriptor` for codeWiki: returns "The versionType should be 'branch' and version cannot not be null"

## Open questions

_(none)_

## Notes for reflection

- ADO "one project wiki per project" limit is a key gotcha for acceptance tests using a shared standing project.
- `DeleteWiki` API works for code wikis only. For project wikis, you must delete the underlying git repository.
- `CreateOrUpdatePage` for code wikis requires `VersionDescriptor` with `versionType:"branch"`. Project wikis don't require it.
- The `version` attribute in `betterado_wiki_page` schema serves as the branch reference for code wikis.
