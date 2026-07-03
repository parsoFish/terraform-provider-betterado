# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_pipeline resource exists in the pipelines package WHEN the developer implements the betterado_pipeline data source and betterado_pipeline_run data source as framework datasource.DataSource types THEN data_pipeline_framework.go and data_pipeline_run_framework.go compile; TestPipelineDataSource* and TestPipelineRunDataSource* unit tests pass under -tags all
- [x] AC2: GIVEN a betterado_pipeline_run data source reads a Run object from GET _apis/pipelines/{id}/runs/{runId} WHEN the unit test verifies the schema THEN state attributes id, name, state, result, created_date, finished_date, pipeline_id are all defined in the schema
