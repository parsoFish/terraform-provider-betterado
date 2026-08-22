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
| `accountId` | `*uuid.UUID` | `account_id` | covered | Exposed as `string` (UUID). |
| `accountName` | `*string` | `account_name` | covered | The organization slug (e.g. `myorg`). |
| `accountUri` | `*string` | `account_uri` | covered | Full URI (e.g. `https://dev.azure.com/myorg`). |
| `accountType` | `*AccountType` | `account_type` | covered | Enum: `personal` \| `organization`. |
| `organizationName` | `*string` | `organization_name` | covered | Display name of the organization. |
| `accountOwner` | `*uuid.UUID` | — | gap-open | Owner UUID. Not surfaced — out of scope for initial cut; low consumer demand. |
| `accountStatus` | `*AccountStatus` | — | gap-open | Enum: `none`, `enabled`, `disabled`, `deleted`, `moved`. Useful but deferred — no consumer use-case identified in this initiative. |
| `createdBy` | `*uuid.UUID` | — | out-of-scope | Audit field; not useful in IaC context. |
| `createdDate` | `*azuredevops.Time` | — | out-of-scope | Audit field. |
| `lastUpdatedBy` | `*uuid.UUID` | — | out-of-scope | Audit field. |
| `lastUpdatedDate` | `*azuredevops.Time` | — | out-of-scope | Audit field. |
| `namespaceId` | `*uuid.UUID` | — | out-of-scope | Internal ADO namespace; no IaC use-case. |
| `newCollectionId` | `*uuid.UUID` | — | out-of-scope | Migration artifact; not relevant to consumers. |
| `hasMoved` | `*bool` | — | out-of-scope | Migration artifact. |
| `properties` | `interface{}` | — | out-of-scope | Unstructured extension bag; schema unpredictable. |
| `statusReason` | `*string` | — | out-of-scope | Free-text reason for account status; redundant with `accountStatus`. |

**Query parameters available in the data source:**

| Parameter | Terraform Attribute | Status | Notes |
|-----------|---------------------|--------|-------|
| `memberId` | `member_id` | covered | Optional UUID to filter by member subject descriptor. |
| `ownerId` | — | gap-open | Owner filter. Deferred — `member_id` covers the primary PAT-based use case. |
| `properties` | — | out-of-scope | Server-side expansion of extra properties; not useful without surfacing `properties` field. |

---

## Profile API — `GET https://app.vssps.visualstudio.com/_apis/profile/profiles/{id}?details=true`

Response object: `Profile`

| Field | Go Type | Terraform Attribute | Status | Notes |
|-------|---------|---------------------|--------|-------|
| `id` | `*uuid.UUID` | — | gap-open | Profile UUID (same as subject descriptor UUID). Profile data source is **not yet in schema** — deferred to a follow-on WI. |
| `coreRevision` | `*int` | — | gap-open | Max revision of core attributes. Deferred. |
| `revision` | `*int` | — | gap-open | Max revision across all attributes. Deferred. |
| `profileState` | `*ProfileState` | — | gap-open | Enum: `custom`, `customReadOnly`, `readOnly`. Deferred. |
| `timeStamp` | `*azuredevops.Time` | — | out-of-scope | Last-modified timestamp; audit field, low IaC value. |
| `coreAttributes` | `*map[string]CoreProfileAttribute` | — | gap-open | Key/value bag of core profile attrs (e.g. `DisplayName`, `EmailAddress`, `PublicAlias`). Deferred — useful but complex to model; follow-on WI should surface `display_name`, `email`, `public_alias` as top-level computed strings. |
| `applicationContainer` | `*AttributesContainer` | — | out-of-scope | Application-specific attributes (e.g. IDE preferences). Not useful in IaC context. |

**Core attribute keys of interest (within `coreAttributes`):**

| Key | Description | Status |
|-----|-------------|--------|
| `DisplayName` | Human-readable name | gap-open — deferred |
| `EmailAddress` | Primary email | gap-open — deferred |
| `PublicAlias` | URL-safe alias | gap-open — deferred |
| `Avatar` | Binary image data | out-of-scope |

---

## Summary

| Surface | Terraform Resource Type | This Initiative | Notes |
|---------|------------------------|-----------------|-------|
| Accounts API (`_apis/accounts`) | `data "betterado_accounts"` | covered | Lists accounts; 5 core fields covered; minor fields deferred. |
| Profile API (`_apis/profile/profiles/{id}`) | `data "betterado_profile"` | gap-open | Profile data source is out of scope for WI-1; follow-on WI should implement `betterado_profile` exposing `display_name`, `email_address`, `public_alias`. |

### Conclusions

1. **Both surfaces are read-only.** Neither the Accounts nor Profile API supports create/update/delete. Both are appropriately modelled as Terraform **data sources**, not resources.
2. **`betterado_accounts` is complete for the primary use-case** (listing accounts accessible to a PAT). The five covered fields (`account_id`, `account_name`, `account_uri`, `account_type`, `organization_name`) cover all fields a Terraform consumer would reference in expressions.
3. **`betterado_profile` is deferred.** The Profile API's most useful fields (`display_name`, `email_address`, `public_alias`) live inside the `coreAttributes` dynamic map, requiring careful modelling. A follow-on work item should implement `data "betterado_profile"` with those top-level computed string attributes.
4. **Gaps tracked:** `account_owner`, `account_status`, `owner_id` query param, and all Profile API fields are marked as gap-open above. None are required for the stated initiative goal.
