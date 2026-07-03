# Dashboard Gap Matrix

> Generated as part of WI-1 (INIT-2026-07-01-migrate-framework-dashboard-extension).
> Covers all fields of the ADO SDK `Dashboard` struct in
> `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/dashboard/models.go`.

## Legend

| Status | Meaning |
|--------|---------|
| **mapped** | Exposed in the Terraform schema; round-trips correctly. |
| **missing** | Not in schema; could be added in a follow-up. |
| **server-computed** | Read-only; set by the ADO service; deferred (see rationale). |
| **writable-deferred** | Writable by the API but explicitly deferred (see rationale). |

---

## Field Matrix

| ADO SDK Field | JSON key | Type | Schema Status | Writable | Notes |
|---------------|----------|------|--------------|----------|-------|
| `Links` | `_links` | `interface{}` | missing | no | Internal HAL links; read-only navigation metadata. Deferred — no user value. |
| `DashboardScope` | `dashboardScope` | `*DashboardScope` | missing | no | Derived from whether `team_id` is set. The API infers scope; not user-settable directly. Deferred. |
| `Description` | `description` | `*string` | **mapped** | yes | Exposed as `description` (Optional, Computed). |
| `ETag` | `eTag` | `*string` | missing | no | Server-managed concurrency token. Deferred — not meaningful to Terraform users. |
| `GroupId` | `groupId` | `*uuid.UUID` | **mapped** (read) | no | Returned as `team_id` on read when team-scoped. Not user-settable directly. |
| `Id` | `id` | `*uuid.UUID` | **mapped** | no | Exposed as `id` (Computed). Set by service at creation time. |
| `LastAccessedDate` | `lastAccessedDate` | `*azuredevops.Time` | missing | no | Server-computed timestamp. Deferred — no user value. |
| `ModifiedBy` | `modifiedBy` | `*uuid.UUID` | missing | no | Server-computed identity. Deferred — no user value. |
| `ModifiedDate` | `modifiedDate` | `*azuredevops.Time` | missing | no | Server-computed timestamp. Deferred — no user value. |
| `Name` | `name` | `*string` | **mapped** | yes | Exposed as `name` (Required). |
| `OwnerId` | `ownerId` | `*uuid.UUID` | **mapped** | no | Exposed as `owner_id` (Computed). Set by service at creation time. |
| `Position` | `position` | `*int` | missing | yes | **writable-deferred**: Position of the dashboard within a dashboard group. Deferred — ordering is typically managed manually in the ADO UI. Will be added in a follow-up WI if demanded. |
| `RefreshInterval` | `refreshInterval` | `*int` | **mapped** | yes | Exposed as `refresh_interval` (Optional, Computed). Valid values: `0` (disabled), `5` (minutes). |
| `Url` | `url` | `*string` | missing | no | Server-provided resource URL. Deferred — not useful in Terraform config. |
| `Widgets` | `widgets` | `*[]Widget` | missing | yes | **writable-deferred**: Dashboard widget configuration. Widgets are a complex nested structure. Deferred to a future WI (`betterado_dashboard_widget` or an embedded block) due to the significant schema complexity and the high risk of breaking idempotency with server-side widget ordering. |

---

## Summary of writable gaps

| Field | Rationale for deferral |
|-------|----------------------|
| `Position` | Dashboard ordering is typically managed via the ADO portal UI; the API field is rarely used via IaC. Low demand, low priority. |
| `Widgets` | Widgets have a deep nested schema (positions, sizes, widget-type-specific settings, artifact/query references). Implementing them correctly requires a dedicated WI to avoid idempotency issues from server-side ordering. |

All other fields are either server-computed (read-only) or mapped. The current schema provides full Create/Read/Update/Delete coverage for the user-writable fields that matter in IaC workflows.
