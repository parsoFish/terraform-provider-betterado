# Extension Gap Matrix — `betterado_extension`

This document maps every field of the ADO SDK `InstalledExtension` struct
(vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement/models.go)
to its Terraform schema status.

## Field coverage

| ADO SDK Field | Type | Terraform Attribute | Status | Writable? | Notes |
|---|---|---|---|---|---|
| `ExtensionId` | `*string` | `extension_id` | **covered** | no (ForceNew) | Required; forces replacement on change |
| `PublisherId` | `*string` | `publisher_id` | **covered** | no (ForceNew) | Required; forces replacement on change |
| `Version` | `*string` | `version` | **covered** | yes | Optional+Computed; may be set to pin a specific version |
| `ExtensionName` | `*string` | `extension_name` | **covered** | no | Computed; display name from ADO Marketplace |
| `PublisherName` | `*string` | `publisher_name` | **covered** | no | Computed; display name from ADO Marketplace |
| `Scopes` | `*[]string` | `scope` | **covered** | no | Computed list; extension-declared OAuth scopes |
| `InstallState.Flags` | `*ExtensionStateFlags` | `disabled` | **covered** | yes | Optional+Computed; true when Flags contains "disabled" |
| `BaseUri` | `*string` | — | **gap-open** | no | Extension contribution base URI; not useful in TF |
| `Constraints` | `*[]ContributionConstraint` | — | **gap-open** | no | Contribution constraints; read-only; not TF-relevant |
| `Contributions` | `*[]Contribution` | — | **gap-open** | no | List of contributions; read-only manifest data |
| `ContributionTypes` | `*[]ContributionType` | — | **gap-open** | no | Contribution types; read-only manifest data |
| `Demands` | `*[]string` | — | **gap-open** | no | Explicit demands; read-only |
| `EventCallbacks` | `*ExtensionEventCallbackCollection` | — | **gap-open** | no | Event callbacks; read-only |
| `FallbackBaseUri` | `*string` | — | **gap-open** | no | Fallback base URI; not TF-relevant |
| `Language` | `*string` | — | **gap-open** | no | Culture name set by gallery; not TF-relevant |
| `Licensing` | `*ExtensionLicensing` | — | **gap-open** | no | Licensing config; read-only |
| `ManifestVersion` | `*float64` | — | **gap-open** | no | Manifest format version; read-only |
| `RestrictedTo` | `*[]string` | — | **gap-open** | no | Default user claims for visibility; read-only |
| `ServiceInstanceType` | `*uuid.UUID` | — | **gap-open** | no | Required VSTS service instance; read-only |
| `Files` | `*[]gallery.ExtensionFile` | — | **gap-open** | no | Files bundled with extension; read-only |
| `Flags` | `*ExtensionFlags` | — | **gap-open** | no | Extension-level flags (e.g. BuiltIn); read-only |
| `LastPublished` | `*azuredevops.Time` | — | **gap-open** | no | Last gallery publish date; read-only, not TF-relevant |
| `RegistrationId` | `*uuid.UUID` | — | **gap-open** | no | Unique extension registration UUID; not TF-relevant |

## Writable gap resolution

All writable attributes of the `InstalledExtension` struct that are relevant
to Terraform management have been covered:

| Gap | Resolution |
|---|---|
| `disabled` (via `InstallState.Flags`) | **Resolved** — covered as `disabled` Optional+Computed bool |
| `version` | **Resolved** — covered as `version` Optional+Computed string; enables pinning |
| Other `InstallState` flags (e.g. `VersionCheckError`, `Warning`) | **gap-deferred** — these flags are set by the ADO platform in error/warning states and cannot be controlled by a Terraform write; deferring since the ADO API does not expose them as writable. |

All remaining gap-open fields are either read-only ADO manifest metadata
(not useful to manage in Terraform) or internal platform fields. No further
Terraform attributes are needed; all writable gaps are either resolved in the
schema or explicitly deferred above.
