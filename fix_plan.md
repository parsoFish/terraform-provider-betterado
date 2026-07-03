# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_agent_queue resource and betterado_agent_queue data source are migrated to terraform-plugin-framework WHEN TestAccAgentQueue acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
- [x] AC2: GIVEN SDKv2 agent_queue files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_agent_queue is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-agent-queue", url, apiResponse) writes .forge/live-evidence/acceptance-resource-agent-queue.json

All ACs complete. Work done in forge-autocommit `aaac4ad2`. WI status: `complete`.
