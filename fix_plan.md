# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the acceptance test HCL for `TestAccReleaseDefinition_basic` omits `retention_policy` and `pre_deploy_approval` blocks WHEN the acceptance test runs against live ADO REST 7.2 THEN the test passes without VS402982 or VS402877 errors (ADO now accepts the apply)
  - Added `retention_policy { days_to_keep=30, releases_to_keep=3, retain_build=true }` to `hclReleaseDefinitionBasic`, `hclReleaseDefinitionWithDeploymentInput`, and `hclReleaseDefinitionWithEnvironmentOptions`
  - Added minimal automated `pre_deploy_approval { approver { id="00000000-...", is_automated=true, rank=1 } }` to those same three HCL functions
  - Schema fields remain Optional (esc-9 compliance)

- [x] AC2: GIVEN the `hclReleaseDefinitionBasic` HCL template includes a minimal `retention_policy` block in its environment WHEN `go test -tags all -count=1 -run TestReleaseDefinition_AccRefresh ./azuredevops/internal/service/release/` is executed THEN new unit tests `TestReleaseDefinition_AccRefresh_*` covering the updated HCL fixture round-trip expand/flatten pass
  - Added `TestReleaseDefinition_AccRefresh_RetentionPolicy` — tests expand/flatten round-trip for retention_policy (days_to_keep, releases_to_keep, retain_build)
  - Added `TestReleaseDefinition_AccRefresh_PreDeployApproval` — tests expand/flatten round-trip for pre_deploy_approval with automated approver
  - Quality gate `go test -tags all -count=1 -run TestReleaseDefinition_AccRefresh ./azuredevops/internal/service/release/` → PASS

- [x] AC3: GIVEN the unit test substrate for the acceptance refresh is green WHEN the full release package unit suite runs THEN `go test -tags all -count=1 -run TestReleaseDefinition ./azuredevops/internal/service/release/` exits 0 with all 11+ existing tests plus new tests passing
  - Full suite: 13 tests total (11 original + 2 new AccRefresh) — all PASS
