package taskagent

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ResourceTaskGroup returns the schema and CRUD for betterado_task_group
func ResourceTaskGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTaskGroupCreate,
		ReadContext:   resourceTaskGroupRead,
		UpdateContext: resourceTaskGroupUpdate,
		DeleteContext: resourceTaskGroupDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"friendly_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"category": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"author": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"icon_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"instance_name_format": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"runs_on": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"version": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"major": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"minor": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"patch": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"is_test": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"label": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "string",
						},
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"help_markdown": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"options": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"visible_rule": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"properties": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"aliases": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"task": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"task_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"task_version": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"task_definition_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "task",
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"always_run": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"continue_on_error": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"condition": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "succeeded()",
						},
						"timeout_in_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"retry_count_on_task_failure": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"inputs": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"environment": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"definition_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func resourceTaskGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	createParam := expandTaskGroupCreate(d)

	created, err := clients.TaskAgentClient.AddTaskGroup(clients.Ctx, taskagent.AddTaskGroupArgs{
		TaskGroup: createParam,
		Project:   &projectID,
	})
	if err != nil {
		return diag.Errorf("creating task group: %+v", err)
	}

	d.SetId(created.Id.String())
	return resourceTaskGroupRead(ctx, d, m)
}

func resourceTaskGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	tgID, err := uuid.Parse(d.Id())
	if err != nil {
		return diag.Errorf("invalid task group ID %q: %+v", d.Id(), err)
	}

	taskGroups, err := clients.TaskAgentClient.GetTaskGroups(clients.Ctx, taskagent.GetTaskGroupsArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("reading task group (ID: %s): %+v", d.Id(), err)
	}

	if taskGroups == nil || len(*taskGroups) == 0 {
		d.SetId("")
		return nil
	}

	tg := (*taskGroups)[0]
	flattenTaskGroup(d, &tg)
	return nil
}

func resourceTaskGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	tgID, err := uuid.Parse(d.Id())
	if err != nil {
		return diag.Errorf("invalid task group ID %q: %+v", d.Id(), err)
	}

	updateParam := expandTaskGroupUpdate(d)
	updateParam.Id = &tgID

	if v, ok := d.GetOk("revision"); ok {
		rev := v.(int)
		updateParam.Revision = &rev
	}

	_, err = clients.TaskAgentClient.UpdateTaskGroup(clients.Ctx, taskagent.UpdateTaskGroupArgs{
		TaskGroup:   updateParam,
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		return diag.Errorf("updating task group (ID: %s): %+v", d.Id(), err)
	}

	return resourceTaskGroupRead(ctx, d, m)
}

func resourceTaskGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	tgID, err := uuid.Parse(d.Id())
	if err != nil {
		return diag.Errorf("invalid task group ID %q: %+v", d.Id(), err)
	}

	err = clients.TaskAgentClient.DeleteTaskGroup(clients.Ctx, taskagent.DeleteTaskGroupArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		return diag.Errorf("deleting task group (ID: %s): %+v", d.Id(), err)
	}

	return nil
}

// ── Expand ────────────────────────────────────────────────────────────────────

func expandTaskGroupCreate(d *schema.ResourceData) *taskagent.TaskGroupCreateParameter {
	param := &taskagent.TaskGroupCreateParameter{
		Name:         converter.String(d.Get("name").(string)),
		FriendlyName: converter.String(d.Get("friendly_name").(string)),
		Category:     converter.String(d.Get("category").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		param.Description = converter.String(v.(string))
	}
	if v, ok := d.GetOk("author"); ok {
		param.Author = converter.String(v.(string))
	}
	if v, ok := d.GetOk("icon_url"); ok {
		param.IconUrl = converter.String(v.(string))
	}
	if v, ok := d.GetOk("instance_name_format"); ok {
		param.InstanceNameFormat = converter.String(v.(string))
	}
	if v, ok := d.GetOk("runs_on"); ok {
		runsOn := expandStringList(v.([]interface{}))
		param.RunsOn = &runsOn
	}
	if v, ok := d.GetOk("version"); ok {
		versionList := v.([]interface{})
		if len(versionList) > 0 {
			param.Version = expandTaskVersion(versionList[0].(map[string]interface{}))
		}
	}
	if v, ok := d.GetOk("input"); ok {
		inputs := expandTaskInputs(v.([]interface{}))
		param.Inputs = &inputs
	}
	if v, ok := d.GetOk("task"); ok {
		tasks := expandTaskGroupSteps(v.([]interface{}))
		param.Tasks = &tasks
	}

	return param
}

func expandTaskGroupUpdate(d *schema.ResourceData) *taskagent.TaskGroupUpdateParameter {
	param := &taskagent.TaskGroupUpdateParameter{
		Name:         converter.String(d.Get("name").(string)),
		FriendlyName: converter.String(d.Get("friendly_name").(string)),
		Category:     converter.String(d.Get("category").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		param.Description = converter.String(v.(string))
	}
	if v, ok := d.GetOk("author"); ok {
		param.Author = converter.String(v.(string))
	}
	if v, ok := d.GetOk("icon_url"); ok {
		param.IconUrl = converter.String(v.(string))
	}
	if v, ok := d.GetOk("instance_name_format"); ok {
		param.InstanceNameFormat = converter.String(v.(string))
	}
	if v, ok := d.GetOk("runs_on"); ok {
		runsOn := expandStringList(v.([]interface{}))
		param.RunsOn = &runsOn
	}
	if v, ok := d.GetOk("version"); ok {
		versionList := v.([]interface{})
		if len(versionList) > 0 {
			param.Version = expandTaskVersion(versionList[0].(map[string]interface{}))
		}
	}
	if v, ok := d.GetOk("input"); ok {
		inputs := expandTaskInputs(v.([]interface{}))
		param.Inputs = &inputs
	}
	if v, ok := d.GetOk("task"); ok {
		tasks := expandTaskGroupSteps(v.([]interface{}))
		param.Tasks = &tasks
	}

	return param
}

func expandTaskVersion(m map[string]interface{}) *taskagent.TaskVersion {
	return &taskagent.TaskVersion{
		Major:  converter.Int(m["major"].(int)),
		Minor:  converter.Int(m["minor"].(int)),
		Patch:  converter.Int(m["patch"].(int)),
		IsTest: converter.Bool(m["is_test"].(bool)),
	}
}

func expandTaskInputs(input []interface{}) []taskagent.TaskInputDefinition {
	inputs := make([]taskagent.TaskInputDefinition, len(input))
	for i, raw := range input {
		m := raw.(map[string]interface{})
		inp := taskagent.TaskInputDefinition{
			Name:         converter.String(m["name"].(string)),
			Label:        converter.String(m["label"].(string)),
			Type:         converter.String(m["type"].(string)),
			DefaultValue: converter.String(m["default_value"].(string)),
			Required:     converter.Bool(m["required"].(bool)),
			HelpMarkDown: converter.String(m["help_markdown"].(string)),
			GroupName:    converter.String(m["group_name"].(string)),
		}
		if opts, ok := m["options"].(map[string]interface{}); ok && len(opts) > 0 {
			optsMap := make(map[string]string, len(opts))
			for k, v := range opts {
				optsMap[k] = v.(string)
			}
			inp.Options = &optsMap
		}
		if vr, ok := m["visible_rule"].(string); ok && vr != "" {
			inp.VisibleRule = converter.String(vr)
		}
		if props, ok := m["properties"].(map[string]interface{}); ok && len(props) > 0 {
			propsMap := make(map[string]string, len(props))
			for k, v := range props {
				propsMap[k] = v.(string)
			}
			inp.Properties = &propsMap
		}
		if aliases, ok := m["aliases"].([]interface{}); ok && len(aliases) > 0 {
			aliasStrs := make([]string, len(aliases))
			for i, a := range aliases {
				aliasStrs[i] = a.(string)
			}
			inp.Aliases = &aliasStrs
		}
		inputs[i] = inp
	}
	return inputs
}

func expandTaskGroupSteps(input []interface{}) []taskagent.TaskGroupStep {
	steps := make([]taskagent.TaskGroupStep, len(input))
	for i, raw := range input {
		m := raw.(map[string]interface{})
		taskID, _ := uuid.Parse(m["task_id"].(string)) //nolint:errcheck

		step := taskagent.TaskGroupStep{
			DisplayName:     converter.String(m["display_name"].(string)),
			Enabled:         converter.Bool(m["enabled"].(bool)),
			AlwaysRun:       converter.Bool(m["always_run"].(bool)),
			ContinueOnError: converter.Bool(m["continue_on_error"].(bool)),
			Condition:       converter.String(m["condition"].(string)),
			Task: &taskagent.TaskDefinitionReference{
				Id:             &taskID,
				VersionSpec:    converter.String(m["task_version"].(string)),
				DefinitionType: converter.String(m["task_definition_type"].(string)),
			},
		}

		if v := m["timeout_in_minutes"].(int); v > 0 {
			step.TimeoutInMinutes = converter.Int(v)
		}
		if v := m["retry_count_on_task_failure"].(int); v > 0 {
			step.RetryCountOnTaskFailure = converter.Int(v)
		}
		if inputs, ok := m["inputs"].(map[string]interface{}); ok && len(inputs) > 0 {
			inputsMap := make(map[string]string, len(inputs))
			for k, v := range inputs {
				inputsMap[k] = v.(string)
			}
			step.Inputs = &inputsMap
		}
		if envVars, ok := m["environment"].(map[string]interface{}); ok && len(envVars) > 0 {
			envMap := make(map[string]string, len(envVars))
			for k, v := range envVars {
				envMap[k] = v.(string)
			}
			step.Environment = &envMap
		}

		steps[i] = step
	}
	return steps
}

func expandStringList(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

// ── Flatten ───────────────────────────────────────────────────────────────────

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
