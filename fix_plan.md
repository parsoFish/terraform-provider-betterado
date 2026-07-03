# Fix Plan

> Checklist for WI-7. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_deployment_group resource is migrated to terraform-plugin-framework WHEN TestAccDeploymentGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
- [x] AC2: GIVEN SDKv2 deployment_group file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_deployment_group is absent from SDKv2 ResourcesMap; source files deleted; framework_provider.go includes NewDeploymentGroupResource; provider_test.go count updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-deployment-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-deployment-group.json

## Status

All ACs implemented in iteration 0 (commit 0ee972ad). Gate was blocked by:
1. 1000-project limit → fixed (iteration 0): switched to standing fixture project.
2. Classic pipelines disabled (iteration 3/commit 51f67ebe): expanded
   `enableClassicPipelinesForFixtureProject` to also set
   `disableClassicDeploymentPipelineCreation=false` (deployment groups need this,
   not just `disableClassicBuildPipelineCreation`) and added org-level PATCH first
   (non-fatal) before the project-level PATCH.
