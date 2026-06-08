# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (2026-06-08)

**STATUS: ALL ACs COMPLETE — tests passing, committed.**

1. Read the full WI spec and the existing source/test files cold (no prior commits).
2. Added three new Optional fields to the `cd_artifact_trigger` Elem.Schema in
   `resource_release_definition.go`:
   - `tag_filter` — TypeList/MaxItems:1 with `pattern` (TypeString, default "") and
     `tags` (TypeList of TypeString) sub-fields
   - `use_build_definition_branch` — TypeBool, default false
   - `create_release_on_build_tagging` — TypeBool, default false
3. Introduced `expandArtifactTriggerConditions(ctMap)` which:
   - Calls existing `expandBranchFiltersForTrigger` for branch_filter entries
   - Adds `tagFilter` map to conditions[0] when tag_filter block is present
   - Adds `useBuildDefinitionBranch: true` to conditions[0] when set
   - Adds `createReleaseOnBuildTagging: true` to conditions[0] when set
   - Creates a synthetic `{sourceBranch: ""}` entry if no branch_filter but extras exist
4. Updated `expandTriggers` to call `expandArtifactTriggerConditions` instead of the
   old branch-filter-only path.
5. Introduced `flattenArtifactTriggerConditions(trigMap)` which reads all four return
   values (branch_filter []interface{}, tagFilter map, useBuildBranch bool, createOnTagging bool)
   from triggerConditions[0].
6. Updated `flattenTriggers` to call `flattenArtifactTriggerConditions` and populate
   the new fields in the state ct map.
7. Added two test functions to `resource_release_definition_test.go`:
   - `TestReleaseDefinition_ArtifactTagFilter_RoundTrip` — exercises AC1+AC2
   - `TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip` — exercises AC3+AC4+AC5+AC6
8. All tests pass: `go test -tags all ./azuredevops/internal/service/release/ → ok`
9. Committed as `feat(release): add tag_filter, use_build_definition_branch, create_release_on_build_tagging to cd_artifact_trigger`

## What worked

- Python script for file edits avoids tab/space issues with the Edit tool (the file uses
  tabs; the Edit tool requires exact match including whitespace).
- The pattern of "extras go on conditions[0], create synthetic entry if no branches" matches
  the WI spec's guidance and works correctly in round-trip tests.
- `schema.TestResourceDataRaw` for unit tests with the full resource schema. Must supply
  all required fields (environment, deploy_phase) even for trigger-focused tests.

## What didn't work

- The Edit tool failed on the initial attempts because the Go source uses hard tabs for
  indentation but the tool requires exact string matching including whitespace. Used a
  Python script to do the replacement reliably.

## Open questions

_(none remaining — all ACs complete)_

## Notes for reflection

_(nothing to flag)_
