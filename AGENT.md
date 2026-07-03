# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2025-07-03)

Created all four required files from scratch:

1. `data_pipeline_framework.go` — `PipelineDataSource` implementing `datasource.DataSource`
   - Schema: `project_id` (required), `pipeline_id` (required Int64), and computed: `id`, `name`, `folder`, `configuration_type`, `revision`, `url`
   - `Read()` calls `PipelinesClient.GetPipeline()`, on 404 sets `id=""` and returns
   - `flattenPipelineDataModel()` maps `*adoPipelines.Pipeline` → `pipelineDataModel`

2. `data_pipeline_framework_test.go` — `//go:build all || data_source_pipeline`
   - `TestPipelineDataSource_Metadata` — asserts TypeName == "betterado_pipeline"
   - `TestPipelineDataSource_Schema` — asserts required attrs (`pipeline_id`, `project_id`) and computed attrs (`id`, `name`, `folder`, `configuration_type`, `revision`, `url`) exist

3. `data_pipeline_run_framework.go` — `PipelineRunDataSource` implementing `datasource.DataSource`
   - Schema: `project_id` (required), `pipeline_id` (required Int64), `run_id` (required Int64), and computed: `id`, `name`, `state`, `result`, `created_date`, `finished_date`
   - `Read()` calls `PipelinesClient.GetRun()`, on error returns diagnostic (not 404-safe per spec)
   - `flattenPipelineRunDataModel()` maps `*adoPipelines.Run` → `pipelineRunDataModel`
   - `azuredevops.Time` wraps `time.Time` via `.Time` field; format with `time.RFC3339`

4. `data_pipeline_run_framework_test.go` — `//go:build all || data_source_pipeline`
   - `TestPipelineRunDataSource_Metadata` — asserts TypeName == "betterado_pipeline_run"
   - `TestPipelineRunDataSource_Schema` — asserts required attrs (`pipeline_id`, `run_id`, `project_id`) and computed attrs (`id`, `name`, `state`, `result`, `created_date`, `finished_date`) exist

## What worked

- Following the `datasource_release_definition_framework.go` pattern exactly for the data source structure
- Using `//go:build all || data_source_pipeline` build tag (both test files)
- `PipelinesClient.GetPipeline()` and `PipelinesClient.GetRun()` are on `client.AggregatedClient.PipelinesClient`
- SDK `azuredevops.Time` has `.Time time.Time` field; access `r.CreatedDate.Time.UTC().Format(time.RFC3339)`
- SDK module is at `third_party/azure-devops-go-api/azuredevops/v7/` in the worktree
- Run struct: `Id`, `Name`, `State (*RunState)`, `Result (*RunResult)`, `CreatedDate (*azuredevops.Time)`, `FinishedDate (*azuredevops.Time)`, `Pipeline (*PipelineReference)` with `Pipeline.Id`

## What didn't work

_(none — first attempt succeeded)_

## Test gate command

```
go test -tags all -run TestPipelineDataSource|TestPipelineRunDataSource ./azuredevops/internal/service/pipelines/
```

All 4 tests pass as of iteration 0:
- `TestPipelineDataSource_Metadata` PASS
- `TestPipelineDataSource_Schema` PASS
- `TestPipelineRunDataSource_Metadata` PASS
- `TestPipelineRunDataSource_Schema` PASS

## Open questions

_(none)_

## Notes for reflection

_(none — straightforward implementation following existing patterns)_
