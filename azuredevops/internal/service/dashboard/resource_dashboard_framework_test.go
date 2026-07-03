package dashboard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Plan modifier unit tests ──────────────────────────────────────────────────

func TestDashboardUseStateForUnknown_Description(t *testing.T) {
	m := dashboardUseStateForUnknown()
	assert.NotEmpty(t, m.Description(context.Background()))
	assert.NotEmpty(t, m.MarkdownDescription(context.Background()))
}

func TestDashboardUseStateForUnknown_UnknownPlanWithKnownState(t *testing.T) {
	m := dashboardUseStateForUnknown()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringUnknown(),
		StateValue: types.StringValue("existing-id"),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	m.PlanModifyString(context.Background(), req, resp)
	// When plan is unknown and state has a value, the state value is used.
	assert.Equal(t, types.StringValue("existing-id"), resp.PlanValue)
}

func TestDashboardUseStateForUnknown_UnknownPlanWithNullState(t *testing.T) {
	m := dashboardUseStateForUnknown()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringUnknown(),
		StateValue: types.StringNull(),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	m.PlanModifyString(context.Background(), req, resp)
	// When state is null (new resource), plan value is left unchanged (unknown).
	assert.True(t, resp.PlanValue.IsUnknown())
}

func TestDashboardUseStateForUnknown_KnownPlanIsUnchanged(t *testing.T) {
	m := dashboardUseStateForUnknown()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue("new-value"),
		StateValue: types.StringValue("old-value"),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	m.PlanModifyString(context.Background(), req, resp)
	// Known plan value is never overwritten.
	assert.Equal(t, types.StringValue("new-value"), resp.PlanValue)
}

// ── dashboardRequiresReplace tests ───────────────────────────────────────────

func TestDashboardRequiresReplace_Description(t *testing.T) {
	m := dashboardRequiresReplace()
	assert.NotEmpty(t, m.Description(context.Background()))
	assert.NotEmpty(t, m.MarkdownDescription(context.Background()))
}

func TestDashboardRequiresReplace_NullState_NoReplace(t *testing.T) {
	m := dashboardRequiresReplace()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue("proj-1"),
		StateValue: types.StringNull(),
	}
	resp := &planmodifier.StringResponse{RequiresReplace: false}
	m.PlanModifyString(context.Background(), req, resp)
	// No replacement when creating (state is null).
	assert.False(t, resp.RequiresReplace)
}

func TestDashboardRequiresReplace_SameValue_NoReplace(t *testing.T) {
	m := dashboardRequiresReplace()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue("proj-1"),
		StateValue: types.StringValue("proj-1"),
	}
	resp := &planmodifier.StringResponse{RequiresReplace: false}
	m.PlanModifyString(context.Background(), req, resp)
	assert.False(t, resp.RequiresReplace)
}

func TestDashboardRequiresReplace_ChangedValue_RequiresReplace(t *testing.T) {
	m := dashboardRequiresReplace()
	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue("proj-2"),
		StateValue: types.StringValue("proj-1"),
	}
	resp := &planmodifier.StringResponse{RequiresReplace: false}
	m.PlanModifyString(context.Background(), req, resp)
	assert.True(t, resp.RequiresReplace)
}

// ── dashboardUseStateForUnknownInt64 tests ───────────────────────────────────

func TestDashboardUseStateForUnknownInt64_Description(t *testing.T) {
	m := dashboardUseStateForUnknownInt64()
	assert.NotEmpty(t, m.Description(context.Background()))
	assert.NotEmpty(t, m.MarkdownDescription(context.Background()))
}

func TestDashboardUseStateForUnknownInt64_UnknownPlanWithKnownState(t *testing.T) {
	m := dashboardUseStateForUnknownInt64()
	req := planmodifier.Int64Request{
		PlanValue:  types.Int64Unknown(),
		StateValue: types.Int64Value(5),
	}
	resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
	m.PlanModifyInt64(context.Background(), req, resp)
	assert.Equal(t, types.Int64Value(5), resp.PlanValue)
}

func TestDashboardUseStateForUnknownInt64_UnknownPlanWithNullState(t *testing.T) {
	m := dashboardUseStateForUnknownInt64()
	req := planmodifier.Int64Request{
		PlanValue:  types.Int64Unknown(),
		StateValue: types.Int64Null(),
	}
	resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
	m.PlanModifyInt64(context.Background(), req, resp)
	assert.True(t, resp.PlanValue.IsUnknown())
}

func TestDashboardUseStateForUnknownInt64_KnownPlanIsUnchanged(t *testing.T) {
	m := dashboardUseStateForUnknownInt64()
	req := planmodifier.Int64Request{
		PlanValue:  types.Int64Value(0),
		StateValue: types.Int64Value(5),
	}
	resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
	m.PlanModifyInt64(context.Background(), req, resp)
	assert.Equal(t, types.Int64Value(0), resp.PlanValue)
}

// ── Resource metadata and schema tests ───────────────────────────────────────

func TestDashboardResource_Metadata(t *testing.T) {
	r := NewDashboardResource()
	req := resource.MetadataRequest{ProviderTypeName: "betterado"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	assert.Equal(t, "betterado_dashboard", resp.TypeName)
}

func TestDashboardResource_Schema_ContainsRequiredAttributes(t *testing.T) {
	r := NewDashboardResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "schema should have no errors")

	attrs := resp.Schema.Attributes
	requiredAttrs := []string{"name", "project_id"}
	for _, attr := range requiredAttrs {
		a, ok := attrs[attr]
		require.True(t, ok, "schema should contain attribute %q", attr)
		assert.True(t, a.IsRequired(), "attribute %q should be required", attr)
	}

	computedAttrs := []string{"id", "owner_id"}
	for _, attr := range computedAttrs {
		a, ok := attrs[attr]
		require.True(t, ok, "schema should contain attribute %q", attr)
		assert.True(t, a.IsComputed(), "attribute %q should be computed", attr)
	}

	optionalAttrs := []string{"team_id", "description", "refresh_interval"}
	for _, attr := range optionalAttrs {
		a, ok := attrs[attr]
		require.True(t, ok, "schema should contain attribute %q", attr)
		assert.True(t, a.IsOptional(), "attribute %q should be optional", attr)
	}
}

func TestDashboardResource_Schema_NoErrors(t *testing.T) {
	r := NewDashboardResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

// ── NewDashboardResource interface compliance ─────────────────────────────────

func TestNewDashboardResource_ImplementsInterfaces(t *testing.T) {
	r := NewDashboardResource()
	require.NotNil(t, r)

	// Verify it implements resource.Resource (static compile-time check).
	var _ resource.Resource = r

	// Verify it implements resource.ResourceWithConfigure via type assertion.
	_, ok := r.(resource.ResourceWithConfigure)
	assert.True(t, ok, "DashboardResource should implement resource.ResourceWithConfigure")

	// Verify it implements resource.ResourceWithImportState via type assertion.
	_, ok = r.(resource.ResourceWithImportState)
	assert.True(t, ok, "DashboardResource should implement resource.ResourceWithImportState")
}
