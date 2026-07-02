# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_build_folder resource registered in the framework provider WHEN a unit test calls Schema() on the framework resource THEN the schema declares project_id, path, and description attributes; no error diagnostics
- [x] AC2: GIVEN the framework resource is registered in framework_provider.go Resources() WHEN go build -mod=vendor . is run THEN the provider binary compiles without error
- [ ] AC3: GIVEN a live ADO environment with TF_ACC=1 WHEN TestAccBuildFolder_Framework_basic runs (apply → read-back → idempotency re-plan → destroy) THEN all steps pass; ExpectNonEmptyPlan is false; CaptureLiveEvidence is called with label 'acceptance-resource' and the folder's REST GET URL
