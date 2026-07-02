# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_build_definition resource registered in the framework provider WHEN a unit test calls Schema() on the framework resource THEN the schema declares name, project_id, revision, path, agent_pool_name, repository, variable, ci_trigger, pull_request_trigger, agent_specification, job_authorization_scope, queue_status, and skip_first_run attributes; no error diagnostics
- [ ] AC2: GIVEN a live ADO environment with TF_ACC=1 WHEN TestAccBuildDefinition_Framework_basic runs (apply → read-back → idempotency re-plan → destroy) THEN all steps pass; ExpectNonEmptyPlan is false; CaptureLiveEvidence is called with label 'acceptance-resource'

## Remaining sub-tasks

- [ ] Remove "betterado_build_definition" from provider.go ResourcesMap (or comment out) to avoid mux conflict
- [ ] Create `azuredevops/internal/acceptancetests/resource_build_definition_framework_test.go` with `TestAccBuildDefinition_Framework_basic`
- [ ] Verify readIntoModel handles repository/variable/trigger fields properly for idempotent read-back
- [ ] Add example HCL at `examples/resources/betterado_build_definition/resource.tf`
- [ ] Run `make docs` then `git checkout -- docs/guides/`
- [ ] Add CHANGELOG entry under `## [Unreleased]`
- [ ] Bump `PROVIDER_VERSION.txt`
