package feed

// framework_defaults.go — inline implementations of terraform-plugin-framework
// plan modifiers used by resource_feed_framework.go. These mirror the
// string plan modifiers available in the release package (which are not
// exported) so the feed package remains self-contained.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ── String plan modifiers ─────────────────────────────────────────────────────

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
	if req.StateValue.IsNull() {
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
