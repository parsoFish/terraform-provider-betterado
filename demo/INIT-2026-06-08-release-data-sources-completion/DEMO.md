# Add release definition revision and history data sources

> _Derived from `demo.json` (ADR 021). Essence:_ Two new read-only Terraform data sources complete the release-definition read surface: `betterado_release_definition_revision` returns the raw JSON payload for a specific numbered revision, and `betterado_release_definition_history` returns the full audit-trail list of every revision. Both are registered in the provider, unit-tested with gomock, and documented.

## Intent & Outcome

> _Assessed intent:_ Two new read-only Terraform data sources complete the release-definition read surface: `betterado_release_definition_revision` returns the raw JSON payload for a specific numbered revision, and `betterado_release_definition_history` returns the full audit-trail list of every revision. Both are registered in the provider, unit-tested with gomock, and documented.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN a gomock ReleaseClient whose GetDefinitionRevision returns an io.ReadCloser with JSON bytes WHEN dataReleaseDefinitionRevisionRead is called with project_id, release_definition_id, revision THEN the resource data json_content attribute contains the drained JSON string and no error is returned | ✓ met | TestDataReleaseDefinitionRevision_Read_ReturnsJSON → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 2 | GIVEN a gomock ReleaseClient whose GetDefinitionRevision returns an error WHEN dataReleaseDefinitionRevisionRead is called THEN an error is returned and the resource ID is cleared | ✓ met | TestDataReleaseDefinitionRevision_Read_Error → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 3 | GIVEN both new data sources registered in provider.go DataSourcesMap WHEN azuredevops.Provider() is called and DataSourcesMap is inspected THEN betterado_release_definition_revision is present (provider_test.go list + count updated) | ✓ met | TestProvider_HasChildDataSources → PASS; betterado_release_definition_revision present in expectedDataSources slice and provider.go DataSourcesMap |
| 4 | GIVEN a gomock ReleaseClient whose GetReleaseDefinitionHistory returns a *[]ReleaseDefinitionRevision slice WHEN dataReleaseDefinitionHistoryRead is called with project_id, release_definition_id THEN the revisions list attribute contains one entry per revision with revision number, changed_by display name, changed_date (RFC3339), change_type string, and comment; no error is returned | ✓ met | TestDataReleaseDefinitionHistory_Read_Populates → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 5 | GIVEN a gomock ReleaseClient whose GetReleaseDefinitionHistory returns an empty slice WHEN dataReleaseDefinitionHistoryRead is called THEN the revisions list is empty and no error is returned | ✓ met | TestDataReleaseDefinitionHistory_Read_Empty → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 6 | GIVEN a gomock ReleaseClient whose GetReleaseDefinitionHistory returns an error WHEN dataReleaseDefinitionHistoryRead is called THEN an error is returned | ✓ met | TestDataReleaseDefinitionHistory_Read_Error → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 7 | GIVEN betterado_release_definition_history registered in provider.go DataSourcesMap WHEN azuredevops.Provider() is called and DataSourcesMap is inspected THEN betterado_release_definition_history is present (provider_test.go list includes it; exact-count assertion passes) | ✓ met | TestProvider_HasChildDataSources → PASS; betterado_release_definition_history present in expectedDataSources slice and provider.go DataSourcesMap |
| 8 | GIVEN docs/data-sources/release_definition_revision.md and docs/data-sources/release_definition_history.md are created WHEN TestDataSourceDocPagesExist runs THEN both files exist and each contains at least 10 non-empty lines; test passes | ✓ met | TestDataSourceDocPagesExist → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s); release_definition_revision.md has 37 lines, release_definition_history.md has 40 lines |
| 9 | GIVEN the new audit test function TestDataSourceDocPagesExist is added to doc_audit_test.go WHEN go test -tags all -run TestDataSourceDocPagesExist ./azuredevops/internal/service/release/ runs THEN exit 0 with PASS line | ✓ met | TestDataSourceDocPagesExist → PASS (go test -tags all -count=1 ./azuredevops/internal/service/release/... ok 0.020s) |
| 10 | GIVEN a live ADO org with TF_ACC=1 WHEN TestAccDataReleaseDefinitionRevision_Basic runs THEN json_content is a non-empty JSON string; idempotency re-plan produces no diff; destroy succeeds | ~ partial | WI-4 status: failed (no live ADO credentials in CI). Acceptance test file azuredevops/internal/acceptancetests/data_release_definition_revision_history_test.go committed; TestAccDataReleaseDefinitionRevision_Basic is skipped when TF_ACC unset (resource.ParallelTest skips automatically). Live run requires TF_ACC=1 + credentials. |
| 11 | GIVEN a live ADO org with credentials set WHEN TestAccDataReleaseDefinitionHistory_Basic runs THEN revisions list is non-empty; each entry has revision >= 1; idempotency re-plan produces no diff; destroy succeeds | ~ partial | WI-4 status: failed (no live ADO credentials in CI). Acceptance test file committed; TestAccDataReleaseDefinitionHistory_Basic is skipped when TF_ACC unset. Live run requires TF_ACC=1 + credentials. |

## Test Evidence

### gomock verifies GetDefinitionRevision SDK call and json_content population

- **Before:** No data source existed; calling GetDefinitionRevision was not surfaced to Terraform consumers.
- **After:** TestDataReleaseDefinitionRevision_Read_ReturnsJSON and _Read_Error both pass (go test -tags all ./azuredevops/internal/service/release/... → ok 0.020s).

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataReleaseDefinitionRevision_Read_ReturnsJSON | test did not exist | PASS | — | new |
| TestDataReleaseDefinitionRevision_Read_Error | test did not exist | PASS | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### gomock verifies GetReleaseDefinitionHistory SDK call, flatten logic, empty slice, and error path

- **Before:** No data source existed; the revision audit trail was inaccessible via Terraform.
- **After:** TestDataReleaseDefinitionHistory_Read_Populates, _Read_Empty, and _Read_Error all pass (go test -tags all ./azuredevops/internal/service/release/... → ok 0.020s).

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataReleaseDefinitionHistory_Read_Populates | test did not exist | PASS | — | new |
| TestDataReleaseDefinitionHistory_Read_Empty | test did not exist | PASS | — | new |
| TestDataReleaseDefinitionHistory_Read_Error | test did not exist | PASS | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### TestDataSourceDocPagesExist confirms both markdown doc pages exist with ≥10 non-empty lines

- **Before:** TestDataSourceDocPagesExist did not exist; doc pages were absent.
- **After:** TestDataSourceDocPagesExist passes: docs/data-sources/release_definition_revision.md (37 lines) and docs/data-sources/release_definition_history.md (40 lines) both present.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataSourceDocPagesExist | test did not exist | PASS | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### TestProvider_HasChildDataSources count assertion passes with both new entries

- **Before:** Provider DataSourcesMap had 45 entries; betterado_release_definition_revision and betterado_release_definition_history were absent.
- **After:** Both entries added to provider.go DataSourcesMap and provider_test.go expectedDataSources; TestProvider_HasChildDataSources passes.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestProvider_HasChildDataSources | 45 data sources | 47 data sources | +4.4% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

## Test Evidence

| test | result | delta |
|---|---|---|
| go test -tags all -count=1 ./azuredevops/internal/service/release/... | pass | +5 new test functions (TestDataReleaseDefinitionRevision_Read_ReturnsJSON, TestDataReleaseDefinitionRevision_Read_Error, TestDataReleaseDefinitionHistory_Read_Populates, TestDataReleaseDefinitionHistory_Read_Empty, TestDataReleaseDefinitionHistory_Read_Error, TestDataSourceDocPagesExist) |
| go test -tags all -count=1 ./azuredevops/internal/service/taskagent/... | pass | 0 (unaffected) |
| TestProvider_HasChildDataSources | pass | +2 entries in expectedDataSources (betterado_release_definition_revision, betterado_release_definition_history) |
| TestAccDataReleaseDefinitionRevision_Basic | skip | +1 new acceptance test (requires TF_ACC=1) |
| TestAccDataReleaseDefinitionHistory_Basic | skip | +1 new acceptance test (requires TF_ACC=1) |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/release/data_release_definition_revision.go` — New data source: DataReleaseDefinitionRevision()
- `azuredevops/internal/service/release/data_release_definition_revision_test.go` — Unit tests for revision data source (gomock)
- `azuredevops/internal/service/release/data_release_definition_history.go` — New data source: DataReleaseDefinitionHistory()
- `azuredevops/internal/service/release/data_release_definition_history_test.go` — Unit tests for history data source (gomock)
- `azuredevops/internal/service/release/doc_audit_test.go` — Added TestDataSourceDocPagesExist gate
- `azuredevops/provider.go` — Registered both new data sources in DataSourcesMap
- `azuredevops/provider_test.go` — Added both entries to expectedDataSources in TestProvider_HasChildDataSources
- `docs/data-sources/release_definition_revision.md` — New documentation page
- `docs/data-sources/release_definition_history.md` — New documentation page

```
.../release/data_release_definition_history.go     | 120 +++++++++++++++
 .../data_release_definition_history_test.go        | 167 +++++++++++++++++++++
 .../release/data_release_definition_revision.go    |  72 +++++++++
 .../data_release_definition_revision_test.go       |  99 ++++++++++++
 .../internal/service/release/doc_audit_test.go     |  67 +++++++++
 azuredevops/provider.go                            |   2 +
 azuredevops/provider_test.go                       |   2 +
 docs/data-sources/release_definition_history.md    |  40 +++++
 docs/data-sources/release_definition_revision.md   |  37 +++++
 9 files changed, 606 insertions(+)
```

## Usage

```
```hcl
# Look up a specific revision of a release definition
data "betterado_release_definition_revision" "snapshot" {
  project_id            = "00000000-0000-0000-0000-000000000000"
  release_definition_id = 42
  revision              = 3
}

output "revision_json" {
  value = data.betterado_release_definition_revision.snapshot.json_content
}

# List the full audit trail of a release definition
data "betterado_release_definition_history" "audit" {
  project_id            = "00000000-0000-0000-0000-000000000000"
  release_definition_id = 42
}

output "last_changed_by" {
  value = data.betterado_release_definition_history.audit.revisions[0].changed_by
}
```
```

## Impact

- Consumers can now read any historical revision of a release definition as raw JSON, enabling diff-based auditing or rollback automation in Terraform.
- The full revision history (who changed what and when) is now queryable directly from Terraform state, supporting compliance and change-management workflows.
- Completes the read surface for release definitions — every SDK read method is now surfaced as a data source.
- Doc pages and a doc-audit gate (TestDataSourceDocPagesExist) guard against future regressions where a data source ships without documentation.
