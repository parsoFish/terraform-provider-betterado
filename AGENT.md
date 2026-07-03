# Agent Memory — WI-7

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (final)

- Gate failure: `go test -tags all -run TestChangelog_HasPipelineApprovalEntry ./azuredevops/internal/service/pipelinesapproval/` returned `[no tests to run]` because the test file didn't exist.
- Created `azuredevops/internal/service/pipelinesapproval/changelog_test.go` with `TestChangelog_HasPipelineApprovalEntry` that reads `../../../../CHANGELOG.md`, locates `## [Unreleased]`, and asserts `betterado_pipeline_approval`, `betterado_pipeline_approvals`, and `### FEATURES` are present.
- Added CHANGELOG.md entry under `## [Unreleased]` → `### FEATURES` (inserted at top of that section before existing entries) describing the resource and data source.
- Bumped `PROVIDER_VERSION.txt` from `1.14.0` to `1.14.1`.
- Test runs and passes: `PASS ok ... 0.003s`.

## What worked

- Build tag `//go:build all || pipelinesapproval` with package `pipelinesapproval_test` matches the pattern in the existing `gap_matrix_test.go` in the same directory.
- Path `../../../../CHANGELOG.md` is correct from the package directory depth.
- Inserting CHANGELOG entry at the TOP of the `### FEATURES` block under `## [Unreleased]` ensures the test finds it within the Unreleased section.
- Bounding the unreleased search by `\n## [` correctly scopes the assertion to only the Unreleased content.

## What didn't work

_(nothing to record — this was a clean run)_

## Open questions

_(none)_

## Notes for reflection

_(none)_
