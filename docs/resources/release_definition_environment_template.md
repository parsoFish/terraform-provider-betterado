# betterado_release_definition_environment_template

Manages an Azure DevOps Release Definition Environment Template.

Environment templates are immutable once created — they cannot be updated in-place.
To change a template, delete it and re-create it (all schema attributes are `ForceNew`).

## Example Usage

```hcl
resource "betterado_release_definition_environment_template" "example" {
  project_id  = "00000000-0000-0000-0000-000000000001"
  name        = "My Deploy Template"
  description = "Template for deploying to production."
  category    = "Deploy"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required, ForceNew) The ID of the Azure DevOps project.
* `name` - (Required, ForceNew) The name of the environment template.
* `description` - (Optional, ForceNew) A description of the environment template.
* `category` - (Optional, ForceNew) The category of the environment template (e.g. `Deploy`, `Build`).
* `environment` - (Optional, ForceNew) JSON-encoded `ReleaseDefinitionEnvironment` used to seed the template.
* `icon_task_id` - (Optional, ForceNew) The UUID of the task whose icon is displayed for this template.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `can_delete` - Whether the template can be deleted.
* `icon_uri` - The icon URI of the template.

## Import

Release Definition Environment Templates can be imported using the `project_id` and the template `id`, separated by a `/`:

```
terraform import betterado_release_definition_environment_template.example 00000000-0000-0000-0000-000000000001/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
```
