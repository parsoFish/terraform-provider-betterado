//go:build all || resource_release_definition

package release

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseDefinitionUpgradeV0_RenamesEnvironmentToStages verifies that the
// v0 upgrader renames the top-level "environment" key to "stages" and removes
// the old key.
func TestReleaseDefinitionUpgradeV0_RenamesEnvironmentToStages(t *testing.T) {
	ctx := context.Background()

	// Build a minimal v0 state JSON that has an "environment" array.
	priorState := map[string]interface{}{
		"schema_version": 0,
		"environment": []interface{}{
			map[string]interface{}{
				"name": "Production",
				"rank": 1,
			},
		},
	}
	priorJSON, err := json.Marshal(priorState)
	require.NoError(t, err)

	upgrader := releaseDefinitionStateUpgraderV0()

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: priorJSON,
		},
	}
	resp := &resource.UpgradeStateResponse{}

	upgrader.StateUpgrader(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "upgrader returned unexpected diagnostics: %v", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue, "DynamicValue should be set by the upgrader")

	// Decode the upgraded JSON and assert key rename.
	var upgraded map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &upgraded))

	assert.Contains(t, upgraded, "stages", "upgraded state must contain 'stages' key")
	assert.NotContains(t, upgraded, "environment", "upgraded state must not contain 'environment' key")

	// Value should be transferred unchanged.
	stages, ok := upgraded["stages"].([]interface{})
	require.True(t, ok, "'stages' value should be an array")
	require.Len(t, stages, 1)
	stageMap, ok := stages[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Production", stageMap["name"])
}

// TestReleaseDefinitionUpgradeV0_MissingEnvironment verifies that the v0
// upgrader succeeds even when the prior state has no "environment" key,
// producing a "stages" list that is empty (or null) and no error.
func TestReleaseDefinitionUpgradeV0_MissingEnvironment(t *testing.T) {
	ctx := context.Background()

	// Prior state has no "environment" key at all.
	priorState := map[string]interface{}{
		"schema_version": 0,
		"name":           "MyDef",
	}
	priorJSON, err := json.Marshal(priorState)
	require.NoError(t, err)

	upgrader := releaseDefinitionStateUpgraderV0()

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

	// Must not contain "environment".
	assert.NotContains(t, upgraded, "environment", "upgraded state must not contain 'environment' key")

	// Must contain "stages" as an empty list.
	stagesRaw, hasstages := upgraded["stages"]
	assert.True(t, hasstages, "upgraded state must contain 'stages' key")
	if hasstages {
		stages, ok := stagesRaw.([]interface{})
		assert.True(t, ok, "'stages' should be an array, got %T", stagesRaw)
		assert.Empty(t, stages, "'stages' should be empty when 'environment' was absent")
	}
}
