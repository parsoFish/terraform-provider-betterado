# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration)

Oriented on prior work. Found that WI-4 was **already completed** by a prior forge-autocommit (`aaac4ad2`) during iteration 1 of the initiative. All acceptance criteria are satisfied:

**AC1 (Framework migration):**
- `resource_agent_queue_framework.go` — full framework resource with CRUD, ImportState, int64 plan modifiers
- `data_agent_queue_framework.go` — full framework data source

**AC2 (SDKv2 deregistration):**
- `resource_agent_queue.go` — DELETED
- `data_agent_queue.go` — DELETED
- `resource_agent_queue_test.go` (SDKv2 unit tests) — DELETED
- `data_agent_queue_test.go` (SDKv2 unit tests) — DELETED
- `framework_provider.go` — has `taskagent.NewAgentQueueResource` (line 207) and `taskagent.NewAgentQueueDataSource` (line 229)
- `provider.go` — has only comments referencing agent_queue, no SDKv2 registrations
- `provider_test.go` — updated; `betterado_agent_queue` absent from both lists; counts correct

**AC3 (CaptureLiveEvidence):**
- `data_agent_queue_test.go` (acceptancetests) has `captureAgentQueueEvidence()` calling `CaptureLiveEvidence("acceptance-resource-agent-queue", url, queue)`

**Quality gate (offline):**
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```
All pass. Pre-existing failures in `build`, `graph`, `identity`, `serviceendpoint` are unrelated to WI-4.

**Acceptance test:**
```
go test -tags all -run TestAccAgentQueue ./azuredevops/internal/acceptancetests/
```
Compiles clean; runs 0 tests without TF_ACC (expected).

## What worked

- Framework migration pattern (same as other taskagent resources in this initiative)
- Custom int64 plan modifiers with `Queue` suffix to avoid naming conflicts with other modifiers in the package
- `captureAgentQueueEvidence` helper function pattern (same as `captureTaskGroupEvidence` in resource_task_group_test.go)
- `getDirectClient()` is in `resource_task_group_test.go` with `all || resource_task_group` build tag; works when running with `-tags all`

## What didn't work

_(nothing to record — all ACs were satisfied on prior iteration)_

## Open questions

_(none)_

## Notes for reflection

- WI-4 was completed by forge-autocommit `aaac4ad2`; the WI frontmatter `status: complete` is correct
- The `splitProjectQualifiedID` helper in `resource_agent_queue_framework.go` avoids importing `strings` package
- Custom `requiresReplaceInt64Modifier` and `useStateForUnknownInt64QueueModifier` use `Queue` suffix to avoid conflict with other plan modifiers in the same package
