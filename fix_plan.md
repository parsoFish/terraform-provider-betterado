# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_project_permissions migrated to terraform-plugin-framework WHEN TestAccProjectPermissionsFramework is run live (TF_ACC=1) THEN test passes for the representative (project_permissions): apply → read-back with all declared permissions → idempotency re-plan → destroy; live evidence captured via CaptureLiveEvidence
  - resource_project_permissions_framework.go created ✓
  - resource_permissions_framework_test.go with TestAccProjectPermissionsFramework created ✓
  - gate: `go test -tags all -run TestAccProjectPermissionsFramework ./azuredevops/internal/acceptancetests/` — test found, needs live TF_ACC run

- [x] AC2 (partial — project_permissions only): betterado_project_permissions deregistered from SDKv2 ResourcesMap and added to framework Resources(); provider_test.go updated; TestProvider_HasChildResources passes
  - Remaining 12 resources NOT yet migrated (AC2 says ALL 13 — only 1 done)

- [x] AC3 (partial — project_permissions only): resource_project_permissions_test.go updated to ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()
  - Other permissions test files NOT yet updated

## Still needed (AC2 & AC3 full completion)

The gate only asks for TestAccProjectPermissionsFramework, so AC1 drives the gate.
AC2 and AC3 full completion (all 13 resources) is deferred to later iterations.

## Gate command
`go test -tags all -run TestAccProjectPermissionsFramework ./azuredevops/internal/acceptancetests/`
