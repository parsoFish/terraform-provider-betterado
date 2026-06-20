//go:build all || resource_release_definition

package release

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestFrameworkReleaseDefinition_expandTopLevel verifies that expand correctly maps top-level
// model fields (name, project_id, path, release_name_format) to the API struct.
func TestFrameworkReleaseDefinition_expandTopLevel(t *testing.T) {
	ctx := context.Background()

	model := &releaseDefinitionModel{
		Name:              types.StringValue("MyDef"),
		ProjectID:         types.StringValue("00000000-0000-0000-0000-000000000001"),
		Path:              types.StringValue(`\`),
		Description:       types.StringValue(""),
		ReleaseNameFormat: types.StringValue("Release-$(rev:r)"),
		Tags:              types.SetValueMust(types.StringType, nil),
		Revision:          types.Int64Value(0),
		Stages:            types.ListValueMust(types.ObjectType{AttrTypes: stageAttrTypes}, nil),
	}

	def, projectID, err := expandReleaseDefinitionFramework(ctx, model)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", projectID)
	require.NotNil(t, def.Name)
	assert.Equal(t, "MyDef", *def.Name)
	require.NotNil(t, def.Path)
	assert.Equal(t, `\`, *def.Path)
	require.NotNil(t, def.ReleaseNameFormat)
	assert.Equal(t, "Release-$(rev:r)", *def.ReleaseNameFormat)
}

// TestFrameworkReleaseDefinition_flattenTopLevel verifies that flatten correctly maps the API
// struct's top-level fields (name, project_id, path, release_name_format) back into the model.
func TestFrameworkReleaseDefinition_flattenTopLevel(t *testing.T) {
	ctx := context.Background()

	projectID := "00000000-0000-0000-0000-000000000002"
	defID := 42
	revision := 3
	def := &releaseapi.ReleaseDefinition{
		Id:                &defID,
		Name:              converter.String("FlattenTest"),
		Path:              converter.String(`\staging`),
		Description:       converter.String("test description"),
		ReleaseNameFormat: converter.String("Release-$(rev:r)"),
		Revision:          &revision,
	}

	var model releaseDefinitionModel
	diags := flattenReleaseDefinitionFramework(ctx, def, projectID, &model)
	require.False(t, diags.HasError(), "flattenReleaseDefinitionFramework returned errors: %s", diags)

	assert.Equal(t, "42", model.ID.ValueString())
	assert.Equal(t, projectID, model.ProjectID.ValueString())
	assert.Equal(t, "FlattenTest", model.Name.ValueString())
	assert.Equal(t, `\staging`, model.Path.ValueString())
	assert.Equal(t, "test description", model.Description.ValueString())
	assert.Equal(t, "Release-$(rev:r)", model.ReleaseNameFormat.ValueString())
	assert.Equal(t, int64(3), model.Revision.ValueInt64())
}

// buildDeployPhaseList is a test helper that constructs a types.List representing a single
// deploy_phase with one workflow_task.
func buildDeployPhaseList(t *testing.T,
	phaseName string, phaseType string, rank int64,
	taskID string, displayName string, versionSpec string,
	enabled bool, alwaysRun bool, condition string,
	timeoutMins int64, retryCount int64,
	inputs map[string]string,
) types.List {
	t.Helper()

	// Build inputs map.
	inputAttrs := map[string]attr.Value{}
	for k, v := range inputs {
		inputAttrs[k] = types.StringValue(v)
	}
	inputsTF, diags := types.MapValue(types.StringType, inputAttrs)
	require.False(t, diags.HasError(), "build inputs map: %s", diags)

	// Build workflow_task object.
	taskObj, diags := types.ObjectValue(workflowTaskAttrTypes, map[string]attr.Value{
		"task_id":                     types.StringValue(taskID),
		"display_name":                types.StringValue(displayName),
		"version_spec":                types.StringValue(versionSpec),
		"enabled":                     types.BoolValue(enabled),
		"always_run":                  types.BoolValue(alwaysRun),
		"condition":                   types.StringValue(condition),
		"timeout_in_minutes":          types.Int64Value(timeoutMins),
		"retry_count_on_task_failure": types.Int64Value(retryCount),
		"inputs":                      inputsTF,
		"definition_type":             types.StringValue("task"),
	})
	require.False(t, diags.HasError(), "build task object: %s", diags)

	taskList, diags := types.ListValue(types.ObjectType{AttrTypes: workflowTaskAttrTypes}, []attr.Value{taskObj})
	require.False(t, diags.HasError(), "build task list: %s", diags)

	// Build empty deployment_input list (not needed for this test).
	// Use nil list — no deployment_input block configured.
	emptyDIList := types.ListValueMust(types.ObjectType{AttrTypes: deploymentInputAttrTypes}, []attr.Value{})

	// Build deploy_phase object.
	phaseObj, diags := types.ObjectValue(deployPhaseAttrTypes, map[string]attr.Value{
		"name":             types.StringValue(phaseName),
		"rank":             types.Int64Value(rank),
		"phase_type":       types.StringValue(phaseType),
		"deployment_input": emptyDIList,
		"workflow_task":    taskList,
	})
	require.False(t, diags.HasError(), "build phase object: %s", diags)

	phaseList, diags := types.ListValue(types.ObjectType{AttrTypes: deployPhaseAttrTypes}, []attr.Value{phaseObj})
	require.False(t, diags.HasError(), "build phase list: %s", diags)

	return phaseList
}

// TestFrameworkReleaseDefinition_expandDeployPhase verifies that expandDeployPhasesFramework
// correctly maps a deploy_phase with workflow_task to the ADO API raw map shape.
func TestFrameworkReleaseDefinition_expandDeployPhase(t *testing.T) {
	ctx := context.Background()

	const taskUUID = "12345678-1234-1234-1234-1234567890ab"

	phaseList := buildDeployPhaseList(t,
		"Run phase",                       // phaseName
		"agentBasedDeployment",            // phaseType
		1,                                 // rank
		taskUUID,                          // taskID
		"My Task",                         // displayName
		"1.*",                             // versionSpec
		true,                              // enabled
		false,                             // alwaysRun
		"succeeded()",                     // condition
		30,                                // timeoutMins (non-default)
		2,                                 // retryCount (non-default)
		map[string]string{"key1": "val1"}, // inputs
	)

	phases, err := expandDeployPhasesFramework(ctx, phaseList)
	require.NoError(t, err)
	require.Len(t, phases, 1)

	phaseMap, ok := phases[0].(map[string]interface{})
	require.True(t, ok, "phase should be map[string]interface{}")

	// Assert top-level phase fields.
	assert.Equal(t, "agentBasedDeployment", phaseMap["phaseType"])
	assert.Equal(t, 1, phaseMap["rank"])
	assert.Equal(t, "Run phase", phaseMap["name"])

	// Assert workflowTasks.
	wfTasksRaw, ok := phaseMap["workflowTasks"]
	require.True(t, ok, "workflowTasks key must be present")
	wfTasks, ok := wfTasksRaw.([]releaseapi.WorkflowTask)
	require.True(t, ok, "workflowTasks should be []releaseapi.WorkflowTask")
	require.Len(t, wfTasks, 1)

	task := wfTasks[0]
	require.NotNil(t, task.TimeoutInMinutes)
	assert.Equal(t, 30, *task.TimeoutInMinutes)
	require.NotNil(t, task.RetryCountOnTaskFailure)
	assert.Equal(t, 2, *task.RetryCountOnTaskFailure)
	require.NotNil(t, task.TaskId)
	assert.Equal(t, taskUUID, task.TaskId.String())
	require.NotNil(t, task.Name)
	assert.Equal(t, "My Task", *task.Name)
	require.NotNil(t, task.Inputs)
	assert.Equal(t, "val1", (*task.Inputs)["key1"])
}

// TestFrameworkReleaseDefinition_flattenDeployPhase verifies that flattenDeployPhasesFramework
// correctly maps the ADO API raw phase map back to the framework types.List.
func TestFrameworkReleaseDefinition_flattenDeployPhase(t *testing.T) {
	ctx := context.Background()

	const taskUUID = "abcdef12-abcd-abcd-abcd-abcdef123456"

	// Simulate the raw API response for a deploy phase.
	rawPhases := &[]interface{}{
		map[string]interface{}{
			"name":      "My Stage Phase",
			"rank":      float64(1),
			"phaseType": "agentBasedDeployment",
			"deploymentInput": map[string]interface{}{
				"condition":             "succeeded()",
				"timeoutInMinutes":      float64(0),
				"skipArtifactsDownload": false,
				"enableAccessToken":     false,
			},
			"workflowTasks": []interface{}{
				map[string]interface{}{
					"taskId":                  taskUUID,
					"name":                    "Flatten Task",
					"version":                 "2.*",
					"enabled":                 true,
					"alwaysRun":               false,
					"condition":               "succeeded()",
					"timeoutInMinutes":        float64(30),
					"retryCountOnTaskFailure": float64(2),
					"inputs": map[string]interface{}{
						"param1": "value1",
					},
				},
			},
		},
	}

	phaseList, diags := flattenDeployPhasesFramework(ctx, rawPhases)
	require.False(t, diags.HasError(), "flattenDeployPhasesFramework returned errors: %s", diags)
	require.False(t, phaseList.IsNull())
	require.False(t, phaseList.IsUnknown())

	var phaseModels []deployPhaseModel
	diags = phaseList.ElementsAs(ctx, &phaseModels, false)
	require.False(t, diags.HasError(), "ElementsAs phaseModels: %s", diags)
	require.Len(t, phaseModels, 1)

	phase := phaseModels[0]
	assert.Equal(t, "agentBasedDeployment", phase.PhaseType.ValueString())
	assert.Equal(t, int64(1), phase.Rank.ValueInt64())
	assert.Equal(t, "My Stage Phase", phase.Name.ValueString())

	// Verify workflow_task.
	require.False(t, phase.WorkflowTask.IsNull())
	var taskModels []workflowTaskModel
	diags = phase.WorkflowTask.ElementsAs(ctx, &taskModels, false)
	require.False(t, diags.HasError(), "ElementsAs taskModels: %s", diags)
	require.Len(t, taskModels, 1)

	task := taskModels[0]
	assert.Equal(t, taskUUID, task.TaskID.ValueString())
	assert.Equal(t, "Flatten Task", task.DisplayName.ValueString())
	assert.Equal(t, "2.*", task.VersionSpec.ValueString())
	assert.Equal(t, true, task.Enabled.ValueBool())
	assert.Equal(t, int64(30), task.TimeoutInMinutes.ValueInt64())
	assert.Equal(t, int64(2), task.RetryCountOnTaskFailure.ValueInt64())

	// Verify inputs.
	var inputMap map[string]string
	diags = task.Inputs.ElementsAs(ctx, &inputMap, false)
	require.False(t, diags.HasError(), "ElementsAs inputs: %s", diags)
	assert.Equal(t, "value1", inputMap["param1"])
}

// ---------------------------------------------------------------------------
// Variables tests
// ---------------------------------------------------------------------------

// TestFrameworkReleaseDefinition_expandVariables verifies that expandVariablesFramework
// correctly converts a MapNestedAttribute value into a ConfigurationVariableValue map.
func TestFrameworkReleaseDefinition_expandVariables(t *testing.T) {
	ctx := context.Background()

	// Build the variables map: {"MY_VAR": {value="hello", is_secret=false, allow_override=true}}
	varObj, diags := types.ObjectValue(variableValueAttrTypes, map[string]attr.Value{
		"value":          types.StringValue("hello"),
		"is_secret":      types.BoolValue(false),
		"allow_override": types.BoolValue(true),
	})
	require.False(t, diags.HasError(), "build variable object: %s", diags)

	varMap, diags := types.MapValue(types.ObjectType{AttrTypes: variableValueAttrTypes}, map[string]attr.Value{
		"MY_VAR": varObj,
	})
	require.False(t, diags.HasError(), "build variables map: %s", diags)

	result, err := expandVariablesFramework(ctx, varMap)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, *result, 1)

	v, ok := (*result)["MY_VAR"]
	require.True(t, ok, "MY_VAR key should be present")
	require.NotNil(t, v.Value)
	assert.Equal(t, "hello", *v.Value)
	require.NotNil(t, v.AllowOverride)
	assert.Equal(t, true, *v.AllowOverride)
	require.NotNil(t, v.IsSecret)
	assert.Equal(t, false, *v.IsSecret)
}

// TestFrameworkReleaseDefinition_flattenVariables verifies that flattenVariablesFramework
// correctly converts a ConfigurationVariableValue map back into a types.Map.
func TestFrameworkReleaseDefinition_flattenVariables(t *testing.T) {
	ctx := context.Background()

	apiVars := map[string]releaseapi.ConfigurationVariableValue{
		"MY_VAR": {
			Value:         converter.String("world"),
			IsSecret:      converter.Bool(false),
			AllowOverride: converter.Bool(false),
		},
	}

	result, diags := flattenVariablesFramework(ctx, &apiVars, types.MapNull(types.ObjectType{AttrTypes: variableValueAttrTypes}))
	require.False(t, diags.HasError(), "flattenVariablesFramework returned errors: %s", diags)
	require.False(t, result.IsNull())
	require.False(t, result.IsUnknown())

	var models map[string]variableValueModel
	diags = result.ElementsAs(ctx, &models, false)
	require.False(t, diags.HasError(), "ElementsAs: %s", diags)
	require.Len(t, models, 1)

	v, ok := models["MY_VAR"]
	require.True(t, ok)
	assert.Equal(t, "world", v.Value.ValueString())
	assert.Equal(t, false, v.IsSecret.ValueBool())
	assert.Equal(t, false, v.AllowOverride.ValueBool())
}

// TestFrameworkReleaseDefinition_expandArtifact verifies that expandArtifactsFramework
// correctly converts an artifactModel list to []releaseapi.Artifact, and that
// flattenArtifactsFramework strips API-computed definition_reference keys.
func TestFrameworkReleaseDefinition_expandArtifact(t *testing.T) {
	ctx := context.Background()

	// Build definition_reference map: {"definition": {"id": "42"}}
	// In the framework model this is a flat map[string]string where values are ArtifactSourceReference.Id strings
	defRefMap, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"definition": types.StringValue("42"),
		"project":    types.StringValue("proj-123"),
	})
	require.False(t, diags.HasError(), "build definition_reference: %s", diags)

	// Build artifact object
	artObj, diags := types.ObjectValue(artifactAttrTypes, map[string]attr.Value{
		"alias":                types.StringValue("_build"),
		"type":                 types.StringValue("Build"),
		"is_primary":           types.BoolValue(true),
		"definition_reference": defRefMap,
	})
	require.False(t, diags.HasError(), "build artifact object: %s", diags)

	artList, diags := types.ListValue(types.ObjectType{AttrTypes: artifactAttrTypes}, []attr.Value{artObj})
	require.False(t, diags.HasError(), "build artifact list: %s", diags)

	// Expand
	artifacts, err := expandArtifactsFramework(ctx, artList)
	require.NoError(t, err)
	require.Len(t, artifacts, 1)

	a := artifacts[0]
	require.NotNil(t, a.Alias)
	assert.Equal(t, "_build", *a.Alias)
	require.NotNil(t, a.Type)
	assert.Equal(t, "Build", *a.Type)
	require.NotNil(t, a.IsPrimary)
	assert.Equal(t, true, *a.IsPrimary)
	require.NotNil(t, a.DefinitionReference)
	defID, ok := (*a.DefinitionReference)["definition"]
	require.True(t, ok)
	require.NotNil(t, defID.Id)
	assert.Equal(t, "42", *defID.Id)

	// Now simulate API response that injects extra computed keys
	a.DefinitionReference = &map[string]releaseapi.ArtifactSourceReference{
		"definition":                  {Id: converter.String("42")},
		"project":                     {Id: converter.String("proj-123")},
		"artifactSourceDefinitionUrl": {Id: converter.String("https://example.com")},
		"defaultVersionType":          {Id: converter.String("latestType")},
	}

	// Flatten and assert the computed keys are stripped
	flatList, diags := flattenArtifactsFramework(ctx, &[]releaseapi.Artifact{a})
	require.False(t, diags.HasError(), "flattenArtifactsFramework: %s", diags)
	require.False(t, flatList.IsNull())

	var flatModels []artifactModel
	diags = flatList.ElementsAs(ctx, &flatModels, false)
	require.False(t, diags.HasError(), "ElementsAs: %s", diags)
	require.Len(t, flatModels, 1)

	fm := flatModels[0]
	assert.Equal(t, "_build", fm.Alias.ValueString())

	var defRef map[string]string
	diags = fm.DefinitionReference.ElementsAs(ctx, &defRef, false)
	require.False(t, diags.HasError(), "ElementsAs defRef: %s", diags)

	// API-computed keys should be stripped
	assert.Equal(t, "42", defRef["definition"])
	assert.Equal(t, "proj-123", defRef["project"])
	_, hasUrl := defRef["artifactSourceDefinitionUrl"]
	assert.False(t, hasUrl, "artifactSourceDefinitionUrl should be stripped")
	_, hasVersion := defRef["defaultVersionType"]
	assert.False(t, hasVersion, "defaultVersionType should be stripped")
}

// TestFrameworkReleaseDefinition_expandTriggers verifies that expandTriggersFramework
// and flattenTriggersFramework round-trip correctly for both trigger types.
func TestFrameworkReleaseDefinition_expandTriggers(t *testing.T) {
	ctx := context.Background()

	// Build cd_artifact_trigger
	cdObj, diags := types.ObjectValue(cdArtifactTriggerAttrTypes, map[string]attr.Value{
		"artifact_alias":                  types.StringValue("_build"),
		"branch_filters":                  types.ListValueMust(types.StringType, []attr.Value{}),
		"tag_filter":                      types.ListValueMust(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{}),
		"use_build_definition_branch":     types.BoolValue(false),
		"create_release_on_build_tagging": types.BoolValue(false),
	})
	require.False(t, diags.HasError(), "build cd trigger: %s", diags)
	cdList, diags := types.ListValue(types.ObjectType{AttrTypes: cdArtifactTriggerAttrTypes}, []attr.Value{cdObj})
	require.False(t, diags.HasError(), "build cd list: %s", diags)

	// Build schedule_trigger
	schObj, diags := types.ObjectValue(scheduleTriggerAttrTypes, map[string]attr.Value{
		"schedule_only_with_changes": types.BoolValue(true),
		"start_hours":                types.Int64Value(2),
		"start_minutes":              types.Int64Value(30),
		"time_zone_id":               types.StringValue("UTC"),
		"days_to_release":            types.Int64Value(0),
	})
	require.False(t, diags.HasError(), "build schedule trigger: %s", diags)
	schList, diags := types.ListValue(types.ObjectType{AttrTypes: scheduleTriggerAttrTypes}, []attr.Value{schObj})
	require.False(t, diags.HasError(), "build sch list: %s", diags)

	// Build trigger model (empty lists for new trigger types not exercised by this test)
	emptySourceRepoList := types.ListValueMust(types.ObjectType{AttrTypes: sourceRepoTriggerAttrTypes}, []attr.Value{})
	emptyCIList := types.ListValueMust(types.ObjectType{AttrTypes: containerImageTriggerAttrTypes}, []attr.Value{})
	emptyPRList := types.ListValueMust(types.ObjectType{AttrTypes: pullRequestTriggerAttrTypes}, []attr.Value{})

	trigObj, diags := types.ObjectValue(triggerAttrTypes, map[string]attr.Value{
		"cd_artifact_trigger":     cdList,
		"schedule_trigger":        schList,
		"source_repo_trigger":     emptySourceRepoList,
		"container_image_trigger": emptyCIList,
		"pull_request_trigger":    emptyPRList,
	})
	require.False(t, diags.HasError(), "build trigger obj: %s", diags)
	trigList, diags := types.ListValue(types.ObjectType{AttrTypes: triggerAttrTypes}, []attr.Value{trigObj})
	require.False(t, diags.HasError(), "build trigger list: %s", diags)

	// Expand
	raw, err := expandTriggersFramework(ctx, trigList)
	require.NoError(t, err)
	require.Len(t, raw, 2, "expect 1 CD + 1 schedule = 2 raw trigger entries")

	// Find the CD trigger and schedule trigger
	var foundCD, foundSched bool
	for _, r := range raw {
		m, ok := r.(map[string]interface{})
		require.True(t, ok)
		switch m["triggerType"] {
		case "artifactSource":
			foundCD = true
			assert.Equal(t, "_build", m["artifactAlias"])
		case "schedule":
			foundSched = true
			sched, ok := m["schedule"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, true, sched["scheduleOnlyWithChanges"])
			assert.Equal(t, 2, sched["startHours"])
			assert.Equal(t, 30, sched["startMinutes"])
		}
	}
	assert.True(t, foundCD, "CD artifact trigger should be present")
	assert.True(t, foundSched, "schedule trigger should be present")

	// Flatten
	flatList, diags := flattenTriggersFramework(ctx, &raw)
	require.False(t, diags.HasError(), "flattenTriggersFramework: %s", diags)
	require.False(t, flatList.IsNull())

	var flatModels []triggerModel
	diags = flatList.ElementsAs(ctx, &flatModels, false)
	require.False(t, diags.HasError(), "ElementsAs: %s", diags)
	require.Len(t, flatModels, 1)

	tm := flatModels[0]

	var cdModels []cdArtifactTriggerModel
	diags = tm.CdArtifactTrigger.ElementsAs(ctx, &cdModels, false)
	require.False(t, diags.HasError(), "ElementsAs cdModels: %s", diags)
	require.Len(t, cdModels, 1)
	assert.Equal(t, "_build", cdModels[0].ArtifactAlias.ValueString())

	var schModels []scheduleTriggerModel
	diags = tm.ScheduleTrigger.ElementsAs(ctx, &schModels, false)
	require.False(t, diags.HasError(), "ElementsAs schModels: %s", diags)
	require.Len(t, schModels, 1)
	assert.Equal(t, true, schModels[0].ScheduleOnlyWithChanges.ValueBool())
	assert.Equal(t, int64(2), schModels[0].StartHours.ValueInt64())
	assert.Equal(t, int64(30), schModels[0].StartMinutes.ValueInt64())
	assert.Equal(t, "UTC", schModels[0].TimeZoneID.ValueString())
}

// TestReleaseDefinitionFramework_UpgradeState verifies that UpgradeState() returns a map with
// exactly one entry keyed at version 0.
func TestReleaseDefinitionFramework_UpgradeState(t *testing.T) {
	ctx := context.Background()
	r := &releaseDefinitionFrameworkResource{}

	upgraders := r.UpgradeState(ctx)
	require.Len(t, upgraders, 1, "UpgradeState must return exactly one upgrader")
	_, ok := upgraders[0]
	assert.True(t, ok, "UpgradeState map must contain key 0")
}

// TestReleaseDefinitionFramework_SchemaVersion verifies that Schema() declares Version 1.
func TestReleaseDefinitionFramework_SchemaVersion(t *testing.T) {
	ctx := context.Background()
	r := &releaseDefinitionFrameworkResource{}

	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "Schema() must not error: %s", resp.Diagnostics)
	assert.Equal(t, int64(1), resp.Schema.Version, "Schema.Version must be 1")
}

// TestFrameworkReleaseDefinition_staleRevisionRetry verifies that retryOnStaleRevision detects
// the stale-revision HTTP 400 error, re-reads the current revision via GetReleaseDefinition,
// and retries UpdateReleaseDefinition exactly once.
//
// Mock sequence:
//  1. First UpdateReleaseDefinition → HTTP 400 / InvalidRequestException / "old copy of the release pipeline"
//  2. GetReleaseDefinition → returns definition with Revision=3 (fresh live revision)
//  3. Second UpdateReleaseDefinition → succeeds, returns updated definition
func TestFrameworkReleaseDefinition_staleRevisionRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := azdosdkmocks.NewMockReleaseClient(ctrl)

	projectID := "00000000-0000-0000-0000-000000000099"
	defID := 7
	initialRevision := 1
	freshRevision := 3

	// The definition to update (starts with stale revision 1).
	def := &releaseapi.ReleaseDefinition{
		Id:       &defID,
		Name:     converter.String("RetryTest"),
		Revision: &initialRevision,
	}

	// 1. First update fails with stale-revision 400.
	conflictMsg := fmt.Sprintf("You are updating an old copy of the release pipeline. Please refresh your copy and try again.")
	statusCode := 400
	conflictErr := azuredevops.WrappedError{
		TypeKey:    converter.String("InvalidRequestException"),
		Message:    &conflictMsg,
		StatusCode: &statusCode,
	}

	// 2. Re-read returns a definition with the fresh revision.
	freshDef := &releaseapi.ReleaseDefinition{
		Id:       &defID,
		Name:     converter.String("RetryTest"),
		Revision: &freshRevision,
	}

	// 3. Second update succeeds and returns the updated definition.
	updatedDef := &releaseapi.ReleaseDefinition{
		Id:       &defID,
		Name:     converter.String("RetryTest"),
		Revision: &freshRevision,
	}

	gomock.InOrder(
		mockClient.EXPECT().
			UpdateReleaseDefinition(ctx, gomock.Any()).
			Return(nil, conflictErr).
			Times(1),
		mockClient.EXPECT().
			GetReleaseDefinition(ctx, gomock.Any()).
			Return(freshDef, nil).
			Times(1),
		mockClient.EXPECT().
			UpdateReleaseDefinition(ctx, gomock.Any()).
			Return(updatedDef, nil).
			Times(1),
	)

	result, err := retryOnStaleRevision(ctx, mockClient, def, projectID, defID)

	require.NoError(t, err, "retryOnStaleRevision should succeed after one retry")
	require.NotNil(t, result)
	assert.Equal(t, freshRevision, *result.Revision, "result should carry the fresh revision from the retry")
	// Confirm the definition's revision was patched to the fresh value before the retry.
	assert.Equal(t, freshRevision, *def.Revision, "def.Revision should be updated to the fresh revision")
}
