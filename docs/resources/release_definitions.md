# Data Source: betterado_release_definitions

Use this data source to list release definitions in a project.

## Example Usage

```hcl
data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_release_definitions" "all" {
  project_id = data.betterado_project.example.id
}

data "betterado_release_definitions" "in_folder" {
  project_id = data.betterado_project.example.id
  path       = "\\MyFolder"
}

output "definition_names" {
  value = [for d in data.betterado_release_definitions.all.release_definitions : d.name]
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The ID of the Azure DevOps project (UUID).
* `path` - (Optional) Filter by folder path (e.g., `\\MyFolder`). If omitted, all definitions in the project are returned.
* `name` - (Optional) Filter by name using a search text match. If omitted, no name filter is applied.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `release_definitions` - A list of release definitions matching the filters. Each element contains:
  * `id` - The numeric ID of the release definition (as a string).
  * `name` - The name of the release definition.
  * `path` - The folder path the definition belongs to.
