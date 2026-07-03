# Fix Plan

> Checklist for WI-7. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: resource_list_framework.go created; registered in framework_provider.go; list acceptance tests use GetMuxedProviderFactories(); PENDING live TF_ACC run
- [x] AC2: resource_field_framework.go created; registered in framework_provider.go; field tests already used GetMuxedProviderFactories(); PENDING live TF_ACC run
- [x] AC3: resource_list.go, resource_field.go, order.go deleted; deregistered from provider.go; provider_test.go updated; TestProvider_HasChildResources PASSES
- [x] AC4: captureListEvidence() added to TestAccWorkitemtrackingprocessList_Basic; calls testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-list", url, apiResponse)
