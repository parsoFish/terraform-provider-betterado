package taskagent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// DataTaskGroup returns the schema and Read for the betterado_task_group data source.
func DataTaskGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataTaskGroupRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"friendly_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"icon_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_name_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"runs_on": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"version": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"major": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"minor": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"patch": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"is_test": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"help_markdown": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"options": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"visible_rule": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"properties": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"aliases": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"task": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_definition_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"always_run": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"continue_on_error": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"condition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timeout_in_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"retry_count_on_task_failure": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"inputs": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"environment": {
							Type:     schema.TypeMap,
							Computed: true,
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

func dataTaskGroupRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	tgID, err := uuid.Parse(d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("invalid task group id %q: %+v", d.Get("id").(string), err)
	}

	taskGroups, err := clients.TaskAgentClient.GetTaskGroups(clients.Ctx, taskagent.GetTaskGroupsArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return fmt.Errorf("task group (ID: %s) not found in project %s", tgID, projectID)
		}
		return fmt.Errorf("reading task group (ID: %s): %+v", tgID, err)
	}

	if taskGroups == nil || len(*taskGroups) == 0 {
		return fmt.Errorf("task group (ID: %s) not found in project %s", tgID, projectID)
	}

	tg := (*taskGroups)[0]
	d.SetId(tg.Id.String())
	flattenTaskGroup(d, &tg)
	return nil
}
