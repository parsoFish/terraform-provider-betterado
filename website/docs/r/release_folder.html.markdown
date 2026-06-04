---
layout: "betterado"
page_title: "BetterADO: betterado_release_folder"
description: |-
  Manages a Release Folder in Azure DevOps.
sidebar_current: "docs-betterado-resource-release-folder"
---

# betterado_release_folder

Manages a Release Folder in Azure DevOps Classic Release (RM) pipelines.

## Example Usage

```hcl
resource "betterado_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "example" {
  project_id  = betterado_project.example.id
  path        = "\\ExampleFolder"
  description = "Folder for production release pipelines"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The ID of the Azure DevOps project in which the release folder will be created. Changing this forces a new resource to be created.

* `path` - (Required) The path of the release folder, starting with a backslash (e.g. `\MyFolder` or `\Parent\Child`). Changing this forces a new resource to be created.

* `description` - (Optional) A description for the release folder. Defaults to `""`.

## Import

Release folders can be imported using the `project_id/path`, e.g.

```shell
terraform import betterado_release_folder.example 00000000-0000-0000-0000-000000000000/\ExampleFolder
```
