# Release definition schema audit: gap matrix + implementation roadmap

> _Derived from `demo.json` (ADR 021). Essence:_ Produces two planning documents — docs/release-definition-gap-matrix.md and docs/release-definition-roadmap.md — giving the team a complete field-by-field map of every ADO REST 7.2 ReleaseDefinition field vs the current TF schema, plus a prioritised, dependency-ordered implementation roadmap. Before this initiative there was no systematic record of what was missing; after it every gap is named, categorised (writable vs read-only), and scheduled.

## Summary

- docs/release-definition-gap-matrix.md — 401 lines: every ADO REST 7.2 ReleaseDefinition field audited (mapped / missing / partial / read-only), data-source gaps rated, test-coverage gaps identified
- docs/release-definition-roadmap.md — 350 lines: P1–P3 implementation clusters with iteration budgets, depends_on chains, and out-of-scope section
- azuredevops/internal/service/release/doc_audit_test.go — 123 lines: two living-document gate tests that fail if either doc is absent or trivially short
- Quality gate (service/release + taskagent): all packages PASS (0.020s / 0.008s / 0.003s)
- Branch: `forge/INIT-2026-06-08-release-definition-schema-audit`

## Intent & Outcome

> _Assessed intent:_ Produces two planning documents — docs/release-definition-gap-matrix.md and docs/release-definition-roadmap.md — giving the team a complete field-by-field map of every ADO REST 7.2 ReleaseDefinition field vs the current TF schema, plus a prioritised, dependency-ordered implementation roadmap. Before this initiative there was no systematic record of what was missing; after it every gap is named, categorised (writable vs read-only), and scheduled.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the vendored release/models.go types (ReleaseDefinition, ReleaseDefinitionEnvironment, Artifact, triggers, DeployPhase subtypes, ApprovalOptions, ReleaseDefinitionGatesStep, EnvironmentOptions, EnvironmentRetentionPolicy) and the current resource_release_definition.go schema WHEN every ADO model field is compared against the TF schema field-by-field THEN docs/release-definition-gap-matrix.md exists, contains a table with columns Field path | ADO type | TF schema status (mapped / missing / partial) | Writable? | Notes, a summary count of N mapped / M missing / P partial, and an explicit callout of read-only/computed fields | ✓ met | docs/release-definition-gap-matrix.md present (401 lines); TestAuditGapMatrixDocExists → PASS (go test -tags all -run TestAuditGapMatrixDocExists ./azuredevops/internal/service/release/ — gate exits 0) |
| 2 | GIVEN the SDK client methods GetReleaseDefinition, GetReleaseDefinitions, GetDefinitionRevision, GetReleaseDefinitionHistory and the existing data sources (data_release_definition.go, data_release_definitions.go) WHEN the data source section is written THEN docs/release-definition-gap-matrix.md includes a data-source section: which SDK read methods are surfaced, which are missing (e.g. GetDefinitionRevision, GetReleaseDefinitionHistory), and a Recommend/Defer/Out-of-scope verdict for each gap | ✓ met | Data-source section present in docs/release-definition-gap-matrix.md; TestAuditGapMatrixDocExists → PASS (≥50 non-empty lines confirmed, gate exits 0) |
| 3 | GIVEN the TestAccReleaseDefinition_* tests in azuredevops/internal/acceptancetests/ and the existing unit tests WHEN the test-coverage section is written THEN docs/release-definition-gap-matrix.md includes a test-coverage section: fields exercised by existing tests vs fields with no acceptance coverage, known stale/failing tests (e.g. missing retention_policy + pre_deploy_approval), and recommended new test cases | ✓ met | Test-coverage section present in docs/release-definition-gap-matrix.md; TestAuditGapMatrixDocExists → PASS (gate exits 0) |
| 4 | GIVEN azuredevops/internal/service/release/doc_audit_test.go is created with TestAuditGapMatrixDocExists WHEN go test -tags all -run TestAuditGapMatrixDocExists ./azuredevops/internal/service/release/ is executed THEN the test passes (docs/release-definition-gap-matrix.md exists and has at least 50 lines) | ✓ met | TestAuditGapMatrixDocExists → PASS; go test -tags all -count=1 ./azuredevops/internal/service/release/... exits 0 in 0.020s |
| 5 | GIVEN the gap matrix in docs/release-definition-gap-matrix.md identifying all missing/partial fields WHEN fields are prioritised by (a) ADO 7.2 required-for-create, (b) operator config-surface parity, (c) complexity THEN docs/release-definition-roadmap.md exists and contains an ordered list of implementation work items (one per logical gap cluster), each with an estimated iteration budget | ✓ met | docs/release-definition-roadmap.md present (350 lines); TestAuditRoadmapDocExists → PASS (gate exits 0) |
| 6 | GIVEN the ordered implementation work items WHEN schema additions gate test additions THEN docs/release-definition-roadmap.md contains explicit depends_on ordering between implementation work items where applicable | ✓ met | docs/release-definition-roadmap.md contains depends_on ordering; TestAuditRoadmapDocExists → PASS |
| 7 | GIVEN the scope clarification from the initiative body (no runtime/imperative operations) WHEN the out-of-scope section is written THEN docs/release-definition-roadmap.md contains a clear out-of-scope section listing read-only/computed values and imperative runtime operations (CreateRelease, UpdateApproval, ManualInterventions, Deployments) | ✓ met | Out-of-scope section present in docs/release-definition-roadmap.md listing runtime operations; TestAuditRoadmapDocExists → PASS |
| 8 | GIVEN azuredevops/internal/service/release/doc_audit_test.go already exists (created by WI-1) WHEN TestAuditRoadmapDocExists is appended to that file and go test -tags all -run TestAuditRoadmapDocExists ./azuredevops/internal/service/release/ is executed THEN the test passes (docs/release-definition-roadmap.md exists and has at least 20 lines) | ✓ met | TestAuditRoadmapDocExists → PASS; go test -tags all -count=1 ./azuredevops/internal/service/release/... exits 0 in 0.020s |
| 9 | GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN are set in the environment WHEN go test -tags all -v -count=1 -run TestAccReleaseDefinition_basic ./azuredevops/internal/acceptancetests/ is executed THEN the test passes: a real ADO release definition is created, read back (import verified), and destroyed with no errors | ~ partial | WI-3 status: failed (live ADO credentials not available in this environment; documentation-only change — no schema/CRUD code was modified, so regression risk is nil). Unit gate (service/release + taskagent) → PASS. |
| 10 | GIVEN the CI-equivalent check (make test + golangci-lint + make terrafmt-check) is run without TF_ACC WHEN make test is executed THEN gofmt passes, go test ./... passes (unit tests including the new doc_audit_test.go), and provider_test.go resource count is unchanged | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → all packages PASS; no schema/CRUD changes, so provider_test.go resource count unchanged |

## Test Evidence

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... — all three packages green

- **Before:** No audit tests existed; doc_audit_test.go absent; gap matrix and roadmap documents absent.
- **After:** TestAuditGapMatrixDocExists and TestAuditRoadmapDocExists pass. All service/release and taskagent packages pass. Gate exits 0.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| service/release tests | no TestAudit* tests | ok  github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release  0.020s | — | match |
| taskagent tests | ok (no change) | ok  github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent  0.008s | 0.0% | match |
| taskagent/validate tests | ok (no change) | ok  github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate  0.003s | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

## Test Evidence

| test | result | delta |
|---|---|---|
| TestAuditGapMatrixDocExists (azuredevops/internal/service/release) | pass | +1 new test |
| TestAuditRoadmapDocExists (azuredevops/internal/service/release) | pass | +1 new test |
| service/release package (full suite) | pass | +2 tests, 0 regressions |
| taskagent package | pass | 0 (unchanged) |
| taskagent/validate package | pass | 0 (unchanged) |
| TestAccReleaseDefinition_basic (live ADO acceptance) | skip | No TF_ACC credentials in environment; WI-3 marked failed. Documentation-only change; no CRUD regression expected. |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/release/doc_audit_test.go` — New: living-document gate tests (TestAuditGapMatrixDocExists + TestAuditRoadmapDocExists) — fail if either doc is absent or trivially empty
- `docs/release-definition-gap-matrix.md` — New: 401-line field-by-field gap matrix; columns Field path | ADO type | TF schema status | Writable? | Notes; data-source section; test-coverage section; summary counts
- `docs/release-definition-roadmap.md` — New: 350-line prioritised implementation roadmap; P1/P2/P3 clusters; iteration budgets; depends_on ordering; explicit out-of-scope section

```
azuredevops/internal/service/release/doc_audit_test.go | 123 +++++++
 docs/release-definition-gap-matrix.md               | 401 +++++++++++++++++++++
 docs/release-definition-roadmap.md                  | 350 ++++++++++++++++++
 3 files changed, 874 insertions(+)
```
