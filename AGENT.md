# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete — all ACs satisfied)

- Read the WI spec and all relevant existing tests to understand the patterns.
- Added `TestAccReleaseDefinition_completeWithNewFields` + `hclReleaseDefinitionCompleteWithNewFields` to `azuredevops/internal/acceptancetests/resource_release_definition_test.go`.
- Test uses `SharedReleaseFixture` (project + build definition pre-provisioned), `runOnServer` phase (no agent queue needed), idempotency step (`PlanOnly: true, ExpectNonEmptyPlan: false`).
- Live gate ran (`TF_ACC=1` was already set) and **passed in ~22 seconds**.
- gofmt and terrafmt-check both pass.
- Committed as `test: add TestAccReleaseDefinition_completeWithNewFields (WI-3)`.

## What worked

- **Appending via `cat >>`** to the file (Edit tool couldn't uniquely match the last line since 4 HCL functions ended with the same pattern).
- **SharedReleaseFixture** provides `ProjectID` + `BuildDefinitionID` — enough to wire the artifact reference. No `AgentQueueID` field; `runOnServer` phase sidesteps the need for one.
- **Combining PR #18 + PR #19 fields** in the same environment block works without conflict. The `condition { name = "ReleaseStarted", condition_type = "event" }` is needed when `environment_trigger` is present alongside `cd_artifact_trigger`.
- Assertions are non-duplicated: the combined test asserts coexistence (e.g. both `source_repo_trigger` and `cd_artifact_trigger` in one `triggers` block), interaction (environment-level PR #18 fields alongside artifact-level PR #19 fields), and idempotency.

## What didn't work

_(nothing failed — first-iteration success)_

## Open questions

- Per the WI spec note: `cd_artifact_trigger.tag_filter` may not round-trip on some ADO instances. The live run succeeded with assertion `triggers.0.cd_artifact_trigger.0.tag_filter.0.tags.0 = "stable"` — so it round-tripped correctly in this environment. If a future run fails, consider `ImportStateVerifyIgnore` per the WI note.

## Notes for reflection

- The WI was well-specified: following the spec HCL shape + assertions exactly produced a first-iteration live pass.
- Pattern: `cat >>` is the safe append approach when the file's last lines are duplicated across functions.
