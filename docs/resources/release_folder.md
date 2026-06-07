# betterado_release_folder

Manages a classic release folder in Azure DevOps.

This resource allows you to create, update, and delete folders used to organise classic release definitions.

## Example Usage

```hcl
resource "betterado_project" "example" {
  name               = "MyProject"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "example" {
  project_id  = betterado_project.example.id
  path        = "\\MyFolder"
  description = "Folder for production release pipelines"
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The ID of the Azure DevOps project in which the folder exists.
- `path` - (Required) The path of the release folder, starting with `\`. For example, `\MyFolder`.
- `description` - (Optional) A human-readable description of the folder.

## Attributes Reference

In addition to the above, the following attributes are exported:

- `id` - The ID of the release folder (same as `path`).

## Import

Release folders can be imported using the project ID and folder path separated by `/`:

```
terraform import betterado_release_folder.example 00000000-0000-0000-0000-000000000000/\\MyFolder
```

---

## Data Source

Use the `betterado_release_folder` data source to read an existing release folder.

### Example Usage

```hcl
data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_release_folder" "example" {
  project_id = data.betterado_project.example.id
  path       = "\\MyFolder"
}

output "folder_description" {
  value = data.betterado_release_folder.example.description
}
```

### Argument Reference

- `project_id` - (Required) The ID of the Azure DevOps project.
- `path` - (Required) The path of the release folder to look up, starting with `\`.

### Attributes Reference

- `id` - The ID of the release folder (same as `path`).
- `description` - The description of the release folder.
