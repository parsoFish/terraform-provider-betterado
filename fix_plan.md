# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the ADO Task Agent REST API v7.1 documentation and each SDKv2 resource schema WHEN compared field-by-field against the Terraform attributes for agent_pool, agent_queue, deployment_group, elastic_pool, environment, environment_resource_kubernetes, variable_group, variable_group_variable, task_group data source THEN docs/taskagent-gap-matrix.md exists, lists every field with status (mapped/partial/missing), and defers unimplemented writable gaps explicitly

**AC1 DONE** — `docs/taskagent-gap-matrix.md` committed in prior iteration (ba0761ab). WI-1 status: complete.

## Gate-tightening fixes (iteration 1)

- [x] `resource_variable_group_variable_test.go`: migrated from project-per-test to fixture project (avoids 1000-project org cap); switched from `GetProviders()` to `GetMuxedProviderFactories()`; use `GetDirectClient()` in check helpers.
- [x] `resource_variable_group_framework.go` + `provider.go`: fixed pre-existing gofmt formatting drift (unblocks `make test` fmtcheck).
