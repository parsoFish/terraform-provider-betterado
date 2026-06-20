//go:build all || resource_task_group_framework

package taskagent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskGroupUpgradeV0_WithTasks verifies that the v0 upgrader passes through
// a state that contains a "task" array without error and preserves the tasks.
func TestTaskGroupUpgradeV0_WithTasks(t *testing.T) {
	ctx := context.Background()

	// Build a minimal v0 state JSON with SDKv2-shaped task list objects.
	priorState := map[string]interface{}{
		"schema_version": 0,
		"id":             "00000000-0000-0000-0000-000000000001",
		"project_id":     "00000000-0000-0000-0000-000000000002",
		"name":           "My Task Group",
		"friendly_name":  "My Task Group",
		"category":       "Build",
		"task": []interface{}{
			map[string]interface{}{
				"display_name":                "Run Script",
				"task_id":                     "00000000-0000-0000-0000-000000000003",
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
	}
	priorJSON, err := json.Marshal(priorState)
	require.NoError(t, err)

	upgrader := taskGroupStateUpgraderV0()

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: priorJSON,
		},
	}
	resp := &resource.UpgradeStateResponse{}

	upgrader.StateUpgrader(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "upgrader returned unexpected diagnostics: %v", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue, "DynamicValue should be set by the upgrader")

	// Decode the upgraded JSON and verify the task list is preserved.
	var upgraded map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &upgraded))

	assert.Contains(t, upgraded, "task", "upgraded state must contain 'task' key")
	tasks, ok := upgraded["task"].([]interface{})
	require.True(t, ok, "'task' value should be an array")
	require.Len(t, tasks, 1, "task list should have exactly one element")

	taskMap, ok := tasks[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Run Script", taskMap["display_name"])
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", taskMap["task_id"])
}

// TestTaskGroupUpgradeV0_EmptyState verifies that the v0 upgrader succeeds when
// the prior state has no "task" key, producing an empty "task" list without
// error or panic.
func TestTaskGroupUpgradeV0_EmptyState(t *testing.T) {
	ctx := context.Background()

	// Prior state has no "task" key at all.
	priorState := map[string]interface{}{
		"schema_version": 0,
		"id":             "00000000-0000-0000-0000-000000000001",
		"project_id":     "00000000-0000-0000-0000-000000000002",
		"name":           "My Task Group",
		"friendly_name":  "My Task Group",
		"category":       "Build",
	}
	priorJSON, err := json.Marshal(priorState)
	require.NoError(t, err)

	upgrader := taskGroupStateUpgraderV0()

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: priorJSON,
		},
	}
	resp := &resource.UpgradeStateResponse{}

	upgrader.StateUpgrader(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "upgrader returned unexpected diagnostics: %v", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue, "DynamicValue should be set by the upgrader")

	// Decode the upgraded JSON.
	var upgraded map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &upgraded))

	// Must contain "task" as an empty list.
	taskRaw, hasTask := upgraded["task"]
	assert.True(t, hasTask, "upgraded state must contain 'task' key")
	if hasTask {
		tasks, ok := taskRaw.([]interface{})
		assert.True(t, ok, "'task' should be an array, got %T", taskRaw)
		assert.Empty(t, tasks, "'task' should be empty when absent from prior state")
	}
}
