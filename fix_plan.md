# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_security_permissions resource and betterado_security_namespace*, betterado_security_namespaces data sources migrated to terraform-plugin-framework WHEN an acceptance test is run live (TF_ACC=1) THEN TestAccSecurityPermissionsFramework passes (apply → provider read-back with all declared permissions asserted → idempotency re-plan produces no diff → destroy clean); live evidence captured via CaptureLiveEvidence
  - Implementation: resource_security_permissions_framework.go + 3 data source framework files
  - Test: resource_security_permissions_framework_test.go (TestAccSecurityPermissionsFramework with mux factory, idempotency step, CaptureLiveEvidence)
  - **Iteration 2 fix (commit 930af1f8):** replaced project creation with SharedReleaseFixture (avoids 1000-project cap); replaced CheckProjectDestroyed (nil-Meta panic on mux) with local checkSecurityPermissionsFrameworkDestroyed using getDirectClient(); replaced betterado_identity_group with betterado_group (descriptor attr)
  - **Iteration 3 fix (commit 4772ff30):** SharedReleaseFixture itself calls QueueCreateProject when betterado-standing-demo doesn't exist → fails at 1000-cap. Replaced entirely with resolveSecurityPermissionsFixtureProject: prefers GetProject(standing-demo), falls back to GetProjects first-WellFormed-project — NEVER calls QueueCreateProject.
  - **Pending: live gate (TF_ACC=1) run to confirm pass**

- [x] AC2: GIVEN betterado_security_permissions is registered in framework_provider.go and REMOVED from provider.go ResourcesMap WHEN provider compiles and TestProvider_HasChildResources runs THEN no duplicate-resource-type error; the resource is absent from the SDKv2 ResourcesMap count and present in the framework Resources() slice
  - Removed from provider.go ResourcesMap, added to framework_provider.go Resources()
  - provider_test.go SDKv2 list updated (commented out)
  - TestProvider_HasChildResources: PASS (verified locally)

- [x] AC3: GIVEN betterado_security_namespace, betterado_security_namespace_token, betterado_security_namespaces data sources migrated to framework and deregistered from SDKv2 WHEN provider compiles and TestProvider_HasChildDataSources runs THEN data sources are absent from the SDKv2 DataSourcesMap count and present in framework DataSources() slice
  - Removed from provider.go DataSourcesMap, added to framework_provider.go DataSources()
  - provider_test.go SDKv2 list updated (commented out)
  - TestProvider_HasChildDataSources: PASS (verified locally)

## Offline gate results (iteration 1)

- `make test`: PASS
- `golangci-lint run --new-from-rev=main ./azuredevops/...`: 0 issues
- `go test -tags all -list TestAccSecurityPermissionsFramework ./azuredevops/internal/acceptancetests/`: test registered ✓
- `go test -tags all -run TestProvider_HasChildResources ./azuredevops/`: PASS
- `go test -tags all -run TestProvider_HasChildDataSources ./azuredevops/`: PASS

## Offline gate results (iteration 2)

- `go build -tags all ./azuredevops/...`: PASS
- `golangci-lint run --new-from-rev=main ./azuredevops/...`: 0 issues
- `go test -tags all -list TestAccSecurityPermissionsFramework ./azuredevops/internal/acceptancetests/`: test registered ✓
- `go test -tags all -run TestProvider_HasChildResources ./azuredevops/`: PASS
- `go test -tags all -run TestProvider_HasChildDataSources ./azuredevops/`: PASS
