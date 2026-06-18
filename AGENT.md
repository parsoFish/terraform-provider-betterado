# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (2026-06-17)

- Read WI-5 spec, existing test file, SharedReleaseFixture definition.
- Confirmed `TestAccReleaseDefinition_withContainerImageTrigger` did not exist yet.
- Confirmed `container_image_trigger` schema was added by WI-4 (commit `7de080fc`).
- Confirmed `SharedFixtureResult` has `BuildDefinitionID` (int) but NO `BuildDefinitionAlias` field — used hardcoded `"_build"` as WI spec instructs when no field exists.
- Confirmed `captureReleaseEvidence` was already implemented; `_blockSyntax` test already called it; added it to `_basic` and to new `_withContainerImageTrigger`.
- Added `TestAccReleaseDefinition_withContainerImageTrigger` + `hclReleaseDefinitionWithContainerImageTrigger` at end of test file.
- Updated gap matrix: marked all 8 previously-writable gaps as 'mapped'; updated section summaries and overall total (103 mapped / 1 partial / 30 missing).
- `go build ./...`, `go vet`, `gofmt -l` all clean.
- Committed as `6cedfd8d`.

## What worked

- `SharedFixtureResult.BuildDefinitionID` (int) works as `%[3]d` in fmt.Sprintf for the HCL tostring() call.
- The Edit tool fails with tab-vs-space mismatches; use Python for tricky replacements in Go files.
- Running `go build ./... && go vet ./... && gofmt -l` is the right CI-equivalent offline check.
- **ALWAYS run `terrafmt-check` (`./scripts/terrafmt.sh`)** — it's a separate gate from gofmt that checks HCL alignment inside `_test.go` files. Use `terrafmt fmt -f <file>` to auto-fix. Run over all test files: `find azuredevops -name "_test.go" | sort | while read f; do terrafmt fmt -f "$f"; done`.

## What didn't work

- Edit tool string matching failed (space vs tab) — had to use Python replace instead.

### Iteration 2 (2026-06-17)

- Discovered terrafmt check was failing on `hclReleaseDefinitionBlockSyntax` (around line 658) — HCL attribute alignment was off.
- Fixed by running `terrafmt fmt -f ./azuredevops/internal/acceptancetests/resource_release_definition_test.go`.
- Also ran terrafmt on ALL test files in the tree — only the one file needed fixing.
- All offline gates now pass: gofmt ✓, go build ✓, go vet ✓, terrafmt-check ✓, acceptance package compiles ✓.
- Committed as `3dd9975b`.

### Iteration 3 (2026-06-17)

- Identified gap: `TestAccReleaseDefinition_basic` was missing an explicit idempotency step (AC1 says "idempotency re-plan produces no diff" but test only had apply + import steps).
- Added idempotency step to `_basic`: `PlanOnly:true, ExpectNonEmptyPlan:false` (after the import step).
- Made `ExpectNonEmptyPlan:false` explicit in `_complete`'s idempotency step (AC3 says "ExpectNonEmptyPlan: false" explicitly).
- All offline gates pass: gofmt ✓, go build ✓, go vet ✓, terrafmt-check ✓, acceptance package compiles ✓.
- Committed as `8c591b62`.

### Iteration 4 (2026-06-17)

- Orientation only — all 5 ACs were completed in prior iterations (1–3).
- All offline gates confirmed clean: `go build ./...` ✓, `go vet ./...` ✓, `gofmt -l` (our files only) ✓, `./scripts/terrafmt.sh` ✓, acceptance package compiles ✓.
- Working tree is clean; 6 commits ahead of origin. No new code changes needed.
- Only remaining step: live gate run by orchestrator (TF_ACC=1).

### Iteration 5 (2026-06-17)

- Orientation only — all 5 ACs confirmed complete in prior iterations.
- All offline gates re-verified clean: `go build ./...` ✓, `go vet ./...` ✓, `gofmt -l` ✓, `./scripts/terrafmt.sh` ✓, acceptance package compiles ✓.
- Working tree clean; 6 commits ahead of origin. No new code changes needed.
- Iteration budget exhausted (5/5). Awaiting orchestrator live gate run (TF_ACC=1).

## Open questions

- The quality gate cmd in the WI is `TF_ACC=1 go test -tags all -run TestAccReleaseDefinition_basic|..._withContainerImageTrigger|..._complete`. This runs live against real ADO. The orchestrator will run that gate; we can't run it offline without credentials.
- Does the `container_image_trigger` actually round-trip in ADO (i.e., does the vsrm API accept `triggerType: "containerImage"` and return it back correctly)? The WI assumes yes (WI-4 author said it's tested). Idempotency check in the test will expose any diff.

## Notes for reflection

- `SharedFixtureResult` does not have a `BuildDefinitionAlias` field. Future WIs that need the artifact alias should document using the hardcoded `"_build"` string (which matches how the fixture wires its artifact).
- Gap matrix "8 writable gaps" in the original footer was slightly undercounted — there were actually 10 new fields mapped (environmentTriggers, schedules, properties, tags, createReleaseOnBuildTagging, use_build_definition_branch, source_repo_trigger, timeoutInMinutes, retryCountOnTaskFailure, overrideInputs, containerImageTrigger), but 3 were from the trigger enhancements sub-category. Updated the matrix to reflect all accurately.
