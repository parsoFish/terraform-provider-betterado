# Extension Gap Matrix — `betterado_extension`

This document maps every field of the ADO SDK `InstalledExtension` struct
(vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement/models.go)
to its Terraform schema status.

## Legend

| Status | Meaning |
|--------|---------|
| `covered` | Exposed in the Terraform schema; round-trips correctly. |
| `gap-open` | Not in schema; could be added in a follow-up. |
| `out-of-scope` | Read-only or internal; not user-configurable via Terraform. |
| `gap-deferred` | Intentionally skipped; reason documented below. |

## Field mapping

| ADO SDK Field | Type | Terraform Attribute | Status | Writable? | Notes |
|---|---|---|---|---|---|
| `ExtensionId` | `*string` | `extension_id` | `covered` | no (ForceNew) | Required; forces replacement on change |
| `PublisherId` | `*string` | `publisher_id` | `covered` | no (ForceNew) | Required; forces replacement on change |
| `Version` | `*string` | `version` | `covered` | yes | Optional+Computed; may be set to pin a specific version |
| `ExtensionName` | `*string` | `extension_name` | `covered` | no | Computed; display name from ADO Marketplace |
| `PublisherName` | `*string` | `publisher_name` | `covered` | no | Computed; display name from ADO Marketplace |
| `Scopes` | `*[]string` | `scope` | `covered` | no | Computed list; extension-declared OAuth scopes |
| `InstallState.Flags` | `*ExtensionStateFlags` | `disabled` | `covered` | yes | Optional+Computed; true when Flags contains "disabled" |
| `BaseUri` | `*string` | — | `out-of-scope` | no | Extension contribution base URI; not useful in TF |
| `Constraints` | `*[]ContributionConstraint` | — | `out-of-scope` | no | Contribution constraints; server-assigned; not TF-relevant |
| `Contributions` | `*[]Contribution` | — | `out-of-scope` | no | List of contributions; server-assigned manifest data |
| `ContributionTypes` | `*[]ContributionType` | — | `out-of-scope` | no | Contribution types; server-assigned manifest data |
| `Demands` | `*[]string` | — | `out-of-scope` | no | Explicit demands; server-assigned |
| `EventCallbacks` | `*ExtensionEventCallbackCollection` | — | `out-of-scope` | no | Event callbacks; server-assigned |
| `FallbackBaseUri` | `*string` | — | `out-of-scope` | no | Fallback base URI; not TF-relevant |
| `Language` | `*string` | — | `out-of-scope` | no | Culture name set by gallery; not TF-relevant |
| `Licensing` | `*ExtensionLicensing` | — | `out-of-scope` | no | Licensing config; server-assigned |
| `ManifestVersion` | `*float64` | — | `out-of-scope` | no | Manifest format version; server-assigned |
| `RestrictedTo` | `*[]string` | — | `out-of-scope` | no | Default user claims for visibility; server-assigned |
| `ServiceInstanceType` | `*uuid.UUID` | — | `out-of-scope` | no | Required VSTS service instance; server-assigned |
| `Files` | `*[]gallery.ExtensionFile` | — | `out-of-scope` | no | Files bundled with extension; server-assigned |
| `Flags` | `*ExtensionFlags` | — | `out-of-scope` | no | Extension-level flags (e.g. BuiltIn); server-assigned |
| `LastPublished` | `*azuredevops.Time` | — | `out-of-scope` | no | Last gallery publish date; server-assigned, not TF-relevant |
| `RegistrationId` | `*uuid.UUID` | — | `out-of-scope` | no | Unique extension registration UUID; not TF-relevant |

## Writable gap resolution

All writable attributes of the `InstalledExtension` struct that are relevant
to Terraform management have been mapped:

| Gap | Resolution |
|---|---|
| `disabled` (via `InstallState.Flags`) | Resolved — surfaced as `disabled` Optional+Computed bool |
| `version` | Resolved — surfaced as `version` Optional+Computed string; enables pinning |
| Other `InstallState` flags (e.g. `VersionCheckError`, `Warning`) | `gap-deferred` — these flags are set by the ADO platform in error/warning states and cannot be controlled by a Terraform write; deferring since the ADO API does not expose them as writable. |

All remaining fields are either server-assigned ADO manifest metadata
(not useful to manage in Terraform) or internal platform fields. No further
Terraform attributes are needed; all writable gaps are either resolved in the
schema or explicitly deferred above.
