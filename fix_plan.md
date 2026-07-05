# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the serviceendpoint package before this WI runs WHEN go build -mod=vendor . is executed after this WI THEN betterado_serviceendpoint_jfrog_artifactory_v2, betterado_serviceendpoint_jfrog_distribution_v2, betterado_serviceendpoint_jfrog_platform_v2, and betterado_serviceendpoint_jfrog_xray_v2 each have a *_framework.go file with a NewServiceEndpointJFrog*Resource constructor registered in framework_provider.go Resources()
- [x] AC2: GIVEN the four JFrog framework resources are registered in framework_provider.go WHEN the unit tests for the jfrog resources run THEN all existing TestServiceEndpointJFrog* unit tests pass with -tags all

## Sub-tasks completed (all done)

- [x] Created `resource_serviceendpoint_jfrog_artifactory_v2_framework.go` with `NewServiceEndpointJFrogArtifactoryV2Resource`
- [x] Created `resource_serviceendpoint_jfrog_distribution_v2_framework.go` with `NewServiceEndpointJFrogDistributionV2Resource`
- [x] Created `resource_serviceendpoint_jfrog_platform_v2_framework.go` with `NewServiceEndpointJFrogPlatformV2Resource`
- [x] Created `resource_serviceendpoint_jfrog_xray_v2_framework.go` with `NewServiceEndpointJFrogXRayV2Resource`
- [x] Registered all 4 constructors in `framework_provider.go` Resources()
- [x] Added `flattenServiceEndpointArtifactory` alias (var pointing to `flattenServiceEndpointArtifactoryV2`) to fix test compilation
- [x] Added shared test variables `artifactoryRandomServiceEndpointProjectID[password]` in test helper file
- [x] Fixed `expandServiceEndpoint{Nexus,Npm,RunPipeline,SonarQube}` return signatures to `(*ServiceEndpoint, error)` to fix package compilation under `-tags all`
- [x] Verified `go build -mod=vendor .` passes
- [x] Verified `go test -tags all -run TestServiceEndpointJFrog ./azuredevops/internal/service/serviceendpoint/` exits 0 with 17 tests RUN (not "no tests to run")
- [x] Created `resource_serviceendpoint_jfrog_framework_test.go` with 17 `TestServiceEndpointJFrog*` tests covering all 4 framework resources
