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
	resp.PlanValue = req.StateValue
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
