# Extension Gap Matrix — `betterado_extension`

This document maps every field of the ADO SDK `InstalledExtension` struct
(vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement/models.go)
to its Terraform schema status.

## Field mapping

| ADO SDK Field | Type | Terraform Attribute | Status | Writable? | Notes |
|---|---|---|---|---|---|
| `ExtensionId` | `*string` | `extension_id` | **mapped** | no (ForceNew) | Required; forces replacement on change |
| `PublisherId` | `*string` | `publisher_id` | **mapped** | no (ForceNew) | Required; forces replacement on change |
| `Version` | `*string` | `version` | **mapped** | yes | Optional+Computed; may be set to pin a specific version |
| `ExtensionName` | `*string` | `extension_name` | **mapped** | no | Computed; display name from ADO Marketplace |
| `PublisherName` | `*string` | `publisher_name` | **mapped** | no | Computed; display name from ADO Marketplace |
| `Scopes` | `*[]string` | `scope` | **mapped** | no | Computed list; extension-declared OAuth scopes |
| `InstallState.Flags` | `*ExtensionStateFlags` | `disabled` | **mapped** | yes | Optional+Computed; true when Flags contains "disabled" |
| `BaseUri` | `*string` | — | **missing** | no | Extension contribution base URI; not useful in TF |
| `Constraints` | `*[]ContributionConstraint` | — | **missing** | no | Contribution constraints; read-only; not TF-relevant |
| `Contributions` | `*[]Contribution` | — | **missing** | no | List of contributions; read-only manifest data |
| `ContributionTypes` | `*[]ContributionType` | — | **missing** | no | Contribution types; read-only manifest data |
| `Demands` | `*[]string` | — | **missing** | no | Explicit demands; read-only |
| `EventCallbacks` | `*ExtensionEventCallbackCollection` | — | **missing** | no | Event callbacks; read-only |
| `FallbackBaseUri` | `*string` | — | **missing** | no | Fallback base URI; not TF-relevant |
| `Language` | `*string` | — | **missing** | no | Culture name set by gallery; not TF-relevant |
| `Licensing` | `*ExtensionLicensing` | — | **missing** | no | Licensing config; read-only |
| `ManifestVersion` | `*float64` | — | **missing** | no | Manifest format version; read-only |
| `RestrictedTo` | `*[]string` | — | **missing** | no | Default user claims for visibility; read-only |
| `ServiceInstanceType` | `*uuid.UUID` | — | **missing** | no | Required VSTS service instance; read-only |
| `Files` | `*[]gallery.ExtensionFile` | — | **missing** | no | Files bundled with extension; read-only |
| `Flags` | `*ExtensionFlags` | — | **missing** | no | Extension-level flags (e.g. BuiltIn); read-only |
| `LastPublished` | `*azuredevops.Time` | — | **missing** | no | Last gallery publish date; read-only, not TF-relevant |
| `RegistrationId` | `*uuid.UUID` | — | **missing** | no | Unique extension registration UUID; not TF-relevant |

## Writable gap resolution

All writable attributes of the `InstalledExtension` struct that are relevant
to Terraform management have been mapped:

| Gap | Resolution |
|---|---|
| `disabled` (via `InstallState.Flags`) | **Resolved** — mapped as `disabled` Optional+Computed bool |
| `version` | **Resolved** — mapped as `version` Optional+Computed string; enables pinning |
| Other `InstallState` flags (e.g. `VersionCheckError`, `Warning`) | **Deferred** — these flags are set by the ADO platform in error/warning states and cannot be controlled by a Terraform write; deferring since the ADO API does not expose them as writable. |

All remaining missing fields are either read-only ADO manifest metadata
(not useful to manage in Terraform) or internal platform fields. No further
Terraform attributes are needed; all writable gaps are either resolved in the
schema or explicitly deferred above.
