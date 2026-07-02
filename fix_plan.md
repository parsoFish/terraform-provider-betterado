# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_project resource implemented as resource.Resource in terraform-plugin-framework WHEN terraform import is run against the betterado-standing-demo project (must NOT create a new project — org is at 1000-project cap) THEN the import succeeds, read-back asserts name/visibility/version_control attributes, and idempotency re-plan shows no diff (ExpectNonEmptyPlan: false)
  - [x] resource_project_framework.go implemented with full CRUD + ImportState by name/UUID
  - [x] Removed ImportStateVerify from TestAccProject_importByName (fixes "resource with ID not found" error from empty pre-import state)
  - [x] checkProjectImportByName verifies all required attributes
  - [x] Step 2 PlanOnly + ExpectNonEmptyPlan:false verifies idempotency
- [ ] AC2: GIVEN data.betterado_project data source implemented in framework WHEN terraform apply runs with a data source lookup by name against betterado-standing-demo THEN data source returns correct project fields; TestAccProject_dataSource_withID and TestAccProject_dataSource_withName pass
- [ ] AC3: GIVEN data.betterado_projects data source implemented in framework WHEN terraform apply runs listing projects THEN TestAccProjects_dataSource passes
- [x] AC4: GIVEN betterado_project, data.betterado_project, data.betterado_projects deregistered from SDKv2 provider.go WHEN TestProvider_HasChildResources and TestProvider_HasChildDataSources run THEN both pass with updated counts (resource removed from ResourcesMap; data sources removed from DataSourcesMap)

## Status

Gate failing: `TestAccProject_importByName` — FIXED in iteration 2 (removed ImportStateVerify from import-only test step; no prior apply means no pre-import state to compare against).

AC2 and AC3 live acceptance tests not yet validated by live gate (gate only runs TestAccProject_importByName per the quality_gate_cmd).
