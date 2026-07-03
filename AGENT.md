# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 2 (this iteration — complete)

**Problem:** `.forge/last-gate-failure.md` showed `[no tests to run]` because `TestPipelineApprovalResource` tests didn't exist yet. Prior iterations had only wired `PipelinesApprovalClient` into `AggregatedClient` and created `gap_matrix_test.go` / docs.

**Action:** Created both mandatory files:
- `azuredevops/internal/service/pipelinesapproval/resource_pipeline_approval_framework.go`
- `azuredevops/internal/service/pipelinesapproval/resource_pipeline_approval_framework_test.go`

**Result:** All 4 ACs satisfied. Gate command passes with 2 tests running:
- `TestPipelineApprovalResource_Metadata` — PASS
- `TestPipelineApprovalResource_Schema` — PASS

## What worked

- Build tag `//go:build all || resource_pipeline_approval` on BOTH files (resource + test) is required — the resource file must include the build tag or it won't compile with `all`.
- Using `stringplanmodifier.RequiresReplace()` and `stringplanmodifier.UseStateForUnknown()` from the framework SDK package (not custom implementations) — simpler and more idiomatic than the custom modifiers in `resource_pipeline_framework.go`.
- The `_test` package suffix (`package pipelinesapproval_test`) requires importing the package being tested explicitly.
- `UpdateApprovals` takes `*[]ApprovalUpdateParameters` (pointer to slice), not a slice directly.
- `GetApproval` takes `*uuid.UUID` for the approval ID.
- The `Approval` struct's comment is in `Steps[*].Comment`, not at the top level.
- Delete is truly a no-op — the method body is empty, no state manipulation needed (framework handles removal automatically).

## What didn't work

_(nothing significant; iteration was straightforward once the scope was understood)_

## Open questions

- The `flattenApproval` picks the first non-empty step comment. If multiple approvers comment, only the first is tracked. This is acceptable per the WI spec which only tracks `status` and `comment`.

## Notes for reflection

- The gate `[no tests to run]` error is a clear signal that the test file simply didn't exist yet.
- Prior iterations focused on client wiring and docs but never created the mandatory `creates:` paths — those should always be created (even as stubs) in the first iteration.
