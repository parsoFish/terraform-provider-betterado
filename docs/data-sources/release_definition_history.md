# Data Source: betterado_release_definition_history

Returns the full revision history (audit trail) of a release definition.

Use this data source to retrieve every change event recorded against a release pipeline, including who changed it, when, what type of change was made, and any associated comment. Useful for compliance, auditing, and change-management workflows.

## Example Usage

```hcl
data "betterado_release_definition_history" "audit" {
  project_id            = "00000000-0000-0000-0000-000000000000"
  release_definition_id = 42
}

output "history" {
  value = data.betterado_release_definition_history.audit.revisions
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The GUID of the Azure DevOps project that owns the release definition.
- `release_definition_id` - (Required) The numeric ID of the release definition whose history should be fetched.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `revisions` - A list of revision objects representing the change history. Each object contains:
  - `revision` - The sequential revision number (integer, starting at 1).
  - `changed_by` - The display name of the identity that made the change.
  - `changed_date` - The UTC timestamp of when the change was recorded (RFC 3339 format).
  - `change_type` - The type of change (e.g. `add`, `update`, `delete`).
  - `comment` - An optional free-text comment supplied by the author at save time. May be empty.

## Import

This data source does not support import — it is read-only.
