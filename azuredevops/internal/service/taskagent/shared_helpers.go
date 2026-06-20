package taskagent

// shared_helpers.go — flatten helpers shared between the task_group data source
// and the framework resource. Relocated from resource_task_group.go (SDK v2
// dead code) after the migration to terraform-plugin-framework.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// flattenTaskGroup writes all task group fields into a SDK v2 ResourceData.
// Called by both the betterado_task_group data source and the framework
// resource's Read path (via adapter shim).
func flattenTaskGroup(d *schema.ResourceData, tg *taskagent.TaskGroup) {
	d.Set("name", converter.ToString(tg.Name, ""))
	d.Set("friendly_name", converter.ToString(tg.FriendlyName, ""))
	d.Set("description", converter.ToString(tg.Description, ""))
	d.Set("category", converter.ToString(tg.Category, ""))
	d.Set("author", converter.ToString(tg.Author, ""))
	d.Set("icon_url", converter.ToString(tg.IconUrl, ""))
	d.Set("instance_name_format", converter.ToString(tg.InstanceNameFormat, ""))
	d.Set("definition_type", converter.ToString(tg.DefinitionType, ""))

	if tg.Revision != nil {
		d.Set("revision", *tg.Revision)
	}
	if tg.RunsOn != nil {
		d.Set("runs_on", *tg.RunsOn)
	}
	if tg.Version != nil {
		d.Set("version", flattenTaskVersion(tg.Version))
	}
	if tg.Inputs != nil {
		d.Set("input", flattenTaskInputs(tg.Inputs))
	}
	if tg.Tasks != nil {
		d.Set("task", flattenTaskGroupSteps(tg.Tasks))
	}
}

func flattenTaskVersion(v *taskagent.TaskVersion) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"major":   converter.ToInt(v.Major, 0),
			"minor":   converter.ToInt(v.Minor, 0),
			"patch":   converter.ToInt(v.Patch, 0),
			"is_test": converter.ToBool(v.IsTest, false),
		},
	}
}

func flattenTaskInputs(inputs *[]taskagent.TaskInputDefinition) []interface{} {
	if inputs == nil {
		return nil
	}
	result := make([]interface{}, len(*inputs))
	for i, inp := range *inputs {
		m := map[string]interface{}{
			"name":          converter.ToString(inp.Name, ""),
			"label":         converter.ToString(inp.Label, ""),
			"type":          converter.ToString(inp.Type, "string"),
			"default_value": converter.ToString(inp.DefaultValue, ""),
			"required":      converter.ToBool(inp.Required, false),
			"help_markdown": converter.ToString(inp.HelpMarkDown, ""),
			"group_name":    converter.ToString(inp.GroupName, ""),
		}
		if inp.Options != nil {
			m["options"] = *inp.Options
		} else {
			m["options"] = map[string]string{}
		}
		m["visible_rule"] = converter.ToString(inp.VisibleRule, "")
		if inp.Properties != nil {
			m["properties"] = *inp.Properties
		} else {
			m["properties"] = map[string]string{}
		}
		if inp.Aliases != nil {
			m["aliases"] = *inp.Aliases
		} else {
			m["aliases"] = []string{}
		}
		result[i] = m
	}
	return result
}

func flattenTaskGroupSteps(steps *[]taskagent.TaskGroupStep) []interface{} {
	if steps == nil {
		return nil
	}
	result := make([]interface{}, len(*steps))
	for i, step := range *steps {
		m := map[string]interface{}{
			"display_name":                converter.ToString(step.DisplayName, ""),
			"enabled":                     converter.ToBool(step.Enabled, true),
			"always_run":                  converter.ToBool(step.AlwaysRun, false),
			"continue_on_error":           converter.ToBool(step.ContinueOnError, false),
			"condition":                   converter.ToString(step.Condition, "succeeded()"),
			"timeout_in_minutes":          converter.ToInt(step.TimeoutInMinutes, 0),
			"retry_count_on_task_failure": converter.ToInt(step.RetryCountOnTaskFailure, 0),
			"task_id":                     "",
			"task_version":                "",
			"task_definition_type":        "task",
		}

		if step.Task != nil {
			if step.Task.Id != nil {
				m["task_id"] = step.Task.Id.String()
			}
			m["task_version"] = converter.ToString(step.Task.VersionSpec, "")
			m["task_definition_type"] = converter.ToString(step.Task.DefinitionType, "task")
		}

		if step.Inputs != nil {
			m["inputs"] = *step.Inputs
		} else {
			m["inputs"] = map[string]string{}
		}
		if step.Environment != nil {
			m["environment"] = *step.Environment
		} else {
			m["environment"] = map[string]string{}
		}

		result[i] = m
	}
	return result
}
