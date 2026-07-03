# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

**Gate failure coming in:** `TestFrameworkProvider_HasPipeline` matched no tests → "no tests to run" rejection.

**State of play from WI-2:**
- `pipelines.NewPipelineResource` was already in `Resources()` in `framework_provider.go`
- `NewPipelineDataSource()` and `NewPipelineRunDataSource()` existed in the pipelines package but were NOT yet registered in `DataSources()`
- No test functions matching `TestFrameworkProvider_HasPipeline*` existed

**Changes made:**
1. Added `pipelines.NewPipelineDataSource` and `pipelines.NewPipelineRunDataSource` to `DataSources()` in `framework_provider.go`
2. Added `TestFrameworkProvider_HasPipelineResource` and `TestFrameworkProvider_HasPipelineDataSource` to `framework_provider_test.go` (with `datasource` import)
3. Created examples: `examples/resources/betterado_pipeline/resource.tf`, `examples/data-sources/betterado_pipeline/data-source.tf`, `examples/data-sources/betterado_pipeline_run/data-source.tf`
4. Ran `make docs` → generated `docs/resources/pipeline.md`, `docs/data-sources/pipeline.md`, `docs/data-sources/pipeline_run.md` (Makefile auto-restores docs/guides/)
5. Updated `CHANGELOG.md` under `## [Unreleased]` with pipeline data source entries
6. Bumped `PROVIDER_VERSION.txt` 1.2.0 → 1.2.1

Gate result: **PASS** — both new tests execute and assert successfully.

## What worked

- `datasource.MetadataResponse` pattern for data source type-name assertion mirrors `resource.MetadataResponse` pattern exactly
- `make docs` handles the `git checkout -- docs/guides/` step automatically via the Makefile target
- The `-run TestFrameworkProvider_HasPipeline` regex matches both `HasPipelineResource` and `HasPipelineDataSource`

## What didn't work

_(nothing — both ACs completed in iteration 0)_

## Open questions

_(none)_

## Notes for reflection

- WI-2 left `NewPipelineResource` registered but forgot DataSources and tests; WI-3 completed the triangle
- "no tests to run" gate rejections happen when test function names don't match the `-run` regex — the solution is always to write the missing test functions
