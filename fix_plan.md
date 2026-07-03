# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_wiki and betterado_wiki_page framework resources are live WHEN the full live acceptance suite runs (TF_ACC=1) for wiki resources THEN TestAccWikiResource_projectWiki, TestAccWikiResource_codeWiki, TestAccWikiPageResource_basic, and TestAccWikiPageResource_update all pass with ExpectNonEmptyPlan: false on each idempotency check step
  - Fixed Delete for projectWiki to use DeleteWiki API directly (was deleting underlying repo, unreliable)
  - Fixed parallel collision: hclWikiPageBasic/Update now use codeWiki (ADO limits to 1 projectWiki per project; parallel tests collided)
- [x] AC2: GIVEN the live acceptance test for betterado_wiki runs WHEN the provider read-back step executes before destroy THEN testutils.CaptureLiveEvidence is called with label 'acceptance-resource-wiki' and a real ADO REST GET URL; .forge/live-evidence/acceptance-resource-wiki.json is written
  - captureWikiEvidence() already implemented in resource_wiki_test.go; calls testutils.CaptureLiveEvidence("acceptance-resource-wiki", url, wikiResp)
- [x] AC3: GIVEN the live acceptance test for betterado_wiki_page runs WHEN the provider read-back step executes before destroy THEN testutils.CaptureLiveEvidence is called with label 'acceptance-resource-wiki-page' and a real ADO REST GET URL; .forge/live-evidence/acceptance-resource-wiki-page.json is written
  - captureWikiPageEvidence() already implemented in resource_wiki_page_test.go; calls testutils.CaptureLiveEvidence("acceptance-resource-wiki-page", url, pageResp)
- [x] AC4: GIVEN all framework migrations are complete WHEN make docs is run then git checkout -- docs/guides/ THEN docs/resources/wiki.md and docs/resources/wiki_page.md reflect the current framework schema; no guides/ files are deleted
  - Created examples/resources/betterado_wiki/resource.tf and betterado_wiki_page/resource.tf
  - Ran make docs; docs/resources/wiki.md and wiki_page.md regenerated with framework schema (no timeouts block)
- [x] AC5: GIVEN a user-visible change was shipped WHEN CHANGELOG.md and PROVIDER_VERSION.txt are inspected THEN CHANGELOG.md has a new entry under ## Unreleased describing the wiki framework migration; PROVIDER_VERSION.txt has a bumped semver
  - CHANGELOG.md: added ## Unreleased section with betterado_wiki and betterado_wiki_page entries
  - PROVIDER_VERSION.txt: bumped 1.2.0 → 1.3.0
