# betterado_release_folder

Manages a release folder in Azure DevOps.

## Example Usage

### Basic example

```hcl
resource "betterado_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "example" {
  project_id = betterado_project.example.id
  path       = "\\MyFolder"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required, ForceNew) The ID of the Azure DevOps project. Changing this forces a new resource to be created.

* `path` - (Required, ForceNew) The full path of the release folder, starting with a backslash (e.g. `\\Parent\\Child`). Changing this forces a new resource to be created.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the release folder in the format `{projectId}/{path}`.

* `name` - The last segment of the folder path (computed from `path`).

## Import

Release folders can be imported using the format `{projectId}/{path}`:

```shell
terraform import betterado_release_folder.example 00000000-0000-0000-0000-000000000000/\\MyFolder
```
