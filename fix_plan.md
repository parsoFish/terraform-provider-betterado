# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_feed_permission resource configured with feed_id, project_id, role=reader, and identity_descriptor pointing to a real ADO group WHEN terraform apply runs via the mux ProtoV6 provider path (GetMuxedProviderFactories) THEN the feed permission is created in ADO, the provider reads it back asserting role and identity_descriptor, an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy succeeds (permission removed)
  - Created resource_feed_permission_framework.go (framework CRUD + import)
  - Registered in framework_provider.go
  - Deregistered from provider.go SDKv2 ResourcesMap (mux requires no duplicates)
  - TestAccFeedPermissionFramework_basic added with ProtoV6ProviderFactories, idempotency step, CheckDestroy
  - **FIXED (iteration 1):** display_name had no plan modifier + wasn't guaranteed known after apply
    → added useStateForUnknown() to schema + ensured Create always sets known display_name value

- [x] AC2: GIVEN the framework betterado_feed_permission resource is live and readable via the Feed Permissions REST API WHEN CaptureLiveEvidence is called inside the acceptance test check step with label acceptance-resource THEN .forge/live-evidence/acceptance-resource.json is written with a real ADO GET URL (dev.azure.com feed permissions endpoint)
  - captureFeedPermissionFrameworkEvidence() calls CaptureLiveEvidence("acceptance-resource", url, permissions)
  - URL format: <AZDO_ORG_SERVICE_URL>/<projectId>/_apis/packaging/feeds/<feedId>/permissions?api-version=7.1

## Remaining
- None identified; awaiting live gate to confirm.
  - Root cause of last gate failure was: "provider still indicated an unknown value for display_name after apply"
  - Fixed by adding useStateForUnknown() plan modifier + always resolving display_name to known value in Create
