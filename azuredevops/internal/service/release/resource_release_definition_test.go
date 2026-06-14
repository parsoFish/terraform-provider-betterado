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

// TestReleaseDefinition_EnvSecretVariables_PreserveOnFlatten verifies that an
// ENVIRONMENT-scoped secret variable (returned null by the API) is preserved
// from the matching environment.<i>.variable state path on flatten. This guards
// the env-level secret perpetual-diff defect: flattenVariables previously read
// the definition-level "variable" key regardless of scope, losing env secrets.
func TestReleaseDefinition_EnvSecretVariables_PreserveOnFlatten(t *testing.T) {
	secretVarName := "ENV_SECRET"
	existingSecretValue := "env-super-secret"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "MyReleaseDefinition",
		"path":                "\\",
		"description":         "",
		"release_name_format": "Release-$(rev:r)",
		"revision":            0,
		"variable":            []interface{}{}, // no definition-level secret
		"variable_groups":     []interface{}{},
		"tags":                []interface{}{},
		"environment": []interface{}{
			map[string]interface{}{
				"id":    0,
				"name":  "Prod",
				"rank":  1,
				"owner": "",
				"variable": []interface{}{
					map[string]interface{}{
						"name":           secretVarName,
						"value":          existingSecretValue,
						"is_secret":      true,
						"allow_override": false,
					},
				},
				"variable_groups":      []interface{}{},
				"condition":            []interface{}{},
				"pre_deploy_approval":  []interface{}{},
				"post_deploy_approval": []interface{}{},
				"deploy_phase":         []interface{}{},
				"retention_policy":     []interface{}{},
				"environment_options":  []interface{}{},
				"execution_policy":     []interface{}{},
			},
		},
		"artifact": []interface{}{},
	})
	resourceData.SetId(strconv.Itoa(testReleaseDefinitionID))

	// API returns the environment with the secret variable having a nil value.
	apiDef := testReleaseDefinition
	envID := 100
	apiEnvVariables := map[string]releaseapi.ConfigurationVariableValue{
		secretVarName: {
			Value:         nil, // secret env variables return nil from the API
			IsSecret:      converter.Bool(true),
			AllowOverride: converter.Bool(false),
		},
	}
	apiDef.Environments = &[]releaseapi.ReleaseDefinitionEnvironment{
		{
			Id:        &envID,
			Name:      converter.String("Prod"),
			Rank:      converter.Int(1),
			Variables: &apiEnvVariables,
		},
	}
	emptyVars := map[string]releaseapi.ConfigurationVariableValue{}
	apiDef.Variables = &emptyVars
	emptyArtifacts := []releaseapi.Artifact{}
	apiDef.Artifacts = &emptyArtifacts

	flattenReleaseDefinition(resourceData, &apiDef, testReleaseDefinitionProjectID.String())

	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	envVars := envMap["variable"].(*schema.Set).List()
	require.Len(t, envVars, 1)
	varMap := envVars[0].(map[string]interface{})
	require.Equal(t, secretVarName, varMap["name"])
	require.Equal(t, existingSecretValue, varMap["value"],
		"env-scoped secret value must be preserved from environment.0.variable state, not lost")
	require.Equal(t, true, varMap["is_secret"])
}

// TestRollbackRedeployErrorHint verifies the TF400898 error is enriched with a
// hint, while unrelated errors and nil pass through untouched.
func TestRollbackRedeployErrorHint(t *testing.T) {
	require.NoError(t, rollbackRedeployErrorHint(nil))

	other := errors.New("some unrelated failure")
	require.Equal(t, other, rollbackRedeployErrorHint(other))

	tf := errors.New("VS800075: TF400898: An internal error occurred")
	enriched := rollbackRedeployErrorHint(tf)
	require.Contains(t, enriched.Error(), "TF400898")
	require.Contains(t, enriched.Error(), "rollbackRedeploy")
	require.Contains(t, enriched.Error(), "successful deployment")
}

// TestReleaseDefinition_WorkflowTaskTimeoutRetry verifies the WI-C fields
// (timeout_in_minutes, retry_count_on_task_failure) expand into the WorkflowTask
// struct and flatten back from it.
func TestReleaseDefinition_WorkflowTaskTimeoutRetry(t *testing.T) {
	in := []interface{}{
		map[string]interface{}{
			"name":                        "Deploy",
			"task_id":                     testReleaseDefinitionWorkflowTaskID.String(),
			"version":                     "2.*",
			"enabled":                     true,
			"always_run":                  false,
			"continue_on_error":           false,
			"condition":                   "succeeded()",
			"definition_type":             "task",
			"inputs":                      map[string]interface{}{},
			"timeout_in_minutes":          15,
			"retry_count_on_task_failure": 3,
		},
	}

	tasks := expandWorkflowTasks(in)
	require.Len(t, tasks, 1)
	require.NotNil(t, tasks[0].TimeoutInMinutes)
	require.Equal(t, 15, *tasks[0].TimeoutInMinutes)
	require.NotNil(t, tasks[0].RetryCountOnTaskFailure)
	require.Equal(t, 3, *tasks[0].RetryCountOnTaskFailure)

	flat := flattenWorkflowTasksFromAPI(&tasks)
	require.Len(t, flat, 1)
	fm := flat[0].(map[string]interface{})
	require.Equal(t, 15, fm["timeout_in_minutes"])
	require.Equal(t, 3, fm["retry_count_on_task_failure"])
}

// TestReleaseDefinition_DeploymentInputOverrideInputs verifies the WI-D field
// (deployment_input.override_inputs) expands into the API payload and flattens
// back from an API-shaped response.
func TestReleaseDefinition_DeploymentInputOverrideInputs(t *testing.T) {
	in := []interface{}{
		map[string]interface{}{
			"queue_id":                      0,
			"timeout_in_minutes":            0,
			"job_cancel_timeout_in_minutes": 1,
			"condition":                     "succeeded()",
			"skip_artifacts_download":       false,
			"enable_access_token":           false,
			"agent_specification":           "",
			"override_inputs": map[string]interface{}{
				"ScriptPath": "./deploy.sh",
			},
		},
	}

	di := expandDeploymentInput(in)
	require.NotNil(t, di)
	overrides, ok := di["overrideInputs"].(map[string]string)
	require.True(t, ok, "overrideInputs should be a map[string]string")
	require.Equal(t, "./deploy.sh", overrides["ScriptPath"])

	// Flatten from an API-shaped response (JSON-decoded → map[string]interface{}).
	apiDI := map[string]interface{}{
		"overrideInputs": map[string]interface{}{
			"ScriptPath": "./deploy.sh",
		},
	}
	flat := flattenDeploymentInput(apiDI)
	require.Len(t, flat, 1)
	fm := flat[0].(map[string]interface{})
	oi, ok := fm["override_inputs"].(map[string]string)
	require.True(t, ok, "override_inputs should be present and a map[string]string")
	require.Equal(t, "./deploy.sh", oi["ScriptPath"])

	// A deployment_input carrying only override_inputs must not be suppressed.
	require.False(t, isDefaultDeploymentInput(flat))
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
	result := flattenDeployPhases(&phases, nil, 0)
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

// ── AccRefresh: retention_policy round-trip ────────────────────────────────

// TestReleaseDefinition_AccRefresh_RetentionPolicy verifies that an environment with a
// retention_policy block round-trips through expandReleaseDefinition → flattenReleaseDefinition
// with days_to_keep, releases_to_keep, and retain_build intact.
// This test uses the same HCL fixture shape as hclReleaseDefinitionBasic (AC1 of WI-1).
func TestReleaseDefinition_AccRefresh_RetentionPolicy(t *testing.T) {
	daysToKeep := 30
	releasesToKeep := 3

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "RetentionPolicyDef",
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
						"name":             "Agent job",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy": []interface{}{
					map[string]interface{}{
						"days_to_keep":     daysToKeep,
						"releases_to_keep": releasesToKeep,
						"retain_build":     true,
					},
				},
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
	require.Len(t, *expanded.Environments, 1)

	env := (*expanded.Environments)[0]
	require.NotNil(t, env.RetentionPolicy, "retention_policy must be expanded to API object")
	require.Equal(t, daysToKeep, *env.RetentionPolicy.DaysToKeep)
	require.Equal(t, releasesToKeep, *env.RetentionPolicy.ReleasesToKeep)
	require.Equal(t, true, *env.RetentionPolicy.RetainBuild)

	// Flatten back to ResourceData
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	// Verify retention_policy round-trip
	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	retPolicyList, ok := envMap["retention_policy"].([]interface{})
	require.True(t, ok, "retention_policy should be present after flatten")
	require.Len(t, retPolicyList, 1)

	retPolicy := retPolicyList[0].(map[string]interface{})
	require.Equal(t, daysToKeep, retPolicy["days_to_keep"],
		"days_to_keep must survive expand/flatten round-trip")
	require.Equal(t, releasesToKeep, retPolicy["releases_to_keep"],
		"releases_to_keep must survive expand/flatten round-trip")
	require.Equal(t, true, retPolicy["retain_build"],
		"retain_build must survive expand/flatten round-trip")
}

// ── AccRefresh: pre_deploy_approval automated approver round-trip ──────────

// TestReleaseDefinition_AccRefresh_PreDeployApproval verifies that an environment with a
// minimal automated pre_deploy_approval block round-trips through
// expandReleaseDefinition → flattenReleaseDefinition with is_automated and rank intact.
// This test uses the same automated-approver shape as hclReleaseDefinitionBasic (AC1 of WI-1).
func TestReleaseDefinition_AccRefresh_PreDeployApproval(t *testing.T) {
	automatedApproverID := "00000000-0000-0000-0000-000000000000"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "PreDeployApprovalDef",
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
								"id":           automatedApproverID,
								"is_automated": true,
								"rank":         1,
							},
						},
						"approval_options": []interface{}{},
					},
				},
				"post_deploy_approval": []interface{}{},
				"deploy_phase": []interface{}{
					map[string]interface{}{
						"name":             "Agent job",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy": []interface{}{
					map[string]interface{}{
						"days_to_keep":     30,
						"releases_to_keep": 3,
						"retain_build":     true,
					},
				},
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
	require.Len(t, *expanded.Environments, 1)

	env := (*expanded.Environments)[0]
	require.NotNil(t, env.PreDeployApprovals, "pre_deploy_approval must be expanded to API object")
	require.NotNil(t, env.PreDeployApprovals.Approvals)
	require.Len(t, *env.PreDeployApprovals.Approvals, 1)

	step := (*env.PreDeployApprovals.Approvals)[0]
	require.NotNil(t, step.Approver)
	require.Equal(t, automatedApproverID, *step.Approver.Id)
	require.Equal(t, true, *step.IsAutomated)
	require.Equal(t, 1, *step.Rank)

	// Flatten back to ResourceData
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	// Verify pre_deploy_approval round-trip
	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	preApprovalList, ok := envMap["pre_deploy_approval"].([]interface{})
	require.True(t, ok, "pre_deploy_approval should be present after flatten")
	require.Len(t, preApprovalList, 1)

	approvalMap := preApprovalList[0].(map[string]interface{})
	approvers := approvalMap["approver"].([]interface{})
	require.Len(t, approvers, 1, "approver count must survive round-trip")

	approverMap := approvers[0].(map[string]interface{})
	require.Equal(t, automatedApproverID, approverMap["id"],
		"approver id must survive expand/flatten round-trip")
	require.Equal(t, true, approverMap["is_automated"],
		"is_automated must survive expand/flatten round-trip")
	require.Equal(t, 1, approverMap["rank"],
		"rank must survive expand/flatten round-trip")
}

// ── Gates: expand/flatten round-trip ──────────────────────────────────────

// TestReleaseDefinition_Gates_ExpandFlatten verifies that an environment with
// pre_deployment_gates and post_deployment_gates blocks (each containing a
// gates_options sub-block with all five fields) round-trips through
// expandReleaseDefinition → flattenReleaseDefinition with all GatesOptions
// fields correctly populated in both directions (AC1, AC2, AC3).
func TestReleaseDefinition_Gates_ExpandFlatten(t *testing.T) {
	timeout := 10
	samplingInterval := 5
	stabilizationTime := 3
	minimumSuccessDuration := 2

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "GatesRoundTripDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
				"pre_deployment_gates": []interface{}{
					map[string]interface{}{
						"gates_options": []interface{}{
							map[string]interface{}{
								"is_enabled":               true,
								"timeout":                  timeout,
								"sampling_interval":        samplingInterval,
								"stabilization_time":       stabilizationTime,
								"minimum_success_duration": minimumSuccessDuration,
							},
						},
					},
				},
				"post_deployment_gates": []interface{}{
					map[string]interface{}{
						"gates_options": []interface{}{
							map[string]interface{}{
								"is_enabled":               false,
								"timeout":                  timeout * 2,
								"sampling_interval":        samplingInterval * 2,
								"stabilization_time":       stabilizationTime * 2,
								"minimum_success_duration": minimumSuccessDuration * 2,
							},
						},
					},
				},
			},
		},
		"artifact": []interface{}{},
	})

	// AC1: Verify expand correctly populates PreDeploymentGates and PostDeploymentGates
	expanded, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.NotNil(t, expanded.Environments)
	require.Len(t, *expanded.Environments, 1)

	env := (*expanded.Environments)[0]

	// Pre-deployment gates assertions
	require.NotNil(t, env.PreDeploymentGates, "PreDeploymentGates must be set after expand")
	require.NotNil(t, env.PreDeploymentGates.GatesOptions, "PreDeploymentGates.GatesOptions must be set after expand")
	preOpts := env.PreDeploymentGates.GatesOptions
	require.Equal(t, true, *preOpts.IsEnabled, "pre IsEnabled must match")
	require.Equal(t, timeout, *preOpts.Timeout, "pre Timeout must match")
	require.Equal(t, samplingInterval, *preOpts.SamplingInterval, "pre SamplingInterval must match")
	require.Equal(t, stabilizationTime, *preOpts.StabilizationTime, "pre StabilizationTime must match")
	require.Equal(t, minimumSuccessDuration, *preOpts.MinimumSuccessDuration, "pre MinimumSuccessDuration must match")

	// Post-deployment gates assertions
	require.NotNil(t, env.PostDeploymentGates, "PostDeploymentGates must be set after expand")
	require.NotNil(t, env.PostDeploymentGates.GatesOptions, "PostDeploymentGates.GatesOptions must be set after expand")
	postOpts := env.PostDeploymentGates.GatesOptions
	require.Equal(t, false, *postOpts.IsEnabled, "post IsEnabled must match")
	require.Equal(t, timeout*2, *postOpts.Timeout, "post Timeout must match")
	require.Equal(t, samplingInterval*2, *postOpts.SamplingInterval, "post SamplingInterval must match")
	require.Equal(t, stabilizationTime*2, *postOpts.StabilizationTime, "post StabilizationTime must match")
	require.Equal(t, minimumSuccessDuration*2, *postOpts.MinimumSuccessDuration, "post MinimumSuccessDuration must match")

	// AC2: Flatten back and verify Terraform state contains correct blocks
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	// Verify pre_deployment_gates round-trip
	preGatesList, ok := envMap["pre_deployment_gates"].([]interface{})
	require.True(t, ok, "pre_deployment_gates should be present after flatten")
	require.Len(t, preGatesList, 1)
	preGatesMap := preGatesList[0].(map[string]interface{})

	preGatesOpts := preGatesMap["gates_options"].([]interface{})
	require.Len(t, preGatesOpts, 1)
	preOptsMap := preGatesOpts[0].(map[string]interface{})
	require.Equal(t, true, preOptsMap["is_enabled"], "pre is_enabled must survive flatten")
	require.Equal(t, timeout, preOptsMap["timeout"], "pre timeout must survive flatten")
	require.Equal(t, samplingInterval, preOptsMap["sampling_interval"], "pre sampling_interval must survive flatten")
	require.Equal(t, stabilizationTime, preOptsMap["stabilization_time"], "pre stabilization_time must survive flatten")
	require.Equal(t, minimumSuccessDuration, preOptsMap["minimum_success_duration"], "pre minimum_success_duration must survive flatten")

	// Verify post_deployment_gates round-trip
	postGatesList, ok := envMap["post_deployment_gates"].([]interface{})
	require.True(t, ok, "post_deployment_gates should be present after flatten")
	require.Len(t, postGatesList, 1)
	postGatesMap := postGatesList[0].(map[string]interface{})

	postGatesOpts := postGatesMap["gates_options"].([]interface{})
	require.Len(t, postGatesOpts, 1)
	postOptsMap := postGatesOpts[0].(map[string]interface{})
	require.Equal(t, false, postOptsMap["is_enabled"], "post is_enabled must survive flatten")
	require.Equal(t, timeout*2, postOptsMap["timeout"], "post timeout must survive flatten")
	require.Equal(t, samplingInterval*2, postOptsMap["sampling_interval"], "post sampling_interval must survive flatten")
	require.Equal(t, stabilizationTime*2, postOptsMap["stabilization_time"], "post stabilization_time must survive flatten")
	require.Equal(t, minimumSuccessDuration*2, postOptsMap["minimum_success_duration"], "post minimum_success_duration must survive flatten")
}

// ── 12b. Gates with tasks: gate {} blocks carrying task {} items ──────────

// TestReleaseDefinition_GatesTasks_ExpandFlatten verifies that pre/post_deployment_gates
// blocks that declare one or more gate {} blocks each carrying task {} items
// round-trip correctly through expandReleaseDefinition → flattenReleaseDefinition
// (AC1, AC2, AC3 for WI-6).
func TestReleaseDefinition_GatesTasks_ExpandFlatten(t *testing.T) {
	gateTaskID := "9a5d4f4e-f7d7-4e6a-a4c5-1b2d4e3f1234"
	gateTaskName := "InvokeRestAPI"
	gateTaskVersion := "0.*"
	gateTaskDefType := "serverGate"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "GatesTasksRoundTripDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
				"pre_deployment_gates": []interface{}{
					map[string]interface{}{
						"gates_options": []interface{}{
							map[string]interface{}{
								"is_enabled":               true,
								"timeout":                  600,
								"sampling_interval":        60,
								"stabilization_time":       0,
								"minimum_success_duration": 0,
							},
						},
						"gate": []interface{}{
							map[string]interface{}{
								"task": []interface{}{
									map[string]interface{}{
										"name":              gateTaskName,
										"task_id":           gateTaskID,
										"version":           gateTaskVersion,
										"enabled":           true,
										"always_run":        false,
										"continue_on_error": false,
										"condition":         "succeeded()",
										"definition_type":   gateTaskDefType,
										"inputs":            map[string]interface{}{},
									},
								},
							},
						},
					},
				},
				"post_deployment_gates": []interface{}{
					map[string]interface{}{
						"gates_options": []interface{}{
							map[string]interface{}{
								"is_enabled":               false,
								"timeout":                  0,
								"sampling_interval":        0,
								"stabilization_time":       0,
								"minimum_success_duration": 0,
							},
						},
						"gate": []interface{}{},
					},
				},
			},
		},
		"artifact": []interface{}{},
	})

	// ── AC1: expand populates Gates ──────────────────────────────────────────
	expanded, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.NotNil(t, expanded.Environments)
	require.Len(t, *expanded.Environments, 1)

	env := (*expanded.Environments)[0]

	// Pre-deployment gates: Gates array must be non-empty
	require.NotNil(t, env.PreDeploymentGates, "PreDeploymentGates must be set")
	require.NotNil(t, env.PreDeploymentGates.Gates, "PreDeploymentGates.Gates must be non-nil")
	require.Len(t, *env.PreDeploymentGates.Gates, 1, "one gate block expected")

	preGate := (*env.PreDeploymentGates.Gates)[0]
	require.NotNil(t, preGate.Tasks, "gate.Tasks must be non-nil")
	require.Len(t, *preGate.Tasks, 1, "one task expected in gate")

	preTask := (*preGate.Tasks)[0]
	require.NotNil(t, preTask.Name)
	require.Equal(t, gateTaskName, *preTask.Name, "task name must match")
	require.NotNil(t, preTask.TaskId)
	require.Equal(t, gateTaskID, preTask.TaskId.String(), "task_id must match")
	require.NotNil(t, preTask.Version)
	require.Equal(t, gateTaskVersion, *preTask.Version, "version must match")
	require.NotNil(t, preTask.DefinitionType)
	require.Equal(t, gateTaskDefType, *preTask.DefinitionType, "definition_type must match")

	// Post-deployment gates: no gate blocks declared, Gates should be nil/empty
	require.NotNil(t, env.PostDeploymentGates, "PostDeploymentGates must be set (has gates_options)")
	// An empty gate list means Gates is not populated
	if env.PostDeploymentGates.Gates != nil {
		require.Len(t, *env.PostDeploymentGates.Gates, 0, "no gate blocks → empty gates array")
	}

	// ── AC2: flatten round-trips gate {} blocks + tasks ───────────────────
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	envList := resourceData.Get("environment").([]interface{})
	require.Len(t, envList, 1)
	envMap := envList[0].(map[string]interface{})

	// pre_deployment_gates should contain gate blocks
	preGatesList, ok := envMap["pre_deployment_gates"].([]interface{})
	require.True(t, ok, "pre_deployment_gates should be present after flatten")
	require.Len(t, preGatesList, 1)
	preGatesMap := preGatesList[0].(map[string]interface{})

	flatGates, ok := preGatesMap["gate"].([]interface{})
	require.True(t, ok, "gate key must be present after flatten")
	require.Len(t, flatGates, 1, "one gate block expected after flatten")

	flatGate := flatGates[0].(map[string]interface{})
	flatTasks, ok := flatGate["task"].([]interface{})
	require.True(t, ok, "task key must be present inside gate after flatten")
	require.Len(t, flatTasks, 1, "one task expected inside gate after flatten")

	flatTask := flatTasks[0].(map[string]interface{})
	require.Equal(t, gateTaskName, flatTask["name"], "task name must survive flatten")
	require.Equal(t, gateTaskID, flatTask["task_id"], "task_id must survive flatten")
	require.Equal(t, gateTaskVersion, flatTask["version"], "version must survive flatten")
	require.Equal(t, gateTaskDefType, flatTask["definition_type"], "definition_type must survive flatten")
}

// ── 12. Triggers: empty triggers block (no panic) ─────────────────────────

// TestReleaseDefinition_Triggers_Empty verifies that a definition with no triggers
// block expands without error and that flattenTriggers handles a nil/empty Triggers
// slice gracefully (no panic).
func TestReleaseDefinition_Triggers_Empty(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "NoTriggersDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		// No "triggers" key — omitted.
	})

	expanded, _, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Nil(t, expanded.Triggers, "Triggers must be nil when no triggers block is provided")

	// flattenTriggers with nil must not panic
	require.NotPanics(t, func() {
		result := flattenTriggers(nil)
		require.Nil(t, result)
	})

	// flattenTriggers with empty slice must not panic
	require.NotPanics(t, func() {
		empty := []interface{}{}
		result := flattenTriggers(&empty)
		require.Nil(t, result)
	})
}

// ── 13. Triggers: CD artifact trigger only ────────────────────────────────

// TestReleaseDefinition_Triggers_ArtifactOnly verifies that a definition with a
// cd_artifact_trigger block in the triggers container (AC1) expands into a
// ReleaseDefinition.Triggers slice containing exactly one artifactSource entry
// with the correct artifactAlias and triggerConditions, and that flattenTriggers
// (AC2) round-trips it back correctly.
func TestReleaseDefinition_Triggers_ArtifactOnly(t *testing.T) {
	artifactAlias := "_myBuild"
	branchInclude := "refs/heads/main"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "ArtifactTriggerDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		"triggers": []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{
					map[string]interface{}{
						"artifact_alias": artifactAlias,
						"branch_filter": []interface{}{
							map[string]interface{}{
								"include": []interface{}{branchInclude},
								"exclude": []interface{}{},
							},
						},
					},
				},
				"schedule_trigger": []interface{}{},
			},
		},
	})

	// AC1: expandReleaseDefinition must produce one artifactSource trigger
	expanded, _, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.NotNil(t, expanded.Triggers)
	require.Len(t, *expanded.Triggers, 1)

	trigRaw := (*expanded.Triggers)[0]
	trigMap, ok := trigRaw.(map[string]interface{})
	require.True(t, ok, "trigger entry must be a map")
	require.Equal(t, "artifactSource", trigMap["triggerType"])
	require.Equal(t, artifactAlias, trigMap["artifactAlias"])

	conditions, ok := trigMap["triggerConditions"].([]map[string]interface{})
	require.True(t, ok, "triggerConditions must be set")
	require.Len(t, conditions, 1)
	require.Equal(t, branchInclude, conditions[0]["sourceBranch"])

	// AC2: flattenTriggers must round-trip back
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, testReleaseDefinitionProjectID.String())

	triggersRaw, ok := resourceData.GetOk("triggers")
	require.True(t, ok, "triggers block must be present in state after flatten")
	triggersList := triggersRaw.([]interface{})
	require.Len(t, triggersList, 1)

	trigsMap := triggersList[0].(map[string]interface{})
	cdTriggers, ok := trigsMap["cd_artifact_trigger"].([]interface{})
	require.True(t, ok)
	require.Len(t, cdTriggers, 1)

	ctMap := cdTriggers[0].(map[string]interface{})
	require.Equal(t, artifactAlias, ctMap["artifact_alias"])

	bfList, ok := ctMap["branch_filter"].([]interface{})
	require.True(t, ok)
	require.Len(t, bfList, 1)
	bfMap := bfList[0].(map[string]interface{})
	includes, ok := bfMap["include"].([]interface{})
	require.True(t, ok)
	require.Len(t, includes, 1)
	require.Equal(t, branchInclude, includes[0])
}

// ── 14. Triggers: schedule trigger only ───────────────────────────────────

// TestReleaseDefinition_Triggers_ScheduleOnly verifies that a definition with a
// schedule_trigger block in the triggers container (AC1) expands into a
// ReleaseDefinition.Triggers slice containing exactly one schedule entry with
// all schedule fields correctly set, and that flattenTriggers (AC2) round-trips it.
// Note: branch_filter was removed from schedule_trigger schema (AC1/WI-9) because
// ADO classic schedule triggers are time-based and ADO does not return branchFilters.
func TestReleaseDefinition_Triggers_ScheduleOnly(t *testing.T) {
	startHours := 2
	startMinutes := 30
	timeZoneID := "UTC"
	daysToRelease := 62 // Mon–Fri (1+2+4+8+16+32 = no, 1+2+4+8+16 = 31... use 62 as arbitrary value)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "ScheduleTriggerDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		"triggers": []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{},
				"schedule_trigger": []interface{}{
					map[string]interface{}{
						// branch_filter removed from schedule_trigger schema (AC1/WI-9)
						"schedule_only_with_changes": true,
						"start_hours":                startHours,
						"start_minutes":              startMinutes,
						"time_zone_id":               timeZoneID,
						"days_to_release":            daysToRelease,
					},
				},
			},
		},
	})

	// AC1: expandReleaseDefinition must produce one schedule trigger
	expanded, _, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.NotNil(t, expanded.Triggers)
	require.Len(t, *expanded.Triggers, 1)

	trigRaw := (*expanded.Triggers)[0]
	trigMap, ok := trigRaw.(map[string]interface{})
	require.True(t, ok, "trigger entry must be a map")
	require.Equal(t, "schedule", trigMap["triggerType"])

	sched, ok := trigMap["schedule"].(map[string]interface{})
	require.True(t, ok, "schedule field must be present")
	require.Equal(t, true, sched["scheduleOnlyWithChanges"])
	require.Equal(t, startHours, sched["startHours"])
	require.Equal(t, startMinutes, sched["startMinutes"])
	require.Equal(t, timeZoneID, sched["timeZoneId"])
	require.Equal(t, daysToRelease, sched["daysToRelease"])

	// AC1/WI-9: branch_filter was REMOVED from schedule_trigger schema.
	// branchFilters must NOT appear in the expanded trigger (no branch filter in schedule triggers).
	_, hasBF := trigMap["branchFilters"]
	require.False(t, hasBF, "branchFilters must NOT appear in expanded schedule trigger (AC1/WI-9)")

	// AC2: flattenTriggers must round-trip back
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, testReleaseDefinitionProjectID.String())

	triggersRaw, ok := resourceData.GetOk("triggers")
	require.True(t, ok, "triggers block must be present in state after flatten")
	triggersList := triggersRaw.([]interface{})
	require.Len(t, triggersList, 1)

	trigsMap := triggersList[0].(map[string]interface{})
	schTriggers, ok := trigsMap["schedule_trigger"].([]interface{})
	require.True(t, ok)
	require.Len(t, schTriggers, 1)

	stMap := schTriggers[0].(map[string]interface{})
	require.Equal(t, true, stMap["schedule_only_with_changes"])
	require.Equal(t, startHours, stMap["start_hours"])
	require.Equal(t, startMinutes, stMap["start_minutes"])
	require.Equal(t, timeZoneID, stMap["time_zone_id"])
	require.Equal(t, daysToRelease, stMap["days_to_release"])

	// AC1/WI-9: branch_filter must NOT appear in flattened schedule_trigger state (no perpetual diff).
	_, hasBFInState := stMap["branch_filter"]
	require.False(t, hasBFInState,
		"branch_filter must NOT appear in flattened schedule_trigger state (AC1/WI-9)")
}

// ── 15. Triggers: both artifact and schedule triggers ─────────────────────

// TestReleaseDefinition_Triggers_ExpandFlatten verifies that a definition with BOTH a
// cd_artifact_trigger and a schedule_trigger in the triggers container correctly
// expands to a Triggers slice containing one artifactSource and one schedule entry
// (AC1), and that flattenReleaseDefinition (AC2) restores both sub-blocks in state.
// This test satisfies AC3 via the `go test -run TestReleaseDefinition_Triggers` prefix match.
func TestReleaseDefinition_Triggers_ExpandFlatten(t *testing.T) {
	artifactAlias := "_myBuild"
	cdBranch := "refs/heads/main"
	startHours := 2
	startMinutes := 0
	timeZoneID := "UTC"
	daysToRelease := 127

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "BothTriggersDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		"triggers": []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{
					map[string]interface{}{
						"artifact_alias": artifactAlias,
						"branch_filter": []interface{}{
							map[string]interface{}{
								"include": []interface{}{cdBranch},
								"exclude": []interface{}{},
							},
						},
					},
				},
				"schedule_trigger": []interface{}{
					map[string]interface{}{
						// branch_filter removed from schedule_trigger schema (AC1/WI-9):
						// ADO classic schedule triggers are time-based and have no branch filter.
						"schedule_only_with_changes": true,
						"start_hours":                startHours,
						"start_minutes":              startMinutes,
						"time_zone_id":               timeZoneID,
						"days_to_release":            daysToRelease,
					},
				},
			},
		},
	})

	// AC1: expandReleaseDefinition must produce two triggers (artifact + schedule)
	expanded, projectID, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.Equal(t, testReleaseDefinitionProjectID.String(), projectID)
	require.NotNil(t, expanded.Triggers)
	require.Len(t, *expanded.Triggers, 2, "Triggers slice must have exactly 2 entries")

	// Verify artifact trigger
	artifactTrigRaw := (*expanded.Triggers)[0]
	atMap, ok := artifactTrigRaw.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "artifactSource", atMap["triggerType"], "first trigger must be artifactSource")
	require.Equal(t, artifactAlias, atMap["artifactAlias"])
	conditions, ok := atMap["triggerConditions"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, conditions, 1)
	require.Equal(t, cdBranch, conditions[0]["sourceBranch"])

	// Verify schedule trigger
	schedTrigRaw := (*expanded.Triggers)[1]
	stMap, ok := schedTrigRaw.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "schedule", stMap["triggerType"], "second trigger must be schedule")
	sched, ok := stMap["schedule"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, sched["scheduleOnlyWithChanges"])
	require.Equal(t, startHours, sched["startHours"])
	require.Equal(t, startMinutes, sched["startMinutes"])
	require.Equal(t, timeZoneID, sched["timeZoneId"])
	require.Equal(t, daysToRelease, sched["daysToRelease"])

	// AC2: flattenReleaseDefinition must restore both sub-blocks
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, projectID)

	triggersRaw, ok := resourceData.GetOk("triggers")
	require.True(t, ok, "triggers block must be present in Terraform state after flatten")
	triggersList := triggersRaw.([]interface{})
	require.Len(t, triggersList, 1)

	trigsMap := triggersList[0].(map[string]interface{})

	// Verify cd_artifact_trigger in state
	cdTriggers, ok := trigsMap["cd_artifact_trigger"].([]interface{})
	require.True(t, ok, "cd_artifact_trigger must be present in state")
	require.Len(t, cdTriggers, 1, "must have exactly one cd_artifact_trigger")
	ctMap := cdTriggers[0].(map[string]interface{})
	require.Equal(t, artifactAlias, ctMap["artifact_alias"],
		"artifact_alias must survive expand/flatten")

	cdBfList, ok := ctMap["branch_filter"].([]interface{})
	require.True(t, ok)
	require.Len(t, cdBfList, 1)
	cdBfMap := cdBfList[0].(map[string]interface{})
	cdIncludes, ok := cdBfMap["include"].([]interface{})
	require.True(t, ok)
	require.Len(t, cdIncludes, 1)
	require.Equal(t, cdBranch, cdIncludes[0],
		"cd_artifact_trigger branch_filter include must survive expand/flatten")

	// Verify schedule_trigger in state
	schTriggers, ok := trigsMap["schedule_trigger"].([]interface{})
	require.True(t, ok, "schedule_trigger must be present in state")
	require.Len(t, schTriggers, 1, "must have exactly one schedule_trigger")
	schMap := schTriggers[0].(map[string]interface{})
	require.Equal(t, true, schMap["schedule_only_with_changes"],
		"schedule_only_with_changes must survive expand/flatten")
	require.Equal(t, startHours, schMap["start_hours"],
		"start_hours must survive expand/flatten")
	require.Equal(t, startMinutes, schMap["start_minutes"],
		"start_minutes must survive expand/flatten")
	require.Equal(t, timeZoneID, schMap["time_zone_id"],
		"time_zone_id must survive expand/flatten")
	require.Equal(t, daysToRelease, schMap["days_to_release"],
		"days_to_release must survive expand/flatten")

	// AC1/WI-9: branch_filter must NOT appear in flattened schedule_trigger state.
	_, hasBFSched := schMap["branch_filter"]
	require.False(t, hasBFSched,
		"branch_filter must NOT appear in flattened schedule_trigger state (AC1/WI-9)")
}

// ── WI-7: round-trip idempotency tests ────────────────────────────────────

// TestReleaseDefinition_RoundTrip verifies that expand → JSON-marshal → JSON-unmarshal → flatten
// produces the same state as the original HCL input (i.e., no perpetual diff).
// This simulates the ADO API round-trip that the live demo proved was broken.
//
// Three sub-tests cover production bugs:
//  1. multipliers round-trips through a comma-joined string (ADO wire format).
//  2. A phase without parallel_execution in HCL produces no parallel_execution block in state.
//  3. schedule_trigger with NO branch_filter round-trips with no residual diff (AC1/WI-9).
func TestReleaseDefinition_RoundTrip(t *testing.T) {
	t.Run("multipliers_comma_string_round_trip", func(t *testing.T) {
		// Simulate what ADO API returns: multipliers as a comma-joined string.
		// The user set ["TargetSlot", "Production"] in HCL; ADO returns "TargetSlot,Production".
		adoParallelExecution := map[string]interface{}{
			"parallelExecutionType": "multiConfiguration",
			"maxNumberOfAgents":     float64(2),
			"continueOnError":       false,
			"multipliers":           "TargetSlot,Production", // ADO wire format: comma-joined string
		}
		result := flattenParallelExecution(adoParallelExecution)
		require.Len(t, result, 1, "flattenParallelExecution must return one element for multiConfiguration")

		peMap := result[0].(map[string]interface{})
		require.Equal(t, "multiConfiguration", peMap["type"])

		multipliers, ok := peMap["multipliers"].([]string)
		require.True(t, ok, "multipliers must be []string in state")
		require.Equal(t, []string{"TargetSlot", "Production"}, multipliers,
			"multipliers must be split from comma-joined string")
	})

	t.Run("multipliers_array_round_trip", func(t *testing.T) {
		// Also verify that when ADO returns multipliers as a JSON array ([]interface{}),
		// the flatten still works correctly (backward compat / other clients).
		adoParallelExecution := map[string]interface{}{
			"parallelExecutionType": "multiConfiguration",
			"maxNumberOfAgents":     float64(1),
			"continueOnError":       false,
			"multipliers":           []interface{}{"x86", "x64"},
		}
		result := flattenParallelExecution(adoParallelExecution)
		require.Len(t, result, 1)
		peMap := result[0].(map[string]interface{})
		multipliers, ok := peMap["multipliers"].([]string)
		require.True(t, ok, "multipliers must be []string")
		require.Equal(t, []string{"x86", "x64"}, multipliers)
	})

	t.Run("no_parallel_execution_produces_no_block", func(t *testing.T) {
		// A deploy phase without parallel_execution in HCL expands with no parallelExecution
		// key. ADO returns {parallelExecutionType: "none"} as the default. The flatten must
		// NOT emit a parallel_execution block in state (which would create a perpetual diff
		// vs HCL that has no parallel_execution block at all).
		adoDeploymentInput := map[string]interface{}{
			"queueId":                   float64(5),
			"timeoutInMinutes":          float64(0),
			"jobCancelTimeoutInMinutes": float64(1),
			"condition":                 "succeeded()",
			"skipArtifactsDownload":     false,
			"enableAccessToken":         false,
			// ADO returns a default parallelExecution even when not configured:
			"parallelExecution": map[string]interface{}{
				"parallelExecutionType": "none",
			},
		}
		result := flattenDeploymentInput(adoDeploymentInput)
		require.Len(t, result, 1)
		diFlat := result[0].(map[string]interface{})

		// parallel_execution must be absent (or empty) — no block should be emitted.
		pe, hasPE := diFlat["parallel_execution"]
		if hasPE {
			// If the key exists it must be an empty/nil list — not a [{type:none}] block.
			asList, ok := pe.([]interface{})
			require.True(t, ok, "parallel_execution must be []interface{} if present")
			require.Len(t, asList, 0,
				"parallel_execution must be empty (no block) for a phase without parallel_execution in HCL")
		}
		// If hasPE is false, that's also correct — no key at all is ideal.
	})

	t.Run("schedule_trigger_no_branch_filter_no_residual_diff", func(t *testing.T) {
		// AC1 / WI-9: branch_filter was REMOVED from schedule_trigger schema because ADO
		// classic schedule triggers are time-based and ADO does NOT return branchFilters
		// in the GET response. This test verifies that a schedule_trigger without any
		// branch_filter round-trips through expand → JSON → flatten with NO residual diff
		// (i.e., no unexpected branch_filter key appears in the flattened state).
		hclState := []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{},
				"schedule_trigger": []interface{}{
					map[string]interface{}{
						"schedule_only_with_changes": true,
						"start_hours":                2,
						"start_minutes":              0,
						"time_zone_id":               "UTC",
						"days_to_release":            62,
					},
				},
			},
		}

		// Step 1: expand
		expanded := expandTriggers(hclState)
		require.Len(t, expanded, 1, "must produce one trigger entry")

		trigMap, ok := expanded[0].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "schedule", trigMap["triggerType"])

		// branchFilters must NOT appear in the expanded trigger (schema removed it).
		_, hasBF := trigMap["branchFilters"]
		require.False(t, hasBF, "branchFilters must NOT appear in expanded schedule trigger (AC1/WI-9)")

		// Step 2: simulate ADO API round-trip (ADO returns schedule trigger without branchFilters).
		jsonBytes, err := json.Marshal(expanded)
		require.NoError(t, err)
		var roundTripped []interface{}
		require.NoError(t, json.Unmarshal(jsonBytes, &roundTripped))

		// Step 3: flatten back — no branch_filter key must appear in the schedule_trigger map.
		triggers := flattenTriggers(&roundTripped)
		require.Len(t, triggers, 1)

		trigsMap, ok := triggers[0].(map[string]interface{})
		require.True(t, ok)
		schTriggers, ok := trigsMap["schedule_trigger"].([]interface{})
		require.True(t, ok)
		require.Len(t, schTriggers, 1)

		st := schTriggers[0].(map[string]interface{})

		// No branch_filter key at all — absence = no residual diff.
		_, hasBFInState := st["branch_filter"]
		require.False(t, hasBFInState,
			"branch_filter must NOT appear in flattened schedule_trigger state (no perpetual diff)")

		// Verify the time-based fields survived the round-trip.
		require.Equal(t, true, st["schedule_only_with_changes"],
			"schedule_only_with_changes must round-trip")
		require.Equal(t, 2, st["start_hours"], "start_hours must round-trip")
		require.Equal(t, "UTC", st["time_zone_id"], "time_zone_id must round-trip")
		require.Equal(t, 62, st["days_to_release"], "days_to_release must round-trip")
	})
}

// ── WI-4: parallel_execution expand/flatten ────────────────────────────────

// TestReleaseDefinition_ParallelExecution_ExpandFlatten covers three sub-cases:
//
//   - AC1: multiConfiguration with maxNumberOfAgents=3 round-trips through
//     expandDeploymentInput → the resulting map carries parallelExecution with the
//     expected ADO camelCase keys.
//   - AC2: flattenDeploymentInput with an ADO multiMachine payload populates
//     parallel_execution.0.type and parallel_execution.0.max_number_of_agents.
//   - AC3: a deployment_input without a parallel_execution block (type "none")
//     does not panic and produces a nil or "none"-typed parallelExecution entry.
func TestReleaseDefinition_ParallelExecution_ExpandFlatten(t *testing.T) {
	t.Run("AC1_expand_multiConfiguration", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"queue_id":                      5,
				"demands":                       []interface{}{},
				"timeout_in_minutes":            0,
				"job_cancel_timeout_in_minutes": 1,
				"condition":                     "succeeded()",
				"skip_artifacts_download":       false,
				"enable_access_token":           false,
				"agent_specification":           "",
				"parallel_execution": []interface{}{
					map[string]interface{}{
						"type":                 "multiConfiguration",
						"max_number_of_agents": 3,
						"multipliers":          []interface{}{"Configuration"},
						"continue_on_error":    false,
					},
				},
			},
		}
		result := expandDeploymentInput(input)
		require.NotNil(t, result, "expandDeploymentInput must return a non-nil map")

		// Verify top-level fields
		require.Equal(t, 5, result["queueId"], "queueId must round-trip")

		// Verify parallelExecution sub-map
		pe, ok := result["parallelExecution"].(map[string]interface{})
		require.True(t, ok, "parallelExecution must be a map[string]interface{}")
		require.Equal(t, "multiConfiguration", pe["parallelExecutionType"],
			"parallelExecutionType must be 'multiConfiguration'")
		require.Equal(t, 3, pe["maxNumberOfAgents"],
			"maxNumberOfAgents must be 3")
		// ADO stores multipliers as a comma-separated string (Multipliers *string in the SDK);
		// expandParallelExecution must send a string, not an array, so ADO stores it correctly.
		multipliersStr, ok := pe["multipliers"].(string)
		require.True(t, ok, "multipliers must be a comma-separated string (ADO wire format)")
		require.Equal(t, "Configuration", multipliersStr)
	})

	t.Run("AC2_flatten_multiMachine", func(t *testing.T) {
		// Simulate the ADO API response map for a multiMachine deployment input
		adoMap := map[string]interface{}{
			"queueId":                   float64(10),
			"timeoutInMinutes":          float64(0),
			"jobCancelTimeoutInMinutes": float64(1),
			"condition":                 "succeeded()",
			"skipArtifactsDownload":     false,
			"enableAccessToken":         false,
			"parallelExecution": map[string]interface{}{
				"parallelExecutionType": "multiMachine",
				"maxNumberOfAgents":     float64(2),
				"continueOnError":       false,
			},
		}
		result := flattenDeploymentInput(adoMap)
		require.Len(t, result, 1, "flattenDeploymentInput must return a slice with one element")

		diFlat := result[0].(map[string]interface{})

		peList, ok := diFlat["parallel_execution"].([]interface{})
		require.True(t, ok, "parallel_execution must be []interface{}")
		require.Len(t, peList, 1, "parallel_execution must have exactly one element")

		peMap := peList[0].(map[string]interface{})
		require.Equal(t, "multiMachine", peMap["type"],
			"deployment_input.0.parallel_execution.0.type must be 'multiMachine'")
		require.Equal(t, 2, peMap["max_number_of_agents"],
			"deployment_input.0.parallel_execution.0.max_number_of_agents must be 2")
	})

	t.Run("AC3_expand_none_no_panic", func(t *testing.T) {
		// deployment_input without a parallel_execution block
		input := []interface{}{
			map[string]interface{}{
				"queue_id":                      1,
				"demands":                       []interface{}{},
				"timeout_in_minutes":            0,
				"job_cancel_timeout_in_minutes": 1,
				"condition":                     "succeeded()",
				"skip_artifacts_download":       false,
				"enable_access_token":           false,
				"agent_specification":           "",
				"parallel_execution":            []interface{}{},
			},
		}
		// Must not panic
		result := expandDeploymentInput(input)
		require.NotNil(t, result, "expandDeploymentInput must return non-nil map even without parallel_execution")

		// parallelExecution key must be absent (not set when pe is empty)
		_, hasPE := result["parallelExecution"]
		require.False(t, hasPE, "parallelExecution must not be present when parallel_execution block is empty")
	})
}

// TestReleaseDefinition_AgentlessPhase_ExpandFlatten covers the runOnServer (agentless)
// deploy-phase variant:
//
//   - AC1: expandDeploymentInput with phase_type=runOnServer produces phaseType: runOnServer
//     in the parent deploy-phase map and the deploymentInput map carries timeout fields
//     but NO queueId key.
//   - AC2: flattenDeploymentInput with an ADO runOnServer payload (no queueId) populates
//     phase_type=runOnServer and timeout_in_minutes without panicking.
//   - AC3: a roundtrip through expandDeployPhases + flattenDeployPhases preserves phaseType
//     and timeout fields for a runOnServer phase alongside a regular agentBasedDeployment.
func TestReleaseDefinition_AgentlessPhase_ExpandFlatten(t *testing.T) {
	t.Run("AC1_expand_runOnServer_no_queueId", func(t *testing.T) {
		// Simulate a deployment_input block for an agentless phase (no queue needed).
		input := []interface{}{
			map[string]interface{}{
				"queue_id":                      0,
				"demands":                       []interface{}{},
				"timeout_in_minutes":            120,
				"job_cancel_timeout_in_minutes": 5,
				"condition":                     "succeeded()",
				"skip_artifacts_download":       false,
				"enable_access_token":           false,
				"agent_specification":           "",
				"parallel_execution":            []interface{}{},
			},
		}
		result := expandDeploymentInput(input, "runOnServer")
		require.NotNil(t, result, "expandDeploymentInput must return a non-nil map for runOnServer")

		// queueId must NOT be present for agentless phases.
		_, hasQueue := result["queueId"]
		require.False(t, hasQueue, "queueId must be absent for runOnServer phase")

		// Timeout fields must be present.
		require.Equal(t, 120, result["timeoutInMinutes"], "timeoutInMinutes must be 120")
		require.Equal(t, 5, result["jobCancelTimeoutInMinutes"], "jobCancelTimeoutInMinutes must be 5")
	})

	t.Run("AC2_flatten_runOnServer_no_queueId_no_panic", func(t *testing.T) {
		// Simulate the ADO API response for a runOnServer phase — no queueId key.
		adoMap := map[string]interface{}{
			"timeoutInMinutes":          float64(120),
			"jobCancelTimeoutInMinutes": float64(5),
			"condition":                 "succeeded()",
			"skipArtifactsDownload":     false,
			"enableAccessToken":         false,
		}
		// Must not panic even though queueId is absent.
		result := flattenDeploymentInput(adoMap)
		require.Len(t, result, 1, "flattenDeploymentInput must return a single-element slice")

		diFlat := result[0].(map[string]interface{})

		// queue_id defaults to 0 when not present in ADO response.
		require.Equal(t, 0, diFlat["queue_id"], "queue_id must default to 0 for agentless phase")

		// Timeout must be populated from the ADO response.
		require.Equal(t, 120, diFlat["timeout_in_minutes"], "timeout_in_minutes must be 120")
	})

	t.Run("AC3_roundtrip_agent_and_agentless_phases", func(t *testing.T) {
		// Build a slice with two phases: one agent-based, one runOnServer.
		phases := []interface{}{
			map[string]interface{}{
				"name":       "Agent Phase",
				"rank":       1,
				"phase_type": "agentBasedDeployment",
				"deployment_input": []interface{}{
					map[string]interface{}{
						"queue_id":                      7,
						"demands":                       []interface{}{},
						"timeout_in_minutes":            30,
						"job_cancel_timeout_in_minutes": 1,
						"condition":                     "succeeded()",
						"skip_artifacts_download":       false,
						"enable_access_token":           false,
						"agent_specification":           "",
						"parallel_execution":            []interface{}{},
					},
				},
				"workflow_task": []interface{}{},
			},
			map[string]interface{}{
				"name":       "Agentless Phase",
				"rank":       2,
				"phase_type": "runOnServer",
				"deployment_input": []interface{}{
					map[string]interface{}{
						"queue_id":                      0,
						"demands":                       []interface{}{},
						"timeout_in_minutes":            120,
						"job_cancel_timeout_in_minutes": 5,
						"condition":                     "succeeded()",
						"skip_artifacts_download":       false,
						"enable_access_token":           false,
						"agent_specification":           "",
						"parallel_execution":            []interface{}{},
					},
				},
				"workflow_task": []interface{}{},
			},
		}

		expanded, err := expandDeployPhases(phases)
		require.NoError(t, err, "expandDeployPhases must not error")
		require.Len(t, expanded, 2)

		// Verify agent-based phase retains queueId.
		agentPhase := expanded[0].(map[string]interface{})
		require.Equal(t, "agentBasedDeployment", agentPhase["phaseType"])
		agentDI := agentPhase["deploymentInput"].(map[string]interface{})
		require.Equal(t, 7, agentDI["queueId"], "agent phase must retain queueId=7")

		// Verify agentless phase has no queueId.
		agentlessPhase := expanded[1].(map[string]interface{})
		require.Equal(t, "runOnServer", agentlessPhase["phaseType"])
		agentlessDI := agentlessPhase["deploymentInput"].(map[string]interface{})
		_, hasQueue := agentlessDI["queueId"]
		require.False(t, hasQueue, "agentless phase must not have queueId in expanded output")
		require.Equal(t, 120, agentlessDI["timeoutInMinutes"], "agentless timeout must be 120")

		// Now marshal the expanded phases and flatten them (simulating the ADO round-trip).
		// Re-marshal through JSON to simulate what ADO would return.
		expandedJSON, err := json.Marshal(expanded)
		require.NoError(t, err)
		var roundTripped []interface{}
		require.NoError(t, json.Unmarshal(expandedJSON, &roundTripped))

		flattened := flattenDeployPhases(&roundTripped, nil, 0)
		require.Len(t, flattened, 2)

		flatAgent := flattened[0].(map[string]interface{})
		require.Equal(t, "agentBasedDeployment", flatAgent["phase_type"])
		flatAgentDI := flatAgent["deployment_input"].([]interface{})[0].(map[string]interface{})
		require.Equal(t, 7, flatAgentDI["queue_id"], "agent phase queue_id must round-trip to 7")

		flatAgentless := flattened[1].(map[string]interface{})
		require.Equal(t, "runOnServer", flatAgentless["phase_type"])
		flatAgentlessDI := flatAgentless["deployment_input"].([]interface{})[0].(map[string]interface{})
		require.Equal(t, 0, flatAgentlessDI["queue_id"], "agentless phase queue_id must be 0")
		require.Equal(t, 120, flatAgentlessDI["timeout_in_minutes"], "agentless timeout must round-trip to 120")
	})
}

// ── TestReleaseDefinition_GatesOptions_RoundTrip ───────────────────────────

// TestReleaseDefinition_GatesOptions_RoundTrip confirms that all five
// ReleaseDefinitionGatesOptions fields (is_enabled, minimum_success_duration,
// sampling_interval, stabilization_time, timeout) round-trip correctly through
// expandDeploymentGates and flattenDeploymentGates.
func TestReleaseDefinition_GatesOptions_RoundTrip(t *testing.T) {
	isEnabled := true
	minimumSuccessDuration := 10
	samplingInterval := 5
	stabilizationTime := 3
	timeout := 120

	// Step 1: Build the Terraform-style input slice and call expandDeploymentGates.
	input := []interface{}{
		map[string]interface{}{
			"gates_options": []interface{}{
				map[string]interface{}{
					"is_enabled":               isEnabled,
					"minimum_success_duration": minimumSuccessDuration,
					"sampling_interval":        samplingInterval,
					"stabilization_time":       stabilizationTime,
					"timeout":                  timeout,
				},
			},
		},
	}

	step := expandDeploymentGates(input)
	require.NotNil(t, step, "expandDeploymentGates must return a non-nil step")
	require.NotNil(t, step.GatesOptions, "GatesOptions must be populated after expand")

	opts := step.GatesOptions
	require.NotNil(t, opts.IsEnabled, "IsEnabled must not be nil")
	require.Equal(t, isEnabled, *opts.IsEnabled, "IsEnabled must match")
	require.NotNil(t, opts.MinimumSuccessDuration, "MinimumSuccessDuration must not be nil")
	require.Equal(t, minimumSuccessDuration, *opts.MinimumSuccessDuration, "MinimumSuccessDuration must match")
	require.NotNil(t, opts.SamplingInterval, "SamplingInterval must not be nil")
	require.Equal(t, samplingInterval, *opts.SamplingInterval, "SamplingInterval must match")
	require.NotNil(t, opts.StabilizationTime, "StabilizationTime must not be nil")
	require.Equal(t, stabilizationTime, *opts.StabilizationTime, "StabilizationTime must match")
	require.NotNil(t, opts.Timeout, "Timeout must not be nil")
	require.Equal(t, timeout, *opts.Timeout, "Timeout must match")

	// Step 2: Simulate the API response by constructing a ReleaseDefinitionGatesStep
	// with the same five values, then call flattenDeploymentGates and assert round-trip.
	apiStep := &releaseapi.ReleaseDefinitionGatesStep{
		GatesOptions: &releaseapi.ReleaseDefinitionGatesOptions{
			IsEnabled:              converter.Bool(isEnabled),
			MinimumSuccessDuration: converter.Int(minimumSuccessDuration),
			SamplingInterval:       converter.Int(samplingInterval),
			StabilizationTime:      converter.Int(stabilizationTime),
			Timeout:                converter.Int(timeout),
		},
	}

	flattened := flattenDeploymentGates(apiStep)
	require.Len(t, flattened, 1, "flattenDeploymentGates must return one block")
	flatMap := flattened[0].(map[string]interface{})

	gatesOptsList, ok := flatMap["gates_options"].([]interface{})
	require.True(t, ok, "gates_options must be present after flatten")
	require.Len(t, gatesOptsList, 1, "gates_options must contain exactly one entry")

	flatOpts := gatesOptsList[0].(map[string]interface{})
	require.Equal(t, isEnabled, flatOpts["is_enabled"], "is_enabled must survive flatten")
	require.Equal(t, minimumSuccessDuration, flatOpts["minimum_success_duration"], "minimum_success_duration must survive flatten")
	require.Equal(t, samplingInterval, flatOpts["sampling_interval"], "sampling_interval must survive flatten")
	require.Equal(t, stabilizationTime, flatOpts["stabilization_time"], "stabilization_time must survive flatten")
	require.Equal(t, timeout, flatOpts["timeout"], "timeout must survive flatten")
}

// ── 21. ArtifactTagFilter round-trip ──────────────────────────────────────

// TestReleaseDefinition_ArtifactTagFilter_RoundTrip verifies that a
// cd_artifact_trigger block with a tag_filter block (AC1 + AC2) round-trips
// correctly through expandTriggers → flattenTriggers.
func TestReleaseDefinition_ArtifactTagFilter_RoundTrip(t *testing.T) {
	artifactAlias := "_myBuild"
	branchInclude := "refs/heads/main"
	tagValue := "stable"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "TagFilterDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		"triggers": []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{
					map[string]interface{}{
						"artifact_alias": artifactAlias,
						"branch_filter": []interface{}{
							map[string]interface{}{
								"include": []interface{}{branchInclude},
								"exclude": []interface{}{},
							},
						},
						"tag_filter": []interface{}{
							map[string]interface{}{
								"tags": []interface{}{tagValue},
							},
						},
						"use_build_definition_branch":     false,
						"create_release_on_build_tagging": false,
					},
				},
				"schedule_trigger": []interface{}{},
			},
		},
	})

	// AC1: expandTriggers must include tags in the first triggerCondition
	expanded, _, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.NotNil(t, expanded.Triggers)
	require.Len(t, *expanded.Triggers, 1)

	trigRaw := (*expanded.Triggers)[0]
	trigMap, ok := trigRaw.(map[string]interface{})
	require.True(t, ok, "trigger entry must be a map")
	require.Equal(t, "artifactSource", trigMap["triggerType"])

	conditions, ok := trigMap["triggerConditions"].([]map[string]interface{})
	require.True(t, ok, "triggerConditions must be []map[string]interface{}")
	require.NotEmpty(t, conditions, "triggerConditions must not be empty")

	firstCond := conditions[0]
	require.Equal(t, branchInclude, firstCond["sourceBranch"], "AC1: sourceBranch must be present")

	// ADO REST 7.1 persists build-tag filtering ONLY as the `tags` array on
	// the condition; the SDK's regex `tagFilter` field is silently dropped by
	// the service (verified live 2026-06-11) and must not be sent.
	require.NotContains(t, firstCond, "tagFilter", "AC1: tagFilter must NOT be sent (ADO drops it)")
	tags, ok := firstCond["tags"].([]string)
	require.True(t, ok, "AC1: condition.tags must be []string")
	require.Equal(t, []string{tagValue}, tags, "AC1: condition.tags must contain the tag value")

	// AC2: flattenTriggers must restore tag_filter block in state
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, testReleaseDefinitionProjectID.String())

	triggersRaw, ok := resourceData.GetOk("triggers")
	require.True(t, ok, "triggers block must be present in state after flatten")
	triggersList := triggersRaw.([]interface{})
	require.Len(t, triggersList, 1)

	trigsMap := triggersList[0].(map[string]interface{})
	cdTriggers, ok := trigsMap["cd_artifact_trigger"].([]interface{})
	require.True(t, ok, "cd_artifact_trigger must be present")
	require.Len(t, cdTriggers, 1)

	ctFlat := cdTriggers[0].(map[string]interface{})
	require.Equal(t, artifactAlias, ctFlat["artifact_alias"])

	tfList, ok := ctFlat["tag_filter"].([]interface{})
	require.True(t, ok, "AC2: tag_filter must be []interface{}")
	require.Len(t, tfList, 1, "AC2: tag_filter must have exactly one entry")

	tfMap := tfList[0].(map[string]interface{})

	var flatTags []interface{}
	switch v := tfMap["tags"].(type) {
	case []interface{}:
		flatTags = v
	default:
		t.Fatalf("AC2: tag_filter.tags has unexpected type %T", tfMap["tags"])
	}
	require.Len(t, flatTags, 1, "AC2: tag_filter.tags must have one entry")
	require.Equal(t, tagValue, flatTags[0], "AC2: tag_filter.tags[0] must round-trip")
}

// ── 22. ArtifactSourceBranchFlags round-trip ──────────────────────────────

// TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip verifies that
// use_build_definition_branch (AC3+AC4) and create_release_on_build_tagging
// (AC5+AC6) round-trip correctly through expandTriggers → flattenTriggers.
func TestReleaseDefinition_ArtifactSourceBranchFlags_RoundTrip(t *testing.T) {
	artifactAlias := "_myBuild"

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseDefinition().Schema, map[string]interface{}{
		"project_id":          testReleaseDefinitionProjectID.String(),
		"name":                "BranchFlagsDef",
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
						"name":             "Agent phase",
						"rank":             1,
						"phase_type":       "agentBasedDeployment",
						"deployment_input": []interface{}{},
						"workflow_task":    []interface{}{},
					},
				},
				"retention_policy":    []interface{}{},
				"environment_options": []interface{}{},
				"execution_policy":    []interface{}{},
			},
		},
		"artifact": []interface{}{},
		"triggers": []interface{}{
			map[string]interface{}{
				"cd_artifact_trigger": []interface{}{
					map[string]interface{}{
						"artifact_alias":                  artifactAlias,
						"branch_filter":                   []interface{}{},
						"tag_filter":                      []interface{}{},
						"use_build_definition_branch":     true,
						"create_release_on_build_tagging": true,
					},
				},
				"schedule_trigger": []interface{}{},
			},
		},
	})

	// AC3: expandTriggers must include useBuildDefinitionBranch: true
	// AC5: expandTriggers must include createReleaseOnBuildTagging: true
	expanded, _, err := expandReleaseDefinition(resourceData)
	require.NoError(t, err)
	require.NotNil(t, expanded.Triggers)
	require.Len(t, *expanded.Triggers, 1)

	trigRaw := (*expanded.Triggers)[0]
	trigMap, ok := trigRaw.(map[string]interface{})
	require.True(t, ok, "trigger entry must be a map")
	require.Equal(t, "artifactSource", trigMap["triggerType"])

	conditions, ok := trigMap["triggerConditions"].([]map[string]interface{})
	require.True(t, ok, "triggerConditions must be []map[string]interface{}")
	require.NotEmpty(t, conditions, "triggerConditions must not be empty (synthetic entry required)")

	firstCond := conditions[0]
	useBuildBranch, ok := firstCond["useBuildDefinitionBranch"].(bool)
	require.True(t, ok, "AC3: useBuildDefinitionBranch must be present")
	require.True(t, useBuildBranch, "AC3: useBuildDefinitionBranch must be true")

	createOnTagging, ok := firstCond["createReleaseOnBuildTagging"].(bool)
	require.True(t, ok, "AC5: createReleaseOnBuildTagging must be present")
	require.True(t, createOnTagging, "AC5: createReleaseOnBuildTagging must be true")

	// AC4 + AC6: flattenTriggers must restore both flags in state
	expandedID := testReleaseDefinitionID
	expanded.Id = &expandedID
	flattenReleaseDefinition(resourceData, expanded, testReleaseDefinitionProjectID.String())

	triggersRaw, ok := resourceData.GetOk("triggers")
	require.True(t, ok, "triggers block must be present in state after flatten")
	triggersList := triggersRaw.([]interface{})
	require.Len(t, triggersList, 1)

	trigsMap := triggersList[0].(map[string]interface{})
	cdTriggers, ok := trigsMap["cd_artifact_trigger"].([]interface{})
	require.True(t, ok, "cd_artifact_trigger must be present")
	require.Len(t, cdTriggers, 1)

	ctFlat := cdTriggers[0].(map[string]interface{})

	// AC4: use_build_definition_branch must be true in state
	useBuildBranchFlat, ok := ctFlat["use_build_definition_branch"].(bool)
	require.True(t, ok, "AC4: use_build_definition_branch must be bool in state")
	require.True(t, useBuildBranchFlat, "AC4: use_build_definition_branch must be true after flatten")

	// AC6: create_release_on_build_tagging must be true in state
	createOnTaggingFlat, ok := ctFlat["create_release_on_build_tagging"].(bool)
	require.True(t, ok, "AC6: create_release_on_build_tagging must be bool in state")
	require.True(t, createOnTaggingFlat, "AC6: create_release_on_build_tagging must be true after flatten")
}

// ── 24. SourceRepoTrigger round-trip ──────────────────────────────────────

// TestReleaseDefinition_SourceRepoTrigger_RoundTrip verifies that a
// source_repo_trigger block (AC1 + AC2) round-trips correctly through
// expandTriggers → flattenTriggers.
func TestReleaseDefinition_SourceRepoTrigger_RoundTrip(t *testing.T) {
	alias := "_myBuild"
	branchFilter := "refs/heads/main"

	// Build the Terraform HCL-equivalent input for expandTriggers.
	hclTriggers := []interface{}{
		map[string]interface{}{
			"cd_artifact_trigger": []interface{}{},
			"schedule_trigger":    []interface{}{},
			"source_repo_trigger": []interface{}{
				map[string]interface{}{
					"alias":          alias,
					"branch_filters": []interface{}{branchFilter},
				},
			},
		},
	}

	// AC1: expandTriggers must emit a triggerType=sourceRepo entry with alias
	// and branchFilters.
	expanded := expandTriggers(hclTriggers)
	require.Len(t, expanded, 1, "AC1: expandTriggers must produce exactly one trigger entry")

	trigRaw := expanded[0]
	trigMap, ok := trigRaw.(map[string]interface{})
	require.True(t, ok, "AC1: trigger entry must be a map[string]interface{}")
	require.Equal(t, "sourceRepo", trigMap["triggerType"], "AC1: triggerType must be sourceRepo")
	require.Equal(t, alias, trigMap["alias"], "AC1: alias must be set correctly")

	bfs, ok := trigMap["branchFilters"].([]string)
	require.True(t, ok, "AC1: branchFilters must be []string")
	require.Equal(t, []string{branchFilter}, bfs, "AC1: branchFilters must contain the expected branch")

	// AC2: flattenTriggers must reconstruct a source_repo_trigger block with
	// alias and branch_filters set correctly.
	flatInput := []interface{}{trigMap}
	flattened := flattenTriggers(&flatInput)
	require.Len(t, flattened, 1, "AC2: flattenTriggers must return a one-element slice")

	trigsMap, ok := flattened[0].(map[string]interface{})
	require.True(t, ok, "AC2: flattened element must be map[string]interface{}")

	srtList, ok := trigsMap["source_repo_trigger"].([]interface{})
	require.True(t, ok, "AC2: source_repo_trigger must be []interface{}")
	require.Len(t, srtList, 1, "AC2: source_repo_trigger must have exactly one entry")

	srt, ok := srtList[0].(map[string]interface{})
	require.True(t, ok, "AC2: source_repo_trigger entry must be map[string]interface{}")
	require.Equal(t, alias, srt["alias"], "AC2: alias must round-trip correctly")

	bfFlat, ok := srt["branch_filters"].([]interface{})
	require.True(t, ok, "AC2: branch_filters must be []interface{}")
	require.Len(t, bfFlat, 1, "AC2: branch_filters must have one entry")
	require.Equal(t, branchFilter, bfFlat[0], "AC2: branch_filters[0] must round-trip correctly")
}
