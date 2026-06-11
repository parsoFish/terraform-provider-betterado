# Add release definition revision and history data sources

> _Derived from `demo.json` (ADR 021). Essence:_ Two new read-only Terraform data sources complete the release-definition read surface: `betterado_release_definition_revision` returns the raw JSON payload for a specific numbered revision, and `betterado_release_definition_history` returns the full audit-trail list of every revision. Both are registered in the provider, unit-tested with gomock, acceptance-tested against live ADO, and documented.

## Intent & Outcome

> _Assessed intent:_ Two new read-only Terraform data sources complete the release-definition read surface: `betterado_release_definition_revision` returns the raw JSON payload for a specific numbered revision, and `betterado_release_definition_history` returns the full audit-trail list of every revision. Both are registered in the provider, unit-tested with gomock, acceptance-tested against live ADO, and documented.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the SDK method GetDefinitionRevision(project, definitionId, revision) returning io.ReadCloser (JSON payload) WHEN the user specifies data.betterado_release_definition_revision with project_id, release_definition_id, revision THEN the data source calls GetDefinitionRevision, returns the raw JSON as a json_content attribute, and unit test with gomock verifies the SDK call path | ✓ met | TestDataReleaseDefinitionRevision_Read_ReturnsJSON → PASS (go test -tags all -v -count=1 ./azuredevops/internal/service/release/... ok 0.022s); TestDataReleaseDefinitionRevision_Read_Error → PASS (same run) |
| 2 | GIVEN the SDK method GetReleaseDefinitionHistory(project, definitionId) returning *[]ReleaseDefinitionRevision WHEN the user specifies data.betterado_release_definition_history with project_id, release_definition_id THEN the data source returns a list of revision objects (each: revision, changed_by, changed_date, change_type, comment); unit test verifies flatten logic | ✓ met | TestDataReleaseDefinitionHistory_Read_Populates → PASS; TestDataReleaseDefinitionHistory_Read_Empty → PASS; TestDataReleaseDefinitionHistory_Read_Error → PASS (go test -tags all -v -count=1 ./azuredevops/internal/service/release/... ok 0.022s) |
| 3 | GIVEN the new data sources WHEN the provider initialises THEN both data sources are registered in provider.go DataSourcesMap, provider_test.go count assertion updated, docs pages added under docs/data-sources/ | ✓ met | TestProvider_HasChildDataSources → PASS (provider.go DataSourcesMap +2 entries); TestDataSourceDocPagesExist → PASS (docs/data-sources/release_definition_revision.md 24 non-empty lines, docs/data-sources/release_definition_history.md 27 non-empty lines) |
| 4 | GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set WHEN TestAccDataReleaseDefinitionRevision_Basic runs against real ADO THEN the test creates a release definition, reads back revision 1 via data.betterado_release_definition_revision, confirms json_content is non-empty, and an idempotency re-plan produces no diff | ~ partial | Acceptance test file azuredevops/internal/acceptancetests/data_release_definition_revision_history_test.go committed; TestAccDataReleaseDefinitionRevision_Basic skips when TF_ACC unset (resource.ParallelTest). WI-2 marked complete by the dev-loop which ran it with live credentials; CI environment lacks TF_ACC — no hallucinated pass. Live confirmation: chore(WI-2) commit b65fb1a5 marks both live acceptance ACs complete. |
| 5 | GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set WHEN TestAccDataReleaseDefinitionHistory_Basic runs against real ADO THEN the test creates a release definition, reads its full history via data.betterado_release_definition_history, confirms at least one revision entry exists with a non-empty revision number, and an idempotency re-plan produces no diff | ~ partial | Acceptance test file committed; TestAccDataReleaseDefinitionHistory_Basic skips when TF_ACC unset. WI-2 marked complete by dev-loop with live credentials (commit 41abfc9d). Live evidence requires TF_ACC=1 + real ADO org; skipped in unifier CI environment — not a false-pass. |

## Test Evidence

### gomock verifies GetDefinitionRevision SDK call and json_content population

- **Before:** No data source existed; calling GetDefinitionRevision was not surfaced to Terraform consumers.
- **After:** TestDataReleaseDefinitionRevision_Read_ReturnsJSON and _Read_Error both pass (go test -tags all -v -count=1 ./azuredevops/internal/service/release/... ok 0.022s).

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataReleaseDefinitionRevision_Read_ReturnsJSON | test did not exist | PASS (0.00s) | — | new |
| TestDataReleaseDefinitionRevision_Read_Error | test did not exist | PASS (0.00s) | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### gomock verifies GetReleaseDefinitionHistory SDK call, flatten logic, empty slice, and error path

- **Before:** No data source existed; the revision audit trail was inaccessible via Terraform.
- **After:** TestDataReleaseDefinitionHistory_Read_Populates, _Read_Empty, and _Read_Error all pass (go test -tags all -v -count=1 ./azuredevops/internal/service/release/... ok 0.022s).

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataReleaseDefinitionHistory_Read_Populates | test did not exist | PASS (0.00s) | — | new |
| TestDataReleaseDefinitionHistory_Read_Empty | test did not exist | PASS (0.00s) | — | new |
| TestDataReleaseDefinitionHistory_Read_Error | test did not exist | PASS (0.00s) | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### TestDataSourceDocPagesExist confirms both markdown doc pages exist with ≥10 non-empty lines

- **Before:** TestDataSourceDocPagesExist did not exist; doc pages were absent.
- **After:** TestDataSourceDocPagesExist passes: docs/data-sources/release_definition_revision.md (24 non-empty lines) and docs/data-sources/release_definition_history.md (27 non-empty lines) both present.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestDataSourceDocPagesExist | test did not exist | PASS (0.00s) — 24 non-empty lines in revision.md, 27 in history.md | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### TestProvider_HasChildDataSources count assertion passes with both new entries

- **Before:** Provider DataSourcesMap lacked betterado_release_definition_revision and betterado_release_definition_history.
- **After:** Both entries added to provider.go DataSourcesMap and provider_test.go expectedDataSources; TestProvider_HasChildDataSources passes.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestProvider_HasChildDataSources | 45 data sources | 47 data sources (+2) | +4.4% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... — all packages pass

- **Before:** The new data-source tests did not exist; the quality gate covered only pre-existing tests in release and taskagent packages.
- **After:** All three packages green: release (0.022s, includes all 6 new test functions), taskagent (0.010s), taskagent/validate (0.004s). Zero failures.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| github.com/.../service/release | ok (pre-existing tests only) | ok 0.022s (+6 new test functions) | — | match |
| github.com/.../service/taskagent | ok 0.010s | ok 0.010s | 0.0% | match |
| github.com/.../service/taskagent/validate | ok 0.004s | ok 0.004s | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

## Test Evidence

| test | result | delta |
|---|---|---|
| TestDataReleaseDefinitionRevision_Read_ReturnsJSON | pass | +1 new test (gomock: GetDefinitionRevision returns io.ReadCloser → json_content populated) |
| TestDataReleaseDefinitionRevision_Read_Error | pass | +1 new test (gomock: GetDefinitionRevision returns error → resource ID cleared) |
| TestDataReleaseDefinitionHistory_Read_Populates | pass | +1 new test (gomock: GetReleaseDefinitionHistory returns *[]ReleaseDefinitionRevision → revisions list flattened) |
| TestDataReleaseDefinitionHistory_Read_Empty | pass | +1 new test (gomock: empty slice → revisions list empty, no error) |
| TestDataReleaseDefinitionHistory_Read_Error | pass | +1 new test (gomock: SDK error → error surfaced) |
| TestDataSourceDocPagesExist | pass | +1 new audit gate (docs/data-sources/release_definition_revision.md: 24 non-empty lines ≥ 10 ✓; docs/data-sources/release_definition_history.md: 27 non-empty lines ≥ 10 ✓) |
| TestProvider_HasChildDataSources | pass | +2 entries in expectedDataSources (betterado_release_definition_revision, betterado_release_definition_history) |
| go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... | pass | All 3 packages green; +6 new test functions in service/release |
| TestAccDataReleaseDefinitionRevision_Basic | skip | +1 new acceptance test (requires TF_ACC=1 + live ADO credentials; WI-2 confirmed live by dev-loop) |
| TestAccDataReleaseDefinitionHistory_Basic | skip | +1 new acceptance test (requires TF_ACC=1 + live ADO credentials; WI-2 confirmed live by dev-loop) |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/release/data_release_definition_revision.go` — New data source: DataReleaseDefinitionRevision() — wraps GetDefinitionRevision, exposes json_content
- `azuredevops/internal/service/release/data_release_definition_revision_test.go` — Unit tests for revision data source (gomock): Read_ReturnsJSON, Read_Error
- `azuredevops/internal/service/release/data_release_definition_history.go` — New data source: DataReleaseDefinitionHistory() — wraps GetReleaseDefinitionHistory, flattens to revisions list
- `azuredevops/internal/service/release/data_release_definition_history_test.go` — Unit tests for history data source (gomock): Read_Populates, Read_Empty, Read_Error
- `azuredevops/internal/service/release/doc_audit_test.go` — Added TestDataSourceDocPagesExist gate (verifies ≥10 non-empty lines in each doc page)
- `azuredevops/provider.go` — Registered betterado_release_definition_revision and betterado_release_definition_history in DataSourcesMap
- `azuredevops/provider_test.go` — Added both entries to expectedDataSources in TestProvider_HasChildDataSources
- `docs/data-sources/release_definition_revision.md` — New documentation page for revision data source
- `docs/data-sources/release_definition_history.md` — New documentation page for history data source
- `examples/data-sources/betterado_release_definition_revision/main.tf` — HCL usage example: look up a specific revision by number, output json_content
- `examples/data-sources/betterado_release_definition_history/main.tf` — HCL usage example: list full revision audit trail, output revisions
- `azuredevops/internal/acceptancetests/data_release_definition_revision_history_test.go` — Live acceptance tests: TestAccDataReleaseDefinitionRevision_Basic, TestAccDataReleaseDefinitionHistory_Basic (require TF_ACC=1)

```
azuredevops/internal/acceptancetests/data_release_definition_revision_history_test.go |  88 +++++
 azuredevops/internal/service/release/data_release_definition_history.go                 | 120 +++++++
 azuredevops/internal/service/release/data_release_definition_history_test.go            | 167 ++++++++++
 azuredevops/internal/service/release/data_release_definition_revision.go                |  72 ++++
 azuredevops/internal/service/release/data_release_definition_revision_test.go           |  99 ++++++
 azuredevops/internal/service/release/doc_audit_test.go                                  |  70 ++++
 azuredevops/provider.go                                                                  |   2 +
 azuredevops/provider_test.go                                                             |   2 +
 docs/data-sources/release_definition_history.md                                          |  40 +++
 docs/data-sources/release_definition_revision.md                                         |  37 +++
 examples/data-sources/betterado_release_definition_history/main.tf                       |  17 +
 examples/data-sources/betterado_release_definition_revision/main.tf                      |  14 +
 12 files changed, 728 insertions(+)
```

## Usage

```
```hcl
# Look up a specific numbered revision of a release definition (returns raw JSON)
data "betterado_release_definition_revision" "snapshot" {
  project_id            = data.betterado_project.my_project.id
  release_definition_id = 42
  revision              = 3
}

output "revision_json" {
  # Full JSON payload for revision 3 — suitable for diffing or archiving
  value = data.betterado_release_definition_revision.snapshot.json_content
}

# List the full audit trail of a release definition
data "betterado_release_definition_history" "audit" {
  project_id            = data.betterado_project.my_project.id
  release_definition_id = 42
}

output "last_changed_by" {
  value = data.betterado_release_definition_history.audit.revisions[0].changed_by
}

output "all_revisions" {
  value = data.betterado_release_definition_history.audit.revisions
  # Each entry: { revision (int), changed_by (string), changed_date (RFC3339),
  #               change_type (string), comment (string) }
}
```
```

## Impact

- Consumers can now read any historical revision of a release definition as raw JSON, enabling diff-based auditing or rollback automation in Terraform.
- The full revision history (who changed what and when) is now queryable directly from Terraform state, supporting compliance and change-management workflows.
- Completes the read surface for release definitions — every SDK read method (`GetDefinitionRevision`, `GetReleaseDefinitionHistory`) is now surfaced as a Terraform data source.
- A doc-audit gate (`TestDataSourceDocPagesExist`) guards against future regressions where a data source ships without a `docs/data-sources/` markdown page.
- Two new HCL example configs (`examples/data-sources/betterado_release_definition_revision/main.tf` and `…history/main.tf`) give users working starting points.
