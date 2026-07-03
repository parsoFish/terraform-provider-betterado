# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

Implemented full migration of `betterado_task_group` data source from SDKv2 to terraform-plugin-framework.

**Root cause of gate failure:** The previous test HCL created a new `betterado_project` resource, which hit the ADO organisation's 1000-project limit. The fix was to switch to `SharedFixtureProjectName` (uses `data "betterado_project"` lookup of the standing demo project, same pattern as `resource_task_group_test.go`).

**Files changed:**
- `azuredevops/internal/service/taskagent/data_task_group_framework.go` — NEW: framework `datasource.DataSource` implementation
- `azuredevops/internal/service/taskagent/data_task_group.go` — DELETED (SDKv2)
- `azuredevops/internal/service/taskagent/data_task_group_test.go` — DELETED (SDKv2 unit tests)
- `azuredevops/internal/provider/framework_provider.go` — added `taskagent.NewTaskGroupDataSource` to `DataSources()`
- `azuredevops/provider.go` — removed `"betterado_task_group": taskagent.DataTaskGroup()` from `DataSourcesMap`
- `azuredevops/provider_test.go` — removed `"betterado_task_group"` from expected SDKv2 data source list
- `azuredevops/internal/acceptancetests/data_task_group_test.go` — switched to `SharedFixtureProjectName`, updated evidence label to `acceptance-resource-task-group-datasource`

**Commit:** d1d23a9d

## What worked

- Reusing `flattenTaskGroupFramework` (already defined in `resource_task_group_framework.go`) via a bridge `taskGroupModel` struct — avoids duplication
- `SharedFixtureProjectName` pattern: use `data "betterado_project"` lookup instead of creating a new project — avoids ADO 1000-project limit
- `NewTaskGroupDataSource` registered in `framework_provider.go` — data source gets provider data via `Configure()` reading `*client.AggregatedClient`

## What didn't work

- Original test HCL created a new `betterado_project` — hit "1000 projects" ADO org limit

## Open questions

_(none)_

## Notes for reflection

_(none)_
