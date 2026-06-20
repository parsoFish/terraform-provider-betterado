//go:build all || resource_release_definition

package release

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// releaseDefinitionStateUpgraderV0 returns a resource.StateUpgrader that
// migrates schema version 0 → 1 for betterado_release_definition.
//
// Schema v0 used an "environment" top-level key; schema v1 renamed it to
// "stages". All other keys are preserved unchanged.
func releaseDefinitionStateUpgraderV0() resource.StateUpgrader {
	return resource.StateUpgrader{
		// No PriorSchema: we use the raw JSON path to avoid redeclaring the
		// old SDK-v2 schema.
		StateUpgrader: upgradeV0StateFunc,
	}
}

// upgradeV0StateFunc is the concrete upgrade logic for schema version 0 → 1.
func upgradeV0StateFunc(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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

	// Decode raw state into a generic map so we can rename the key without
	// needing to redeclare the full v0 schema.
	var raw map[string]interface{}
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Prior State JSON",
			fmt.Sprintf("An error occurred while parsing the prior resource state JSON: %s", err),
		)
		return
	}

	// Rename "environment" → "stages". If absent, ensure "stages" is
	// present as an empty list so the current schema is satisfied.
	if envVal, ok := raw["environment"]; ok {
		raw["stages"] = envVal
		delete(raw, "environment")
	} else if _, hasStages := raw["stages"]; !hasStages {
		raw["stages"] = []interface{}{}
	}

	// Re-marshal the modified map to JSON.
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
