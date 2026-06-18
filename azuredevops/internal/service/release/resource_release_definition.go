package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/distributedtaskcommon"
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
				// Tags are not persisted by the release definitions API (silently ignored on write).
				// Computed prevents permanent plan diff since the API always returns [].
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"stages": {
				Type:       schema.TypeList,
				Required:   true,
				MinItems:   1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
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
							Type:       schema.TypeSet,
							Optional:   true,
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
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
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
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: approvalSchema(),
							},
						},
						"post_deploy_approval": {
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: approvalSchema(),
							},
						},
						"deploy_phase": {
							Type:       schema.TypeList,
							Required:   true,
							MinItems:   1,
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
									"deployment_input": {
										Type:       schema.TypeList,
										Optional:   true,
										MaxItems:   1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"queue_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  0,
												},
												"demands": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"timeout_in_minutes": {
													Type:         schema.TypeInt,
													Optional:     true,
													Default:      0,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"job_cancel_timeout_in_minutes": {
													Type:         schema.TypeInt,
													Optional:     true,
													Default:      1,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"condition": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "succeeded()",
												},
												"skip_artifacts_download": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"enable_access_token": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"agent_specification": {
													Type:     schema.TypeString,
													Optional: true,
													// e.g. "ubuntu-latest", "windows-2022", "macOS-14"
												},
												"override_inputs": {
													Type:        schema.TypeMap,
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Description: "Phase-level task input overrides (e.g. for parameterised task groups).",
												},
												"parallel_execution": {
													Type:       schema.TypeList,
													Optional:   true,
													MaxItems:   1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      "none",
																ValidateFunc: validation.StringInSlice([]string{"none", "multiConfiguration", "multiMachine"}, false),
															},
															"max_number_of_agents": {
																Type:         schema.TypeInt,
																Optional:     true,
																Default:      0,
																ValidateFunc: validation.IntAtLeast(0),
															},
															"multipliers": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"continue_on_error": {
																Type:     schema.TypeBool,
																Optional: true,
																Default:  false,
															},
														},
													},
												},
											},
										},
									},
									"workflow_task": {
										Type:       schema.TypeList,
										Optional:   true,
										Elem: &schema.Resource{
											Schema: workflowTaskSchema(),
										},
									},
								},
							},
						},
						"retention_policy": {
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
							MaxItems:   1,
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
						"environment_options": {
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"email_notification_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "OnlyOnFailure",
										ValidateFunc: validation.StringInSlice([]string{"OnlyOnFailure", "Always", "Never"}, false),
									},
									"email_recipients": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true, // API defaults to "release.environment.owner;release.creator"
									},
									"skip_artifacts_download": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"timeout_in_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"enable_access_token": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"publish_deployment_status": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"badge_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"auto_link_work_items": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"pull_request_deployment_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"execution_policy": {
							Type:       schema.TypeList,
							Optional:   true,
							Computed:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrency_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"queue_depth_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntInSlice([]int{0, 1}),
									},
								},
							},
						},
						"pre_deployment_gates": {
							Type:       schema.TypeList,
							Optional:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: deploymentGatesSchema(),
							},
						},
						"post_deployment_gates": {
							Type:       schema.TypeList,
							Optional:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: deploymentGatesSchema(),
							},
						},
						"environment_trigger": {
							Type:       schema.TypeList,
							Optional:   true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"definition_environment_id": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										// ADO automatically fills definition_environment_id with the
										// parent environment's ID when it is left unset (0). Mark as
										// Computed so Terraform uses the value ADO returns rather than
										// treating 0 vs <envID> as a perpetual drift.
									},
									"trigger_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "",
										ValidateFunc: validation.StringInSlice([]string{"undefined", "deploymentGroupRedeploy", "rollbackRedeploy"}, false),
									},
									"trigger_content": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
								},
							},
						},
						"schedule": {
							Type:       schema.TypeList,
							Optional:   true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days_to_release": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      127,
										ValidateFunc: validation.IntBetween(0, 127),
									},
									"start_hours": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 23),
									},
									"start_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 59),
									},
									"time_zone_id": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "UTC",
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"job_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"process_parameters": {
							Type:       schema.TypeList,
							Optional:   true,
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input": {
										Type:       schema.TypeList,
										Optional:   true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"default_value": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "",
												},
												"parameter_type": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "",
												},
											},
										},
									},
								},
							},
						},
						"properties": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
			"triggers": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cd_artifact_trigger": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"artifact_alias": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"branch_filter": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"include": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"exclude": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									// Build-tag filter for the artifact trigger. ADO REST 7.1 only
									// persists the `tags` list on artifactSource trigger conditions
									// (the UI's "Build tags"); the SDK's regex `tagFilter.pattern`
									// field is silently DROPPED by the service on create/update
									// (verified live 2026-06-11), so it is not exposed here.
									"tag_filter": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"tags": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"use_build_definition_branch": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"create_release_on_build_tagging": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"schedule_trigger": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// branch_filter is intentionally NOT present for schedule_trigger.
									// ADO classic schedule triggers are time-based and have no branch filter;
									// ADO does not return branchFilters for schedule triggers in the GET
									// response, so a branch_filter block would perpetually diff.
									// (cd_artifact_trigger keeps its branch_filter — that one is real.)
									"schedule_only_with_changes": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"start_hours": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 23),
									},
									"start_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 59),
									},
									"time_zone_id": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "UTC",
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"days_to_release": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      127,
										ValidateFunc: validation.IntBetween(0, 127),
									},
								},
							},
						},
						"source_repo_trigger": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alias": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"branch_filters": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
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
			Type:       schema.TypeList,
			Optional:   true,
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
		"approval_options": {
			Type:       schema.TypeList,
			Optional:   true,
			Computed:   true,
			MaxItems:   1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"required_approver_count": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"release_creator_can_be_approver": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"enforce_identity_revalidation": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"timeout_in_minutes": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"execution_order": {
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "beforeGates",
						ValidateFunc: validation.StringInSlice([]string{"beforeGates", "afterSuccessfulGates", "afterGatesAlways"}, false),
					},
					"auto_triggered_and_previous_environment_approved_can_be_skipped": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
		},
	}
}

func deploymentGatesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"gates_options": {
			Type:       schema.TypeList,
			Optional:   true,
			MaxItems:   1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"is_enabled": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"timeout": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"sampling_interval": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"stabilization_time": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"minimum_success_duration": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						ValidateFunc: validation.IntAtLeast(0),
					},
				},
			},
		},
		"gate": {
			Type:       schema.TypeList,
			Optional:   true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"task": {
						Type:       schema.TypeList,
						Required:   true,
						MinItems:   1,
						Elem: &schema.Resource{
							Schema: workflowTaskSchema(),
						},
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
		"timeout_in_minutes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Per-task timeout in minutes (0 = use the phase default).",
		},
		"retry_count_on_task_failure": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Number of times to retry the task on failure.",
		},
	}
}

// CRUD Operations

// rollbackRedeployErrorHint detects ADO's TF400898 failure, which occurs when an
// environment_trigger of type "rollbackRedeploy" is configured on an environment
// that has never had a successful deployment. This is a service-side constraint
// (not a provider bug), so the raw error is wrapped with actionable guidance
// rather than surfaced as an opaque internal error.
func rollbackRedeployErrorHint(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "TF400898") {
		return fmt.Errorf("%w\n\nhint: ADO rejects a rollbackRedeploy environment_trigger on an "+
			"environment that has never had a successful deployment. Remove the rollbackRedeploy "+
			"trigger until the stage has deployed at least once, then add it back", err)
	}
	return err
}

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
		return diag.Errorf("creating release definition: %+v", rollbackRedeployErrorHint(err))
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
		// The API returns HTTP 400 with a message about "old copy of the release pipeline" when
		// there is a revision conflict (not HTTP 409). Retry once with a fresh revision.
		if strings.Contains(err.Error(), "old copy of the release pipeline") {
			freshDef, readErr := clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
				Project:      &projectID,
				DefinitionId: &defID,
			})
			if readErr != nil {
				return diag.Errorf("re-reading release definition after revision conflict (ID: %d): %+v", defID, readErr)
			}
			releaseDefinition.Revision = freshDef.Revision
			_, err = clients.ReleaseClient.UpdateReleaseDefinition(clients.Ctx, releaseapi.UpdateReleaseDefinitionArgs{
				ReleaseDefinition: releaseDefinition,
				Project:           &projectID,
			})
		}
		if err != nil {
			return diag.Errorf("updating release definition (ID: %d): %+v", defID, rollbackRedeployErrorHint(err))
		}
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
	if v, ok := d.GetOk("stages"); ok {
		envs, err := expandStages(v.([]interface{}))
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

	// Triggers
	if v, ok := d.GetOk("triggers"); ok {
		triggers := expandTriggers(v.([]interface{}))
		def.Triggers = &triggers
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

func expandStages(input []interface{}) ([]releaseapi.ReleaseDefinitionEnvironment, error) {
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

		// Environment options
		if envOpts, ok := envMap["environment_options"].([]interface{}); ok && len(envOpts) > 0 {
			env.EnvironmentOptions = expandEnvironmentOptions(envOpts)
		}

		// Execution policy
		if execPolicy, ok := envMap["execution_policy"].([]interface{}); ok && len(execPolicy) > 0 {
			env.ExecutionPolicy = expandExecutionPolicy(execPolicy)
		}

		// Pre-deployment gates
		if preGates, ok := envMap["pre_deployment_gates"].([]interface{}); ok && len(preGates) > 0 {
			env.PreDeploymentGates = expandDeploymentGates(preGates)
		}

		// Post-deployment gates
		if postGates, ok := envMap["post_deployment_gates"].([]interface{}); ok && len(postGates) > 0 {
			env.PostDeploymentGates = expandDeploymentGates(postGates)
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

		// Environment triggers
		if triggers, ok := envMap["environment_trigger"].([]interface{}); ok && len(triggers) > 0 {
			expanded := expandEnvironmentTriggers(triggers)
			env.EnvironmentTriggers = &expanded
		}

		// Schedules
		if scheds, ok := envMap["schedule"].([]interface{}); ok && len(scheds) > 0 {
			expanded := expandEnvironmentSchedules(scheds)
			env.Schedules = &expanded
		}

		// Process parameters
		if pp, ok := envMap["process_parameters"].([]interface{}); ok && len(pp) > 0 {
			env.ProcessParameters = expandProcessParameters(pp)
		}

		// Properties
		if props, ok := envMap["properties"].(map[string]interface{}); ok && len(props) > 0 {
			env.Properties = props
		}

		envs[i] = env
	}
	return envs, nil
}

func expandEnvironmentTriggers(input []interface{}) []releaseapi.EnvironmentTrigger {
	triggers := make([]releaseapi.EnvironmentTrigger, len(input))
	for i, v := range input {
		tMap := v.(map[string]interface{})
		t := releaseapi.EnvironmentTrigger{}
		if id, ok := tMap["definition_environment_id"].(int); ok {
			t.DefinitionEnvironmentId = converter.Int(id)
		}
		if tt, ok := tMap["trigger_type"].(string); ok && tt != "" {
			ttVal := releaseapi.EnvironmentTriggerType(tt)
			t.TriggerType = &ttVal
		}
		if tc, ok := tMap["trigger_content"].(string); ok && tc != "" {
			t.TriggerContent = converter.String(tc)
		}
		triggers[i] = t
	}
	return triggers
}

func expandEnvironmentSchedules(input []interface{}) []releaseapi.ReleaseSchedule {
	scheds := make([]releaseapi.ReleaseSchedule, len(input))
	for i, v := range input {
		sMap := v.(map[string]interface{})
		s := releaseapi.ReleaseSchedule{}
		if d, ok := sMap["days_to_release"].(int); ok {
			sd := releaseapi.ScheduleDays(strconv.Itoa(d))
			s.DaysToRelease = &sd
		}
		if h, ok := sMap["start_hours"].(int); ok {
			s.StartHours = converter.Int(h)
		}
		if m, ok := sMap["start_minutes"].(int); ok {
			s.StartMinutes = converter.Int(m)
		}
		if tz, ok := sMap["time_zone_id"].(string); ok && tz != "" {
			s.TimeZoneId = converter.String(tz)
		}
		if jid, ok := sMap["job_id"].(string); ok && jid != "" {
			parsed, err := uuid.Parse(jid)
			if err == nil {
				s.JobId = &parsed
			}
		}
		scheds[i] = s
	}
	return scheds
}

func expandProcessParameters(input []interface{}) *distributedtaskcommon.ProcessParameters {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	ppMap := input[0].(map[string]interface{})
	pp := &distributedtaskcommon.ProcessParameters{}
	if rawInputs, ok := ppMap["input"].([]interface{}); ok && len(rawInputs) > 0 {
		inputs := make([]distributedtaskcommon.TaskInputDefinitionBase, len(rawInputs))
		for i, ri := range rawInputs {
			iMap := ri.(map[string]interface{})
			inp := distributedtaskcommon.TaskInputDefinitionBase{}
			if name, ok := iMap["name"].(string); ok {
				inp.Name = converter.String(name)
			}
			if dv, ok := iMap["default_value"].(string); ok {
				inp.DefaultValue = converter.String(dv)
			}
			if pt, ok := iMap["parameter_type"].(string); ok {
				inp.Type = converter.String(pt)
			}
			inputs[i] = inp
		}
		pp.Inputs = &inputs
	}
	return pp
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

	if opts, ok := approvalMap["approval_options"].([]interface{}); ok && len(opts) > 0 && opts[0] != nil {
		result.ApprovalOptions = expandApprovalOptions(opts)
	}

	return result
}

func expandApprovalOptions(input []interface{}) *releaseapi.ApprovalOptions {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	optMap := input[0].(map[string]interface{})
	opts := &releaseapi.ApprovalOptions{
		ReleaseCreatorCanBeApprover:                             converter.Bool(optMap["release_creator_can_be_approver"].(bool)),
		EnforceIdentityRevalidation:                             converter.Bool(optMap["enforce_identity_revalidation"].(bool)),
		TimeoutInMinutes:                                        converter.Int(optMap["timeout_in_minutes"].(int)),
		AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped: converter.Bool(optMap["auto_triggered_and_previous_environment_approved_can_be_skipped"].(bool)),
	}

	if exOrder, ok := optMap["execution_order"].(string); ok && exOrder != "" {
		order := releaseapi.ApprovalExecutionOrder(exOrder)
		opts.ExecutionOrder = &order
	}

	// required_approver_count: only set if explicitly provided (nil = all required)
	if v, ok := optMap["required_approver_count"]; ok && v != nil {
		if count, ok := v.(int); ok {
			opts.RequiredApproverCount = converter.Int(count)
		}
	}

	return opts
}

func expandDeployPhases(input []interface{}) ([]interface{}, error) {
	phases := make([]interface{}, len(input))
	for i, v := range input {
		phaseMap := v.(map[string]interface{})
		phaseRaw := map[string]interface{}{
			"name":      phaseMap["name"].(string),
			"rank":      phaseMap["rank"].(int),
			"phaseType": phaseMap["phase_type"].(string),
		}

		if deplInput, ok := phaseMap["deployment_input"].([]interface{}); ok && len(deplInput) > 0 {
			phaseRaw["deploymentInput"] = expandDeploymentInput(deplInput, phaseMap["phase_type"].(string))
		}

		if tasks, ok := phaseMap["workflow_task"].([]interface{}); ok && len(tasks) > 0 {
			wfTasks := expandWorkflowTasks(tasks)
			phaseRaw["workflowTasks"] = wfTasks
		}

		phases[i] = phaseRaw
	}
	return phases, nil
}

func expandDeploymentInput(input []interface{}, phaseType ...string) map[string]interface{} {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	diMap := input[0].(map[string]interface{})
	di := map[string]interface{}{
		"timeoutInMinutes":          diMap["timeout_in_minutes"].(int),
		"jobCancelTimeoutInMinutes": diMap["job_cancel_timeout_in_minutes"].(int),
		"condition":                 diMap["condition"].(string),
		"skipArtifactsDownload":     diMap["skip_artifacts_download"].(bool),
		"enableAccessToken":         diMap["enable_access_token"].(bool),
	}

	// Only include queueId for agent-based phases; agentless (runOnServer) phases do not use a queue.
	pt := ""
	if len(phaseType) > 0 {
		pt = phaseType[0]
	}
	queueID := diMap["queue_id"].(int)
	if pt != "runOnServer" && queueID != 0 {
		di["queueId"] = queueID
	}

	if demands, ok := diMap["demands"].([]interface{}); ok && len(demands) > 0 {
		demandStrs := make([]string, len(demands))
		for j, d := range demands {
			demandStrs[j] = d.(string)
		}
		di["demands"] = demandStrs
	}

	if spec, ok := diMap["agent_specification"].(string); ok && spec != "" {
		di["agentSpecification"] = map[string]interface{}{
			"identifier": spec,
		}
	}

	if oi, ok := diMap["override_inputs"].(map[string]interface{}); ok && len(oi) > 0 {
		overrides := make(map[string]string, len(oi))
		for k, v := range oi {
			overrides[k], _ = v.(string)
		}
		di["overrideInputs"] = overrides
	}

	if pe, ok := diMap["parallel_execution"].([]interface{}); ok && len(pe) > 0 {
		di["parallelExecution"] = expandParallelExecution(pe)
	}

	return di
}

// expandParallelExecution converts the parallel_execution block to the ADO map shape.
func expandParallelExecution(input []interface{}) map[string]interface{} {
	if len(input) == 0 || input[0] == nil {
		return map[string]interface{}{"parallelExecutionType": "none"}
	}
	peMap := input[0].(map[string]interface{})
	peType, _ := peMap["type"].(string)
	if peType == "" {
		peType = "none"
	}
	result := map[string]interface{}{
		"parallelExecutionType": peType,
	}
	if maxAgents, ok := peMap["max_number_of_agents"].(int); ok && maxAgents > 0 {
		result["maxNumberOfAgents"] = maxAgents
	}
	if multipliers, ok := peMap["multipliers"].([]interface{}); ok && len(multipliers) > 0 {
		multStrs := make([]string, len(multipliers))
		for i, m := range multipliers {
			multStrs[i], _ = m.(string)
		}
		// ADO stores multipliers as a comma-separated string (Multipliers *string in the SDK),
		// not as a JSON array. Sending an array causes ADO to silently drop it.
		result["multipliers"] = strings.Join(multStrs, ",")
	}
	if coe, ok := peMap["continue_on_error"].(bool); ok {
		result["continueOnError"] = coe
	}
	return result
}

func expandWorkflowTasks(input []interface{}) []releaseapi.WorkflowTask {
	tasks := make([]releaseapi.WorkflowTask, len(input))
	for i, v := range input {
		taskMap := v.(map[string]interface{})
		taskID, _ := uuid.Parse(taskMap["task_id"].(string)) //nolint:errcheck
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

		if v, ok := taskMap["timeout_in_minutes"].(int); ok {
			task.TimeoutInMinutes = converter.Int(v)
		}
		if v, ok := taskMap["retry_count_on_task_failure"].(int); ok {
			task.RetryCountOnTaskFailure = converter.Int(v)
		}

		tasks[i] = task
	}
	return tasks
}

func expandRetentionPolicy(input []interface{}) *releaseapi.EnvironmentRetentionPolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	rpMap := input[0].(map[string]interface{})
	return &releaseapi.EnvironmentRetentionPolicy{
		DaysToKeep:     converter.Int(rpMap["days_to_keep"].(int)),
		ReleasesToKeep: converter.Int(rpMap["releases_to_keep"].(int)),
		RetainBuild:    converter.Bool(rpMap["retain_build"].(bool)),
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

		if defRef, ok := artMap["definition_reference"].(map[string]interface{}); ok && len(defRef) > 0 {
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

func expandEnvironmentOptions(input []interface{}) *releaseapi.EnvironmentOptions {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	optMap := input[0].(map[string]interface{})
	opts := &releaseapi.EnvironmentOptions{
		EmailNotificationType:        converter.String(optMap["email_notification_type"].(string)),
		SkipArtifactsDownload:        converter.Bool(optMap["skip_artifacts_download"].(bool)),
		TimeoutInMinutes:             converter.Int(optMap["timeout_in_minutes"].(int)),
		EnableAccessToken:            converter.Bool(optMap["enable_access_token"].(bool)),
		PublishDeploymentStatus:      converter.Bool(optMap["publish_deployment_status"].(bool)),
		BadgeEnabled:                 converter.Bool(optMap["badge_enabled"].(bool)),
		AutoLinkWorkItems:            converter.Bool(optMap["auto_link_work_items"].(bool)),
		PullRequestDeploymentEnabled: converter.Bool(optMap["pull_request_deployment_enabled"].(bool)),
	}
	// The API rejects empty string for EmailRecipients; omit when blank so the API
	// uses its default ("release.environment.owner;release.creator").
	if v, ok := optMap["email_recipients"].(string); ok && v != "" {
		opts.EmailRecipients = converter.String(v) //nolint:staticcheck
	}
	return opts
}

func expandExecutionPolicy(input []interface{}) *releaseapi.EnvironmentExecutionPolicy {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	epMap := input[0].(map[string]interface{})
	return &releaseapi.EnvironmentExecutionPolicy{
		ConcurrencyCount: converter.Int(epMap["concurrency_count"].(int)),
		QueueDepthCount:  converter.Int(epMap["queue_depth_count"].(int)),
	}
}

func expandDeploymentGates(input []interface{}) *releaseapi.ReleaseDefinitionGatesStep {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	gatesMap := input[0].(map[string]interface{})
	step := &releaseapi.ReleaseDefinitionGatesStep{}

	if opts, ok := gatesMap["gates_options"].([]interface{}); ok && len(opts) > 0 && opts[0] != nil {
		optMap := opts[0].(map[string]interface{})
		step.GatesOptions = &releaseapi.ReleaseDefinitionGatesOptions{
			IsEnabled:              converter.Bool(optMap["is_enabled"].(bool)),
			Timeout:                converter.Int(optMap["timeout"].(int)),
			SamplingInterval:       converter.Int(optMap["sampling_interval"].(int)),
			StabilizationTime:      converter.Int(optMap["stabilization_time"].(int)),
			MinimumSuccessDuration: converter.Int(optMap["minimum_success_duration"].(int)),
		}
	}

	if gateBlocks, ok := gatesMap["gate"].([]interface{}); ok && len(gateBlocks) > 0 {
		gates := make([]releaseapi.ReleaseDefinitionGate, 0, len(gateBlocks))
		for _, gb := range gateBlocks {
			if gb == nil {
				continue
			}
			gateMap := gb.(map[string]interface{})
			gate := releaseapi.ReleaseDefinitionGate{}
			if taskList, ok := gateMap["task"].([]interface{}); ok && len(taskList) > 0 {
				tasks := expandWorkflowTasks(taskList)
				gate.Tasks = &tasks
			}
			gates = append(gates, gate)
		}
		step.Gates = &gates
	}

	return step
}

// isNonDefaultGatesOptions returns true when at least one field in GatesOptions
// differs from its default (is_enabled=false, timeout=0, samplingInterval=0,
// stabilizationTime=0, minimumSuccessDuration=0). Used to distinguish a real
// gates_options block from ADO's always-present but empty-configured step.
func isNonDefaultGatesOptions(o *releaseapi.ReleaseDefinitionGatesOptions) bool {
	if o == nil {
		return false
	}
	if o.IsEnabled != nil && *o.IsEnabled {
		return true
	}
	if o.Timeout != nil && *o.Timeout != 0 {
		return true
	}
	if o.SamplingInterval != nil && *o.SamplingInterval != 0 {
		return true
	}
	if o.StabilizationTime != nil && *o.StabilizationTime != 0 {
		return true
	}
	if o.MinimumSuccessDuration != nil && *o.MinimumSuccessDuration != 0 {
		return true
	}
	return false
}

func flattenDeploymentGates(step *releaseapi.ReleaseDefinitionGatesStep) []interface{} {
	if step == nil {
		return nil
	}

	// ADO always returns a ReleaseDefinitionGatesStep object even for environments
	// that have no gates configured (all options at default, no gate tasks).
	// Suppress such empty steps to avoid perpetual diffs on environments that do
	// not have a pre/post_deployment_gates block in HCL.
	//
	// We consider a step "empty" when:
	//   - no gate tasks are configured (step.Gates is nil or empty), AND
	//   - GatesOptions is either nil or has only default values (is_enabled=false,
	//     timeout=0, samplingInterval=0, stabilizationTime=0, minimumSuccessDuration=0)
	hasGates := step.Gates != nil && len(*step.Gates) > 0
	hasNonDefaultOptions := step.GatesOptions != nil && isNonDefaultGatesOptions(step.GatesOptions)

	if !hasGates && !hasNonDefaultOptions {
		return nil
	}

	gatesMap := map[string]interface{}{}

	if step.GatesOptions != nil {
		o := step.GatesOptions
		optMap := map[string]interface{}{
			"is_enabled":               false,
			"timeout":                  0,
			"sampling_interval":        0,
			"stabilization_time":       0,
			"minimum_success_duration": 0,
		}
		if o.IsEnabled != nil {
			optMap["is_enabled"] = *o.IsEnabled
		}
		if o.Timeout != nil {
			optMap["timeout"] = *o.Timeout
		}
		if o.SamplingInterval != nil {
			optMap["sampling_interval"] = *o.SamplingInterval
		}
		if o.StabilizationTime != nil {
			optMap["stabilization_time"] = *o.StabilizationTime
		}
		if o.MinimumSuccessDuration != nil {
			optMap["minimum_success_duration"] = *o.MinimumSuccessDuration
		}
		gatesMap["gates_options"] = []interface{}{optMap}
	}

	if step.Gates != nil && len(*step.Gates) > 0 {
		flatGates := make([]interface{}, 0, len(*step.Gates))
		for _, g := range *step.Gates {
			gateFlat := map[string]interface{}{
				"task": []interface{}{},
			}
			if g.Tasks != nil && len(*g.Tasks) > 0 {
				gateFlat["task"] = flattenWorkflowTasksFromAPI(g.Tasks)
			}
			flatGates = append(flatGates, gateFlat)
		}
		gatesMap["gate"] = flatGates
	}

	return []interface{}{gatesMap}
}

// flattenWorkflowTasksFromAPI converts a *[]releaseapi.WorkflowTask (from the ADO API response)
// into a []interface{} suitable for setting in Terraform state.
func flattenWorkflowTasksFromAPI(tasks *[]releaseapi.WorkflowTask) []interface{} {
	if tasks == nil {
		return nil
	}
	result := make([]interface{}, 0, len(*tasks))
	for _, t := range *tasks {
		flat := map[string]interface{}{
			"name":                        "",
			"task_id":                     "",
			"version":                     "1.*",
			"enabled":                     true,
			"always_run":                  false,
			"continue_on_error":           false,
			"condition":                   "succeeded()",
			"definition_type":             "task",
			"inputs":                      map[string]string{},
			"timeout_in_minutes":          0,
			"retry_count_on_task_failure": 0,
		}
		if t.Name != nil {
			flat["name"] = *t.Name
		}
		if t.TaskId != nil {
			flat["task_id"] = t.TaskId.String()
		}
		if t.Version != nil {
			flat["version"] = *t.Version
		}
		if t.Enabled != nil {
			flat["enabled"] = *t.Enabled
		}
		if t.AlwaysRun != nil {
			flat["always_run"] = *t.AlwaysRun
		}
		if t.ContinueOnError != nil {
			flat["continue_on_error"] = *t.ContinueOnError
		}
		if t.Condition != nil {
			flat["condition"] = *t.Condition
		}
		if t.DefinitionType != nil {
			flat["definition_type"] = *t.DefinitionType
		}
		if t.Inputs != nil && len(*t.Inputs) > 0 {
			flat["inputs"] = *t.Inputs
		}
		if t.TimeoutInMinutes != nil {
			flat["timeout_in_minutes"] = *t.TimeoutInMinutes
		}
		if t.RetryCountOnTaskFailure != nil {
			flat["retry_count_on_task_failure"] = *t.RetryCountOnTaskFailure
		}
		result = append(result, flat)
	}
	return result
}

// Flatten functions: API response → Terraform state

func flattenReleaseDefinition(d *schema.ResourceData, def *releaseapi.ReleaseDefinition, projectID string) {
	d.SetId(strconv.Itoa(*def.Id))
	d.Set("project_id", projectID)
	d.Set("name", def.Name)
	d.Set("path", def.Path)
	d.Set("description", def.Description)
	d.Set("release_name_format", def.ReleaseNameFormat)
	d.Set("revision", def.Revision)

	// Variables
	if def.Variables != nil {
		d.Set("variable", flattenVariables(def.Variables, collectSecretValues(d, "variable")))
	}

	// Variable groups
	if def.VariableGroups != nil {
		d.Set("variable_groups", flattenVariableGroups(def.VariableGroups))
	}

	// Tags
	if def.Tags != nil {
		d.Set("tags", *def.Tags)
	}

	// Environments
	if def.Environments != nil {
		d.Set("stages", flattenStages(def.Environments, d))
	}

	// Artifacts
	if def.Artifacts != nil {
		d.Set("artifact", flattenArtifacts(def.Artifacts, d))
	}

	// Triggers
	if def.Triggers != nil && len(*def.Triggers) > 0 {
		d.Set("triggers", flattenTriggers(def.Triggers))
	}
}

// collectSecretValues reads secret variable values from the Terraform state at
// the given state key (e.g. "variable" for definition-level, or
// "stages.0.variable" for a stage-scoped variable set). The ADO API
// returns null for secret values, so the in-state value must be preserved on
// flatten to avoid a perpetual diff. The key MUST match the scope of the
// variables being flattened — using the definition-level key for an
// environment's variables loses env-scoped secrets (the original defect).
func collectSecretValues(d *schema.ResourceData, stateKey string) map[string]string {
	existingVars := make(map[string]string)
	if v, ok := d.GetOk(stateKey); ok {
		for _, item := range v.(*schema.Set).List() {
			varMap := item.(map[string]interface{})
			if varMap["is_secret"].(bool) {
				existingVars[varMap["name"].(string)] = varMap["value"].(string)
			}
		}
	}
	return existingVars
}

// flattenVariables converts an API variable map into Terraform state. Secret
// values (returned null by ADO) are restored from existingSecrets, which the
// caller builds from the correct state scope via collectSecretValues.
func flattenVariables(variables *map[string]releaseapi.ConfigurationVariableValue, existingSecrets map[string]string) []interface{} {
	if variables == nil {
		return nil
	}

	result := make([]interface{}, 0, len(*variables))
	for name, v := range *variables {
		varMap := map[string]interface{}{
			"name":           name,
			"value":          "",
			"is_secret":      false,
			"allow_override": false,
		}
		if v.Value != nil {
			varMap["value"] = *v.Value
		}
		if v.IsSecret != nil {
			varMap["is_secret"] = *v.IsSecret
			if *v.IsSecret {
				// Secret values come back null — use state value
				if stateVal, ok := existingSecrets[name]; ok {
					varMap["value"] = stateVal
				}
			}
		}
		if v.AllowOverride != nil {
			varMap["allow_override"] = *v.AllowOverride
		}
		result = append(result, varMap)
	}
	return result
}

func flattenVariableGroups(groups *[]int) []int {
	if groups == nil {
		return nil
	}
	return *groups
}

func flattenStages(envs *[]releaseapi.ReleaseDefinitionEnvironment, d *schema.ResourceData) []interface{} {
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

		// Conditions
		if env.Conditions != nil {
			envMap["condition"] = flattenConditions(env.Conditions)
		}

		// Pre-deploy approvals
		if env.PreDeployApprovals != nil {
			envMap["pre_deploy_approval"] = flattenApprovals(env.PreDeployApprovals)
		}

		// Post-deploy approvals
		if env.PostDeployApprovals != nil {
			envMap["post_deploy_approval"] = flattenApprovals(env.PostDeployApprovals)
		}

		// Deploy phases
		if env.DeployPhases != nil {
			envMap["deploy_phase"] = flattenDeployPhases(env.DeployPhases, d, i)
		}

		// Retention policy
		if env.RetentionPolicy != nil {
			envMap["retention_policy"] = flattenRetentionPolicy(env.RetentionPolicy)
		}

		// Environment options
		if env.EnvironmentOptions != nil {
			envMap["environment_options"] = flattenEnvironmentOptions(env.EnvironmentOptions)
		}

		// Execution policy
		if env.ExecutionPolicy != nil {
			envMap["execution_policy"] = flattenExecutionPolicy(env.ExecutionPolicy)
		}

		// Pre-deployment gates
		if env.PreDeploymentGates != nil {
			envMap["pre_deployment_gates"] = flattenDeploymentGates(env.PreDeploymentGates)
		}

		// Post-deployment gates
		if env.PostDeploymentGates != nil {
			envMap["post_deployment_gates"] = flattenDeploymentGates(env.PostDeploymentGates)
		}

		// Variables — preserve stage-scoped secrets from the matching state path
		// (stages.<i>.variable), not the definition-level "variable" key.
		if env.Variables != nil {
			envKey := fmt.Sprintf("stages.%d.variable", i)
			envMap["variable"] = flattenVariables(env.Variables, collectSecretValues(d, envKey))
		}

		// Variable groups — always include (even if empty) so that the Terraform state
		// matches HCL's implicit zero-value of [] for Optional TypeList. Omitting it
		// causes a perpetual diff (state absent → plan adds []).
		if env.VariableGroups != nil {
			envMap["variable_groups"] = flattenVariableGroups(env.VariableGroups)
		} else {
			envMap["variable_groups"] = []int{}
		}

		// Environment triggers
		if env.EnvironmentTriggers != nil && len(*env.EnvironmentTriggers) > 0 {
			envMap["environment_trigger"] = flattenEnvironmentTriggers(env.EnvironmentTriggers)
		}

		// Schedules
		if env.Schedules != nil && len(*env.Schedules) > 0 {
			envMap["schedule"] = flattenEnvironmentSchedules(env.Schedules)
		}

		// Process parameters
		if env.ProcessParameters != nil {
			envMap["process_parameters"] = flattenProcessParameters(env.ProcessParameters)
		}

		// Properties
		// ADO REST returns properties in a typed wrapper:
		//   {"key": {"$type": "System.String", "$value": "<actual-value>"}}
		// We unwrap the $value field to recover the plain string the caller set.
		if env.Properties != nil {
			if props, ok := env.Properties.(map[string]interface{}); ok {
				propsStr := make(map[string]string, len(props))
				for k, v := range props {
					switch cv := v.(type) {
					case string:
						propsStr[k] = cv
					case map[string]interface{}:
						// Unwrap ADO's {"$type":"System.String","$value":"..."} wrapper.
						if val, ok := cv["$value"].(string); ok {
							propsStr[k] = val
						} else {
							propsStr[k] = fmt.Sprintf("%v", v)
						}
					default:
						propsStr[k] = fmt.Sprintf("%v", v)
					}
				}
				envMap["properties"] = propsStr
			}
		}

		result[i] = envMap
	}
	return result
}

func flattenEnvironmentTriggers(triggers *[]releaseapi.EnvironmentTrigger) []interface{} {
	if triggers == nil {
		return nil
	}
	result := make([]interface{}, len(*triggers))
	for i, t := range *triggers {
		tMap := map[string]interface{}{
			"definition_environment_id": 0,
			"trigger_type":              "",
			"trigger_content":           "",
		}
		if t.DefinitionEnvironmentId != nil {
			tMap["definition_environment_id"] = *t.DefinitionEnvironmentId
		}
		if t.TriggerType != nil {
			tMap["trigger_type"] = string(*t.TriggerType)
		}
		if t.TriggerContent != nil {
			tMap["trigger_content"] = *t.TriggerContent
		}
		result[i] = tMap
	}
	return result
}

func flattenEnvironmentSchedules(scheds *[]releaseapi.ReleaseSchedule) []interface{} {
	if scheds == nil {
		return nil
	}
	result := make([]interface{}, len(*scheds))
	for i, s := range *scheds {
		sMap := map[string]interface{}{
			"days_to_release": 127,
			"start_hours":     0,
			"start_minutes":   0,
			"time_zone_id":    "UTC",
			"job_id":          "",
		}
		if s.DaysToRelease != nil {
			if d, err := strconv.Atoi(string(*s.DaysToRelease)); err == nil {
				sMap["days_to_release"] = d
			}
		}
		if s.StartHours != nil {
			sMap["start_hours"] = *s.StartHours
		}
		if s.StartMinutes != nil {
			sMap["start_minutes"] = *s.StartMinutes
		}
		if s.TimeZoneId != nil {
			sMap["time_zone_id"] = *s.TimeZoneId
		}
		if s.JobId != nil {
			sMap["job_id"] = s.JobId.String()
		}
		result[i] = sMap
	}
	return result
}

func flattenProcessParameters(pp *distributedtaskcommon.ProcessParameters) []interface{} {
	if pp == nil {
		return nil
	}
	ppMap := map[string]interface{}{}
	if pp.Inputs != nil && len(*pp.Inputs) > 0 {
		inputs := make([]interface{}, len(*pp.Inputs))
		for i, inp := range *pp.Inputs {
			iMap := map[string]interface{}{
				"name":           "",
				"default_value":  "",
				"parameter_type": "",
			}
			if inp.Name != nil {
				iMap["name"] = *inp.Name
			}
			if inp.DefaultValue != nil {
				iMap["default_value"] = *inp.DefaultValue
			}
			if inp.Type != nil {
				iMap["parameter_type"] = *inp.Type
			}
			inputs[i] = iMap
		}
		ppMap["input"] = inputs
	}
	return []interface{}{ppMap}
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
		steps := make([]interface{}, len(*approvals.Approvals))
		for i, step := range *approvals.Approvals {
			stepMap := map[string]interface{}{
				"id":           "00000000-0000-0000-0000-000000000000",
				"is_automated": false,
				"rank":         1,
			}
			if step.Approver != nil && step.Approver.Id != nil && *step.Approver.Id != "" {
				stepMap["id"] = *step.Approver.Id
			}
			if step.IsAutomated != nil {
				stepMap["is_automated"] = *step.IsAutomated
			}
			if step.Rank != nil {
				stepMap["rank"] = *step.Rank
			}
			steps[i] = stepMap
		}
		approvalMap["approver"] = steps
	}

	if approvals.ApprovalOptions != nil {
		approvalMap["approval_options"] = flattenApprovalOptions(approvals.ApprovalOptions)
	}

	return []interface{}{approvalMap}
}

func flattenApprovalOptions(opts *releaseapi.ApprovalOptions) []interface{} {
	if opts == nil {
		return nil
	}
	optMap := map[string]interface{}{
		"release_creator_can_be_approver": false,
		"enforce_identity_revalidation":   false,
		"timeout_in_minutes":              0,
		"execution_order":                 "beforeGates",
		"auto_triggered_and_previous_environment_approved_can_be_skipped": false,
	}
	if opts.RequiredApproverCount != nil {
		optMap["required_approver_count"] = *opts.RequiredApproverCount
	}
	if opts.ReleaseCreatorCanBeApprover != nil {
		optMap["release_creator_can_be_approver"] = *opts.ReleaseCreatorCanBeApprover
	}
	if opts.EnforceIdentityRevalidation != nil {
		optMap["enforce_identity_revalidation"] = *opts.EnforceIdentityRevalidation
	}
	if opts.TimeoutInMinutes != nil {
		optMap["timeout_in_minutes"] = *opts.TimeoutInMinutes
	}
	if opts.ExecutionOrder != nil {
		optMap["execution_order"] = string(*opts.ExecutionOrder)
	}
	if opts.AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped != nil {
		optMap["auto_triggered_and_previous_environment_approved_can_be_skipped"] = *opts.AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped
	}
	return []interface{}{optMap}
}

// flattenDeployPhases converts ADO deploy phases back to Terraform state.
// d and envIdx are used to check whether each phase's corresponding HCL block had a
// deployment_input sub-block: if yes, we always emit deployment_input (even when all
// values are at their defaults); if no, we suppress all-default deployment_input blocks
// to avoid perpetual diffs on phases whose HCL does not set deployment_input.
func flattenDeployPhases(phases *[]interface{}, d *schema.ResourceData, envIdx int) []interface{} {
	if phases == nil {
		return nil
	}

	result := make([]interface{}, 0, len(*phases))
	for phaseIdx, phase := range *phases {
		// The API returns deploy phases as generic interface{} which unmarshals as map[string]interface{}
		phaseData, err := json.Marshal(phase)
		if err != nil {
			continue
		}
		var phaseMap map[string]interface{}
		if err := json.Unmarshal(phaseData, &phaseMap); err != nil {
			continue
		}

		flatPhase := map[string]interface{}{
			"name":       "",
			"rank":       1,
			"phase_type": "agentBasedDeployment",
		}

		if name, ok := phaseMap["name"].(string); ok {
			flatPhase["name"] = name
		}
		if rank, ok := phaseMap["rank"].(float64); ok {
			flatPhase["rank"] = int(rank)
		}
		if pt, ok := phaseMap["phaseType"].(string); ok {
			flatPhase["phase_type"] = pt
		}

		// Deployment input — check whether the corresponding HCL phase block has a
		// deployment_input sub-block. If it does, always emit deployment_input (even
		// all-defaults). If it doesn't, suppress all-default entries to avoid perpetual
		// diffs. ADO always returns a deploymentInput object for every phase.
		if di, ok := phaseMap["deploymentInput"].(map[string]interface{}); ok {
			flat := flattenDeploymentInput(di)
			hclHasDI := hclPhaseHasDeploymentInput(d, envIdx, phaseIdx)
			if hclHasDI || !isDefaultDeploymentInput(flat) {
				flatPhase["deployment_input"] = flat
			}
		}

		// Workflow tasks
		if wfTasks, ok := phaseMap["workflowTasks"].([]interface{}); ok {
			flatPhase["workflow_task"] = flattenWorkflowTasks(wfTasks)
		}

		result = append(result, flatPhase)
	}
	return result
}

// hclPhaseHasDeploymentInput returns true if the phase at [envIdx][phaseIdx] in the
// current Terraform resource data has a non-empty deployment_input block.
func hclPhaseHasDeploymentInput(d *schema.ResourceData, envIdx, phaseIdx int) bool {
	if d == nil {
		return false
	}
	key := fmt.Sprintf("stages.%d.deploy_phase.%d.deployment_input.#", envIdx, phaseIdx)
	count, ok := d.GetOk(key)
	if !ok {
		return false
	}
	if n, ok := count.(int); ok && n > 0 {
		return true
	}
	return false
}

func flattenDeploymentInput(di map[string]interface{}) []interface{} {
	if di == nil {
		return nil
	}
	diMap := map[string]interface{}{
		"queue_id":                      0,
		"timeout_in_minutes":            0,
		"job_cancel_timeout_in_minutes": 1,
		"condition":                     "succeeded()",
		"skip_artifacts_download":       false,
		"enable_access_token":           false,
		"agent_specification":           "",
	}

	if queueID, ok := di["queueId"].(float64); ok {
		diMap["queue_id"] = int(queueID)
	}
	if timeout, ok := di["timeoutInMinutes"].(float64); ok {
		diMap["timeout_in_minutes"] = int(timeout)
	}
	if jcTimeout, ok := di["jobCancelTimeoutInMinutes"].(float64); ok {
		diMap["job_cancel_timeout_in_minutes"] = int(jcTimeout)
	}
	if cond, ok := di["condition"].(string); ok {
		diMap["condition"] = cond
	}
	if skip, ok := di["skipArtifactsDownload"].(bool); ok {
		diMap["skip_artifacts_download"] = skip
	}
	if eat, ok := di["enableAccessToken"].(bool); ok {
		diMap["enable_access_token"] = eat
	}

	// Agent specification
	if agentSpec, ok := di["agentSpecification"].(map[string]interface{}); ok {
		if identifier, ok := agentSpec["identifier"].(string); ok {
			diMap["agent_specification"] = identifier
		}
	}

	// Override inputs — only set when populated so phases without overrides in HCL
	// don't acquire a spurious empty map in state (perpetual diff).
	if oi, ok := di["overrideInputs"].(map[string]interface{}); ok && len(oi) > 0 {
		overrides := make(map[string]string, len(oi))
		for k, v := range oi {
			overrides[k], _ = v.(string)
		}
		diMap["override_inputs"] = overrides
	}

	// Demands — always include the key (empty or populated) so that the Terraform
	// state matches HCL's implicit zero-value of [] for Optional TypeList.
	// Omitting it causes a perpetual diff (state absent → plan adds []).
	demandStrs := []interface{}{}
	if demands, ok := di["demands"].([]interface{}); ok {
		for _, d := range demands {
			if s, ok := d.(string); ok {
				demandStrs = append(demandStrs, s)
			}
		}
	}
	diMap["demands"] = demandStrs

	// Parallel execution — only set the key when there is a meaningful block.
	// flattenParallelExecution returns nil for a "none"-typed or absent entry so
	// that phases without an explicit parallel_execution block in HCL don't
	// acquire a spurious empty block in state (which would cause a perpetual diff).
	if pe, ok := di["parallelExecution"].(map[string]interface{}); ok {
		if flat := flattenParallelExecution(pe); flat != nil {
			diMap["parallel_execution"] = flat
		}
	}

	return []interface{}{diMap}
}

// isDefaultDeploymentInput returns true when the flattened deployment_input contains
// only default values (queue_id=0, timeout=0, jcTimeout=1, condition="succeeded()",
// skip=false, eat=false, no demands, no parallel_execution, no agent_specification).
// Used by flattenDeployPhases to avoid emitting a spurious deployment_input block for
// phases whose HCL does not set one (ADO always returns a deploymentInput object).
func isDefaultDeploymentInput(flat []interface{}) bool {
	if len(flat) == 0 || flat[0] == nil {
		return true
	}
	diMap, ok := flat[0].(map[string]interface{})
	if !ok {
		return true
	}
	if v, ok := diMap["queue_id"].(int); ok && v != 0 {
		return false
	}
	if v, ok := diMap["timeout_in_minutes"].(int); ok && v != 0 {
		return false
	}
	if v, ok := diMap["job_cancel_timeout_in_minutes"].(int); ok && v != 1 {
		return false
	}
	if v, ok := diMap["condition"].(string); ok && v != "" && v != "succeeded()" {
		return false
	}
	if v, ok := diMap["skip_artifacts_download"].(bool); ok && v {
		return false
	}
	if v, ok := diMap["enable_access_token"].(bool); ok && v {
		return false
	}
	if v, ok := diMap["agent_specification"].(string); ok && v != "" {
		return false
	}
	if demands, ok := diMap["demands"].([]string); ok && len(demands) > 0 {
		return false
	}
	if _, hasPE := diMap["parallel_execution"]; hasPE {
		return false
	}
	if oi, ok := diMap["override_inputs"].(map[string]string); ok && len(oi) > 0 {
		return false
	}
	return true
}

// flattenParallelExecution converts the ADO parallelExecution map to Terraform state.
// Returns nil (not an empty slice) when the type is "none" so that phases that did
// not declare a parallel_execution block in HCL don't get a spurious empty block
// injected into state.
func flattenParallelExecution(pe map[string]interface{}) []interface{} {
	if pe == nil {
		return nil
	}
	peType := "none"
	if t, ok := pe["parallelExecutionType"].(string); ok {
		peType = t
	}
	// Only emit a block when parallelExecution is meaningfully configured.
	// A bare "none" type with all defaults is indistinguishable from "not set"
	// from the HCL author's perspective, so suppress it to avoid perpetual diffs.
	if peType == "none" {
		return nil
	}
	peFlat := map[string]interface{}{
		"type":                 peType,
		"max_number_of_agents": 0,
		"continue_on_error":    false,
		"multipliers":          []string{},
	}
	if maxAgents, ok := pe["maxNumberOfAgents"].(float64); ok {
		peFlat["max_number_of_agents"] = int(maxAgents)
	}
	if coe, ok := pe["continueOnError"].(bool); ok {
		peFlat["continue_on_error"] = coe
	}
	// The ADO API stores multipliers as a comma-joined string (e.g. "x86,x64"),
	// NOT as a JSON array. Handle both forms for robustness.
	switch v := pe["multipliers"].(type) {
	case string:
		if v != "" {
			parts := strings.Split(v, ",")
			cleaned := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					cleaned = append(cleaned, p)
				}
			}
			if len(cleaned) > 0 {
				peFlat["multipliers"] = cleaned
			}
		}
	case []interface{}:
		if len(v) > 0 {
			multStrs := make([]string, len(v))
			for i, m := range v {
				multStrs[i], _ = m.(string)
			}
			peFlat["multipliers"] = multStrs
		}
	}
	return []interface{}{peFlat}
}

func flattenWorkflowTasks(tasks []interface{}) []interface{} {
	if tasks == nil {
		return nil
	}
	result := make([]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		flat := map[string]interface{}{
			"name":                        "",
			"task_id":                     "",
			"version":                     "1.*",
			"enabled":                     true,
			"always_run":                  false,
			"continue_on_error":           false,
			"condition":                   "succeeded()",
			"definition_type":             "task",
			"inputs":                      map[string]string{},
			"timeout_in_minutes":          0,
			"retry_count_on_task_failure": 0,
		}

		if name, ok := taskMap["name"].(string); ok {
			flat["name"] = name
		}

		// task_id comes from nested "task" object or direct "taskId" field
		if taskRef, ok := taskMap["task"].(map[string]interface{}); ok {
			if id, ok := taskRef["id"].(string); ok {
				flat["task_id"] = id
			}
			if ver, ok := taskRef["versionSpec"].(string); ok {
				flat["version"] = ver
			}
			if dt, ok := taskRef["definitionType"].(string); ok {
				flat["definition_type"] = dt
			}
		}
		if taskID, ok := taskMap["taskId"].(string); ok {
			flat["task_id"] = taskID
		}

		if enabled, ok := taskMap["enabled"].(bool); ok {
			flat["enabled"] = enabled
		}
		if ar, ok := taskMap["alwaysRun"].(bool); ok {
			flat["always_run"] = ar
		}
		if coe, ok := taskMap["continueOnError"].(bool); ok {
			flat["continue_on_error"] = coe
		}
		if cond, ok := taskMap["condition"].(string); ok {
			flat["condition"] = cond
		}
		if ver, ok := taskMap["version"].(string); ok {
			flat["version"] = ver
		}
		if dt, ok := taskMap["definitionType"].(string); ok {
			flat["definition_type"] = dt
		}

		// Inputs
		if inputs, ok := taskMap["inputs"].(map[string]interface{}); ok && len(inputs) > 0 {
			inputMap := make(map[string]string)
			for k, v := range inputs {
				inputMap[k], _ = v.(string)
			}
			flat["inputs"] = inputMap
		}

		if v, ok := taskMap["timeoutInMinutes"].(int); ok {
			flat["timeout_in_minutes"] = v
		}
		if v, ok := taskMap["retryCountOnTaskFailure"].(int); ok {
			flat["retry_count_on_task_failure"] = v
		}

		result = append(result, flat)
	}
	return result
}

func flattenRetentionPolicy(rp *releaseapi.EnvironmentRetentionPolicy) []interface{} {
	if rp == nil {
		return nil
	}
	rpMap := map[string]interface{}{
		"days_to_keep":     30,
		"releases_to_keep": 3,
		"retain_build":     true,
	}
	if rp.DaysToKeep != nil {
		rpMap["days_to_keep"] = *rp.DaysToKeep
	}
	if rp.ReleasesToKeep != nil {
		rpMap["releases_to_keep"] = *rp.ReleasesToKeep
	}
	if rp.RetainBuild != nil {
		rpMap["retain_build"] = *rp.RetainBuild
	}
	return []interface{}{rpMap}
}

func flattenEnvironmentOptions(opts *releaseapi.EnvironmentOptions) []interface{} {
	if opts == nil {
		return nil
	}
	optMap := map[string]interface{}{
		"email_notification_type":         "OnlyOnFailure",
		"email_recipients":                "",
		"skip_artifacts_download":         false,
		"timeout_in_minutes":              0,
		"enable_access_token":             false,
		"publish_deployment_status":       false,
		"badge_enabled":                   false,
		"auto_link_work_items":            false,
		"pull_request_deployment_enabled": false,
	}
	if opts.EmailNotificationType != nil { //nolint:staticcheck
		optMap["email_notification_type"] = *opts.EmailNotificationType //nolint:staticcheck
	}
	if opts.EmailRecipients != nil { //nolint:staticcheck
		optMap["email_recipients"] = *opts.EmailRecipients //nolint:staticcheck
	}
	if opts.SkipArtifactsDownload != nil { //nolint:staticcheck
		optMap["skip_artifacts_download"] = *opts.SkipArtifactsDownload //nolint:staticcheck
	}
	if opts.TimeoutInMinutes != nil { //nolint:staticcheck
		optMap["timeout_in_minutes"] = *opts.TimeoutInMinutes //nolint:staticcheck
	}
	if opts.EnableAccessToken != nil { //nolint:staticcheck
		optMap["enable_access_token"] = *opts.EnableAccessToken //nolint:staticcheck
	}
	if opts.PublishDeploymentStatus != nil {
		optMap["publish_deployment_status"] = *opts.PublishDeploymentStatus
	}
	if opts.BadgeEnabled != nil {
		optMap["badge_enabled"] = *opts.BadgeEnabled
	}
	if opts.AutoLinkWorkItems != nil {
		optMap["auto_link_work_items"] = *opts.AutoLinkWorkItems
	}
	if opts.PullRequestDeploymentEnabled != nil {
		optMap["pull_request_deployment_enabled"] = *opts.PullRequestDeploymentEnabled
	}
	return []interface{}{optMap}
}

func flattenExecutionPolicy(ep *releaseapi.EnvironmentExecutionPolicy) []interface{} {
	if ep == nil {
		return nil
	}
	epMap := map[string]interface{}{
		"concurrency_count": 1,
		"queue_depth_count": 0,
	}
	if ep.ConcurrencyCount != nil {
		epMap["concurrency_count"] = *ep.ConcurrencyCount
	}
	if ep.QueueDepthCount != nil {
		epMap["queue_depth_count"] = *ep.QueueDepthCount
	}
	return []interface{}{epMap}
}

// expandTriggers converts a Terraform "triggers" block into a []interface{} that
// matches ADO's polymorphic Triggers wire format.
func expandTriggers(input []interface{}) []interface{} {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	trigMap := input[0].(map[string]interface{})
	result := make([]interface{}, 0)

	// CD artifact triggers — one entry per cd_artifact_trigger block
	if cdTriggers, ok := trigMap["cd_artifact_trigger"].([]interface{}); ok {
		for _, raw := range cdTriggers {
			if raw == nil {
				continue
			}
			ctMap := raw.(map[string]interface{})
			trigEntry := map[string]interface{}{
				"triggerType":   "artifactSource",
				"artifactAlias": ctMap["artifact_alias"].(string),
			}
			// Branch filters + tag_filter / use_build_definition_branch / create_release_on_build_tagging
			trigEntry["triggerConditions"] = expandArtifactTriggerConditions(ctMap)
			result = append(result, trigEntry)
		}
	}

	// Schedule triggers — one entry per schedule_trigger block
	if schTriggers, ok := trigMap["schedule_trigger"].([]interface{}); ok {
		for _, raw := range schTriggers {
			if raw == nil {
				continue
			}
			stMap := raw.(map[string]interface{})
			schedule := map[string]interface{}{
				"scheduleOnlyWithChanges": stMap["schedule_only_with_changes"].(bool),
				"startHours":              stMap["start_hours"].(int),
				"startMinutes":            stMap["start_minutes"].(int),
				"timeZoneId":              stMap["time_zone_id"].(string),
				"daysToRelease":           stMap["days_to_release"].(int),
			}
			trigEntry := map[string]interface{}{
				"triggerType": "schedule",
				"schedule":    schedule,
			}
			// branch_filter was removed from schedule_trigger schema (AC1/WI-9):
			// ADO classic schedule triggers are time-based; ADO does not return
			// branchFilters in the GET response, so we never set them here.
			result = append(result, trigEntry)
		}
	}

	// Source repo triggers — one entry per source_repo_trigger block
	if srtEntries, ok := trigMap["source_repo_trigger"].([]interface{}); ok {
		for _, raw := range srtEntries {
			if raw == nil {
				continue
			}
			srtMap := raw.(map[string]interface{})
			trigEntry := map[string]interface{}{
				"triggerType": "sourceRepo",
				"alias":       srtMap["alias"].(string),
			}
			if bfs, ok := srtMap["branch_filters"].([]interface{}); ok && len(bfs) > 0 {
				branchFilters := make([]string, 0, len(bfs))
				for _, bf := range bfs {
					branchFilters = append(branchFilters, bf.(string))
				}
				trigEntry["branchFilters"] = branchFilters
			}
			result = append(result, trigEntry)
		}
	}

	return result
}

// expandBranchFiltersForTrigger converts a branch_filter block into a slice of
// ArtifactFilter-style maps for the cd_artifact_trigger triggerConditions field.
// Deprecated: use expandArtifactTriggerConditions instead.
func expandBranchFiltersForTrigger(bfMap map[string]interface{}) []map[string]interface{} {
	var filters []map[string]interface{}
	if includes, ok := bfMap["include"].([]interface{}); ok {
		for _, v := range includes {
			filters = append(filters, map[string]interface{}{
				"sourceBranch": v.(string),
			})
		}
	}
	return filters
}

// expandArtifactTriggerConditions builds the triggerConditions slice for a
// cd_artifact_trigger block, incorporating branch_filter, tag_filter,
// use_build_definition_branch, and create_release_on_build_tagging.
func expandArtifactTriggerConditions(ctMap map[string]interface{}) []map[string]interface{} {
	// Start with branch filter conditions (one entry per include branch)
	var conditions []map[string]interface{}
	if bf, ok := ctMap["branch_filter"].([]interface{}); ok && len(bf) > 0 && bf[0] != nil {
		bfMap := bf[0].(map[string]interface{})
		conditions = expandBranchFiltersForTrigger(bfMap)
	}

	// Gather the extra per-trigger fields. ADO REST 7.1 persists build-tag
	// filtering as the `tags` string array on the ArtifactFilter condition;
	// the SDK's regex `tagFilter` field is silently dropped by the service
	// (verified live 2026-06-11), so only `tags` is sent.
	var tags []string
	if tf, ok := ctMap["tag_filter"].([]interface{}); ok && len(tf) > 0 && tf[0] != nil {
		tfBlock := tf[0].(map[string]interface{})
		if rawTags, ok := tfBlock["tags"].([]interface{}); ok {
			for _, t := range rawTags {
				tags = append(tags, t.(string))
			}
		}
	}
	useBuildBranch, _ := ctMap["use_build_definition_branch"].(bool)
	createOnTagging, _ := ctMap["create_release_on_build_tagging"].(bool)

	hasExtras := len(tags) > 0 || useBuildBranch || createOnTagging

	if !hasExtras {
		return conditions
	}

	// Emit the extra fields on the first condition entry.
	// If there are no branch-filter conditions, create a synthetic entry with
	// an empty sourceBranch so the extras have a container.
	if len(conditions) == 0 {
		conditions = []map[string]interface{}{
			{"sourceBranch": ""},
		}
	}
	if len(tags) > 0 {
		conditions[0]["tags"] = tags
	}
	if useBuildBranch {
		conditions[0]["useBuildDefinitionBranch"] = true
	}
	if createOnTagging {
		conditions[0]["createReleaseOnBuildTagging"] = true
	}
	return conditions
}

// flattenTriggers converts the ADO polymorphic Triggers []interface{} back into
// a Terraform "triggers" block.
func flattenTriggers(triggers *[]interface{}) []interface{} {
	if triggers == nil || len(*triggers) == 0 {
		return nil
	}

	var cdArtifactTriggers []interface{}
	var scheduleTriggers []interface{}
	var sourceRepoTriggers []interface{}

	for _, raw := range *triggers {
		// Marshal to JSON and back to get a clean map[string]interface{}
		data, err := json.Marshal(raw)
		if err != nil {
			continue
		}
		var trigMap map[string]interface{}
		if err := json.Unmarshal(data, &trigMap); err != nil {
			continue
		}

		trigType, _ := trigMap["triggerType"].(string)
		switch trigType {
		case "artifactSource":
			ct := map[string]interface{}{
				"artifact_alias":                  "",
				"use_build_definition_branch":     false,
				"create_release_on_build_tagging": false,
			}
			if alias, ok := trigMap["artifactAlias"].(string); ok {
				ct["artifact_alias"] = alias
			}
			// Flatten triggerConditions → branch_filter include list + new fields
			bf, tagFilter, useBuildBranch, createOnTagging := flattenArtifactTriggerConditions(trigMap)
			ct["branch_filter"] = bf
			if tagFilter != nil {
				ct["tag_filter"] = []interface{}{tagFilter}
			} else {
				ct["tag_filter"] = []interface{}{}
			}
			ct["use_build_definition_branch"] = useBuildBranch
			ct["create_release_on_build_tagging"] = createOnTagging
			cdArtifactTriggers = append(cdArtifactTriggers, ct)

		case "schedule":
			st := map[string]interface{}{
				"schedule_only_with_changes": false,
				"start_hours":                0,
				"start_minutes":              0,
				"time_zone_id":               "UTC",
				"days_to_release":            127,
			}
			if sched, ok := trigMap["schedule"].(map[string]interface{}); ok {
				if v, ok := sched["scheduleOnlyWithChanges"].(bool); ok {
					st["schedule_only_with_changes"] = v
				}
				if v, ok := sched["startHours"].(float64); ok {
					st["start_hours"] = int(v)
				}
				if v, ok := sched["startMinutes"].(float64); ok {
					st["start_minutes"] = int(v)
				}
				if v, ok := sched["timeZoneId"].(string); ok {
					st["time_zone_id"] = v
				}
				if v, ok := sched["daysToRelease"].(float64); ok {
					st["days_to_release"] = int(v)
				}
			}
			// branch_filter was removed from schedule_trigger schema (AC1/WI-9):
			// ADO does not return branchFilters for schedule triggers, so we never
			// populate branch_filter in state (it would perpetually diff otherwise).
			scheduleTriggers = append(scheduleTriggers, st)

		case "sourceRepo":
			srt := map[string]interface{}{
				"alias":          "",
				"branch_filters": []interface{}{},
			}
			if alias, ok := trigMap["alias"].(string); ok {
				srt["alias"] = alias
			}
			if bfs, ok := trigMap["branchFilters"].([]interface{}); ok {
				srt["branch_filters"] = bfs
			}
			sourceRepoTriggers = append(sourceRepoTriggers, srt)
		}
	}

	triggersMap := map[string]interface{}{
		"cd_artifact_trigger": cdArtifactTriggers,
		"schedule_trigger":    scheduleTriggers,
		"source_repo_trigger": sourceRepoTriggers,
	}
	return []interface{}{triggersMap}
}

// flattenArtifactTriggerConditions converts a cd artifact trigger's triggerConditions
// into:
//   - branch_filter block ([]interface{})
//   - tag_filter map (map[string]interface{}) or nil
//   - use_build_definition_branch bool
//   - create_release_on_build_tagging bool
func flattenArtifactTriggerConditions(trigMap map[string]interface{}) ([]interface{}, map[string]interface{}, bool, bool) {
	var includes []interface{}
	var tagFilter map[string]interface{}
	var useBuildBranch, createOnTagging bool

	if conditions, ok := trigMap["triggerConditions"].([]interface{}); ok {
		for i, c := range conditions {
			if cMap, ok := c.(map[string]interface{}); ok {
				if branch, ok := cMap["sourceBranch"].(string); ok && branch != "" {
					includes = append(includes, branch)
				}
				// Only read extras from the first condition entry. ADO stores
				// build-tag filtering as the `tags` array on the condition;
				// the regex `tagFilter` field is never persisted by the
				// service, so only `tags` is read back.
				if i == 0 {
					var tags []interface{}
					if v, ok := cMap["tags"].([]interface{}); ok {
						for _, t := range v {
							if s, ok := t.(string); ok {
								tags = append(tags, s)
							}
						}
					}
					if len(tags) > 0 {
						tagFilter = map[string]interface{}{
							"tags": tags,
						}
					}
					if v, ok := cMap["useBuildDefinitionBranch"].(bool); ok {
						useBuildBranch = v
					}
					if v, ok := cMap["createReleaseOnBuildTagging"].(bool); ok {
						createOnTagging = v
					}
				}
			}
		}
	}

	// Only emit a branch_filter block when there are actual include entries.
	// Returning nil (no block) when empty matches HCL configs that omit branch_filter
	// entirely, preventing a perpetual idempotency diff where state has an empty
	// block but the config does not declare one.
	var bf []interface{}
	if len(includes) > 0 {
		bf = []interface{}{map[string]interface{}{
			"include": includes,
			"exclude": []interface{}{},
		}}
	}
	return bf, tagFilter, useBuildBranch, createOnTagging
}

func flattenArtifacts(artifacts *[]releaseapi.Artifact, d *schema.ResourceData) []interface{} {
	if artifacts == nil {
		return nil
	}

	// Build a set of user-configured definition_reference keys per artifact index
	// so we can filter out API-computed keys like "artifactSourceDefinitionUrl".
	configuredKeys := map[int]map[string]bool{}
	if v, ok := d.GetOk("artifact"); ok {
		for i, raw := range v.([]interface{}) {
			artMap := raw.(map[string]interface{})
			if dr, ok := artMap["definition_reference"].(map[string]interface{}); ok {
				keys := make(map[string]bool, len(dr))
				for k := range dr {
					keys[k] = true
				}
				configuredKeys[i] = keys
			}
		}
	}

	result := make([]interface{}, len(*artifacts))
	for i, a := range *artifacts {
		artMap := map[string]interface{}{
			"alias":      "",
			"type":       "",
			"is_primary": false,
		}
		if a.Alias != nil {
			artMap["alias"] = *a.Alias
		}
		if a.Type != nil {
			artMap["type"] = *a.Type
		}
		if a.IsPrimary != nil {
			artMap["is_primary"] = *a.IsPrimary
		}
		if a.DefinitionReference != nil {
			defRef := make(map[string]string)
			allowed := configuredKeys[i]
			for k, v := range *a.DefinitionReference {
				// Only include keys the user configured; skip API-computed keys
				if allowed != nil && !allowed[k] {
					continue
				}
				if v.Id != nil {
					defRef[k] = *v.Id
				}
			}
			artMap["definition_reference"] = defRef
		}
		result[i] = artMap
	}
	return result
}
