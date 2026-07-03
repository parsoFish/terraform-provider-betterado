# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new file azuredevops/internal/service/servicehook/resource_servicehook_storage_queue_pipelines_framework.go implementing resource.Resource WHEN the framework resource's Configure() is called with a non-nil ProviderData THEN it stores *client.AggregatedClient (not a stub); panic-free under the mux
- [x] AC2: GIVEN the framework implementation of betterado_servicehook_storage_queue_pipelines WHEN Create/Read/Update/Delete are exercised THEN the resource calls clients.ServiceHooksClient CRUD methods with the correct subscription shape; 404 in Read clears the ID and returns nil (no error)
- [x] AC3: GIVEN the SDKv2 registration of betterado_servicehook_storage_queue_pipelines in provider.go ResourcesMap WHEN the framework resource is registered in framework_provider.go Resources() THEN the SDKv2 entry is REMOVED from provider.go ResourcesMap in the same commit; provider_test.go resource count updated; no 'Duplicate resource type' at apply

## Notes

All three ACs were completed in iteration 0. The gate test
`TestServicehookStorageQueuePipelinesFramework_Configure` passes (3 subtests run).
`go build -mod=vendor .` passes clean. provider_test.go `TestProvider_HasChildResources` passes.
