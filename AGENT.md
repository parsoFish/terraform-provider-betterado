# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration)

**Goal:** Convert all `environment { }` block HCL to `stages = [{ }]` array syntax in
`resource_release_definition_test.go`; add `TestAccReleaseDefinition_stagesArraySyntax`.

**What was done:**
1. Rewrote `azuredevops/internal/acceptancetests/resource_release_definition_test.go` from
   scratch — all HCL templates now use `stages = [{ ... }]` array syntax; every
   `TestCheckResourceAttr` path updated from `environment.*` to `stages.0.*` / `stages.1.*`.
2. Added `TestAccReleaseDefinition_stagesArraySyntax` (WI-3/AC1) and its HCL template
   `hclReleaseDefinitionStagesArraySyntax` using two stages, non-default retention_policy,
   and a deployment_input with `queue_id = fixture.AgentQueueID`.
3. Added `AgentQueueID int` to `SharedFixtureResult` in `shared_fixtures.go` (resolved via
   `resolveAgentQueueID` which was already in the file). Also populated it in `SharedReleaseFixture`.
4. Added `ConfigMode: schema.SchemaConfigModeAttr` to ALL 19 TypeList/TypeSet fields with
   `Elem: *schema.Resource` nested inside the `stages` block (and its sub-schemas
   `approvalSchema`, `deploymentGatesSchema`). This was REQUIRED by the Terraform Plugin SDK
   `InternalValidate` — when a parent TypeList uses `ConfigMode:attr`, all children that are
   TypeList/TypeSet with `Elem:*schema.Resource` must also use `ConfigMode:attr`.

**Verification:**
- `go build ./...` → clean
- `go vet ./...` → clean
- `go test -count=1 -run TestAccGroupDataSource_ReadersResolvesWithProjectID ./azuredevops/internal/acceptancetests/` → PASS (no InternalValidate error)
- `go test ./azuredevops/internal/service/release/` → PASS

**Committed as:**
`test: convert all release-definition acc tests to stages array syntax (WI-3)`

## What worked

- **SDK ConfigMode rule**: When `stages` (TypeList) has `ConfigMode: SchemaConfigModeAttr`,
  every child TypeList/TypeSet with `Elem: *schema.Resource` at ANY nesting depth also
  needs `ConfigMode: SchemaConfigModeAttr`. This is enforced by InternalValidate at
  provider startup (even without TF_ACC). The error message names the offending field:
  `resource betterado_release_definition: environment_options: in *schema.Resource with
  ConfigMode of attribute, so must also have ConfigMode of attribute`.

- **Python file-replace approach**: For bulk tab-indented Go edits, using a Python script
  to replace exact byte patterns (using `\t` in strings) is more reliable than the Edit
  tool which sometimes mismatches indentation.

- **`SharedFixtureResult.AgentQueueID`**: Adding the field to the fixture result lets tests
  use a real queue_id without needing a `data "betterado_agent_queue"` block (saves an
  extra API call and resource dependency in the HCL).

## What didn't work

- **Edit tool with tab characters**: The Edit tool's old_string matching was inconsistent
  with tab-indented Go files. Used a Python subprocess-based approach instead.

## Open questions

- AC3 (ado-demo skill verification) is a live ADO run that the orchestrator handles. No
  action needed from Ralph — the commit is ready for the acceptance gate.

## Iteration 1 (Ralph loop, iteration 0 of WI-3 task invocation)

**Goal:** Orient, verify all gates still pass, confirm no rework needed.

**What was done:**
1. Re-ran all code gates: `go build ./...`, `go vet ./...`, `gofmtcheck.sh`, unit tests — all clean.
2. Confirmed zero remaining `environment { }` HCL blocks or `"environment.*"` attribute paths.
3. Confirmed `TestAccReleaseDefinition_stagesArraySyntax` is present with correct shape.
4. `make test` (which runs `go test -v ./...`) was too slow/blocked — skipped in favor of targeted gate runs above.
5. Updated `fix_plan.md` with verification summary.

**Status:** No code changes needed. AC1 + AC2 complete. AC3 awaits orchestrator TF_ACC gate.

## Notes for reflection

- The SDK InternalValidate rule `ConfigMode parent → all children need ConfigMode` is a
  footgun that has now bitten this codebase. Consider documenting this in a schema
  convention note (brain/themes or a comment in the resource file).
- The `resolveAgentQueueID` function was already in `shared_fixtures.go` but wasn't exposed
  in `SharedFixtureResult`. Worth noting as a pattern: if a helper function exists, expose
  its result in the fixture struct so tests can use it without data sources.
