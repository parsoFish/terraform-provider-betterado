# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2025-06, initial)

- Read WI-2.md: requires `TestAccReleaseDefinition_withContainerImageTrigger` + `hclReleaseDefinitionWithContainerImageTrigger` in `resource_release_definition_test.go`.
- Checked existing patterns: `captureReleaseEvidence(tfNode)` already in the file (line ~508); it calls `testutils.CaptureLiveEvidence("acceptance-resource", vsrmURL, def)` — this is exactly what AC2 requires (the WI spec pseudocode names it `captureAndWriteLiveEvidence` but the existing function is the right one).
- Used `SharedReleaseFixture(t)` (same as other fixture-based tests like `TestAccReleaseDefinition_basic`).
- HCL uses DockerHub artifact (`type = "DockerHub"`) per WI spec guidance, `runOnServer` agentless stage (no queue needed).
- Appended test + HCL at end of file (lines 2554–2683).
- All offline gates pass: `go build -tags all ./...`, `go vet -tags all ./acceptancetests/...`, `golangci-lint run ./acceptancetests/...`, `terrafmt diff` (no changes), unit tests pass.
- Committed: 8405e536.

## What worked

- `captureReleaseEvidence(tfNode)` is the existing evidence capture function (not `captureAndWriteLiveEvidence`).
- The WI spec pseudocode function names are illustrative only — use what's in the file.
- `SharedFixtureResult.ProjectID` is the only field needed for this HCL (no BuildDefID, no WorkItemQueryID).
- DockerHub artifact + `runOnServer` agentless stage avoids any queue dependency.

## What didn't work

_(none — completed in one iteration)_

## Open questions

- Will ADO accept `containerImageTrigger` paired with `DockerHub` artifact type in the live API? The WI spec says yes; if not, switch to `AzureContainerRepository` with a real ACR endpoint. This requires live TF_ACC run to verify.

## Notes for reflection

- WI-2 was straightforward once the existing `captureReleaseEvidence` pattern was identified. The key insight: the file already had all helper functions; only the test function + HCL fixture needed adding.
