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
   - **Fix**: Simplified `Delete` to call `DeleteWiki` directly for ALL wiki types (both `codeWiki` and `projectWiki`). This is the correct API call and matches what ADO expects.

**Other ACs completed in this iteration:**
- AC2/AC3: `captureWikiEvidence` and `captureWikiPageEvidence` were already implemented in prior WI iterations. No changes needed.
- AC4: Created `examples/resources/betterado_wiki/resource.tf` and `examples/resources/betterado_wiki_page/resource.tf`, ran `make docs` to regenerate `docs/resources/wiki.md` and `docs/resources/wiki_page.md` with framework schema (no timeouts block).
- AC5: Added `## Unreleased` section to `CHANGELOG.md` with wiki migration entries; bumped `PROVIDER_VERSION.txt` from `1.2.0` to `1.3.0`.

## What worked

- Using `DeleteWiki` API directly for projectWiki (not repo deletion)
- Switching page test HCL to `codeWiki` to avoid parallel test collisions

## What didn't work

- Original projectWiki Delete: deleting underlying git repository via `GitReposClient.DeleteRepository` - doesn't reliably remove the wiki

## Open questions

_(none)_

## Notes for reflection

- ADO "one project wiki per project" limit is a key gotcha for acceptance tests using a shared standing project. Document this in the wiki gap matrix or testing conventions.
- `DeleteWiki` API works for both project and code wikis - no need to differentiate in delete logic.
