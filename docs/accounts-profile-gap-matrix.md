# Accounts & Profile API Gap Matrix

> **Initiative:** INIT-2026-07-01-new-api-accounts-profile
> **Updated:** 2026-07-04
> **Scope:** ADO Accounts API (`_apis/accounts`) + ADO Profile API (`_apis/profile/profiles/{id}`)
> **Conclusion:** Both surfaces are **read-only data sources** — neither API supports write/create/delete operations via the Terraform provider pattern.

---

## Accounts API — `GET https://app.vssps.visualstudio.com/_apis/accounts`

Response object: `Account`

| Field | Go Type | Terraform Attribute | Status | Notes |
|-------|---------|---------------------|--------|-------|
| `accountId` | `*uuid.UUID` | `account_id` | ✅ Implemented | Exposed as `string` (UUID). |
| `accountName` | `*string` | `account_name` | ✅ Implemented | The organization slug (e.g. `myorg`). |
| `accountUri` | `*string` | `account_uri` | ✅ Implemented | Full URI (e.g. `https://dev.azure.com/myorg`). |
| `accountType` | `*AccountType` | `account_type` | ✅ Implemented | Enum: `personal` \| `organization`. |
| `organizationName` | `*string` | `organization_name` | ✅ Implemented | Display name of the organization. |
| `accountOwner` | `*uuid.UUID` | — | ⚠️ Gap | Owner UUID. Not surfaced — out of scope for initial cut; low consumer demand. |
| `accountStatus` | `*AccountStatus` | — | ⚠️ Gap | Enum: `none`, `enabled`, `disabled`, `deleted`, `moved`. Useful but deferred — no consumer use-case identified in this initiative. |
| `createdBy` | `*uuid.UUID` | — | 🚫 Out of scope | Audit field; not useful in IaC context. |
| `createdDate` | `*azuredevops.Time` | — | 🚫 Out of scope | Audit field. |
| `lastUpdatedBy` | `*uuid.UUID` | — | 🚫 Out of scope | Audit field. |
| `lastUpdatedDate` | `*azuredevops.Time` | — | 🚫 Out of scope | Audit field. |
| `namespaceId` | `*uuid.UUID` | — | 🚫 Out of scope | Internal ADO namespace; no IaC use-case. |
| `newCollectionId` | `*uuid.UUID` | — | 🚫 Out of scope | Migration artifact; not relevant to consumers. |
| `hasMoved` | `*bool` | — | 🚫 Out of scope | Migration artifact. |
| `properties` | `interface{}` | — | 🚫 Out of scope | Unstructured extension bag; schema unpredictable. |
| `statusReason` | `*string` | — | 🚫 Out of scope | Free-text reason for account status; redundant with `accountStatus`. |

**Query parameters supported by the data source:**

| Parameter | Terraform Attribute | Status | Notes |
|-----------|---------------------|--------|-------|
| `memberId` | `member_id` | ✅ Implemented | Optional UUID to filter by member subject descriptor. |
| `ownerId` | — | ⚠️ Gap | Owner filter. Deferred — `member_id` covers the primary PAT-based use case. |
| `properties` | — | 🚫 Out of scope | Server-side expansion of extra properties; not useful without surfacing `properties` field. |

---

## Profile API — `GET https://app.vssps.visualstudio.com/_apis/profile/profiles/{id}?details=true`

Response object: `Profile`

| Field | Go Type | Terraform Attribute | Status | Notes |
|-------|---------|---------------------|--------|-------|
| `id` | `*uuid.UUID` | — | ⚠️ Gap | Profile UUID (same as subject descriptor UUID). Profile data source is **not implemented in this initiative** — deferred to a follow-on WI. |
| `coreRevision` | `*int` | — | ⚠️ Gap | Max revision of core attributes. Deferred. |
| `revision` | `*int` | — | ⚠️ Gap | Max revision across all attributes. Deferred. |
| `profileState` | `*ProfileState` | — | ⚠️ Gap | Enum: `custom`, `customReadOnly`, `readOnly`. Deferred. |
| `timeStamp` | `*azuredevops.Time` | — | 🚫 Out of scope | Last-modified timestamp; audit field, low IaC value. |
| `coreAttributes` | `*map[string]CoreProfileAttribute` | — | ⚠️ Gap | Key/value bag of core profile attrs (e.g. `DisplayName`, `EmailAddress`, `PublicAlias`). Deferred — useful but complex to model; follow-on WI should surface `display_name`, `email`, `public_alias` as top-level computed strings. |
| `applicationContainer` | `*AttributesContainer` | — | 🚫 Out of scope | Application-specific attributes (e.g. IDE preferences). Not useful in IaC context. |

**Core attribute keys of interest (within `coreAttributes`):**

| Key | Description | Status |
|-----|-------------|--------|
| `DisplayName` | Human-readable name | ⚠️ Gap — deferred |
| `EmailAddress` | Primary email | ⚠️ Gap — deferred |
| `PublicAlias` | URL-safe alias | ⚠️ Gap — deferred |
| `Avatar` | Binary image data | 🚫 Out of scope |

---

## Summary

| Surface | Terraform Resource Type | This Initiative | Notes |
|---------|------------------------|-----------------|-------|
| Accounts API (`_apis/accounts`) | `data "betterado_accounts"` | ✅ **Implemented** | Lists accounts; 5 core fields implemented; minor fields deferred. |
| Profile API (`_apis/profile/profiles/{id}`) | `data "betterado_profile"` | ⚠️ **Gap — not yet implemented** | Profile data source is out of scope for WI-1; follow-on WI should implement `betterado_profile` exposing `display_name`, `email_address`, `public_alias`. |

### Conclusions

1. **Both surfaces are read-only.** Neither the Accounts nor Profile API supports create/update/delete. Both are appropriately modelled as Terraform **data sources**, not resources.
2. **`betterado_accounts` is complete for the primary use-case** (listing accounts accessible to a PAT). The five implemented fields (`account_id`, `account_name`, `account_uri`, `account_type`, `organization_name`) cover all fields a Terraform consumer would reference in expressions.
3. **`betterado_profile` is deferred.** The Profile API's most useful fields (`display_name`, `email_address`, `public_alias`) live inside the `coreAttributes` dynamic map, requiring careful modelling. A follow-on work item should implement `data "betterado_profile"` with those top-level computed string attributes.
4. **Gaps tracked:** `account_owner`, `account_status`, `owner_id` query param, and all Profile API fields are marked as gaps above. None are required for the stated initiative goal.
