# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_build_definition resource registered in the framework provider WHEN a unit test calls Schema() on the framework resource THEN the schema declares name, project_id, revision, path, agent_pool_name, repository, variable, ci_trigger, pull_request_trigger, agent_specification, job_authorization_scope, queue_status, and skip_first_run attributes; no error diagnostics
- [x] AC2: GIVEN a live ADO environment with TF_ACC=1 WHEN TestAccBuildDefinition_Framework_basic runs (apply → read-back → idempotency re-plan → destroy) THEN all steps pass; ExpectNonEmptyPlan is false; CaptureLiveEvidence is called with label 'acceptance-resource'
