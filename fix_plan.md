# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN serviceendpoint_jenkins, serviceendpoint_argocd, serviceendpoint_incomingwebhook, serviceendpoint_externaltfs, serviceendpoint_azuredevops exist in the SDKv2 provider WHEN all five migrated to terraform-plugin-framework THEN framework resource files exist for all five; all deregistered from provider.go ResourcesMap; registered in framework_provider.go Resources(); no Duplicate resource type panic; CI-equivalent gate passes (make test green, golangci-lint clean on changed code)
- [x] AC2: GIVEN acceptance tests for migrated CI/CD tool resources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go counts updated for removed SDKv2 resources
- [x] AC3: GIVEN framework resource files compile correctly WHEN go build -mod=vendor . is run THEN provider binary builds without errors; no import cycles; TypeName for each resource uses req.ProviderTypeName + suffix pattern

## Completion summary (iteration 0)

All ACs completed in prior iteration (commit 95ea70cd):

- **AC1**: Framework resource files created for all 5 endpoints:
  - `resource_serviceendpoint_jenkins_framework.go`
  - `resource_serviceendpoint_argocd_framework.go`
  - `resource_serviceendpoint_incomingwebhook_framework.go`
  - `resource_serviceendpoint_externaltfs_framework.go`
  - `resource_serviceendpoint_azuredevops_framework.go`
  - All 5 deregistered from `provider.go` (with comments noting migration)
  - All 5 registered in `framework_provider.go` Resources() (lines 221-225)
  - `TestProvider_HasChildResources` passes
  - `golangci-lint --new-from-rev=main` 0 issues

- **AC2**: All acceptance test files updated to use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
  - provider_test.go expectedResources list updated (5 names removed)

- **AC3**: `go build -mod=vendor .` clean; all TypeName use `req.ProviderTypeName + "_serviceendpoint_<suffix>"` pattern

Gate: `make test` clean (no FAIL lines), `golangci-lint` 0 issues, `go build` clean.
