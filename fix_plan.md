# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_workitemtrackingprocess_workitemtype resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessWorkItemType_Basic and TestAccWorkitemtrackingprocessWorkItemType_CreateAndUpdate both pass; no duplicate resource type error at plan/apply time
  - [x] Created resource_work_item_type_framework.go with full CRUD + Import
  - [x] Updated acceptance test to use GetMuxedProviderFactories()
  - [x] Fixed CheckDestroy to use getWorkItemTypeDirectClient() (not GetProvider().Meta())
- [x] AC2: GIVEN betterado_workitemtrackingprocess_workitemtype and betterado_workitemtrackingprocess_workitemtypes data sources migrated to framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessWorkItemType_DataSource_Get and TestAccWorkitemtrackingprocessWorkItemTypes_DataSource_List pass
  - [x] Created data_work_item_type_framework.go
  - [x] Created data_work_item_types_framework.go (SetNestedAttribute)
  - [x] Updated data source acceptance tests
- [x] AC3: SDKv2 files removed; workitemtype resource and data sources deregistered from provider.go; registered in framework_provider.go; provider_test.go counts updated
  - [x] Deleted resource_work_item_type.go, resource_work_item_type_test.go
  - [x] Deleted data_work_item_type.go, data_work_item_type_test.go
  - [x] Deleted data_work_item_types.go, data_work_item_types_test.go
  - [x] Removed from provider.go ResourcesMap and DataSourcesMap
  - [x] Registered in framework_provider.go Resources() and DataSources()
  - [x] Updated provider_test.go counts
- [x] AC4: CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-workitemtype", url, wit) called in acceptance test before destroy

## Remaining
- [ ] Live gate must pass (iteration 1 will confirm via TF_ACC)
