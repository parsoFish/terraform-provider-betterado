# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-9): GIVEN the schedule_trigger schema currently exposes a branch_filter block WHEN a schedule_trigger is configured with a branch_filter and applied THEN it must NOT perpetually diff — ADO classic schedule triggers are time-based and have no branch filter, so REMOVE branch_filter from the schedule_trigger schema (and its expand/flatten). cd_artifact_trigger keeps its branch_filter. Add/extend TestReleaseDefinition_RoundTrip to assert a schedule_trigger round-trips with no residual diff.
  - PROVEN: branch_filter removed from schedule_trigger schema in resource_release_definition.go; TestReleaseDefinition_Triggers_ScheduleOnly and _Triggers_ExpandFlatten updated to assert no branch_filter in flattened state; TestReleaseDefinition_RoundTrip subtest renamed to schedule_trigger_no_branch_filter_no_residual_diff; all tests pass.
- [x] AC2 (WI-9): GIVEN a deploy phase deployment_input WHEN the exhaustive acceptance test (TestAccReleaseDefinition_complete) configures it THEN it sets agent_specification to a real non-default value (e.g. "ubuntu-22.04") and the live read confirms it persisted (agentSpecification.identifier)
  - PROVEN: agent_specification = "ubuntu-22.04" in hclReleaseDefinitionComplete; checkReleaseDefinitionAgentSpecification() added to assert agentSpecification.identifier persisted via live API; committed in ce11e6ba.
- [x] AC3 (WI-9): GIVEN the pre/post deployment gate "Query Work Items" task needs a real queryId WHEN the exhaustive acceptance test builds the gate THEN a real shared work-item query is created in the test project (via the ADO API in the test setup, or a provider resource if one exists) and its id is set as the gate task's queryId input — so the gate is complete (not an empty queryId). If a query genuinely cannot be created in TF/test-setup, use a self-contained ServerGate task instead and note why.
  - PROVEN: betterado_workitemquery resource "All Work Items - Gate Check" added to hclReleaseDefinitionComplete; betterado_workitemquery.gate_query.id referenced as queryId for both pre/post gate tasks; committed in ce11e6ba.
- [ ] AC4 (WI-9): GIVEN the live acceptance gate WHEN TF_ACC=1 go test -run TestAccReleaseDefinition_complete -timeout 30m runs THEN it applies, every option (incl. agent_specification + a real gate queryId) persists, the idempotency check (ExpectNonEmptyPlan:false) passes with NO residual schedule_trigger diff, and it destroys cleanly
  - PENDING: requires live ADO credentials (TF_ACC=1 gate); code changes are in place.

## Unifier sub-phase status

- [x] Quality gate (offline): `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` exits 0 (22 top-level test functions pass)
- [x] demo.json authored and validated (demo/INIT-2026-06-05-complete-release-definition/demo.json)
- [x] DEMO.md + DEMO.html rendered via `forge demo render`
- [x] .forge/pr-description.md updated with WI-9 additions
- [ ] Commit + push (in progress this iteration)
