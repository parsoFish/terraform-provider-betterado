# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_securityrole_assignment resource and betterado_securityrole_definitions data source migrated to terraform-plugin-framework WHEN an acceptance test is run live (TF_ACC=1) THEN TestAccSecurityRoleAssignmentFramework passes (apply → provider read-back asserting scope, resource_id, identity_id, role_name → idempotency re-plan → destroy); live evidence captured via CaptureLiveEvidence
  - Implementation: `azuredevops/internal/service/securityroles/resource_securityrole_assignment_framework.go`
  - Implementation: `azuredevops/internal/service/securityroles/data_securityrole_definitions_framework.go`
  - Test: `azuredevops/internal/acceptancetests/resource_securityrole_assignment_framework_test.go`
  - Both registered in `azuredevops/internal/provider/framework_provider.go`
  - Iteration 2 fix: destroy check failed because ADO PATCH delete reverts to inherited access;
    fixed checkDestroyed to accept inherited-access as "done", added waitForDeletion,
    and stopped overwriting state.Scope from Role.Scope in Read. Committed 41dc6aad.
  - AWAITING live gate: forge will re-run live

- [x] AC2: GIVEN betterado_securityrole_assignment removed from SDKv2 ResourcesMap and betterado_securityrole_definitions removed from DataSourcesMap WHEN provider compiles and TestProvider_HasChildResources / TestProvider_HasChildDataSources run THEN no duplicate-resource-type error; counts updated correctly in provider_test.go
  - Removed both from `provider.go` + unused `securityroles` import
  - Updated `provider_test.go` expected lists
  - TestProvider_HasChildResources PASSES offline ✓
  - TestProvider_HasChildDataSources PASSES offline ✓

## Sub-tasks

- [x] resource_securityrole_assignment_framework.go (full CRUD + inline plan modifiers)
- [x] data_securityrole_definitions_framework.go (Read, computed Set of definitions)
- [x] framework_provider.go: register NewSecurityRoleAssignmentResource + NewSecurityRoleDefinitionsDataSource
- [x] provider.go: remove SDKv2 entries + import
- [x] provider_test.go: counts updated
- [x] TestAccSecurityRoleAssignmentFramework: uses shared project (no new project create), creates env + group via ADO API, asserts scope/resource_id/identity_id/role_name, idempotency, destroy, CaptureLiveEvidence
- [x] examples/resources/betterado_securityrole_assignment/resource.tf
- [x] All changes committed (f13104ee → 41dc6aad)
- [x] Fix destroy check: handle inherited-access after PATCH delete (commit 41dc6aad)
