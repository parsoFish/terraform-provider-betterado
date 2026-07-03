# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 5 (final)

**Problem:** Gate failure was a build error in `framework_provider_test.go`:
- `undefined: datasource` at lines 153, 161, 162, 197, 208, 209 — the `datasource` import was missing from the test file.
- `TestFrameworkProvider_HasPipelineApprovalResources` test didn't exist yet.

**Root cause of prior iterations' failure:** The pipelinesapproval production source files (`resource_pipeline_approval_framework.go`, `data_pipeline_approvals_framework.go`) had build constraints (`//go:build all || resource_pipeline_approval` etc.) that prevented them from being included in a plain `go build`. Only test files should carry build tags in this project.

**Fixes applied:**
1. Added `"github.com/hashicorp/terraform-plugin-framework/datasource"` import to `framework_provider_test.go`.
2. Added `TestFrameworkProvider_HasPipelineApprovalResources` test checking both `betterado_pipeline_approval` resource and `betterado_pipeline_approvals` data source.
3. Removed build constraints from `resource_pipeline_approval_framework.go` and `data_pipeline_approvals_framework.go` (production files must have no build tags).
4. Added `pipelinesapproval` import to `framework_provider.go`.
5. Added `pipelinesapproval.NewPipelineApprovalResource` to `Resources()` slice.
6. Added `pipelinesapproval.NewPipelineApprovalsDataSource` to `DataSources()` slice.

**Verified:**
- `go build -mod=vendor .` → exits 0 ✓
- `go test -tags all -run TestFrameworkProvider_HasPipelineApprovalResources ./azuredevops/internal/provider/` → PASS ✓
- `grep pipeline_approval azuredevops/provider.go` → no output ✓

## What worked

- Removing build constraints from production source files (only test files carry build tags in this repo's convention)
- Adding `datasource` import to the test file
- Wiring constructors into framework_provider.go using the established pattern

## What didn't work

- Prior iterations added build tags to production files — this broke `go build` without `-tags all`

## Open questions

_(none — all ACs complete)_
