# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the schedule_trigger schema currently exposes a branch_filter block WHEN a schedule_trigger is configured with a branch_filter and applied THEN it must NOT perpetually diff — ADO classic schedule triggers are time-based and have no branch filter, so REMOVE branch_filter from the schedule_trigger schema (and its expand/flatten). cd_artifact_trigger keeps its branch_filter. Add/extend TestReleaseDefinition_RoundTrip to assert a schedule_trigger round-trips with no residual diff.
  - Schema already correct (branch_filter absent from schedule_trigger)
  - Fixed unit tests: removed stale branchInclude/branch_filter assertions, replaced with AC1 assertions (no branchFilters in expanded trigger, no branch_filter in flattened state)
  - TestReleaseDefinition_Triggers_ScheduleOnly, TestReleaseDefinition_Triggers_ExpandFlatten, TestReleaseDefinition_RoundTrip/schedule_trigger_no_branch_filter_no_residual_diff all PASS
- [x] AC2: GIVEN a deploy phase deployment_input WHEN the exhaustive acceptance test (TestAccReleaseDefinition_complete) configures it THEN it sets agent_specification to a real non-default value (e.g. "ubuntu-22.04") and the live read confirms it persisted (agentSpecification.identifier)
  - Added agent_specification = "ubuntu-22.04" to hclReleaseDefinitionComplete deployment_input
  - Added TestCheckResourceAttr check for agent_specification
  - Added checkReleaseDefinitionAgentSpecification("ubuntu-22.04") API-level check
- [x] AC3: GIVEN the pre/post deployment gate "Query Work Items" task needs a real queryId WHEN the exhaustive acceptance test builds the gate THEN a real shared work-item query is created in the test project (via the ADO API in the test setup, or a provider resource if one exists) and its id is set as the gate task's queryId input — so the gate is complete (not an empty queryId).
  - Added betterado_workitemquery "gate_query" resource under "Shared Queries" in hclReleaseDefinitionComplete
  - Both pre_deployment_gates and post_deployment_gates now use betterado_workitemquery.gate_query.id as queryId
- [ ] AC4: GIVEN the live acceptance gate WHEN TF_ACC=1 go test -run TestAccReleaseDefinition_complete -timeout 30m runs THEN it applies, every option (incl. agent_specification + a real gate queryId) persists, the idempotency check (ExpectNonEmptyPlan:false) passes with NO residual schedule_trigger diff, and it destroys cleanly
  - Requires live ADO environment (TF_ACC=1)
