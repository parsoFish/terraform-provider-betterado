# Dashboard Gap Matrix

> Generated as part of WI-1 (INIT-2026-07-01-migrate-framework-dashboard-extension).
> Covers all fields of the ADO SDK `Dashboard` struct in
> `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/dashboard/models.go`.

## Legend

| Status | Meaning |
|--------|---------|
| `covered` | Exposed in the Terraform schema; round-trips correctly. |
| `gap-open` | Not in schema; could be added in a follow-up. |
| `out-of-scope` | Read-only; set by the ADO service; not user-configurable. |
| `gap-deferred` | Writable by the API but explicitly deferred (see rationale). |

---

## Field Matrix

| ADO SDK Field | JSON key | Type | Schema Status | Writable | Notes |
|---------------|----------|------|--------------|----------|-------|
| `Links` | `_links` | `interface{}` | `gap-deferred` | no | Internal HAL links; server-assigned navigation metadata. Deferred — no user value. |
| `DashboardScope` | `dashboardScope` | `*DashboardScope` | `gap-deferred` | no | Derived from whether `team_id` is set. The API infers scope; not user-settable directly. Deferred. |
| `Description` | `description` | `*string` | `covered` | yes | Exposed as `description` (Optional, Computed). |
| `ETag` | `eTag` | `*string` | `gap-deferred` | no | Server-managed concurrency token. Deferred — not meaningful to Terraform users. |
| `GroupId` | `groupId` | `*uuid.UUID` | `covered` | no | Returned as `team_id` on read when team-scoped. Not user-settable directly. |
| `Id` | `id` | `*uuid.UUID` | `covered` | no | Exposed as `id` (Computed). Set by service at creation time. |
| `LastAccessedDate` | `lastAccessedDate` | `*azuredevops.Time` | `gap-deferred` | no | Server-computed timestamp. Deferred — no user value. |
| `ModifiedBy` | `modifiedBy` | `*uuid.UUID` | `gap-deferred` | no | Server-computed identity. Deferred — no user value. |
| `ModifiedDate` | `modifiedDate` | `*azuredevops.Time` | `gap-deferred` | no | Server-computed timestamp. Deferred — no user value. |
| `Name` | `name` | `*string` | `covered` | yes | Exposed as `name` (Required). |
| `OwnerId` | `ownerId` | `*uuid.UUID` | `covered` | no | Exposed as `owner_id` (Computed). Set by service at creation time. |
| `Position` | `position` | `*int` | `gap-deferred` | yes | Position of the dashboard within a dashboard group. Deferred — ordering is typically managed manually in the ADO UI. Will be added in a follow-up WI if demanded. |
| `RefreshInterval` | `refreshInterval` | `*int` | `covered` | yes | Exposed as `refresh_interval` (Optional, Computed). Valid values: `0` (disabled), `5` (minutes). |
| `Url` | `url` | `*string` | `gap-deferred` | no | Server-provided resource URL. Deferred — not useful in Terraform config. |
| `Widgets` | `widgets` | `*[]Widget` | `gap-deferred` | yes | Dashboard widget configuration. Widgets are a complex nested structure. Deferred to a future WI (`betterado_dashboard_widget` or an embedded block) due to the significant schema complexity and the high risk of breaking idempotency with server-side widget ordering. |

---

## Summary of deferred gaps

| Field | Rationale for deferral |
|-------|----------------------|
| `Position` | Dashboard ordering is typically managed via the ADO portal UI; the API field is rarely used via IaC. Low demand, low priority. |
| `Widgets` | Widgets have a deep nested schema (positions, sizes, widget-type-specific settings, artifact/query references). Implementing them correctly requires a dedicated WI to avoid idempotency issues from server-side ordering. |

All other fields are either server-computed (not user-configurable) or covered. The current schema provides full Create/Read/Update/Delete coverage for the user-writable fields that matter in IaC workflows.
