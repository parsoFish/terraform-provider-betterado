# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete — all ACs done)

1. Read WI-1.md, acceptance test file, unit test file, and resource_go source to understand structure.
2. Updated three HCL fixture functions in `azuredevops/internal/acceptancetests/resource_release_definition_test.go`:
   - `hclReleaseDefinitionBasic` — added `retention_policy` + `pre_deploy_approval` blocks
   - `hclReleaseDefinitionWithDeploymentInput` — same additions
   - `hclReleaseDefinitionWithEnvironmentOptions` — same additions
3. Added two new unit tests to `azuredevops/internal/service/release/resource_release_definition_test.go`:
   - `TestReleaseDefinition_AccRefresh_RetentionPolicy` — expand/flatten round-trip for retention_policy
   - `TestReleaseDefinition_AccRefresh_PreDeployApproval` — expand/flatten round-trip for minimal automated pre_deploy_approval
4. Ran quality gate: `go test -tags all -count=1 -run TestReleaseDefinition_AccRefresh ./azuredevops/internal/service/release/` → PASS
5. Ran full suite: `go test -tags all -count=1 -run TestReleaseDefinition ./azuredevops/internal/service/release/` → 13 PASS
6. Committed as: `test: fix VS402982/VS402877 – add retention_policy + pre_deploy_approval to acc-test HCL and add AccRefresh unit tests`

## What worked

- The expand/flatten functions for `retention_policy` and `pre_deploy_approval` already existed and were correct. Only test fixtures and unit tests needed to be added.
- Using `schema.TestResourceDataRaw` with the exact same map shape as the existing 11 tests. Key: provide `deployment_input: []interface{}{}` (empty) rather than omitting the key when there's no deployment input — avoids nil panics in flatten.
- The all-zeros UUID `"00000000-0000-0000-0000-000000000000"` is a valid UUID per Go's `uuid.Parse` so expandApprovals works correctly.

## What didn't work

_(none — first iteration was successful)_

## Open questions

_(none)_

## Notes for reflection

- ADO REST 7.2 now mandates `retention_policy` (VS402982) and pre/post approvals (VS402877) on environments. The fix is test-HCL-only; schema fields remain Optional per esc-9.
- The `hclReleaseDefinitionWithApprovalOptions` and `hclReleaseDefinitionComplete` functions already had the required blocks (approval options test had pre_deploy_approval, complete had retention_policy + pre_deploy_approval on Production env). Only the three basic/no-approval templates needed updating.
