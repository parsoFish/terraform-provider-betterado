# Fix Plan

> Checklist for WI-8. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 sub-task: Fix schema so HCL block syntax works for condition/action — changed from SetNestedAttribute (Attributes map) to SetNestedBlock (Blocks map) using NestedBlockObject
- [ ] AC1: GIVEN betterado_workitemtrackingprocess_rule resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessRule_Basic, TestAccWorkitemtrackingprocessRule_Update, TestAccWorkitemtrackingprocessRule_ConditionTypes, TestAccWorkitemtrackingprocessRule_ConditionGroupMembership, TestAccWorkitemtrackingprocessRule_ActionTypes, TestAccWorkitemtrackingprocessRule_HideTargetField, and TestAccWorkitemtrackingprocessRule_DisallowValue all pass
- [x] AC2: GIVEN SDKv2 resource_rule.go removed and deregistered from provider.go WHEN provider.go ResourcesMap inspected THEN betterado_workitemtrackingprocess_rule registered ONLY in framework_provider.go; orphaned SDKv2 file deleted; provider_test.go resource count updated (done in previous commit 2bedf727)
- [x] AC3: GIVEN live acceptance test running WHEN rule resource is read back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-rule", url, apiResponse) is called in TestAccWorkitemtrackingprocessRule_Basic (done in previous commit 2bedf727)
