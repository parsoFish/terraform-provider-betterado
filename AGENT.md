# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0
- Created `azuredevops/internal/acceptancetests/resource_task_group_test.go` with `TestAccTaskGroup_basic`.
- Built on the environment acceptance test pattern (`resource_environment_test.go`) for `checkTaskGroupDestroyed` using `clients.TaskAgentClient.GetTaskGroups`.
- Used `testutils.GetProvider().Meta().(*client.AggregatedClient)` to get client in CheckDestroy.
- CmdLine@2 task_id `d9bafed4-0b18-4f58-968d-86655b4d2ce9` as the stable built-in ADO task.
- Build tag: `//go:build (all || resource_task_group) && !exclude_resource_task_group`
- Commit: `a01f0915`

## What worked

- Standard acceptance test structure mirrored from `resource_environment_test.go`.
- `testutils.GenerateResourceName()` → `test-acc-<random>` satisfies unique/UUID-prefixed name requirement.
- `go test -tags all -list TestAccTaskGroup_basic` → found, compiles clean.
- `gofmt`, `go vet`, `golangci-lint` all clean.
- Inline HCL uses `betterado_project.test` (not `.project`) — consistent with env test pattern.

## What didn't work

_(nothing — clean first iteration)_

## Open questions

- Live execution awaits orchestrator running gate with `TF_ACC=1` + credentials.

## Notes for reflection

- `checkTaskGroupDestroyed` pattern mirrors `checkEnvironmentDestroyed` — good template for future resource tests.
- `GetTaskGroups` returns `*[]TaskGroup`; `len(*taskGroups) > 0` confirms existence after non-404 response.
- `utils.ResponseWasNotFound(err)` is the idiomatic 404-check (`azuredevops/internal/utils/HttpResponse.go`).
