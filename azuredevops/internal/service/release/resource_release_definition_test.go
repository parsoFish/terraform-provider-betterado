//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"context"
	"encoding/json"
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
				"id":                   0,
				"name":                 "placeholder",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
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
				"id":                   0,
				"name":                 "Production",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
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
				"id":                   0,
				"name":                 "Production",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
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
				"id":                   0,
				"name":                 "Production",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	diags := resourceReleaseDefinitionUpdate(context.Background(), resourceData, clients)
	require.Empty(t, diags)
}

// ── 6. Update revision-conflict retry ─────────────────────────────────────

// TestReleaseDefinition_Update_RevisionRetryOnConflict verifies that when
// UpdateReleaseDefinition returns a "old copy of the release pipeline" error,
// resourceReleaseDefinitionUpdate re-reads via GetReleaseDefinition (once, for
// the fresh revision) and then retries UpdateReleaseDefinition a second time.
func TestReleaseDefinition_Update_RevisionRetryOnConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	conflictMessage := "old copy of the release pipeline"
	conflictErr := azuredevops.WrappedError{
		Message:    &conflictMessage,
		StatusCode: func() *int { code := 400; return &code }(),
	}

	freshRevision := 2
	freshDef := testReleaseDefinition
	freshDef.Revision = &freshRevision

	// First UpdateReleaseDefinition fails with revision conflict
	gomock.InOrder(
		releaseClient.
			EXPECT().
			UpdateReleaseDefinition(clients.Ctx, gomock.Any()).
			Return(nil, conflictErr).
			Times(1),
		// GetReleaseDefinition re-reads the fresh revision (once, for conflict handling)
		releaseClient.
			EXPECT().
			GetReleaseDefinition(clients.Ctx, gomock.Any()).
			Return(&freshDef, nil).
			Times(1),
		// Second UpdateReleaseDefinition succeeds with fresh revision
		releaseClient.
			EXPECT().
			UpdateReleaseDefinition(clients.Ctx, gomock.Any()).
			Return(&freshDef, nil).
			Times(1),
		// Final GetReleaseDefinition called by resourceReleaseDefinitionRead
		releaseClient.
			EXPECT().
			GetReleaseDefinition(clients.Ctx, gomock.Any()).
			Return(&freshDef, nil).
			Times(1),
	)

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
				"id":                   0,
				"name":                 "Production",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	diags := resourceReleaseDefinitionUpdate(context.Background(), resourceData, clients)
	require.Empty(t, diags)
}

// ── 7. Secret variables preserve value on flatten ─────────────────────────

// TestReleaseDefinition_SecretVariables_PreserveOnFlatten verifies that when the
// API returns a secret variable with a nil value, flattenReleaseDefinition
// preserves the value already held in Terraform state rather than overwriting it
// with an empty string.
func TestReleaseDefinition_SecretVariables_PreserveOnFlatten(t *testing.T) {
	secretVarName := "MY_SECRET"
	existingSecretValue := "super-secret-value"

	// Set up ResourceData with existing state value for the secret variable
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable": []interface{}{
			map[string]interface{}{
				"name":           secretVarName,
				"value":          existingSecretValue,
				"is_secret":      true,
				"allow_override": false,
			},
		},
		"variable_groups": []interface{}{},
		"tags":            []interface{}{},
		"environment":     []interface{}{},
		"artifact":        []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	// API returns the definition with the secret variable having nil value
	apiDef := testReleaseDefinition
	apiVariables := map[string]releaseapi.ConfigurationVariableValue{
		secretVarName: {
			Value:         nil, // secret variables return nil from the API
			IsSecret:      converter.Bool(true),
			AllowOverride: converter.Bool(false),
		},
	}
	apiDef.Variables = &apiVariables
	// Remove environments to keep the test focused on variables
	emptyEnvs := []releaseapi.ReleaseDefinitionEnvironment{}
	apiDef.Environments = &emptyEnvs
	emptyArtifacts := []releaseapi.Artifact{}
	apiDef.Artifacts = &emptyArtifacts

	flattenReleaseDefinition(resourceData, &apiDef, testReleaseDefinitionProjectID.String())

	// After flatten, retrieve the variable set and verify the secret value was preserved
	vars, ok := resourceData.GetOk("variable")
	require.True(t, ok, "variable set should be non-empty after flatten")

	varList := vars.(*schema.Set).List()
	require.Len(t, varList, 1)

	varMap := varList[0].(map[string]interface{})
	require.Equal(t, secretVarName, varMap["name"])
	require.Equal(t, existingSecretValue, varMap["value"],
		"secret variable value must be preserved from state, not overwritten with empty string")
	require.Equal(t, true, varMap["is_secret"])
}

// ── 8. Deep-nested environment expand/flatten round-trip ─────────────────

// TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten verifies that an
// environment with deploy_phase → deployment_input (queue ID, demands) →
// workflow_task (inputs map) round-trips intact through expandReleaseDefinition
// followed by flattenReleaseDefinition.
func TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten(t *testing.T) {
	taskInputKey := "scriptPath"
	taskInputValue := "./scripts/deploy.sh"
	taskDisplayName := "Deploy Script"
	queueID := 7
	demand1 := "Agent.OS -equals Linux"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "DeepNestedDef",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":                   0,
				"name":                 "Staging",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":       "Deploy Phase",
						"rank":       1,
						"phase_type": "agentBasedDeployment",
						"deployment_input": []interface{}{
							map[string]interface{}{
								"queue_id":                      queueID,
								"demands":                       []interface{}{demand1},
								"timeout_in_minutes":            10,
								"job_cancel_timeout_in_minutes": 5,
								"condition":                     "succeeded()",
								"skip_artifacts_download":       false,
								"enable_access_token":           false,
								"agent_specification":           "",
							},
						},
						"workflow_task": []interface{}{
							map[string]interface{}{
								"name":              taskDisplayName,
								"task_id":           testReleaseDefinitionWorkflowTaskID.String(),
								"version":           "2.*",
								"enabled":           true,
								"always_run":        false,
								"continue_on_error": false,
								"condition":         "succeeded()",
								"definition_type":   "task",
								"inputs": map[string]interface{}{
									taskInputKey: taskInputValue,
								},
							},
						},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})

	// Expand → flatten round-trip
	expanded, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.NotNil(t, expanded.Environments)
	require.Len(t, *expanded.Environments, 1)

	// Set an ID so flattenReleaseDefinition can call SetId
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID

	flattenReleaseDefinition(resourceData, expanded, projectID)

	// Verify deploy phase round-trip
	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	phases := envMap["deploy_phase"].([]interface{})
	require.Len(t, phases, 1)
	phase := phases[0].(map[string]interface{})

	// Verify workflow task display name and task ID survive the round-trip
	wfTasks, ok := phase["workflow_task"].([]interface{})
	require.True(t, ok, "workflow_task should be present")
	require.Len(t, wfTasks, 1)
	task := wfTasks[0].(map[string]interface{})
	require.Equal(t, taskDisplayName, task["name"])
	require.Equal(t, testReleaseDefinitionWorkflowTaskID.String(), task["task_id"])

	// Verify workflow task input key/value pair survives.
	// Note: Terraform SDK normalises TypeMap(TypeString) → map[string]interface{} on Get,
	// so we accept either map[string]interface{} or map[string]string.
	switch inputsTyped := task["inputs"].(type) {
	case map[string]interface{}:
		val, ok := inputsTyped[taskInputKey]
		require.True(t, ok, "input key %q must be present", taskInputKey)
		require.Equal(t, taskInputValue, val.(string))
	case map[string]string:
		require.Equal(t, taskInputValue, inputsTyped[taskInputKey])
	default:
		require.Fail(t, "inputs must be map[string]interface{} or map[string]string, got %T", task["inputs"])
	}
}

// ── 9. Artifacts: API-computed keys are filtered out ─────────────────────

// TestReleaseDefinition_Artifacts_DefinitionReferenceFiltering verifies that
// flattenArtifacts strips API-computed keys (e.g. artifactSourceDefinitionUrl)
// from definitionReference, keeping only the keys the user originally configured.
func TestReleaseDefinition_Artifacts_DefinitionReferenceFiltering(t *testing.T) {
	// Set up ResourceData with user-configured artifact keys
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "ArtifactFilterDef",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment":         []interface{}{},
		"artifact": []interface{}{
			map[string]interface{}{
				"alias":      "_myBuild",
				"type":       "Build",
				"is_primary": true,
				"definition_reference": map[string]interface{}{
					"definition": "1",
					"project":    testReleaseDefinitionProjectID.String(),
				},
			},
		},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	// API response includes an extra computed key "artifactSourceDefinitionUrl"
	apiArtifacts := []releaseapi.Artifact{
		{
			Alias:     converter.String("_myBuild"),
			Type:      converter.String("Build"),
			IsPrimary: converter.Bool(true),
			DefinitionReference: &map[string]releaseapi.ArtifactSourceReference{
				"definition":                  {Id: converter.String("1")},
				"project":                     {Id: converter.String(testReleaseDefinitionProjectID.String())},
				"artifactSourceDefinitionUrl": {Id: converter.String("https://dev.azure.com/...")},
			},
		},
	}

	flattened := flattenArtifacts(&apiArtifacts, resourceData)
	require.Len(t, flattened, 1)

	artMap := flattened[0].(map[string]interface{})
	defRef, ok := artMap["definition_reference"].(map[string]string)
	require.True(t, ok, "definition_reference should be map[string]string")

	// User-configured keys must be present
	require.Equal(t, "1", defRef["definition"])
	require.Equal(t, testReleaseDefinitionProjectID.String(), defRef["project"])

	// Computed key must be absent
	_, hasComputedKey := defRef["artifactSourceDefinitionUrl"]
	require.False(t, hasComputedKey, "API-computed key artifactSourceDefinitionUrl must not appear in Terraform state")
}

// ── 10. Approval options round-trip ────────────────────────────────────────

// TestReleaseDefinition_ApprovalOptions_RoundTrip verifies that an environment
// with pre_deploy_approval (one approver + ApprovalOptions) and a post_deploy_approval
// block round-trips through expandReleaseDefinition → flattenReleaseDefinition
// with approver count, release_creator_can_be_approver, and required_approver_count
// intact.
func TestReleaseDefinition_ApprovalOptions_RoundTrip(t *testing.T) {
	approverUUID := uuid.New()
	requiredApproverCount := 1

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "ApprovalRoundTripDef",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{},
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":              0,
				"name":            "Production",
				"rank":            1,
				"owner":           "",
				"variable":        []interface{}{},
				"variable_groups": []interface{}{},
				"condition":       []interface{}{},
				"pre_deploy_approval": []interface{}{
					map[string]interface{}{
						"approver": []interface{}{
							map[string]interface{}{
								"id":           approverUUID.String(),
								"is_automated": false,
								"rank":         1,
							},
						},
						"approval_options": []interface{}{
							map[string]interface{}{
								"required_approver_count":         requiredApproverCount,
								"release_creator_can_be_approver": true,
								"enforce_identity_revalidation":   false,
								"timeout_in_minutes":              0,
								"execution_order":                 "beforeGates",
								"auto_triggered_and_previous_environment_approved_can_be_skipped": false,
							},
						},
					},
				},
				"post_deploy_approval": []interface{}{
					map[string]interface{}{
						"approver":         []interface{}{},
						"approval_options": []interface{}{},
					},
				},
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
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})

	// Expand to API object
	expanded, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.NotNil(t, expanded.Environments)

	env := (*expanded.Environments)[0]
	require.NotNil(t, env.PreDeployApprovals)
	require.NotNil(t, env.PreDeployApprovals.Approvals)
	require.Len(t, *env.PreDeployApprovals.Approvals, 1)
	require.NotNil(t, env.PreDeployApprovals.ApprovalOptions)
	require.Equal(t, true, *env.PreDeployApprovals.ApprovalOptions.ReleaseCreatorCanBeApprover)
	require.Equal(t, requiredApproverCount, *env.PreDeployApprovals.ApprovalOptions.RequiredApproverCount)

	// Flatten back to ResourceData
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	// Verify the pre_deploy_approval round-trip
	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	preApproval := envMap["pre_deploy_approval"].([]interface{})
	require.Len(t, preApproval, 1)
	approvalMap := preApproval[0].(map[string]interface{})

	approvers := approvalMap["approver"].([]interface{})
	require.Len(t, approvers, 1, "approver count must survive round-trip")
	approverMap := approvers[0].(map[string]interface{})
	require.Equal(t, approverUUID.String(), approverMap["id"])

	approvalOpts := approvalMap["approval_options"].([]interface{})
	require.Len(t, approvalOpts, 1)
	opts := approvalOpts[0].(map[string]interface{})
	require.Equal(t, true, opts["release_creator_can_be_approver"],
		"release_creator_can_be_approver must survive round-trip")
	require.Equal(t, requiredApproverCount, opts["required_approver_count"],
		"required_approver_count must survive round-trip")
}

// ── 11. Deploy phases: JSON marshal/unmarshal survives DisplayName ─────────

// TestReleaseDefinition_DeployPhases_JSONMarshalUnmarshal verifies that
// flattenDeployPhases correctly unmarshals DeploymentInput from interface{} via
// the JSON round-trip path, and that the workflow task's DisplayName survives.
func TestReleaseDefinition_DeployPhases_JSONMarshalUnmarshal(t *testing.T) {
	displayName := "Run Integration Tests"
	taskIDStr := testReleaseDefinitionWorkflowTaskID.String()

	// Construct a deploy phase as an AgentBasedDeployPhase struct, then marshal it
	// to interface{} (simulating what the API SDK does).
	phaseType := releaseapi.DeployPhaseTypes("agentBasedDeployment")
	queueID := 3
	phase := releaseapi.AgentBasedDeployPhase{
		Name:      converter.String("Test Phase"),
		Rank:      converter.Int(1),
		PhaseType: &phaseType,
		DeploymentInput: &releaseapi.AgentDeploymentInput{
			QueueId: converter.Int(queueID),
		},
		WorkflowTasks: &[]releaseapi.WorkflowTask{
			{
				Name:    converter.String(displayName),
				TaskId:  &testReleaseDefinitionWorkflowTaskID,
				Version: converter.String("3.*"),
				Enabled: converter.Bool(true),
			},
		},
	}

	// Marshal to JSON and back to interface{} — simulates the SDK's decode path
	phaseBytes, err := json.Marshal(phase)
	require.NoError(t, err)
	var phaseIface interface{}
	err = json.Unmarshal(phaseBytes, &phaseIface)
	require.NoError(t, err)

	phases := []interface{}{phaseIface}
	result := flattenDeployPhases(&phases)
	require.Len(t, result, 1)

	flatPhase := result[0].(map[string]interface{})
	wfTasks, ok := flatPhase["workflow_task"].([]interface{})
	require.True(t, ok, "workflow_task should be present")
	require.Len(t, wfTasks, 1)

	task := wfTasks[0].(map[string]interface{})
	require.Equal(t, displayName, task["name"], "workflow task DisplayName must survive JSON marshal/unmarshal round-trip")
	require.Equal(t, taskIDStr, task["task_id"], "workflow task ID must survive JSON marshal/unmarshal round-trip")
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
				"id":                   0,
				"name":                 "Production",
				"rank":                 1,
				"owner":                "",
				"variable":             []interface{}{},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
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
				"environment_options": []interface{}{},
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
