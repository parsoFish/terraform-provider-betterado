package approvalsandchecks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ── Static default helpers ────────────────────────────────────────────────────

// staticCheckBool returns a defaults.Bool that always returns v.
type staticCheckBoolDefault struct{ value bool }

func staticCheckBool(v bool) defaults.Bool { return staticCheckBoolDefault{value: v} }

func (d staticCheckBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}
func (d staticCheckBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}
func (d staticCheckBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// staticCheckInt64 returns a defaults.Int64 that always returns v.
type staticCheckInt64Default struct{ value int64 }

func staticCheckInt64(v int64) defaults.Int64 { return staticCheckInt64Default{value: v} }

func (d staticCheckInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", d.value)
}
func (d staticCheckInt64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%d`", d.value)
}
func (d staticCheckInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.value)
}

// staticCheckString returns a defaults.String that always returns v.
type staticCheckStringDefault struct{ value string }

func staticCheckString(v string) defaults.String { return staticCheckStringDefault{value: v} }

func (d staticCheckStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}
func (d staticCheckStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}
func (d staticCheckStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// checkUseStateForUnknownString holds the prior state for string attributes.
type checkUseStateForUnknownString struct{}

func checkUseStateForUnknown() planmodifier.String { return checkUseStateForUnknownString{} }

func (m checkUseStateForUnknownString) Description(_ context.Context) string {
	return "uses prior state for unknown values"
}
func (m checkUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "uses prior state for unknown values"
}
func (m checkUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// If there is no prior state (create), keep the plan value Unknown so the
	// provider can set any value after apply without causing an inconsistency.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// checkUseStateForUnknownInt64 holds the prior state for int64 attributes.
type checkUseStateForUnknownInt64 struct{}

func checkUseStateForUnknownInt64Val() planmodifier.Int64 { return checkUseStateForUnknownInt64{} }

func (m checkUseStateForUnknownInt64) Description(_ context.Context) string {
	return "uses prior state for unknown values"
}
func (m checkUseStateForUnknownInt64) MarkdownDescription(_ context.Context) string {
	return "uses prior state for unknown values"
}
func (m checkUseStateForUnknownInt64) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// If there is no prior state (create), keep the plan value Unknown so the
	// provider can set any value after apply without causing an inconsistency.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// checkVersionPlanModifier handles the 'version' field, which the ADO server
// increments on every write. During a no-op re-plan (no other attribute changes),
// it preserves the state value so the plan appears empty (idempotency). During
// an actual update (other attributes differ between plan and state), it keeps the
// version as Unknown so the post-apply API value (n+1) does not conflict with a
// planned value of n.
type checkVersionPlanModifier struct{}

func checkVersionPlanModifierFn() planmodifier.Int64 { return checkVersionPlanModifier{} }

func (m checkVersionPlanModifier) Description(_ context.Context) string {
	return "preserves version for no-op re-plans; keeps Unknown when other attributes change"
}
func (m checkVersionPlanModifier) MarkdownDescription(_ context.Context) string {
	return "preserves version for no-op re-plans; keeps Unknown when other attributes change"
}
func (m checkVersionPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		// Already known (user-specified); do not override.
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		// Create path: no prior state, keep Unknown so the provider can write any value.
		return
	}
	// At this point we have a prior known version in state and Unknown in the plan.
	// Determine whether this is a no-op re-plan (nothing else changed) or an actual
	// update (other attributes changed → version will be incremented by the server).
	// We detect changes by comparing the full plan and state tftypes.Value trees,
	// ignoring the 'version' attribute itself.
	planRaw := req.Plan.Raw
	stateRaw := req.State.Raw
	if planRaw.IsNull() || !planRaw.IsKnown() || stateRaw.IsNull() || !stateRaw.IsKnown() {
		// Cannot compare — err on the side of preserving state (safe for re-plans).
		resp.PlanValue = req.StateValue
		return
	}
	// Walk both values, comparing non-version attributes.
	otherChanged := checkOtherAttrsChanged(planRaw, stateRaw, req.Path)
	if !otherChanged {
		// No-op re-plan: preserve the state version for idempotency.
		resp.PlanValue = req.StateValue
	}
	// else: actual update — keep Unknown so the post-apply version (n+1) is accepted.
}

// checkOtherAttrsChanged returns true if the plan and state tftypes.Value differ
// in any top-level attribute EXCEPT the one at excludePath.
func checkOtherAttrsChanged(plan, state tftypes.Value, excludePath path.Path) bool {
	if !plan.Type().Is(tftypes.Object{}) || !state.Type().Is(tftypes.Object{}) {
		return !plan.Equal(state)
	}
	var planAttrs map[string]tftypes.Value
	var stateAttrs map[string]tftypes.Value
	if err := plan.As(&planAttrs); err != nil {
		return true
	}
	if err := state.As(&stateAttrs); err != nil {
		return true
	}
	// Determine the attribute name to exclude (the version attribute's last step).
	excludeAttr := ""
	steps := excludePath.Steps()
	if len(steps) == 1 {
		if attr, ok := steps[0].(path.PathStepAttributeName); ok {
			excludeAttr = string(attr)
		}
	}
	for k, pv := range planAttrs {
		if k == excludeAttr {
			continue
		}
		sv, ok := stateAttrs[k]
		if !ok {
			return true
		}
		// Skip Unknown plan values (other computed fields also being planned).
		if !pv.IsKnown() {
			continue
		}
		if !pv.Equal(sv) {
			return true
		}
	}
	return false
}

// checkRequiresReplace forces recreation when a string attribute changes.
type checkRequiresReplaceModifier struct{}

func checkRequiresReplace() planmodifier.String { return checkRequiresReplaceModifier{} }

func (m checkRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m checkRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m checkRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Import state helper ───────────────────────────────────────────────────────

// importCheckState provides a generic ImportState for check resources.
// Import ID format: <project_id>/<check_id>
func importCheckState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected format: <project_id>/<check_id>, got: %s", req.ID))
		return
	}
	projectID := parts[0]
	checkIDStr := parts[1]
	if _, err := strconv.Atoi(checkIDStr); err != nil {
		resp.Diagnostics.AddError("Invalid check ID in import",
			fmt.Sprintf("check ID must be an integer, got: %s", checkIDStr))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), checkIDStr)...)
}
