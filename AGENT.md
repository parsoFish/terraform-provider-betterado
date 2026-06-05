# Agent Memory — WI-9

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (this iteration)

**AC1 — schedule_trigger branch_filter removal:**
- The schema (`resource_release_definition.go` ~line 527-568) was already correct: `schedule_trigger` has NO `branch_filter` field. cd_artifact_trigger keeps its branch_filter.
- The expand (`expandTriggers`) and flatten (`flattenTriggers`) logic was already correct: no branchFilters in schedule triggers.
- The unit tests however still had STALE code:
  - `TestReleaseDefinition_Triggers_ScheduleOnly`: still asserted `branchFilters` was in the expanded trigger AND `branch_filter` was in the flattened state — both wrong, and referencing undefined `branchInclude` variable (compile error)
  - `TestReleaseDefinition_Triggers_ExpandFlatten`: still had `branch_filter` in the `schedule_trigger` input data AND asserted `branch_filter` survived flatten — both wrong
- **Fix**: Removed stale branch_filter assertions, replaced with correct AC1 assertions (absent branch_filter). Used Python for tab-based replacements since the Edit tool struggled with tab indentation.
- All three relevant tests now pass: ScheduleOnly, ExpandFlatten, RoundTrip/schedule_trigger_no_branch_filter

**AC2 — agent_specification in acceptance test:**
- `agent_specification` schema field already exists (`resource_release_definition.go` ~line 283)
- expand: already sets `agentSpecification: { identifier: spec }` when non-empty
- flatten: already reads `agentSpecification.identifier` → `agent_specification`
- Just needed to ADD `agent_specification = "ubuntu-22.04"` to the HCL template and checks
- Added `checkReleaseDefinitionAgentSpecification("ubuntu-22.04")` API-level check

**AC3 — real shared query for gate queryId:**
- `betterado_workitemquery` resource EXISTS in the provider (`azuredevops/internal/service/workitemtracking/resource_workitemquery.go`)
- Supports `area = "Shared Queries"` to create a shared query
- Added `betterado_workitemquery "gate_query"` resource to `hclReleaseDefinitionComplete`
- Both pre_deployment_gates and post_deployment_gates now reference `betterado_workitemquery.gate_query.id` as queryId

**AC4 — live acceptance test:**
- Not testable without TF_ACC=1 environment
- All code changes are in place; needs live ADO run

## What worked

- Python string replacement for tab-indented Go test files (Edit tool fails to match tab patterns)
- `betterado_workitemquery` with `area = "Shared Queries"` is the right TF resource for AC3
- The schema/expand/flatten for agent_specification was already correct; just the HCL template was missing the field

## What didn't work

- Edit tool with tabs in the test file — the `old_string` patterns didn't match even when visually correct; had to use Python replacement

## Open questions

_(nothing blocking)_

## Notes for reflection

- Prior iterations left stale test code that still asserted `branch_filter` in `schedule_trigger` after the schema was fixed — a pattern to watch for in future WIs
- The `betterado_workitemquery` resource for creating shared queries is a useful test pattern when gate tasks need ADO-side state
