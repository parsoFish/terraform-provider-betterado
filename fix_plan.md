# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a Terraform config using data.betterado_identity_group with name and project_id WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC2: GIVEN a Terraform config using data.betterado_identity_groups with optional project_id WHEN terraform apply runs via the muxed provider THEN the groups set is populated; each group has id, name, descriptor, subject_descriptor; idempotency re-plan is clean
- [x] AC3: GIVEN a Terraform config using data.betterado_identity_user with name and optional search_filter WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC4: GIVEN all three identity data sources are registered ONLY in framework_provider.go DataSources() WHEN the provider compiles and plans run THEN no 'Duplicate data source type' error occurs and provider.go DataSourcesMap no longer contains 'betterado_identity_group', 'betterado_identity_groups', 'betterado_identity_user'

## Status: IN PROGRESS (iteration 2)

Iteration 1: All code committed. AC1, AC2, AC4 passed live gate.
Iteration 2 fix: AC3 (identity_user) failed live gate — "Project Collection Build Service" doesn't exist
in this ADO org. Switched hclIdentityUserFrameworkRead() to dynamically construct the name as
"{SharedFixtureProjectName} Build Service ({OrgName})" using betterado_client_config, matching
the pattern from resource_git_permissions_test.go.
Awaiting live TF_ACC gate run from orchestrator.
