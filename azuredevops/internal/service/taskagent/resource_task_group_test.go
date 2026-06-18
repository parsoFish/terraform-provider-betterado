//go:build (all || resource_task_group) && !exclude_resource_task_group
// +build all resource_task_group
// +build !exclude_resource_task_group

package taskagent

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Package-level fixtures ─────────────────────────────────────────────────

var testTaskGroupProjectID = uuid.New()
var testTaskGroupID = uuid.New()
var testTaskStepTaskID = uuid.New()

var testTaskGroupRevision = 1
var testTaskVersionMajor = 1
var testTaskVersionMinor = 0
var testTaskVersionPatch = 0
var testTaskVersionIsTest = false

var testTaskGroup = taskagent.TaskGroup{
	Id:           &testTaskGroupID,
	Name:         converter.String("MyTaskGroup"),
	FriendlyName: converter.String("My Task Group"),
	Description:  converter.String("A test task group"),
	Category:     converter.String("Build"),
	Author:       converter.String("tester"),
	Revision:     &testTaskGroupRevision,
	Version: &taskagent.TaskVersion{
		Major:  &testTaskVersionMajor,
		Minor:  &testTaskVersionMinor,
		Patch:  &testTaskVersionPatch,
		IsTest: &testTaskVersionIsTest,
	},
	Tasks: &[]taskagent.TaskGroupStep{
		{
			DisplayName:     converter.String("Run Script"),
			Enabled:         converter.Bool(true),
			AlwaysRun:       converter.Bool(false),
			ContinueOnError: converter.Bool(false),
			Condition:       converter.String("succeeded()"),
			Task: &taskagent.TaskDefinitionReference{
				Id:             &testTaskStepTaskID,
				VersionSpec:    converter.String("1.*"),
				DefinitionType: converter.String("task"),
			},
		},
	},
}

// ── 1. Roundtrip ──────────────────────────────────────────────────────────

// TestTaskGroup_ExpandFlatten_Roundtrip verifies that flattenTaskGroup followed
// by expandTaskGroupCreate preserves the key fields of a TaskGroup.
func TestTaskGroup_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           testTaskGroupProjectID.String(),
		"name":                 "",
		"friendly_name":        "",
		"category":             "",
		"description":          "",
		"author":               "",
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   0,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "placeholder",
				"task_id":                     uuid.New().String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        0,
		"definition_type": "",
	})

	flattenTaskGroup(resourceData, &testTaskGroup)

	result := expandTaskGroupCreate(resourceData)
	require.NotNil(t, result)
	require.Equal(t, converter.ToString(testTaskGroup.Name, ""), converter.ToString(result.Name, ""))
	require.Equal(t, converter.ToString(testTaskGroup.FriendlyName, ""), converter.ToString(result.FriendlyName, ""))
	require.Equal(t, converter.ToString(testTaskGroup.Category, ""), converter.ToString(result.Category, ""))
	require.NotNil(t, result.Version)
	require.Equal(t, *testTaskGroup.Version.Major, *result.Version.Major)
	require.Equal(t, *testTaskGroup.Version.Minor, *result.Version.Minor)
	require.Equal(t, *testTaskGroup.Version.Patch, *result.Version.Patch)
	require.NotNil(t, result.Tasks)
	require.Equal(t, 1, len(*result.Tasks))
	require.Equal(t, converter.ToString((*testTaskGroup.Tasks)[0].DisplayName, ""), converter.ToString((*result.Tasks)[0].DisplayName, ""))
}

// ── 2. Create error propagation ────────────────────────────────────────────

// TestTaskGroup_Create_DoesNotSwallowError verifies that an error from
// AddTaskGroup surfaces as a non-empty Diagnostics (mirrors the environment
// test pattern).
func TestTaskGroup_Create_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	projectID := testTaskGroupProjectID.String()

	taskAgentClient.
		EXPECT().
		AddTaskGroup(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("AddTaskGroup() Failed")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           projectID,
		"name":                 "MyTaskGroup",
		"friendly_name":        "My Task Group",
		"category":             "Build",
		"description":          "",
		"author":               "",
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   1,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     testTaskStepTaskID.String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        0,
		"definition_type": "",
	})

	diags := resourceTaskGroupCreate(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "AddTaskGroup() Failed")
}

// ── 3. Read clears ID on 404 ───────────────────────────────────────────────

// TestTaskGroup_Read_ClearsIdOn404 verifies that when GetTaskGroups returns a
// 404 WrappedError, resourceTaskGroupRead clears the resource ID and returns
// no diagnostics (graceful drift detection).
func TestTaskGroup_Read_ClearsIdOn404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	notFoundStatusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &notFoundStatusCode,
	}

	taskAgentClient.
		EXPECT().
		GetTaskGroups(clients.Ctx, gomock.Any()).
		Return(nil, notFoundErr).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           testTaskGroupProjectID.String(),
		"name":                 "MyTaskGroup",
		"friendly_name":        "My Task Group",
		"category":             "Build",
		"description":          "",
		"author":               "",
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   1,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     testTaskStepTaskID.String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        0,
		"definition_type": "",
	})
	resourceData.SetId(testTaskGroupID.String())

	diags := resourceTaskGroupRead(context.Background(), resourceData, clients)
	require.Empty(t, diags)
	require.Equal(t, "", resourceData.Id())
}

// ── 4. Update calls SDK with args ──────────────────────────────────────────

// TestTaskGroup_Update_CallsSDKWithArgs verifies that resourceTaskGroupUpdate
// calls UpdateTaskGroup exactly once and then re-reads via GetTaskGroups.
func TestTaskGroup_Update_CallsSDKWithArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	updatedGroup := testTaskGroup

	taskAgentClient.
		EXPECT().
		UpdateTaskGroup(clients.Ctx, gomock.Any()).
		Return(&updatedGroup, nil).
		Times(1)

	taskAgentClient.
		EXPECT().
		GetTaskGroups(clients.Ctx, gomock.Any()).
		Return(&[]taskagent.TaskGroup{updatedGroup}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           testTaskGroupProjectID.String(),
		"name":                 "MyTaskGroup",
		"friendly_name":        "My Task Group",
		"category":             "Build",
		"description":          "",
		"author":               "",
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   1,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     testTaskStepTaskID.String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        testTaskGroupRevision,
		"definition_type": "",
	})
	resourceData.SetId(testTaskGroupID.String())

	diags := resourceTaskGroupUpdate(context.Background(), resourceData, clients)
	require.Empty(t, diags)
}

// ── 5. Delete surfaces API error ───────────────────────────────────────────

// TestTaskGroup_Delete_SurfacesAPIError verifies that an error from
// DeleteTaskGroup is surfaced as non-empty Diagnostics.
func TestTaskGroup_Delete_SurfacesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &client.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	taskAgentClient.
		EXPECT().
		DeleteTaskGroup(clients.Ctx, gomock.Any()).
		Return(errors.New("DeleteTaskGroup() Failed")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           testTaskGroupProjectID.String(),
		"name":                 "MyTaskGroup",
		"friendly_name":        "My Task Group",
		"category":             "Build",
		"description":          "",
		"author":               "",
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   1,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     testTaskStepTaskID.String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        0,
		"definition_type": "",
	})
	resourceData.SetId(testTaskGroupID.String())

	diags := resourceTaskGroupDelete(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "DeleteTaskGroup() Failed")
}

// ── 6. icon_url round-trip ─────────────────────────────────────────────────

// TestTaskGroup_ExpandFlatten_IconUrl verifies that icon_url is round-trippable:
// expandTaskGroupCreate produces a non-nil IconUrl from state, and flattenTaskGroup
// writes IconUrl back to state.
func TestTaskGroup_ExpandFlatten_IconUrl(t *testing.T) {
	iconURL := "https://example.com/icon.png"

	resourceData := schema.TestResourceDataRaw(t, ResourceTaskGroup().Schema, map[string]interface{}{
		"project_id":           testTaskGroupProjectID.String(),
		"name":                 "MyTaskGroup",
		"friendly_name":        "My Task Group",
		"category":             "Build",
		"description":          "",
		"author":               "",
		"icon_url":             iconURL,
		"instance_name_format": "",
		"runs_on":              []interface{}{},
		"version": []interface{}{
			map[string]interface{}{
				"major":   1,
				"minor":   0,
				"patch":   0,
				"is_test": false,
			},
		},
		"input": []interface{}{},
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     testTaskStepTaskID.String(),
				"task_version":                "1.*",
				"task_definition_type":        "task",
				"enabled":                     true,
				"always_run":                  false,
				"continue_on_error":           false,
				"condition":                   "succeeded()",
				"timeout_in_minutes":          0,
				"retry_count_on_task_failure": 0,
				"inputs":                      map[string]interface{}{},
				"environment":                 map[string]interface{}{},
			},
		},
		"revision":        0,
		"definition_type": "",
	})

	// Expand: icon_url in state → IconUrl in API struct
	result := expandTaskGroupCreate(resourceData)
	require.NotNil(t, result)
	require.NotNil(t, result.IconUrl, "expandTaskGroupCreate must set IconUrl when icon_url is non-empty")
	require.Equal(t, iconURL, *result.IconUrl)

	// Flatten: IconUrl in API struct → icon_url in state
	tgWithIcon := testTaskGroup
	tgWithIcon.IconUrl = converter.String(iconURL)
	flattenTaskGroup(resourceData, &tgWithIcon)
	require.Equal(t, iconURL, resourceData.Get("icon_url").(string))
}

// ── 7. input extended fields round-trip ────────────────────────────────────

// TestTaskGroup_ExpandFlatten_InputExtendedFields verifies that visible_rule,
// properties, and aliases in an input block are round-trippable through
// expandTaskInputs and flattenTaskInputs.
func TestTaskGroup_ExpandFlatten_InputExtendedFields(t *testing.T) {
	visibleRule := "targetType = filePath"
	propKey := "key"
	propVal := "val"
	alias := "alt_name"

	inputRaw := []interface{}{
		map[string]interface{}{
			"name":          "myInput",
			"label":         "My Input",
			"type":          "string",
			"default_value": "",
			"required":      false,
			"help_markdown": "",
			"group_name":    "",
			"options":       map[string]interface{}{},
			"visible_rule":  visibleRule,
			"properties":    map[string]interface{}{propKey: propVal},
			"aliases":       []interface{}{alias},
		},
	}

	// Expand: extended fields in raw map → SDK struct fields
	expanded := expandTaskInputs(inputRaw)
	require.Len(t, expanded, 1)

	inp := expanded[0]
	require.NotNil(t, inp.VisibleRule, "expandTaskInputs must set VisibleRule")
	require.Equal(t, visibleRule, *inp.VisibleRule)

	require.NotNil(t, inp.Properties, "expandTaskInputs must set Properties")
	require.Equal(t, propVal, (*inp.Properties)[propKey])

	require.NotNil(t, inp.Aliases, "expandTaskInputs must set Aliases")
	require.Len(t, *inp.Aliases, 1)
	require.Equal(t, alias, (*inp.Aliases)[0])

	// Flatten: SDK struct → map, assert extended fields come back
	flattened := flattenTaskInputs(&[]taskagent.TaskInputDefinition{inp})
	require.Len(t, flattened, 1)

	m := flattened[0].(map[string]interface{})
	require.Equal(t, visibleRule, m["visible_rule"].(string))

	props := m["properties"].(map[string]string)
	require.Equal(t, propVal, props[propKey])

	aliases := m["aliases"].([]string)
	require.Len(t, aliases, 1)
	require.Equal(t, alias, aliases[0])
}
