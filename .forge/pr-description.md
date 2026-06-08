## Why

Before this initiative there was no systematic record of the gap between the ADO REST 7.2 `ReleaseDefinition` API surface and what the `betterado_release_definition` Terraform resource actually maps. Implementation work has been proceeding piecemeal — each WI discovered gaps at code-time, causing iteration budget overruns and scope surprises. This audit closes that loop: it produces a single source of truth for remaining work so that future implementation initiatives can be spec-driven from day one rather than discovery-driven mid-loop.

## What

Three files are added; no existing code is modified:

- **`docs/release-definition-gap-matrix.md`** (401 lines) — field-by-field gap table comparing every ADO REST 7.2 `ReleaseDefinition`, `ReleaseDefinitionEnvironment`, `Artifact`, `WorkflowTask`, `ApprovalOptions`, `ReleaseDefinitionGatesStep`, `EnvironmentOptions`, `EnvironmentRetentionPolicy`, and related types against the current TF schema. Columns: Field path | ADO type | TF schema status (mapped / missing / partial) | Writable? | Notes. Includes a summary count, an explicit read-only/computed callout, a data-source section (SDK methods vs surfaced data sources, each rated Recommend/Defer/Out-of-scope), and a test-coverage section listing fields exercised by `TestAccReleaseDefinition_*`, known stale tests (`retention_policy`, `pre_deploy_approval`), and recommended new test cases.

- **`docs/release-definition-roadmap.md`** (350 lines) — prioritised, dependency-ordered implementation roadmap derived from the gap matrix. P1 (required-for-create / high-operator-demand), P2 (config-surface parity), P3 (rarely-used / complex nesting). Each cluster carries an estimated iteration budget calibrated against `work-item-completion-by-domain` data, explicit `depends_on` relationships, and a clear out-of-scope section (read-only/computed values; imperative runtime operations: CreateRelease, UpdateApproval, ManualInterventions, Deployments).

- **`azuredevops/internal/service/release/doc_audit_test.go`** (123 lines) — two living-document gate tests (`TestAuditGapMatrixDocExists`, `TestAuditRoadmapDocExists`) that fail fast if either document is absent or trivially empty (< 50 lines / < 20 lines respectively), ensuring the docs stay tracked with the code.

## How

1. **WI-1** read `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/release/models.go` and `azuredevops/internal/service/release/resource_release_definition.go` field-by-field, then authored `docs/release-definition-gap-matrix.md` and the initial `doc_audit_test.go` with `TestAuditGapMatrixDocExists`.

2. **WI-2** read the gap matrix, grouped missing/partial fields into logical implementation clusters, prioritised by ADO required-for-create → operator parity → complexity, and authored `docs/release-definition-roadmap.md`. Appended `TestAuditRoadmapDocExists` to `doc_audit_test.go`.

3. **WI-3** (live acceptance verification) was attempted but could not complete due to absent ADO credentials (`TF_ACC` environment). Because the initiative is documentation-only (no schema or CRUD code changes), there is no provider regression risk. The CI-equivalent unit gate (`go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`) passes green.
