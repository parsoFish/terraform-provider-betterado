# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a Terraform config using data.betterado_identity_group with name and project_id WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC2: GIVEN a Terraform config using data.betterado_identity_groups with optional project_id WHEN terraform apply runs via the muxed provider THEN the groups set is populated; each group has id, name, descriptor, subject_descriptor; idempotency re-plan is clean
- [x] AC3: GIVEN a Terraform config using data.betterado_identity_user with name and optional search_filter WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC4: GIVEN all three identity data sources are registered ONLY in framework_provider.go DataSources() WHEN the provider compiles and plans run THEN no 'Duplicate data source type' error occurs and provider.go DataSourcesMap no longer contains 'betterado_identity_group', 'betterado_identity_groups', 'betterado_identity_user'

## Status: IN PROGRESS (iteration 3)

Iteration 1: All code committed. AC1, AC2, AC4 passed live gate.
Iteration 2 fix: AC3 (identity_user) failed live gate — "{SharedFixtureProjectName} Build Service" not found
because the per-project build service only exists after a pipeline run in that project.
Iteration 3 fix: Switched to "Project Collection Build Service ({OrgName})" — the collection-level
system identity that is permanently present in every ADO org without needing any pipeline runs.
Awaiting live TF_ACC gate run from orchestrator.
