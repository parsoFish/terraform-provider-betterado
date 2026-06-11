# Data Source: betterado_release_definition_revision

Looks up a specific historical revision of a release definition by its revision number and returns the raw JSON payload.

Use this data source to retrieve the exact JSON snapshot of a release pipeline at a given revision, enabling auditing, diffing, and restoring past configurations.

## Example Usage

```hcl
data "betterado_release_definition_revision" "snapshot" {
  project_id            = "00000000-0000-0000-0000-000000000000"
  release_definition_id = 42
  revision              = 3
}

output "revision_snapshot" {
  value = data.betterado_release_definition_revision.snapshot.json_content
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The GUID of the Azure DevOps project that owns the release definition.
- `release_definition_id` - (Required) The numeric ID of the release definition whose revision should be fetched.
- `revision` - (Required) The revision number to retrieve. Revision numbers start at 1 and increment with each save.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `json_content` - The raw JSON payload returned by the Azure DevOps Release Management API for the specified revision. This is an opaque string suitable for archiving, diffing with `jsondiff`, or feeding into `templatefile`.

## Import

This data source does not support import — it is read-only.
