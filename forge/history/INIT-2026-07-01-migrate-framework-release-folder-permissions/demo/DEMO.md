# Demo — INIT-2026-07-01-migrate-framework-release-folder-permissions

> **Migrate release folder, permissions, and 5 data sources to terraform-plugin-framework**

## Essence

Two SDKv2 resources (`betterado_release_folder`, `betterado_release_definition_permissions`) and five SDKv2 data sources (`betterado_release_definition`, `betterado_release_definition_history`, `betterado_release_definition_revision`, `betterado_release_definitions`, `betterado_release_folder` data source) are now served by the mux provider via terraform-plugin-framework. Live acceptance tests ran (TF_ACC=1): `TestAccReleaseFolderFramework`, `TestAccReleaseDefinitionPermissionsFramework`, all six `TestAccData*` data-source tests, and `TestAccMuxSdkv2Passthrough` — all passed. Live evidence captured at 2026-07-01T21:25:43Z: ADO REST GET of release folder `\MuxSmokeTest` confirmed existence post-apply at vsrm.dev.azure.com.

## Diff stat

28 files changed, 2116 insertions(+), 206 deletions(-)

---

## Checkpoint 1 — Offline quality gate

**Caption:** Offline unit tests for release and taskagent packages pass on branch HEAD (the gate forge ran, verbatim)

**Command (before/after evidence):**
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```

| | |
|---|---|
| **Before (main)** | Framework datasource and resource files did not exist; only SDKv2 paths compiled |
| **After (HEAD)** | `ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.006s` \| `ok .../taskagent 0.005s` \| `ok .../taskagent/validate 0.004s` — all three packages green |

---

## Checkpoint 2 — Live resource read-back

**Caption:** Live release folder created via framework resource; ADO REST GET confirms folder exists at vsrm endpoint

**Live evidence (captured 2026-07-01T21:25:43Z):**

- **REST GET:** `https://vsrm.dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/release/folders%5CMuxSmokeTest?api-version=7.1`
- **Response:**
  ```json
  {
    "createdBy": {
      "displayName": "david.g.parsonson",
      "uniqueName": "david.g.parsonson@gmail.com",
      "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1"
    },
    "createdOn": "2026-07-01T21:25:42.927Z",
    "lastChangedDate": "0001-01-01T00:00:00Z",
    "path": "\\MuxSmokeTest"
  }
  ```

| | |
|---|---|
| **Before (main)** | `betterado_release_folder` was SDKv2-only; no framework path existed |
| **After (HEAD)** | Folder `\MuxSmokeTest` created via mux→framework provider path; GET response confirms path, createdBy, and createdOn match the apply. `TestAccMuxSdkv2Passthrough` idempotency re-plan: `ExpectNonEmptyPlan: false` → PASS |

---

## Intent & Outcome — AC Evaluations

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| AC1 | GIVEN `betterado_release_folder` block, mux active WHEN `terraform apply` THEN folder created via vsrm API; read-back succeeds; re-plan `ExpectNonEmptyPlan: false`; destroy clean | **met** | `TestAccReleaseFolderFramework` live (TF_ACC=1): apply → idempotency re-plan → destroy all passed. Live GET at vsrm endpoint confirmed folder existence. |
| AC2 | GIVEN `TestAccReleaseFolderFramework` with TF_ACC=1 WHEN cycle completes THEN test passes; `CaptureLiveEvidence` called; `.forge/live-evidence/acceptance-resource.json` written | **met** | Test passed. `.forge/live-evidence/acceptance-resource.json` present (capturedAt: 2026-07-01T21:25:43Z, path: `\MuxSmokeTest`). |
| AC3 | GIVEN CI-equivalent gate (no TF_ACC) WHEN `make test && golangci-lint && terrafmt-check` THEN all pass, zero new lint findings (WI-1) | **met** | `go test` → ok (3 packages). `golangci-lint --new-from-rev=main` + `terrafmt-check` passed on all changed framework files. |
| AC4 | GIVEN `betterado_release_definition_permissions` with 13 ACL bits, mux active WHEN `terraform apply` THEN ACL applied idempotently; `TestAccReleaseDefinitionPermissionsFramework` passes live | **met** | `TestAccReleaseDefinitionPermissionsFramework` live (TF_ACC=1): all 13 ACL bits exercised, `ExpectNonEmptyPlan: false`, destroy clean. |
| AC5 | GIVEN live acceptance test WHEN read-back executes THEN `CaptureLiveEvidence("acceptance-resource", …)` called; `.forge/live-evidence/acceptance-resource.json` written | **met** | `.forge/live-evidence/acceptance-resource.json` written at 2026-07-01T21:25:43Z via `captureReleaseDefinitionPermissionsFrameworkEvidence`. |
| AC6 | GIVEN CI-equivalent gate (no TF_ACC) WHEN gate runs THEN all pass, zero new lint (WI-2) | **met** | `go test` (offline) + `golangci-lint --new-from-rev=main` + `terrafmt-check` green on `resource_release_definition_permissions_framework.go`. |
| AC7 | GIVEN 5 data sources migrated to `datasource.DataSource`, registered in `DataSources()` WHEN each read live (TF_ACC=1) THEN same fields as SDKv2; all 6 `TestAccData*` tests pass | **met** | All 6 `TestAccData*` passed live (TF_ACC=1, `GetMuxProviderFactories()`). Five datasource framework files in branch diff; all registered in `DataSources()`. |
| AC8 | GIVEN acceptance tests updated to `GetMuxProviderFactories()` WHEN any run under mux THEN compile + pass; `ExpectNonEmptyPlan: false` | **met** | `data_release_definition_test.go`, `data_release_definition_revision_history_test.go`, `data_release_folder_test.go` updated and compiled/passed. |
| AC9 | GIVEN CI-equivalent gate (no TF_ACC) WHEN gate runs THEN all pass, zero new lint (WI-3) | **met** | `go test` → ok; `golangci-lint` + `terrafmt-check` passed on all five datasource framework files. |
| AC10 | GIVEN `TestAccMuxSdkv2Passthrough` WHEN run after migration THEN passes; no SDKv2 regression | **met** | `TestAccMuxSdkv2Passthrough` passed live (TF_ACC=1). Migrated to `ProtoV6ProviderFactories` + `SharedReleaseFixture`. |
| AC11 | GIVEN `make docs` + `git checkout -- docs/guides/` WHEN docs inspected THEN every migrated attribute documented; guides restored | **met** | Branch diff: 7 doc files updated (`docs/resources/release_folder.md`, `docs/resources/release_definition_permissions.md`, 5 data-source docs). `docs/guides/` restored. |
| AC12 | GIVEN `examples/resources/betterado_release_folder/resource.tf` + `data-source.tf` exist WHEN reviewed THEN valid HCL, embedded in docs | **met** | Both example files present in branch diff (renamed to correct tfplugindocs convention). Embedded in `docs/resources/release_folder.md`. |
| AC13 | GIVEN `CHANGELOG.md` + `PROVIDER_VERSION.txt` inspected WHEN complete THEN `CHANGELOG.md` has `## Unreleased` for 2 resources + 5 data sources; version bumped one minor | **met** | `CHANGELOG.md` has `## [Unreleased]` with all 7 entries. `PROVIDER_VERSION.txt` = `1.1.0`. |

---

## Test evidence

| Test | Result |
|------|--------|
| `go test -tags all -count=1 ./azuredevops/internal/service/release/...` (offline) | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/...` (offline) | pass |
| `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/validate/...` (offline) | pass |
| `TestAccReleaseFolderFramework` (TF_ACC=1, live) | pass |
| `TestAccReleaseDefinitionPermissionsFramework` (TF_ACC=1, live) | pass |
| `TestAccDataReleaseDefinition_ById` (TF_ACC=1, live, mux provider) | pass |
| `TestAccDataReleaseDefinition_ByName` (TF_ACC=1, live, mux provider) | pass |
| `TestAccDataReleaseDefinitions_List` (TF_ACC=1, live, mux provider) | pass |
| `TestAccDataReleaseDefinitionRevision_Basic` (TF_ACC=1, live, mux provider) | pass |
| `TestAccDataReleaseDefinitionHistory_Basic` (TF_ACC=1, live, mux provider) | pass |
| `TestAccDataReleaseFolder_Basic` (TF_ACC=1, live, mux provider) | pass |
| `TestAccMuxSdkv2Passthrough` (TF_ACC=1, live, regression gate) | pass |
