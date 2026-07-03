# Demo — INIT-2026-07-01-migrate-framework-wiki

**Migrate betterado_wiki and betterado_wiki_page to terraform-plugin-framework**

> Both wiki resources (`betterado_wiki`, `betterado_wiki_page`) are migrated from Terraform Plugin SDK v2 to terraform-plugin-framework, served via the mux provider. SDKv2 implementations are deleted; live acceptance tests pass with idempotency verified (`ExpectNonEmptyPlan: false`); docs regenerated; changelog updated.

---

## Intent & Outcome

| Criterion | Verdict | Evidence |
|---|---|---|
| docs/wiki-gap-matrix.md lists every ADO Wiki API v7.1 field with status and resolution notes | ✅ met | `TestWikiGapMatrix_FileExists` → pass; docs/wiki-gap-matrix.md present (64 lines) with full field audit. |
| betterado_wiki framework resource: apply + idempotency + destroy via GetMuxedProviderFactories() | ✅ met | `TestAccWikiResource_projectWiki` and `TestAccWikiResource_codeWiki` → pass (live TF_ACC=1, 2026-07-03); `ExpectNonEmptyPlan: false` on each idempotency step. |
| SDKv2 resource_wiki.go and resource_wiki_test.go deleted; provider compiles; provider_test.go decremented | ✅ met | resource_wiki.go deleted (199 lines), resource_wiki_test.go deleted (107 lines); `go build -mod=vendor .` succeeds; betterado_wiki absent from expectedResources. |
| betterado_wiki_page framework resource: apply + update + idempotency + destroy via GetMuxedProviderFactories() | ✅ met | `TestAccWikiPageResource_basic` and `TestAccWikiPageResource_update` → pass (live TF_ACC=1, 2026-07-03); `ExpectNonEmptyPlan: false` on each idempotency step. |
| SDKv2 resource_wiki_page.go deleted; provider compiles; provider_test.go decremented | ✅ met | resource_wiki_page.go deleted (158 lines); `go build -mod=vendor .` succeeds; betterado_wiki_page absent from expectedResources. |
| Full live acceptance suite (4 tests) passes with ExpectNonEmptyPlan: false | ✅ met | All four acceptance tests pass (live TF_ACC=1, 2026-07-03); offline gate `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → ok. |
| CaptureLiveEvidence called for betterado_wiki; .forge/live-evidence/acceptance-resource-wiki.json written | ✅ met | .forge/live-evidence/acceptance-resource-wiki.json written at 2026-07-03T09:09:31Z; url=`https://dev.azure.com/davidgparsonson/.../_apis/wiki/wikis/d06c666f...?api-version=7.1`. |
| CaptureLiveEvidence called for betterado_wiki_page; .forge/live-evidence/acceptance-resource-wiki-page.json written | ✅ met | .forge/live-evidence/acceptance-resource-wiki-page.json written at 2026-07-03T09:09:34Z; url=`https://dev.azure.com/davidgparsonson/.../_apis/wiki/wikis/b5573ed1.../pages/30?includeContent=true&api-version=7.1`. |
| docs/resources/wiki.md and docs/resources/wiki_page.md reflect framework schema; no guides/ deleted | ✅ met | Both docs updated in diff (42+33 lines changed); examples created; docs/guides/ not in diff. |
| CHANGELOG.md has Unreleased entry; PROVIDER_VERSION.txt bumped | ✅ met | CHANGELOG.md `## [Unreleased]` has FEATURES entries for both wiki resources; PROVIDER_VERSION.txt bumped (2 lines changed in diff). |

---

## Checkpoints

### quality-gate

> Offline unit/integration gate — release and taskagent packages green

**Command:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`

---

### wiki-gap-matrix

> docs/wiki-gap-matrix.md exists and lists every ADO Wiki API v7.1 field

**Command:** `go test -tags all -count=1 -run TestWikiGapMatrix_FileExists ./azuredevops/internal/service/wiki/`

| | Before | After |
|---|---|---|
| | No wiki-gap-matrix.md existed; TestWikiGapMatrix_FileExists did not exist. | docs/wiki-gap-matrix.md present with full field audit; TestWikiGapMatrix_FileExists passes. |

---

### live-evidence-wiki

> Live ADO REST GET response for betterado_wiki (codeWiki) created by acceptance test

| | Before | After |
|---|---|---|
| | No live evidence captured; resource ran under SDKv2. | CaptureLiveEvidence called during TestAccWikiResource_codeWiki; acceptance-resource-wiki.json written with real ADO GET response. |

**Live evidence captured at 2026-07-03T09:09:31Z**

URL: `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/wiki/wikis/d06c666f-9d11-4076-8c1e-6e99e92dca48?api-version=7.1`

```json
{
  "mappedPath": "/",
  "name": "test-acc-6f33gi6m4z",
  "projectId": "6ddb680c-093d-4953-9561-2266eb7af800",
  "repositoryId": "b32d4e7b-ca08-4afc-b1f0-a4c39638ed12",
  "type": "codeWiki",
  "id": "d06c666f-9d11-4076-8c1e-6e99e92dca48",
  "remoteUrl": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_wiki/wikis/d06c666f-9d11-4076-8c1e-6e99e92dca48",
  "url": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/wiki/wikis/d06c666f-9d11-4076-8c1e-6e99e92dca48",
  "versions": [{"version": "master"}]
}
```

---

### live-evidence-wiki-page

> Live ADO REST GET response for betterado_wiki_page created by acceptance test

| | Before | After |
|---|---|---|
| | No live evidence captured; resource ran under SDKv2. | CaptureLiveEvidence called during TestAccWikiPageResource_basic; acceptance-resource-wiki-page.json written with real ADO GET response. |

**Live evidence captured at 2026-07-03T09:09:34Z**

URL: `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/wiki/wikis/b5573ed1-659c-4be6-bd88-9c380427333c/pages/30?includeContent=true&api-version=7.1`

```json
{
  "eTag": ["\"314d32ee03a8bd75dbe1b398b0f6309caff88f0b\""],
  "page": {
    "content": "contentupdate",
    "gitItemPath": "/page%2Dpath.md",
    "id": 30,
    "order": 0,
    "path": "/page-path",
    "remoteUrl": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_wiki/wikis/b5573ed1-659c-4be6-bd88-9c380427333c?pagePath=%2Fpage-path",
    "subPages": [],
    "url": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/wiki/wikis/b5573ed1-659c-4be6-bd88-9c380427333c/pages/%2Fpage-path"
  }
}
```

---

*Generated from `demo.json` — do not hand-edit.*
