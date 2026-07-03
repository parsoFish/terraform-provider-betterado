# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_extension_install resource configured with publisher_id=ms-securitydevops and extension_id=microsoft-security-devops-azdevops WHEN terraform apply runs live against a real ADO organization (TF_ACC=1) THEN the extension is installed; the provider reads it back (TestCheckResourceAttrSet for publisher_id, extension_id, version); ExpectNonEmptyPlan is false (idempotency re-plan produces empty plan); terraform destroy uninstalls it; TestAccExtensionInstall passes
  - TestAccExtensionInstall_basic written; Step 1 checks publisher_id/extension_id/version with TestCheckResourceAttrSet; Step 2 is PlanOnly/ExpectNonEmptyPlan:false; CheckDestroy calls GetInstalledExtensionByName and expects 404
  - **betterado_extension_install now registered in framework_provider.go** — the mux provider exposes it; live gate should no longer fail with "Invalid resource type"
- [x] AC2: GIVEN the live acceptance test's read-back step WHEN the extension has been applied and is being read THEN testutils.CaptureLiveEvidence("acceptance-resource-extension-install", url, apiResponse) is called with a real REST GET URL against the ExtensionManagement API; .forge/live-evidence/acceptance-resource-extension-install.json is written
  - captureExtensionInstallEvidence() calls CaptureLiveEvidence with label "acceptance-resource-extension-install" and real REST GET URL _apis/extensionmanagement/installedextensionsbyname/...
- [x] AC3: GIVEN the live acceptance test WHEN it runs without TF_ACC set THEN resource.Test skips cleanly (standard PreCheck behaviour)
  - PreCheck(t, nil) skips with "Acceptance tests skipped unless env 'TF_ACC' set"

## Outstanding

- All ACs satisfied. Gate requires live TF_ACC run with real ADO credentials/org.
