package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/tfhelper"
)

// ResourceReleaseDefinition schema and implementation for release definition resource
func ResourceReleaseDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReleaseDefinitionCreate,
		ReadContext:   resourceReleaseDefinitionRead,
		UpdateContext: resourceReleaseDefinitionUpdate,
		DeleteContext: resourceReleaseDefinitionDelete,
		Importer:      tfhelper.ImportProjectQualifiedResource(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"project_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "\\",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"release_name_format": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Release-$(rev:r)",
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"variable": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"is_secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"allow_override": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"variable_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environment": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"rank": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"owner": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsUUID,
						},
						"variable": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
									"is_secret": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"allow_override": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"variable_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"condition": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"condition_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"event", "environmentState", "artifact", "undefined"}, false),
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
								},
							},
						},
						"pre_deploy_approval": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: approvalSchema(),
							},
						},
						"post_deploy_approval": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: approvalSchema(),
							},
						},
						"deploy_phase": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"rank": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"phase_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "agentBasedDeployment",
										ValidateFunc: validation.StringInSlice([]string{"agentBasedDeployment", "runOnServer", "machineGroupBasedDeployment"}, false),
									},
									"workflow_task": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: workflowTaskSchema(),
										},
									},
								},
							},
						},
						"retention_policy": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days_to_keep": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      30,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"releases_to_keep": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      3,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"retain_build": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
								},
							},
						},
					},
				},
			},
			"artifact": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Build", "Jenkins", "GitHub", "Nuget", "Team Build (external)", "ExternalTFSBuild", "Git", "TFVC", "ExternalTfsXamlBuild"}, false),
						},
						"is_primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"definition_reference": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func approvalSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"approver": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.IsUUID,
					},
					"is_automated": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"rank": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      1,
						ValidateFunc: validation.IntAtLeast(1),
					},
				},
			},
		},
	}
}

func workflowTaskSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"task_id": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		"version": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1.*",
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
		"definition_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "task",
		},
		"inputs": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

// CRUD Operations

func resourceReleaseDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	releaseDefinition, projectID, err := expandReleaseDefinition(d)
	if err != nil {
		return diag.Errorf("expanding release definition: %+v", err)
	}

	createdDef, err := clients.ReleaseClient.CreateReleaseDefinition(clients.Ctx, releaseapi.CreateReleaseDefinitionArgs{
		ReleaseDefinition: releaseDefinition,
		Project:           &projectID,
	})
	if err != nil {
		return diag.Errorf("creating release definition: %+v", err)
	}

	d.SetId(strconv.Itoa(*createdDef.Id))
	return resourceReleaseDefinitionRead(ctx, d, m)
}

func resourceReleaseDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID, definitionID, err := tfhelper.ParseProjectIDAndResourceID(d)
	if err != nil {
		return diag.FromErr(err)
	}

	def, err := clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &definitionID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("reading release definition (ID: %d): %+v", definitionID, err)
	}

	flattenReleaseDefinition(d, def, projectID)
	return nil
}

func resourceReleaseDefinitionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	releaseDefinition, projectID, err := expandReleaseDefinition(d)
	if err != nil {
		return diag.Errorf("expanding release definition for update: %+v", err)
	}

	defID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid definition ID: %s", d.Id())
	}
	releaseDefinition.Id = &defID

	// Revision is required for updates (optimistic concurrency)
	if v, ok := d.GetOk("revision"); ok {
		rev := v.(int)
		releaseDefinition.Revision = &rev
	}

	_, err = clients.ReleaseClient.UpdateReleaseDefinition(clients.Ctx, releaseapi.UpdateReleaseDefinitionArgs{
		ReleaseDefinition: releaseDefinition,
		Project:           &projectID,
	})
	if err != nil {
		return diag.Errorf("updating release definition (ID: %d): %+v", defID, err)
	}

	return resourceReleaseDefinitionRead(ctx, d, m)
}

func resourceReleaseDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)

	defID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid definition ID: %s", d.Id())
	}

	err = clients.ReleaseClient.DeleteReleaseDefinition(clients.Ctx, releaseapi.DeleteReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if err != nil {
		return diag.Errorf("deleting release definition (ID: %d): %+v", defID, err)
	}

	return nil
}

// Expand functions: Terraform state → API objects

func expandReleaseDefinition(d *schema.ResourceData) (*releaseapi.ReleaseDefinition, string, error) {
	projectID := d.Get("project_id").(string)

	def := &releaseapi.ReleaseDefinition{
		Name:              converter.String(d.Get("name").(string)),
		Path:              converter.String(d.Get("path").(string)),
		Description:       converter.String(d.Get("description").(string)),
		ReleaseNameFormat: converter.String(d.Get("release_name_format").(string)),
	}

	// Variables
	if v, ok := d.GetOk("variable"); ok {
		def.Variables = expandVariables(v.(*schema.Set).List())
	}

	// Variable groups
	if v, ok := d.GetOk("variable_groups"); ok {
		vgs := expandVariableGroups(v.([]interface{}))
		def.VariableGroups = &vgs
	}

	// Tags
	if v, ok := d.GetOk("tags"); ok {
		tags := expandTags(v.(*schema.Set).List())
		def.Tags = &tags
	}

	// Environments
	if v, ok := d.GetOk("environment"); ok {
		envs, err := expandEnvironments(v.([]interface{}))
		if err != nil {
			return nil, "", err
		}
		def.Environments = &envs
	}

	// Artifacts
	if v, ok := d.GetOk("artifact"); ok {
		artifacts := expandArtifacts(v.([]interface{}))
		def.Artifacts = &artifacts
	}

	return def, projectID, nil
}

func expandVariables(input []interface{}) *map[string]releaseapi.ConfigurationVariableValue {
	variables := make(map[string]releaseapi.ConfigurationVariableValue)
	for _, v := range input {
		varMap := v.(map[string]interface{})
		name := varMap["name"].(string)
		variables[name] = releaseapi.ConfigurationVariableValue{
			Value:         converter.String(varMap["value"].(string)),
			IsSecret:      converter.Bool(varMap["is_secret"].(bool)),
			AllowOverride: converter.Bool(varMap["allow_override"].(bool)),
		}
	}
	return &variables
}

func expandVariableGroups(input []interface{}) []int {
	result := make([]int, len(input))
	for i, v := range input {
		result[i] = v.(int)
	}
	return result
}

func expandTags(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

func expandEnvironments(input []interface{}) ([]releaseapi.ReleaseDefinitionEnvironment, error) {
	envs := make([]releaseapi.ReleaseDefinitionEnvironment, len(input))
	for i, v := range input {
		envMap := v.(map[string]interface{})
		env := releaseapi.ReleaseDefinitionEnvironment{
			Name: converter.String(envMap["name"].(string)),
			Rank: converter.Int(envMap["rank"].(int)),
		}

		// Owner
		if ownerID, ok := envMap["owner"].(string); ok && ownerID != "" {
			env.Owner = &webapi.IdentityRef{
				Id: converter.String(ownerID),
			}
		}

		// Conditions
		if conditions, ok := envMap["condition"].([]interface{}); ok && len(conditions) > 0 {
			conds := expandConditions(conditions)
			env.Conditions = &conds
		}

		// Pre-deploy approvals
		if preApprovals, ok := envMap["pre_deploy_approval"].([]interface{}); ok && len(preApprovals) > 0 {
			env.PreDeployApprovals = expandApprovals(preApprovals)
		}

		// Post-deploy approvals
		if postApprovals, ok := envMap["post_deploy_approval"].([]interface{}); ok && len(postApprovals) > 0 {
			env.PostDeployApprovals = expandApprovals(postApprovals)
		}

		// Deploy phases
		if phases, ok := envMap["deploy_phase"].([]interface{}); ok && len(phases) > 0 {
			expandedPhases, err := expandDeployPhases(phases)
			if err != nil {
				return nil, err
			}
			env.DeployPhases = &expandedPhases
		}

		// Retention policy
		if retention, ok := envMap["retention_policy"].([]interface{}); ok && len(retention) > 0 {
			env.RetentionPolicy = expandRetentionPolicy(retention)
		}

		// Variables
		if vars, ok := envMap["variable"]; ok {
			env.Variables = expandVariables(vars.(*schema.Set).List())
		}

		// Variable groups
		if vgs, ok := envMap["variable_groups"].([]interface{}); ok && len(vgs) > 0 {
			groups := expandVariableGroups(vgs)
			env.VariableGroups = &groups
		}

		envs[i] = env
	}
	return envs, nil
}

func expandConditions(input []interface{}) []releaseapi.Condition {
	conditions := make([]releaseapi.Condition, len(input))
	for i, v := range input {
		condMap := v.(map[string]interface{})
		condType := releaseapi.ConditionType(condMap["condition_type"].(string))
		conditions[i] = releaseapi.Condition{
			Name:          converter.String(condMap["name"].(string)),
			ConditionType: &condType,
			Value:         converter.String(condMap["value"].(string)),
		}
	}
	return conditions
}

func expandApprovals(input []interface{}) *releaseapi.ReleaseDefinitionApprovals {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	approvalMap := input[0].(map[string]interface{})
	result := &releaseapi.ReleaseDefinitionApprovals{}

	if approvers, ok := approvalMap["approver"].([]interface{}); ok && len(approvers) > 0 {
		steps := make([]releaseapi.ReleaseDefinitionApprovalStep, len(approvers))
		for i, a := range approvers {
			aMap := a.(map[string]interface{})
			steps[i] = releaseapi.ReleaseDefinitionApprovalStep{
				Approver: &webapi.IdentityRef{
					Id: converter.String(aMap["id"].(string)),
				},
				IsAutomated: converter.Bool(aMap["is_automated"].(bool)),
				Rank:        converter.Int(aMap["rank"].(int)),
			}
		}
		result.Approvals = &steps
	}

	return result
}

func expandDeployPhases(input []interface{}) ([]interface{}, error) {
	phases := make([]interface{}, len(input))
	for i, v := range input {
		phaseMap := v.(map[string]interface{})
		phaseType := releaseapi.DeployPhaseTypes(phaseMap["phase_type"].(string))
		phase := releaseapi.AgentBasedDeployPhase{
			Name:      converter.String(phaseMap["name"].(string)),
			Rank:      converter.Int(phaseMap["rank"].(int)),
			PhaseType: &phaseType,
		}

		if tasks, ok := phaseMap["workflow_task"].([]interface{}); ok && len(tasks) > 0 {
			wfTasks := expandWorkflowTasks(tasks)
			phase.WorkflowTasks = &wfTasks
		}

		phases[i] = phase
	}
	return phases, nil
}

func expandWorkflowTasks(input []interface{}) []releaseapi.WorkflowTask {
	tasks := make([]releaseapi.WorkflowTask, len(input))
	for i, v := range input {
		taskMap := v.(map[string]interface{})
		taskID, _ := uuid.Parse(taskMap["task_id"].(string))
		task := releaseapi.WorkflowTask{
			Name:            converter.String(taskMap["name"].(string)),
			TaskId:          &taskID,
			Version:         converter.String(taskMap["version"].(string)),
			Enabled:         converter.Bool(taskMap["enabled"].(bool)),
			AlwaysRun:       converter.Bool(taskMap["always_run"].(bool)),
			ContinueOnError: converter.Bool(taskMap["continue_on_error"].(bool)),
			Condition:       converter.String(taskMap["condition"].(string)),
			DefinitionType:  converter.String(taskMap["definition_type"].(string)),
		}

		if inputs, ok := taskMap["inputs"].(map[string]interface{}); ok && len(inputs) > 0 {
			inputMap := make(map[string]string)
			for k, val := range inputs {
				inputMap[k] = val.(string)
			}
			task.Inputs = &inputMap
		}

		tasks[i] = task
	}
	return tasks
}

func expandRetentionPolicy(input []interface{}) *releaseapi.EnvironmentRetentionPolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	retMap := input[0].(map[string]interface{})
	return &releaseapi.EnvironmentRetentionPolicy{
		DaysToKeep:     converter.Int(retMap["days_to_keep"].(int)),
		ReleasesToKeep: converter.Int(retMap["releases_to_keep"].(int)),
		RetainBuild:    converter.Bool(retMap["retain_build"].(bool)),
	}
}

func expandArtifacts(input []interface{}) []releaseapi.Artifact {
	artifacts := make([]releaseapi.Artifact, len(input))
	for i, v := range input {
		artMap := v.(map[string]interface{})
		artifact := releaseapi.Artifact{
			Alias:     converter.String(artMap["alias"].(string)),
			Type:      converter.String(artMap["type"].(string)),
			IsPrimary: converter.Bool(artMap["is_primary"].(bool)),
		}

		if defRef, ok := artMap["definition_reference"].(map[string]interface{}); ok {
			refs := make(map[string]releaseapi.ArtifactSourceReference)
			for k, val := range defRef {
				refs[k] = releaseapi.ArtifactSourceReference{
					Id: converter.String(val.(string)),
				}
			}
			artifact.DefinitionReference = &refs
		}

		artifacts[i] = artifact
	}
	return artifacts
}

// Flatten functions: API response → Terraform state

func flattenReleaseDefinition(d *schema.ResourceData, def *releaseapi.ReleaseDefinition, projectID string) {
	d.Set("name", def.Name)
	d.Set("project_id", projectID)
	d.Set("path", def.Path)
	d.Set("description", def.Description)
	d.Set("release_name_format", def.ReleaseNameFormat)
	d.Set("revision", def.Revision)

	if def.Variables != nil {
		d.Set("variable", flattenVariables(def.Variables))
	}

	if def.VariableGroups != nil {
		d.Set("variable_groups", *def.VariableGroups)
	}

	if def.Tags != nil {
		d.Set("tags", *def.Tags)
	}

	if def.Environments != nil {
		d.Set("environment", flattenEnvironments(def.Environments))
	}

	if def.Artifacts != nil {
		d.Set("artifact", flattenArtifacts(def.Artifacts))
	}
}

func flattenVariables(variables *map[string]releaseapi.ConfigurationVariableValue) []interface{} {
	if variables == nil {
		return nil
	}
	result := make([]interface{}, 0, len(*variables))
	for name, v := range *variables {
		varMap := map[string]interface{}{
			"name":           name,
			"is_secret":      false,
			"allow_override": false,
			"value":          "",
		}
		if v.IsSecret != nil {
			varMap["is_secret"] = *v.IsSecret
		}
		if v.AllowOverride != nil {
			varMap["allow_override"] = *v.AllowOverride
		}
		// Secret values are not returned by the API
		if v.Value != nil && !*v.IsSecret {
			varMap["value"] = *v.Value
		}
		result = append(result, varMap)
	}
	return result
}

func flattenEnvironments(envs *[]releaseapi.ReleaseDefinitionEnvironment) []interface{} {
	if envs == nil {
		return nil
	}
	result := make([]interface{}, len(*envs))
	for i, env := range *envs {
		envMap := map[string]interface{}{
			"name": env.Name,
			"rank": env.Rank,
		}

		if env.Id != nil {
			envMap["id"] = *env.Id
		}

		if env.Owner != nil && env.Owner.Id != nil {
			envMap["owner"] = *env.Owner.Id
		}

		if env.Conditions != nil {
			envMap["condition"] = flattenConditions(env.Conditions)
		}

		if env.PreDeployApprovals != nil {
			envMap["pre_deploy_approval"] = flattenApprovals(env.PreDeployApprovals)
		}

		if env.PostDeployApprovals != nil {
			envMap["post_deploy_approval"] = flattenApprovals(env.PostDeployApprovals)
		}

		if env.DeployPhases != nil {
			envMap["deploy_phase"] = flattenDeployPhases(env.DeployPhases)
		}

		if env.RetentionPolicy != nil {
			envMap["retention_policy"] = flattenRetentionPolicy(env.RetentionPolicy)
		}

		if env.Variables != nil {
			envMap["variable"] = flattenVariables(env.Variables)
		}

		if env.VariableGroups != nil {
			envMap["variable_groups"] = *env.VariableGroups
		}

		result[i] = envMap
	}
	return result
}

func flattenConditions(conditions *[]releaseapi.Condition) []interface{} {
	if conditions == nil {
		return nil
	}
	result := make([]interface{}, len(*conditions))
	for i, c := range *conditions {
		condMap := map[string]interface{}{
			"name":           "",
			"condition_type": "",
			"value":          "",
		}
		if c.Name != nil {
			condMap["name"] = *c.Name
		}
		if c.ConditionType != nil {
			condMap["condition_type"] = string(*c.ConditionType)
		}
		if c.Value != nil {
			condMap["value"] = *c.Value
		}
		result[i] = condMap
	}
	return result
}

func flattenApprovals(approvals *releaseapi.ReleaseDefinitionApprovals) []interface{} {
	if approvals == nil {
		return nil
	}
	approvalMap := map[string]interface{}{}

	if approvals.Approvals != nil {
		approvers := make([]interface{}, len(*approvals.Approvals))
		for i, step := range *approvals.Approvals {
			aMap := map[string]interface{}{
				"is_automated": false,
				"rank":         1,
				"id":           "",
			}
			if step.Approver != nil && step.Approver.Id != nil {
				aMap["id"] = *step.Approver.Id
			}
			if step.IsAutomated != nil {
				aMap["is_automated"] = *step.IsAutomated
			}
			if step.Rank != nil {
				aMap["rank"] = *step.Rank
			}
			approvers[i] = aMap
		}
		approvalMap["approver"] = approvers
	}

	return []interface{}{approvalMap}
}

func flattenDeployPhases(phases *[]interface{}) []interface{} {
	if phases == nil {
		return nil
	}
	result := make([]interface{}, len(*phases))
	for i, p := range *phases {
		phaseMap := map[string]interface{}{
			"name":       "",
			"rank":       1,
			"phase_type": "agentBasedDeployment",
		}

		// The API returns deploy phases as map[string]interface{} (since the SDK type is []interface{})
		if pMap, ok := p.(map[string]interface{}); ok {
			if name, ok := pMap["name"].(string); ok {
				phaseMap["name"] = name
			}
			if rank, ok := pMap["rank"].(json.Number); ok {
				if r, err := rank.Int64(); err == nil {
					phaseMap["rank"] = int(r)
				}
			} else if rank, ok := pMap["rank"].(float64); ok {
				phaseMap["rank"] = int(rank)
			}
			if pt, ok := pMap["phaseType"].(string); ok {
				phaseMap["phase_type"] = pt
			}
			if tasks, ok := pMap["workflowTasks"].([]interface{}); ok {
				phaseMap["workflow_task"] = flattenWorkflowTasksFromMap(tasks)
			}
		}

		result[i] = phaseMap
	}
	return result
}

func flattenWorkflowTasksFromMap(tasks []interface{}) []interface{} {
	result := make([]interface{}, len(tasks))
	for i, t := range tasks {
		taskMap := map[string]interface{}{
			"name":              "",
			"task_id":           "",
			"version":           "1.*",
			"enabled":           true,
			"always_run":        false,
			"continue_on_error": false,
			"condition":         "succeeded()",
			"definition_type":   "task",
			"inputs":            map[string]interface{}{},
		}

		if tMap, ok := t.(map[string]interface{}); ok {
			if name, ok := tMap["name"].(string); ok {
				taskMap["name"] = name
			}
			if taskID, ok := tMap["taskId"].(string); ok {
				taskMap["task_id"] = taskID
			}
			if version, ok := tMap["version"].(string); ok {
				taskMap["version"] = version
			}
			if enabled, ok := tMap["enabled"].(bool); ok {
				taskMap["enabled"] = enabled
			}
			if alwaysRun, ok := tMap["alwaysRun"].(bool); ok {
				taskMap["always_run"] = alwaysRun
			}
			if continueOnError, ok := tMap["continueOnError"].(bool); ok {
				taskMap["continue_on_error"] = continueOnError
			}
			if condition, ok := tMap["condition"].(string); ok {
				taskMap["condition"] = condition
			}
			if defType, ok := tMap["definitionType"].(string); ok {
				taskMap["definition_type"] = defType
			}
			if inputs, ok := tMap["inputs"].(map[string]interface{}); ok {
				strInputs := make(map[string]interface{})
				for k, v := range inputs {
					strInputs[k] = fmt.Sprintf("%v", v)
				}
				taskMap["inputs"] = strInputs
			}
		}

		result[i] = taskMap
	}
	return result
}

func flattenRetentionPolicy(policy *releaseapi.EnvironmentRetentionPolicy) []interface{} {
	if policy == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"days_to_keep":     policy.DaysToKeep,
			"releases_to_keep": policy.ReleasesToKeep,
			"retain_build":     policy.RetainBuild,
		},
	}
}

func flattenArtifacts(artifacts *[]releaseapi.Artifact) []interface{} {
	if artifacts == nil {
		return nil
	}
	result := make([]interface{}, len(*artifacts))
	for i, a := range *artifacts {
		artMap := map[string]interface{}{
			"alias":      a.Alias,
			"type":       a.Type,
			"is_primary": a.IsPrimary,
		}

		if a.DefinitionReference != nil {
			refs := make(map[string]interface{})
			for k, v := range *a.DefinitionReference {
				if v.Id != nil {
					refs[k] = *v.Id
				}
			}
			artMap["definition_reference"] = refs
		}

		result[i] = artMap
	}
	return result
}
