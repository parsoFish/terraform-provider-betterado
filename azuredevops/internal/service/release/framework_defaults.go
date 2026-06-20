package release

// framework_defaults.go — inline implementations of terraform-plugin-framework
// schema defaults and plan modifiers whose sub-packages are not vendored in
// this project. Only the types used by resource_release_definition_framework.go
// are implemented here.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Bool default ────────────────────────────────────────────────────────────

type staticBoolDefault struct{ value bool }

func staticBool(v bool) defaults.Bool { return staticBoolDefault{value: v} }

func (s staticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", s.value)
}

func (s staticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", s.value)
}

func (s staticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// ── String default ───────────────────────────────────────────────────────────

type staticStringDefault struct{ value string }

func staticString(v string) defaults.String { return staticStringDefault{value: v} }

func (s staticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", s.value)
}

func (s staticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", s.value)
}

func (s staticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(s.value)
}

// ── Int64 default ────────────────────────────────────────────────────────────

type staticInt64Default struct{ value int64 }

func staticInt64(v int64) defaults.Int64 { return staticInt64Default{value: v} }

func (s staticInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", s.value)
}

func (s staticInt64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%d`", s.value)
}

func (s staticInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(s.value)
}

// ── String plan modifiers ────────────────────────────────────────────────────

// useStateForUnknownString is equivalent to stringplanmodifier.UseStateForUnknown().
type useStateForUnknownString struct{}

func useStateForUnknown() planmodifier.String { return useStateForUnknownString{} }

func (useStateForUnknownString) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (useStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (useStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// On create there is no prior state; leave the plan value as Unknown so
	// Terraform shows "(known after apply)" rather than null. Setting it to
	// a null state value would confuse the post-apply consistency check.
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// useStateForEmptyString is a plan modifier that substitutes the prior state
// value when the plan value is an empty string (the schema Default("") fires
// for omitted Optional+Computed string attributes). This keeps attributes whose
// value is ADO-generated (e.g. deploy-phase name) idempotent across apply
// cycles without requiring an explicit value in the HCL config.
type useStateForEmptyString struct{}

func useStateForEmpty() planmodifier.String { return useStateForEmptyString{} }

func (useStateForEmptyString) Description(_ context.Context) string {
	return "use prior state value when plan value is empty"
}

func (useStateForEmptyString) MarkdownDescription(_ context.Context) string {
	return "use prior state value when plan value is empty"
}

func (useStateForEmptyString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only act when the plan value is the empty string (default placeholder).
	if req.PlanValue.ValueString() != "" {
		return
	}
	// No prior state (create); leave the plan as-is so provider can default it.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.StateValue.ValueString() == "" {
		return
	}
	// Substitute the non-empty prior state value so plans stay idempotent.
	resp.PlanValue = req.StateValue
}

// ── List plan modifiers ──────────────────────────────────────────────────────

// useStateForUnknownList is equivalent to listplanmodifier.UseStateForUnknown().
// It substitutes the prior state value when the plan value is Unknown — this
// suppresses spurious "(known after apply)" for Optional+Computed list attributes
// whose values are populated by the Read function (e.g. ADO-default fields that
// the user didn't set in HCL). If the plan value is NOT Unknown (config
// explicitly sets it), the modifier does nothing and the config value wins.
type useStateForUnknownList struct{}

func useStateForUnknownListModifier() planmodifier.List { return useStateForUnknownList{} }

func (useStateForUnknownList) Description(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownList) MarkdownDescription(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownList) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// No prior state (create) — leave Unknown so Terraform shows "(known after apply)".
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Map plan modifiers ───────────────────────────────────────────────────────

// useStateForUnknownMap is equivalent to mapplanmodifier.UseStateForUnknown().
type useStateForUnknownMap struct{}

func useStateForUnknownMapModifier() planmodifier.Map { return useStateForUnknownMap{} }

func (useStateForUnknownMap) Description(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownMap) MarkdownDescription(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownMap) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// No prior state (create) — leave Unknown.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Set plan modifiers ───────────────────────────────────────────────────────

// useStateForUnknownSet is equivalent to setplanmodifier.UseStateForUnknown().
type useStateForUnknownSet struct{}

func useStateForUnknownSetModifier() planmodifier.Set { return useStateForUnknownSet{} }

func (useStateForUnknownSet) Description(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownSet) MarkdownDescription(_ context.Context) string {
	return "use prior state value when plan is unknown"
}

func (useStateForUnknownSet) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// No prior state (create) — leave Unknown so Terraform shows "(known after apply)".
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// requiresReplaceString is equivalent to stringplanmodifier.RequiresReplace().
type requiresReplaceString struct{}

func requiresReplace() planmodifier.String { return requiresReplaceString{} }

func (requiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (requiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (requiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}
