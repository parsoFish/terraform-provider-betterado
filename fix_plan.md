# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_pipeline resource and data sources exist in azuredevops/internal/service/pipelines/ WHEN framework_provider.go Resources() and DataSources() are updated to include the new pipeline types THEN TestFrameworkProvider_HasPipelineResource passes; grep of azuredevops/provider.go confirms zero new SDKv2 entries for pipeline types; provider_test.go includes a test for betterado_pipeline_run data source
  - [x] pipelines.NewPipelineResource already in Resources() (done in WI-2)
  - [x] Added pipelines.NewPipelineDataSource and pipelines.NewPipelineRunDataSource to DataSources()
  - [x] TestFrameworkProvider_HasPipelineResource and TestFrameworkProvider_HasPipelineDataSource added → gate PASSES
  - [x] No betterado_pipeline in SDKv2 provider.go confirmed
- [x] AC2: GIVEN the implementation is complete WHEN make docs runs THEN docs created; guides restored; CHANGELOG updated; PROVIDER_VERSION bumped
  - [x] Created examples/resources/betterado_pipeline/resource.tf
  - [x] Created examples/data-sources/betterado_pipeline/data-source.tf
  - [x] Created examples/data-sources/betterado_pipeline_run/data-source.tf
  - [x] make docs ran successfully → docs/resources/pipeline.md, docs/data-sources/pipeline.md, docs/data-sources/pipeline_run.md created
  - [x] docs/guides/ restored (make docs Makefile target runs git checkout -- docs/guides/ automatically)
  - [x] CHANGELOG.md updated with betterado_pipeline and betterado_pipeline_run data source entries under ## [Unreleased]
  - [x] PROVIDER_VERSION.txt bumped from 1.2.0 → 1.2.1
