# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a Terraform config using data.betterado_identity_group with name and project_id WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC2: GIVEN a Terraform config using data.betterado_identity_groups with optional project_id WHEN terraform apply runs via the muxed provider THEN the groups set is populated; each group has id, name, descriptor, subject_descriptor; idempotency re-plan is clean
- [x] AC3: GIVEN a Terraform config using data.betterado_identity_user with name and optional search_filter WHEN terraform apply runs via the muxed provider THEN descriptor and subject_descriptor are populated; idempotency re-plan is clean
- [x] AC4: GIVEN all three identity data sources are registered ONLY in framework_provider.go DataSources() WHEN the provider compiles and plans run THEN no 'Duplicate data source type' error occurs and provider.go DataSourcesMap no longer contains 'betterado_identity_group', 'betterado_identity_groups', 'betterado_identity_user'

## Status: COMPLETE (iteration 4)

Iteration 1: All code committed. AC1, AC2, AC4 passed live gate.
Iteration 2 fix: AC3 (identity_user) failed live gate — "{SharedFixtureProjectName} Build Service" not found
because the per-project build service only exists after a pipeline run in that project.
Iteration 3 fix: Switched to "Project Collection Build Service ({OrgName})" — but live gate still failed.
Root cause: validateIdentityUser checked ProviderDisplayName (a GUID for service accounts), not CustomDisplayName
(the actual human-readable "Project Collection Build Service (davidgparsonson)" name).

Iteration 4 fix: Fixed validateIdentityUser in data_identity_user.go to check BOTH ProviderDisplayName AND
CustomDisplayName for the DisplayName filter case. All three acceptance tests now pass live:
- TestAccIdentityDataSources_Framework/IdentityUser PASS (4.75s)
- TestAccIdentityDataSources_Framework/IdentityGroups PASS (4.39s)
- TestAccIdentityDataSources_Framework/IdentityGroup PASS (4.76s)
