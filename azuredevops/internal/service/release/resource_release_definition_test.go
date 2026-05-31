//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Package-level fixtures ─────────────────────────────────────────────────

var testReleaseDefinitionProjectID = uuid.New()
var testReleaseDefinitionID = 42
var testReleaseDefinitionRevision = 1
var testReleaseDefinitionEnvID = 1
var testReleaseDefinitionEnvRank = 1
var testReleaseDefinitionWorkflowTaskID = uuid.New()

var testReleaseDefinition = releaseapi.ReleaseDefinition{
	Id:                &testReleaseDefinitionID,
	Name:              converter.String("MyReleaseDefinition"),
	Path:              converter.String("\\"),
	Description:       converter.String("A test release definition"),
	ReleaseNameFormat: converter.String("Release-$(rev:r)"),
	Revision:          &testReleaseDefinitionRevision,
	Variables: &map[string]releaseapi.ConfigurationVariableValue{
		"MY_VAR": {
			Value:         converter.String("my-value"),
			IsSecret:      converter.Bool(false),
			AllowOverride: converter.Bool(false),
		},
	},
	Environments: &[]releaseapi.ReleaseDefinitionEnvironment{
		{
			Id:   &testReleaseDefinitionEnvID,
			Name: converter.String("Production"),
			Rank: &testReleaseDefinitionEnvRank,
			DeployPhases: &[]interface{}{
				map[string]interface{}{
					"name":      "Agent phase",
					"rank":      float64(1),
					"phaseType": "agentBasedDeployment",
					"workflowTasks": []interface{}{
						map[string]interface{}{
							"name":            "Run Script",
							"taskId":          testReleaseDefinitionWorkflowTaskID.String(),
							"version":         "1.*",
							"enabled":         true,
							"alwaysRun":       false,
							"continueOnError": false,
							"condition":       "succeeded()",
							"definitionType":  "task",
						},
					},
				},
			},
		},
	},
	Artifacts: &[]releaseapi.Artifact{
		{
			Alias:     converter.String("_myBuild"),
			Type:      converter.String("Build"),
			IsPrimary: converter.Bool(true),
			DefinitionReference: &map[string]releaseapi.ArtifactSourceReference{
				"definition": {Id: converter.String("1")},
				"project":    {Id: converter.String(testReleaseDefinitionProjectID.String())},
			},
		},
	},
}

// ── 1. Roundtrip ──────────────────────────────────────────────────────────

// TestReleaseDefinition_ExpandFlatten_Roundtrip verifies that flattenReleaseDefinition
// followed by expandReleaseDefinition preserves key fields of a ReleaseDefinition.
func TestReleaseDefinition_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                  0,
				"name":                "placeholder",
				"rank":                1,
				"owner":               "",
				"variable":            []interface{}{},
				"variable_groups":     []interface{}{},
				"condition":           []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Agent phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      1,
								"demands":                       []interface{}{},
								"timeout_in_minutes":            0,
								"job_cancel_timeout_in_minutes": 1,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{
							map[string]interface{}{
								"name":              "Run Script",
								"task_id":           testReleaseDefinitionWorkflowTaskID.String(),
								"version":           "1.*",
								"enabled":           true,
								"always_run":        false,
								"continue_on_error": false,
								"condition":         "succeeded()",
								"definition_type":   "task",
								"inputs":            map[string]interface{}{},
							},
						},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})

	flattenReleaseDefinition(resourceData, &testReleaseDefinition, testReleaseDefinitionProjectID.String())

	result, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.Equal(t, converter.ToString(testReleaseDefinition.Name, ""), converter.ToString(result.Name, ""))
	require.Equal(t, converter.ToString(testReleaseDefinition.Path, ""), converter.ToString(result.Path, ""))
	require.Equal(t, converter.ToString(testReleaseDefinition.ReleaseNameFormat, ""), converter.ToString(result.ReleaseNameFormat, ""))
}

// ── 2. Create error propagation ────────────────────────────────────────────

// TestReleaseDefinition_Create_DoesNotSwallowError verifies that an error from
// CreateReleaseDefinition surfaces as a non-empty Diagnostics.
func TestReleaseDefinition_Create_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	releaseClient.
		EXPECT().
		CreateReleaseDefinition(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("CreateReleaseDefinition() Failed")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                  0,
				"name":                "Production",
				"rank":                1,
				"owner":               "",
				"variable":            []interface{}{},
				"variable_groups":     []interface{}{},
				"condition":           []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Agent phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      1,
								"demands":                       []interface{}{},
								"timeout_in_minutes":            0,
								"job_cancel_timeout_in_minutes": 1,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})

	diags := resourceReleaseDefinitionCreate(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "CreateReleaseDefinition() Failed")
}

// ── 3. Read clears ID on 404 ───────────────────────────────────────────────

// TestReleaseDefinition_Read_ClearsIdOn404 verifies that when GetReleaseDefinition
// returns a 404 WrappedError, resourceReleaseDefinitionRead clears the resource ID
// and returns no diagnostics (graceful drift detection).
func TestReleaseDefinition_Read_ClearsIdOn404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	notFoundStatusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &notFoundStatusCode,
	}

	releaseClient.
		EXPECT().
		GetReleaseDefinition(clients.Ctx, gomock.Any()).
		Return(nil, notFoundErr).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                  0,
				"name":                "Production",
				"rank":                1,
				"owner":               "",
				"variable":            []interface{}{},
				"variable_groups":     []interface{}{},
				"condition":           []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Agent phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      1,
								"demands":                       []interface{}{},
								"timeout_in_minutes":            0,
								"job_cancel_timeout_in_minutes": 1,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	diags := resourceReleaseDefinitionRead(context.Background(), resourceData, clients)
	require.Empty(t, diags)
	require.Equal(t, "", resourceData.Id())
}

// ── 4. Update calls SDK with args ──────────────────────────────────────────

// TestReleaseDefinition_Update_CallsSDKWithArgs verifies that resourceReleaseDefinitionUpdate
// calls UpdateReleaseDefinition exactly once and then re-reads via GetReleaseDefinition.
func TestReleaseDefinition_Update_CallsSDKWithArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	updatedDef := testReleaseDefinition

	releaseClient.
		EXPECT().
		UpdateReleaseDefinition(clients.Ctx, gomock.Any()).
		Return(&updatedDef, nil).
		Times(1)

	releaseClient.
		EXPECT().
		GetReleaseDefinition(clients.Ctx, gomock.Any()).
		Return(&updatedDef, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            testReleaseDefinitionRevision,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                  0,
				"name":                "Production",
				"rank":                1,
				"owner":               "",
				"variable":            []interface{}{},
				"variable_groups":     []interface{}{},
				"condition":           []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Agent phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      1,
								"demands":                       []interface{}{},
								"timeout_in_minutes":            0,
								"job_cancel_timeout_in_minutes": 1,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	diags := resourceReleaseDefinitionUpdate(context.Background(), resourceData, clients)
	require.Empty(t, diags)
}

// ── 5. Delete surfaces API error ───────────────────────────────────────────

// TestReleaseDefinition_Delete_SurfacesAPIError verifies that an error from
// DeleteReleaseDefinition is surfaced as non-empty Diagnostics.
func TestReleaseDefinition_Delete_SurfacesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	releaseClient.
		EXPECT().
		DeleteReleaseDefinition(clients.Ctx, gomock.Any()).
		Return(errors.New("DeleteReleaseDefinition() Failed")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                  0,
				"name":                "Production",
				"rank":                1,
				"owner":               "",
				"variable":            []interface{}{},
				"variable_groups":     []interface{}{},
				"condition":           []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Agent phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      1,
								"demands":                       []interface{}{},
								"timeout_in_minutes":            0,
								"job_cancel_timeout_in_minutes": 1,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	diags := resourceReleaseDefinitionDelete(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "DeleteReleaseDefinition() Failed")
}
