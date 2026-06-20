//go:build all || resource_task_group_framework

package taskagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// taskGroupStateUpgraderV0 returns a resource.StateUpgrader that migrates
// schema version 0 → 1 for betterado_task_group.
//
// The SDKv2 resource stored task, input and version as TypeList blocks.
// The framework resource uses ListNestedAttribute — the JSON shape is
// structurally the same, so the upgrader is a pass-through: it decodes the
// raw JSON, ensures that the "task" key is present (defaulting to an empty
// list if absent), re-encodes and returns the result.
func taskGroupStateUpgraderV0() resource.StateUpgrader {
	return resource.StateUpgrader{
		// No PriorSchema: we use the raw JSON path to avoid redeclaring the
		// old SDK-v2 schema.
		StateUpgrader: upgradeTaskGroupV0StateFunc,
	}
}

// upgradeTaskGroupV0StateFunc is the concrete upgrade logic for schema
// version 0 → 1.
func upgradeTaskGroupV0StateFunc(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil {
		resp.Diagnostics.AddError(
			"Missing Prior State",
			"A state upgrade was requested but no prior state was found. "+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	rawJSON := req.RawState.JSON
	if rawJSON == nil {
		resp.Diagnostics.AddError(
			"Prior State Not in JSON Format",
			"The prior resource state is not in JSON format (possible flatmap state from Terraform < 0.12). "+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	// Decode raw state into a generic map so we can manipulate it without
	// needing to redeclare the full v0 schema.
	var raw map[string]interface{}
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Prior State JSON",
			fmt.Sprintf("An error occurred while parsing the prior resource state JSON: %s", err),
		)
		return
	}

	// No key renames are required between SDK-v2 and framework for task_group.
	// Ensure "task" is present as an empty list if the prior state omitted it,
	// so the current schema's Required list attribute is satisfied.
	if _, ok := raw["task"]; !ok {
		raw["task"] = []interface{}{}
	}

	// Re-marshal the (possibly modified) map to JSON.
	upgradedJSON, err := json.Marshal(raw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Serialise Upgraded State",
			fmt.Sprintf("An error occurred while serialising the upgraded resource state: %s", err),
		)
		return
	}

	// Return the upgraded state via DynamicValue so the framework can
	// unmarshal it against the current schema type without requiring us to
	// populate the full tfsdk.State model.
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: upgradedJSON,
	}
}
