# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new file azuredevops/internal/service/servicehook/resource_servicehook_webhook_tfs_framework.go implementing resource.Resource WHEN the framework resource's Configure() is called with a non-nil ProviderData THEN it stores *client.AggregatedClient; panic-free under the mux
- [x] AC2: GIVEN the framework implementation of betterado_servicehook_webhook_tfs WHEN Create/Read/Update/Delete are exercised THEN the resource calls clients.ServiceHooksClient CRUD methods with the correct subscription shape including all 19 TFS event type blocks; 404 in Read clears the resource and returns nil
- [x] AC3: GIVEN the SDKv2 registration of betterado_servicehook_webhook_tfs in provider.go ResourcesMap WHEN the framework resource is registered in framework_provider.go Resources() THEN the SDKv2 entry is REMOVED from provider.go ResourcesMap in the same commit; no 'Duplicate resource type' at apply; provider_test.go count updated

All ACs complete. Committed as a79292d3.
