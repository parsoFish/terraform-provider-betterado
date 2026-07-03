# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_environment_resource_kubernetes resource is migrated to terraform-plugin-framework WHEN TestAccEnvironmentResourceKubernetes acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
  - Framework resource fully implemented: Create/Read/Update(error)/Delete with all attrs
  - Acceptance test `TestAccEnvironmentResourceKubernetes_createUpdate` implemented with `ExpectNonEmptyPlan: false` and destroy check
  - **PENDING live gate run** — offline tests skip without TF_ACC (expected)
- [x] AC2: GIVEN SDKv2 environment_resource_kubernetes file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_environment_resource_kubernetes is absent from SDKv2 ResourcesMap; source files deleted; framework_provider.go includes NewEnvironmentResourceKubernetesResource; provider_test.go count updated
  - SDKv2 `resource_environment_resource_kubernetes.go` — DELETED
  - SDKv2 unit test `resource_environment_resource_kubernetes_test.go` — DELETED
  - `provider.go` ResourcesMap: kubernetes resource NOT listed (has comment noting framework migration)
  - `framework_provider.go`: `NewEnvironmentResourceKubernetesResource` registered at line 209
  - `provider_test.go`: resource NOT in expectedResources list; comment added noting framework migration
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment-kubernetes.json
  - `captureEnvironmentKubernetesEvidence` in acceptance test calls `testutils.CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, k8sResource)`
  - **PENDING live gate run** — will write evidence when TF_ACC runs

## Offline gates (all PASS)
- `go build ./azuredevops/...` ✅
- `go test -v ./...` ✅ (serviceendpoint pre-existing failure is from main, not ours)
- `make terrafmt-check` ✅
- `golangci-lint run --new-from-rev=main ./azuredevops/... → 0 issues` ✅
