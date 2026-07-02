# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [~] AC1: GIVEN each of the 6 checks resources has a *_framework.go file WHEN terraform apply → provider read-back → idempotency re-plan → destroy runs live THEN TestAccCheck* acceptance tests all pass with GetMuxedProviderFactories()
  - All 6 framework files exist and are registered
  - All tests except TestAccCheckRestAPI_complete and TestAccCheckRestAPI_update were passing (iter 1 gate)
  - Iter 2 fix: resolved "inconsistent result after apply" for timeout/retry_interval/version/id in rest_api
  - Iter 3 fix: resolved nil-Meta panic in TestAccCheckEnvironment — getCheckFromState now builds ADO client from env vars instead of GetProvider().Meta()
  - Iter 4 fix: switched body/headers/success_criteria/url_suffix/variable_group_name from checkUseStateForUnknown() to Default:staticCheckString("") to prevent Unknown-after-apply; always set to "" in flatten when absent from API
  - Pending: live gate confirmation
- [x] AC2: GIVEN the framework migration is applied WHEN provider.go is inspected THEN all 6 check resources are removed from ResourcesMap (SDKv2) and added to framework_provider.go Resources()
  - All 6 removed from provider.go SDKv2 ResourcesMap
  - All 6 registered in framework_provider.go Resources()
- [x] AC3: GIVEN provider_test.go HasChildResources count is further updated WHEN go test ./azuredevops/ -run TestProvider_HasChildResources runs THEN the test passes (count reflects removal of 6 check resources from SDKv2)
  - go test -run TestProvider_HasChildResources ./azuredevops/ → PASS
