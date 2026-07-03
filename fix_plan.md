# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_environment_resource_kubernetes resource is migrated to terraform-plugin-framework WHEN TestAccEnvironmentResourceKubernetes acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
  - [x] Create resource_environment_resource_kubernetes_framework.go
  - [x] Register in framework_provider.go
  - [x] Deregister from provider.go
  - [x] Delete SDKv2 source files
  - [x] Update acceptance test: rename function to TestAccEnvironmentResourceKubernetes_createUpdate, add ExpectNonEmptyPlan: false, add attribute checks (namespace, cluster_name)

- [x] AC2: GIVEN SDKv2 environment_resource_kubernetes file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_environment_resource_kubernetes is absent from SDKv2 ResourcesMap; source files deleted; framework_provider.go includes NewEnvironmentResourceKubernetesResource; provider_test.go count updated
  - [x] betterado_environment_resource_kubernetes removed from provider.go ResourcesMap
  - [x] SDKv2 source files deleted
  - [x] NewEnvironmentResourceKubernetesResource added to framework_provider.go
  - [x] betterado_environment_resource_kubernetes removed from provider_test.go list

- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment-kubernetes.json
  - [x] captureEnvironmentKubernetesEvidence() added — calls CaptureLiveEvidence with label "acceptance-resource-environment-kubernetes"

## Status: awaiting live gate (TF_ACC=1)

All code changes committed in baacdbf7. Offline gates pass:
- make test: PASS (0 failures)
- golangci-lint --new-from-rev=main: 0 issues
- make terrafmt-check: PASS
- make docs: PASS

Live gate will run TestAccEnvironmentResourceKubernetes_createUpdate with TF_ACC=1 to write live evidence.
