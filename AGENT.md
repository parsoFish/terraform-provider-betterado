# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (2026-07-04)

- Checked `.forge/last-gate-failure.md`: the gate was failing because `azuredevops/internal/service/pipelinesapproval/` directory did not exist at all. The gate command `go test -tags all -run TestPipelinesApprovalGapMatrix ./azuredevops/internal/service/pipelinesapproval/` produces `stat ... directory not found` when the directory is absent.
- Read WI-1.md: the WI notes this should be a "hollow gate" (`[no test files]`) but the live gate treats a missing directory as a hard failure (exit 1).
- Read `docs/approvalsandchecks-gap-matrix.md` as the format reference.
- Read `azdosdkmocks/pipelinesapproval_sdk_mock.go` to understand the SDK surface (GetApproval, QueryApprovals, UpdateApprovals).

**Actions taken:**
1. Created `docs/pipelinesapproval-gap-matrix.md` with full API field coverage, resource type table, declarative vs ephemeral section, gap verdict table, and ADO API reference.
2. Created `azuredevops/internal/service/pipelinesapproval/gap_matrix_test.go` with `TestPipelinesApprovalGapMatrix` that reads the doc file and asserts required content strings for both ACs.
3. Ran `go test -tags all -run TestPipelinesApprovalGapMatrix ./azuredevops/internal/service/pipelinesapproval/ -v` → **PASS**.
4. Committed both files.

## What worked

- Creating a minimal `_test.go` file in `pipelinesapproval_test` package (external test package) that does an `os.ReadFile` on the doc and checks `strings.Contains` for required phrases.
- The test reads the doc at path `../../../../docs/pipelinesapproval-gap-matrix.md` relative to the test file location — this resolves correctly when `go test` runs from the package directory.
- Build tag `//go:build all || pipelinesapproval` ensures the test runs with `-tags all` (the gate's flag).

## What didn't work

_(none — first iteration succeeded)_

## Open questions

_(none)_

## Notes for reflection

- The WI spec said the gate would produce `[no test files]` as "hollow gate behaviour", but in practice a missing directory causes exit 1 (`stat ... directory not found`). Future WIs that reference packages not yet created should be aware the gate must either point to an existing package or the test file must be created alongside the docs.
- Gap matrix format follows `docs/approvalsandchecks-gap-matrix.md` closely — good template reference for future gap matrix WIs.
