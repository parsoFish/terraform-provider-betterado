## Why

The `betterado_release_definition` resource's trigger configuration only covered the basic build-completion trigger. Three ADO 7.2 trigger capabilities were missing from the schema:

1. **Artifact tag filters** — operators needed to scope a CD trigger to specific build tags (e.g. `stable`, `release`) rather than firing on every successful build. Without this, every build in the pipeline kicked off a release, creating unwanted noise.
2. **Source branch and build-tagging flags** — `use_build_definition_branch` (inherit the triggering build's branch) and `create_release_on_build_tagging` (fire when the build is tagged, not completed) were silently accepted by Terraform but never written to or read from the ADO API, so any plan referencing them diverged on refresh.
3. **Source-repo trigger** — the `sourceRepo` trigger type (fire when a commit lands on a target branch) had no schema representation at all, preventing branch-driven release workflows from being declared in HCL.

All three gaps meant customers couldn't express their real ADO trigger configurations in Terraform, and any import of an existing release definition with these fields set would immediately produce a non-empty plan on re-apply.

## What

Changes in this PR (`3 files changed, 603 insertions(+), 194 deletions(-)`):

- **`azuredevops/internal/service/release/resource_release_definition.go`** — extends `cd_artifact_trigger` with `tag_filter` (nested block: `pattern` string + `tags` string list), `use_build_definition_branch` bool, and `create_release_on_build_tagging` bool. Adds a new `source_repo_trigger` block (`alias` + `branch_filters` string list) to the `triggers` block. Expand and flatten functions handle all new fields; empty `branch_filters` and nil tag slices are suppressed to prevent residual diffs.
- **`azuredevops/internal/service/release/resource_release_definition_test.go`** — adds `TestReleaseDefinition_ArtifactTagFilter_RoundTrip`, `TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip`, and `TestReleaseDefinition_SourceRepoTrigger_RoundTrip`. Existing tests refactored to use shared builders (net +252 lines of test logic, significant structural cleanup).
- **`azuredevops/internal/acceptancetests/resource_release_definition_test.go`** — adds `TestAccReleaseDefinition_triggerEnhancements`, which uses `SharedReleaseFixture` to stand up a release definition with all four new trigger fields against a live ADO org, asserts every field survives the ADO read-back, confirms idempotency (step 2: `PlanOnly: true, ExpectNonEmptyPlan: false`), and destroys cleanly.

## How

**Schema layer:** `tag_filter` is a `TypeList` block (max 1) nested inside `cd_artifact_trigger`. It carries `pattern` (optional string) and `tags` (optional `TypeList` of strings). The ADO SDK model maps `Tags` to `ArtifactFilter.Tags *[]string` and `Pattern` to `ArtifactFilter.TagFilter.Pattern`. The two boolean flags map directly to `ArtifactFilter.UseBuildDefinitionBranch` and `ArtifactFilter.CreateReleaseOnBuildTagging`. `source_repo_trigger` maps to `SourceRepoTrigger` in the ADO SDK with `TriggerType: release.ReleaseDefinitionTriggerType("sourceRepo")`, `Alias`, and `BranchFilters`.

**Flatten correctness:** ADO 7.1 does not persist the SDK `tagFilter` regex field — only the `tags` list survives a GET. The flatten reads from `Tags` (not `TagFilter.Pattern`) to prevent a perpetual diff on the `pattern` attribute. A `tag_filter` block is only emitted when the tags list is non-empty; an empty list is suppressed rather than written as `tags = []`, which would differ from the absent block that ADO returns.

**Idempotency fixes:** The initiative's second commit (`fix(release): tag_filter round-trips via condition tags`) addressed two residual-diff sources found during acceptance testing: (a) `branch_filters` on `cd_artifact_trigger` was written as `[]` when ADO returned null, fixed by nil-checking before flatten; (b) `variable_groups` on the definition was normalised to handle ADO returning `null` vs the config's `[]`.

**Testing approach:** Unit round-trip tests construct a `schema.ResourceData` with the new fields set, call `expandTriggers`, call `flattenTriggers`, and assert the output matches the input — no ADO credentials required, runs in `< 50 ms`. The acceptance test uses `SharedReleaseFixture` (shared project + build definition, created once per test run) and a `runOnServer` deploy phase to avoid needing a real agent queue, making it compatible with the ADO free tier.
