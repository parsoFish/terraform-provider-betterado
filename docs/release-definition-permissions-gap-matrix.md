# Release Definition Permissions Gap Matrix

> **TF resource:** `betterado_release_definition_permissions`
>   (`azuredevops/internal/service/permissions/resource_release_definition_permissions.go`)
> **Security namespace:** `ReleaseManagement` (SDK alias `ReleaseManagement2`)
> **Namespace ID:** `c788c23e-1b46-4162-8f5e-d7585343b5de`
> **ADO host:** `https://vsrm.dev.azure.com/{org}` (engine); ACLs via `https://dev.azure.com/{org}`
> **Corrected 2026-06-19** against the LIVE namespace
> (`GET _apis/securitynamespaces`). An earlier revision listed fabricated action
> names (`ViewReleasePipeline`, `QueueRelease`, `ManageReleasesSettings`) and wrong
> bit values — those names are NOT accepted by ADO.

---

## 1. Namespace Overview

The `ReleaseManagement` namespace governs access to classic release definitions (the
VSRM engine, distinct from YAML pipelines). Each permission is a bit in the `Allow`/`Deny`
mask of an `AccessControlEntry`. The `permissions = {}` map on the resource is keyed by the
namespace's **action `name`** (below); values are `Allow` / `Deny` / `NotSet`
(case-insensitive, normalised to lowercase in state). Keys are validated by ADO at apply.

---

## 2. Permission Bit Table (authoritative — live namespace)

| Action name (HCL key) | Bit | Writable | Display name | Live-proven |
|---|---|---|---|---|
| `ViewReleaseDefinition` | 1 | Yes | View release pipeline | covered |
| `EditReleaseDefinition` | 2 | Yes | Edit release pipeline | covered |
| `DeleteReleaseDefinition` | 4 | Yes | Delete release pipeline | covered |
| `ManageReleaseApprovers` | 8 | Yes | Manage release approvers | covered |
| `ManageReleases` | 16 | Yes | Manage releases | covered |
| `ViewReleases` | 32 | Yes | View releases | covered |
| `CreateReleases` | 64 | Yes | Create releases | covered |
| `EditReleaseEnvironment` | 128 | Yes | Edit release stage | covered |
| `DeleteReleaseEnvironment` | 256 | Yes | Delete release stage | covered |
| `AdministerReleasePermissions` | 512 | Yes | Administer release permissions | covered |
| `DeleteReleases` | 1024 | Yes | Delete releases | covered |
| `ManageDeployments` | 2048 | Yes | Manage deployments | covered |
| `ManageReleaseSettings` | 4096 | Yes | Manage release settings | covered |
| `ManageTaskHubExtension` | 8192 | Yes | Manage TaskHub Extension | extension-scoped — not used at definition level |

> **All 13 release-definition bits are Writable.** `ManageTaskHubExtension` is present in
> the namespace but is an extension/org-level concern, not a definition ACE.
> "standing demo" = applied + idempotent in `demo/standing/` (`permissions.tf`).

---

## 3. Token Format

| Scope | Token format | Example |
|---|---|---|
| Project-level (all definitions) | `{projectId}` | `6ddb680c-…` |
| Definition-level (one definition) | `{projectId}/{releaseDefinitionId}` | `6ddb680c-…/2` |

**Source:** live ADO, org `davidgparsonson` (mirrored in
`resource_release_definition_permissions.go`). The `{projectId}/{definitionId}` format
matches the `Build` namespace; the longer `ReleaseManagement2/Project/…` form was not observed.

---

## 4. Example HCL — every writable definition key

```hcl
resource "betterado_release_definition_permissions" "example" {
  project_id            = betterado_project.example.id
  principal             = data.betterado_group.readers.id
  release_definition_id = betterado_release_definition.example.id

  permissions = {
    ViewReleaseDefinition        = "Allow"
    EditReleaseDefinition        = "Allow"
    DeleteReleaseDefinition      = "Deny"
    ManageReleaseApprovers       = "Allow"
    ManageReleases               = "Allow"
    ViewReleases                 = "Allow"
    CreateReleases               = "Allow"
    EditReleaseEnvironment       = "Allow"
    DeleteReleaseEnvironment     = "Deny"
    AdministerReleasePermissions = "Deny"
    DeleteReleases               = "Deny"
    ManageDeployments            = "Allow"
    ManageReleaseSettings        = "Allow"
  }
}
```

---

## 5. Implementation Reference

- **Resource:** `azuredevops/internal/service/permissions/resource_release_definition_permissions.go`
- **Namespace ID:** `securityhelper.SecurityNamespaceIDValues.ReleaseManagement2` (`c788c23e-…`)
- **Token:** `releaseDefinitionTokenFormat = "%s/%d"` (projectID, releaseDefinitionID)
- **Acceptance tests (live):** `azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go`
  — `TestAccReleaseDefinitionPermissions_SetPermissions` / `_UpdatePermissions`
- **Standing demo (all 13 keys, live + idempotent):** `demo/standing/permissions.tf`

---

## 6. Coverage Summary

The acceptance suite asserts 4 keys (`ViewReleases`, `EditReleaseEnvironment`,
`DeleteReleases`, `CreateReleases`) through full apply→read→update cycles. The standing
demo (`demo/standing/`) applies **all 13** writable definition keys and re-plans clean
(idempotent), proving each round-trips. A `TestAccReleaseDefinitionPermissions_AllBits`
asserting the remaining nine in the suite proper remains a **Recommend**.
