# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN betterado_workitemtrackingprocess_workitemtype resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessWorkItemType_Basic and TestAccWorkitemtrackingprocessWorkItemType_CreateAndUpdate both pass; no duplicate resource type error at plan/apply time
- [ ] AC2: GIVEN betterado_workitemtrackingprocess_workitemtype and betterado_workitemtrackingprocess_workitemtypes data sources migrated to framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessWorkItemType_DataSource_Get and TestAccWorkitemtrackingprocessWorkItemTypes_DataSource_List pass
- [ ] AC3: GIVEN SDKv2 resource_work_item_type.go, data_work_item_type.go, data_work_item_types.go and their unit test files removed and deregistered WHEN provider.go ResourcesMap and DataSourcesMap inspected THEN workitemtype resource and both workitemtype data sources registered ONLY in framework_provider.go; orphaned SDKv2 files deleted; provider_test.go counts updated
- [ ] AC4: GIVEN live acceptance test running WHEN work item type is read back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-workitemtype", url, apiResponse) is called

## Sub-tasks completed

- [x] Framework resource `resource_work_item_type_framework.go` implemented (prior iterations)
- [x] Framework data sources `data_work_item_type_framework.go` and `data_work_item_types_framework.go` implemented (prior iterations)
- [x] SDKv2 files (`resource_work_item_type.go`, `data_work_item_type.go`, `data_work_item_types.go`, unit tests) deleted (prior iterations)
- [x] Provider registrations updated: removed from `provider.go`, added to `framework_provider.go` (prior iterations)
- [x] Acceptance tests updated to use `ProtoV6ProviderFactories` and direct clients (prior iterations)
- [x] `captureWorkItemTypeEvidence` / `CaptureLiveEvidence` wired in acceptance test (AC4) (prior iterations)
- [x] **Iteration 4 fix**: Added VS1640142 to `ResponseWasNotFound` in `HttpResponse.go` — the destroy check was failing because after the process is deleted, reading a work item type returns HTTP 400 with `VS1640142: Work item type not found or you do not have permission in the process`, which was not recognized as a not-found error.

## Root cause of live gate failure (iteration 4 entry)

All three test failures had the same pattern:
```
Error running post-test destroy, there may be dangling resources:
error reading work item type <name> after destroy: VS1640142: Work item type not found or you do not have permission in the process <id>
```

This is the `checkWorkItemTypeDestroyed` destroy check calling `GetProcessWorkItemType` after resources are torn down. After the parent process is deleted, the API returns HTTP 400 with `VS1640142` instead of 404. `utils.ResponseWasNotFound` only handled HTTP 400 with `VS800075` and `VS402806` — not `VS1640142`. The fix adds `VS1640142` to the list.
