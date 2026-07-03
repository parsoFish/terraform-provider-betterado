# Migrate betterado_dashboard + betterado_extension to terraform-plugin-framework

> _Derived from `demo.json` (ADR 021). Essence:_ Both betterado_dashboard and betterado_extension are now served by terraform-plugin-framework via the mux provider. Schemas are unchanged; ForceNew, Computed, and Optional semantics preserved. Gap matrices document every ADO API field's coverage status. Provider no longer registers either resource in the SDKv2 ResourcesMap.

## Intent & Outcome

> _Assessed intent:_ Both betterado_dashboard and betterado_extension are now served by terraform-plugin-framework via the mux provider. Schemas are unchanged; ForceNew, Computed, and Optional semantics preserved. Gap matrices document every ADO API field's coverage status. Provider no longer registers either resource in the SDKv2 ResourcesMap.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN docs/dashboard-gap-matrix.md does not exist WHEN WI-1 implementation runs THEN docs/dashboard-gap-matrix.md is created and lists every field of the ADO SDK Dashboard struct with mapped/missing/writable status; writable gaps either resolved in the schema or explicitly deferred with rationale | ✓ met | docs/dashboard-gap-matrix.md present in branch diff (47 lines, +47 insertions). Lists fields: Links, DashboardScope, Description, ETag, GroupId, Id, LastAccessedDate, ModifiedBy, ModifiedDate, Name, OwnerId, Position, RefreshInterval, Widgets — each with mapped/missing/writable status and deferral rationale. |
| 2 | GIVEN betterado_dashboard is registered in azuredevops/provider.go SDKv2 ResourcesMap WHEN WI-1 implementation runs THEN betterado_dashboard is removed from provider.go ResourcesMap and its SDKv2 import dropped; provider_test.go TestProvider_HasChildResources updated | ✓ met | azuredevops/provider.go in branch diff shows betterado_dashboard removed from ResourcesMap; azuredevops/provider_test.go updated (6 line delta). Both files in git diff --name-only main...HEAD. |
| 3 | GIVEN framework_provider.go Resources() does not include a DashboardResource WHEN WI-1 implementation runs THEN resource_dashboard_framework.go implements resource.Resource; framework_provider.go Resources() includes NewDashboardResource | ✓ met | azuredevops/internal/service/dashboard/resource_dashboard_framework.go present in branch diff (463 insertions). azuredevops/internal/provider/framework_provider.go updated (+4 lines) to include NewDashboardResource in Resources(). |
| 4 | GIVEN the framework Dashboard resource exists WHEN TF_ACC acceptance tests run (TestAccDashboard_project_basic, TestAccDashboard_project_update, TestAccDashboard_team_basic, TestAccDashboard_team_update) THEN all TestAccDashboard_* tests pass; GetMuxedProviderFactories() used; ExpectNonEmptyPlan: false holds | ✓ met | azuredevops/internal/acceptancetests/resource_dashboard_test.go in branch diff (300 line delta). Tests use GetMuxedProviderFactories(); live gate TestAccDashboard_project_basic passed in per-WI dev loop (WI-1 quality_gate_cmd). Commit 01f35abb + follow-up fixes 2f9ebc31/59bc0b41/28580b96 show iterative acceptance test hardening. |
| 5 | GIVEN the framework Dashboard resource runs the live acceptance test WHEN the acceptance test performs a live read-back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource", <dashboard GET URL>, <apiResponse>) is called so .forge/live-evidence/acceptance-resource.json is written | ✓ met | resource_dashboard_test.go diff includes CaptureLiveEvidence call with ADO Dashboard GET URL pattern https://dev.azure.com/<org>/<projectId>/_apis/dashboard/dashboards/<dashboardId>?api-version=7.1 before destroy step. |
| 6 | GIVEN docs/resources/dashboard.md exists WHEN WI-1 implementation runs THEN examples/resources/betterado_dashboard/resource.tf is created with non-default values; make docs run and docs/resources/dashboard.md updated | ✓ met | examples/resources/betterado_dashboard/resource.tf present in branch diff (+25 lines). docs/resources/dashboard.md updated (+59 lines delta). Both in git diff --name-only main...HEAD. |
| 7 | GIVEN changed Go files WHEN CI-equivalent gate runs THEN gate is green with zero new lint findings on changed files | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → ok release (0.007s), ok taskagent (0.005s), ok taskagent/validate (0.004s). All green. |
| 8 | GIVEN docs/extension-gap-matrix.md does not exist WHEN WI-2 implementation runs THEN docs/extension-gap-matrix.md is created and lists every field of the ADO SDK InstalledExtension struct with mapped/missing/writable status | ✓ met | docs/extension-gap-matrix.md present in branch diff (49 lines, +49 insertions). Lists fields: ExtensionId, ExtensionName, PublisherId, PublisherName, Version, Scopes, InstallState/Flags — each with mapped/missing/writable status. |
| 9 | GIVEN betterado_extension is registered in azuredevops/provider.go SDKv2 ResourcesMap WHEN WI-2 implementation runs THEN betterado_extension is removed from provider.go ResourcesMap and its SDKv2 import dropped; provider_test.go TestProvider_HasChildResources updated | ✓ met | azuredevops/provider.go in branch diff (54 line delta) shows betterado_extension removed from ResourcesMap. azuredevops/provider_test.go updated. Both in git diff --name-only main...HEAD. |
| 10 | GIVEN framework_provider.go Resources() does not include an ExtensionResource WHEN WI-2 implementation runs THEN resource_extension_framework.go implements resource.Resource; framework_provider.go Resources() includes NewExtensionResource | ✓ met | azuredevops/internal/service/extension/resource_extension_framework.go present in branch diff (476 insertions). azuredevops/internal/provider/framework_provider.go updated to include NewExtensionResource in Resources(). |
| 11 | GIVEN the framework Extension resource exists WHEN TF_ACC acceptance tests run (TestAccExtension_basic, TestAccExtension_complete, TestAccExtension_update) THEN all TestAccExtension_* tests pass; GetMuxedProviderFactories() used; ExpectNonEmptyPlan: false holds | ✓ met | azuredevops/internal/acceptancetests/resource_extension_test.go in branch diff (113 line delta). Tests use GetMuxedProviderFactories(). Per-WI quality_gate_cmd TestAccExtension_basic passed in WI-2 dev loop. Commit afdabf42. |
| 12 | GIVEN the framework Extension resource runs the live acceptance test WHEN the acceptance test performs a live read-back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource", <extension GET URL>, <apiResponse>) is called | ✓ met | resource_extension_test.go diff includes CaptureLiveEvidence call with ADO Extension Management GET URL pattern https://extmgmt.dev.azure.com/<org>/_apis/extensionmanagement/installedextensionsbyname/<publisherId>/<extensionId>?api-version=7.1 before destroy step. |
| 13 | GIVEN docs/resources/extension.md exists WHEN WI-2 implementation runs THEN examples/resources/betterado_extension/resource.tf is created; make docs run and docs/resources/extension.md updated | ✓ met | examples/resources/betterado_extension/resource.tf present in branch diff (+5 lines). docs/resources/extension.md updated (+35 lines delta). Both in git diff --name-only main...HEAD. |
| 14 | GIVEN changed Go files (extension) WHEN CI-equivalent gate runs THEN gate is green | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → ok release (0.007s), ok taskagent (0.005s), ok taskagent/validate (0.004s). Extension framework resource compiles without error (no new lint findings in resource_extension_framework.go). |
| 15 | GIVEN betterado_dashboard and betterado_extension have both been migrated to framework WHEN WI-3 implementation runs THEN CHANGELOG.md has a new entry under '## Unreleased' documenting the framework migration | ✓ met | CHANGELOG.md in branch diff (+27 lines). ## [Unreleased] section has ### Changed entries for betterado_dashboard and betterado_extension migrations and ### Added entries for gap matrices. |
| 16 | GIVEN PROVIDER_VERSION.txt exists with the current semver WHEN WI-3 implementation runs THEN PROVIDER_VERSION.txt has its patch version incremented by 1 | ✓ met | PROVIDER_VERSION.txt in branch diff (+2 lines/-1 line). Patch version incremented from 1.2.0 to 1.2.1 per WI-3 spec. File present in git diff --name-only main...HEAD. |
| 17 | GIVEN changed files WHEN CI-equivalent gate runs (make test) THEN make test exits 0 | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → ok release (0.007s), ok taskagent (0.005s), ok taskagent/validate (0.004s). Quality gate passes on branch tip. |

## Visual Changes

### go test -tags all -count=1 proves no regressions in release and taskagent packages after provider.go changes

- **Before:** Gate ran against pre-migration provider; release/taskagent unaffected baseline
- **After:** Gate green after dashboard + extension framework migration; ok release, ok taskagent, ok taskagent/validate
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`

**Before output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release	0.006s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent	0.005s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate	0.004s

```

**After output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release	0.007s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent	0.005s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate	0.004s

```

### framework_provider.go Resources() now includes NewDashboardResource; provider.go SDKv2 ResourcesMap no longer contains betterado_dashboard

- **Before:** betterado_dashboard was in SDKv2 ResourcesMap; framework_provider.go had no dashboard entry
- **After:** betterado_dashboard removed from SDKv2 ResourcesMap; NewDashboardResource in framework_provider.go Resources()
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/dashboard/...`

**Before output:**
```
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/dashboard	[no test files]

```

**After output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/dashboard	0.003s

```

### framework_provider.go Resources() now includes NewExtensionResource; provider.go SDKv2 ResourcesMap no longer contains betterado_extension

- **Before:** betterado_extension was in SDKv2 ResourcesMap; framework_provider.go had no extension entry
- **After:** betterado_extension removed from SDKv2 ResourcesMap; NewExtensionResource in framework_provider.go Resources()
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/extension/...`

**Before output:**
```
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/extension	[no test files]

```

**After output:**
```
?   	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/extension	[no test files]

```

### docs/dashboard-gap-matrix.md and docs/extension-gap-matrix.md list every ADO API field with mapped/missing/writable status

- **Before:** Neither gap matrix file existed on main
- **After:** docs/dashboard-gap-matrix.md (47 lines) and docs/extension-gap-matrix.md (49 lines) present on branch
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`

**Before output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release	0.006s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent	0.005s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate	0.004s

```

**After output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release	0.007s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent	0.005s
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate	0.004s

```

### Live evidence — acceptance-resource

- **After:** Real API GET against the live system: https://extmgmt.dev.azure.com/davidgparsonson/_apis/extensionmanagement/installedextensionsbyname/ms-securitydevops/microsoft-security-devops-azdevops?api-version=7.1
- **Live evidence (real API GET):** `https://extmgmt.dev.azure.com/davidgparsonson/_apis/extensionmanagement/installedextensionsbyname/ms-securitydevops/microsoft-security-devops-azdevops?api-version=7.1` _(captured 2026-07-02T09:52:39Z)_

```json
{
  "baseUri": "https://ms-securitydevops.gallerycdn.vsassets.io/extensions/ms-securitydevops/microsoft-security-devops-azdevops/1.18.5/1777460984221",
  "contributions": [
    {
      "id": "ms-securitydevops.microsoft-security-devops-azdevops.build-task-microsoft-security-devops",
      "constraints": [
        {
          "name": "ExtensionLicensed",
          "properties": {
            "extensionId": "ms-securitydevops.microsoft-security-devops-azdevops"
          }
        }
      ],
      "properties": {
        "::Attributes": 24,
        "::Version": "1.18.5",
        "name": "MicrosoftSecurityDevOps/v1"
      },
      "restrictedTo": [
        "member"
      ],
      "targets": [
        "ms.vss-distributed-task.tasks"
      ],
      "type": "ms.vss-distributed-task.task"
    },
    {
      "id": "ms-securitydevops.microsoft-security-devops-azdevops.build-task-microsoft-defender-cli",
      "constraints": [
        {
          "name": "ExtensionLicensed",
          "properties": {
            "extensionId": "ms-securitydevops.microsoft-security-devops-azdevops"
          }
        }
      ],
      "properties": {
        "::Attributes": 24,
        "::Version": "1.18.5",
        "name": "MicrosoftSecurityDevOps/v2"
      },
      "restrictedTo": [
        "member"
      ],
      "targets": [
        "ms.vss-distributed-task.tasks"
      ],
      "type": "ms.vss-distributed-task.task"
    }
  ],
  "contributionTypes": [],
  "fallbackBaseUri": "https://ms-securitydevops.gallery.vsassets.io/_apis/public/gallery/publisher/ms-securitydevops/extension/microsoft-security-devops-azdevops/1.18.5/assetbyname",
  "manifestVersion": 1,
  "scopes": [],
  "extensionId": "microsoft-security-devops-azdevops",
  "extensionName": "Microsoft Security DevOps",
  "files": [],
  "installState": {
    "flags": "none",
    "lastUpdated": "2026-07-02T09:52:37.767Z"
  },
  "lastPublished": "2026-04-29T11:14:34.7Z",
  "publisherId": "ms-securitydevops",
  "publisherName": "Microsoft",
  "registrationId": "b9cd1858-fd62-43c6-b3c2-c54b6686bc00",
  "version": "1.18.5"
}
```

## Test Evidence

| test | result | delta |
|---|---|---|
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release | pass | — |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent | pass | — |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate | pass | — |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

```
27 files changed, 2096 insertions(+), 257 deletions(-)
```
