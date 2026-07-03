package dashboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	azdodashboard "github.com/microsoft/azure-devops-go-api/azuredevops/v7/dashboard"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/dashboardextras"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*DashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*DashboardResource)(nil)
	_ resource.ResourceWithImportState = (*DashboardResource)(nil)
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// useStateForUnknownString copies the prior state value when the plan value is
// unknown (prevents perpetual diffs for computed-only attributes).
type dashboardUseStateForUnknownString struct{}

func dashboardUseStateForUnknown() planmodifier.String { return dashboardUseStateForUnknownString{} }

func (m dashboardUseStateForUnknownString) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m dashboardUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m dashboardUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// dashboardRequiresReplaceString marks the resource for replacement when the
// attribute value changes.
type dashboardRequiresReplaceString struct{}

func dashboardRequiresReplace() planmodifier.String { return dashboardRequiresReplaceString{} }
func (m dashboardRequiresReplaceString) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m dashboardRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m dashboardRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// dashboardUseStateForUnknownInt64Struct copies the prior state value when
// the plan value is unknown for Int64 attributes.
type dashboardUseStateForUnknownInt64Struct struct{}

func dashboardUseStateForUnknownInt64() planmodifier.Int64 {
	return dashboardUseStateForUnknownInt64Struct{}
}

func (m dashboardUseStateForUnknownInt64Struct) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m dashboardUseStateForUnknownInt64Struct) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m dashboardUseStateForUnknownInt64Struct) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Resource struct ───────────────────────────────────────────────────────────

// DashboardResource is the terraform-plugin-framework implementation of
// betterado_dashboard.
type DashboardResource struct {
	client *client.AggregatedClient
}

// NewDashboardResource returns a new resource.Resource for betterado_dashboard.
func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *DashboardResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_dashboard"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					dashboardUseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Dashboard.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Project.",
				PlanModifiers: []planmodifier.String{
					dashboardRequiresReplace(),
				},
			},
			"team_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the team associated with the Dashboard. If not provided, the Dashboard is project-scoped.",
				PlanModifiers: []planmodifier.String{
					dashboardRequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the Dashboard.",
			},
			"refresh_interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The interval for the client to automatically refresh the Dashboard. Expressed in minutes. Valid values: 0 (disabled), 5.",
				PlanModifiers: []planmodifier.Int64{
					dashboardUseStateForUnknownInt64(),
				},
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the owner of the Dashboard.",
				PlanModifiers: []planmodifier.String{
					dashboardUseStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── State model ───────────────────────────────────────────────────────────────

// dashboardModel is the Terraform state model for betterado_dashboard.
type dashboardModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ProjectID       types.String `tfsdk:"project_id"`
	TeamID          types.String `tfsdk:"team_id"`
	Description     types.String `tfsdk:"description"`
	RefreshInterval types.Int64  `tfsdk:"refresh_interval"`
	OwnerID         types.String `tfsdk:"owner_id"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	params := azdodashboard.CreateDashboardArgs{
		Project: &projectID,
		Dashboard: &azdodashboard.Dashboard{
			Name:        converter.String(model.Name.ValueString()),
			Description: converter.String(model.Description.ValueString()),
		},
	}

	if !model.RefreshInterval.IsNull() && !model.RefreshInterval.IsUnknown() {
		ri := int(model.RefreshInterval.ValueInt64())
		params.Dashboard.RefreshInterval = &ri
	}

	if !model.TeamID.IsNull() && !model.TeamID.IsUnknown() && model.TeamID.ValueString() != "" {
		params.Team = converter.String(model.TeamID.ValueString())
	}

	created, err := r.client.DashboardClient.CreateDashboard(r.client.Ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Creating Dashboard in Azure DevOps", err.Error())
		return
	}

	model.ID = types.StringValue(created.Id.String())

	found := r.readIntoModel(&model)
	if !found {
		resp.Diagnostics.AddError("Error reading Dashboard after create", "Dashboard was not found immediately after creation")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := r.readIntoModel(&model)
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = stateModel.ID

	dashboardID, err := uuid.Parse(stateModel.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Dashboard ID", err.Error())
		return
	}

	projectID := plan.ProjectID.ValueString()

	// First fetch the existing dashboard (required to preserve server-set fields).
	getArgs := azdodashboard.GetDashboardArgs{
		Project:     &projectID,
		DashboardId: &dashboardID,
	}
	if !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() && plan.TeamID.ValueString() != "" {
		getArgs.Team = converter.String(plan.TeamID.ValueString())
	}

	existing, err := r.client.DashboardClient.GetDashboard(r.client.Ctx, getArgs)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dashboard", err.Error())
		return
	}

	existing.Name = converter.String(plan.Name.ValueString())
	existing.Description = converter.String(plan.Description.ValueString())

	if !plan.RefreshInterval.IsNull() && !plan.RefreshInterval.IsUnknown() {
		ri := int(plan.RefreshInterval.ValueInt64())
		existing.RefreshInterval = &ri
	}

	updateArgs := dashboardextras.UpdateDashboardArgs{
		Project:   &projectID,
		Dashboard: existing,
	}
	if !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() && plan.TeamID.ValueString() != "" {
		updateArgs.Team = converter.String(plan.TeamID.ValueString())
	}

	_, err = r.client.DashboardClientExtra.UpdateDashboard(r.client.Ctx, updateArgs)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dashboard", err.Error())
		return
	}

	found := r.readIntoModel(&plan)
	if !found {
		resp.Diagnostics.AddError("Error reading Dashboard after update", "Dashboard was not found after update")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Dashboard ID", err.Error())
		return
	}

	projectID := model.ProjectID.ValueString()
	params := azdodashboard.DeleteDashboardArgs{
		Project:     &projectID,
		DashboardId: &dashboardID,
	}

	if !model.TeamID.IsNull() && !model.TeamID.IsUnknown() && model.TeamID.ValueString() != "" {
		params.Team = converter.String(model.TeamID.ValueString())
	}

	err = r.client.DashboardClient.DeleteDashboard(r.client.Ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "VS402433") {
			resp.Diagnostics.AddError(
				"Cannot delete last team dashboard",
				fmt.Sprintf("You can't delete the last dashboard of a team. You can delete the team directly. Dashboard ID: %s. Error: %s", model.ID.ValueString(), err.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("Error deleting Dashboard", err.Error())
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) < 2 || len(idParts) > 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Unexpected ID format (%q), expected: <projectId>/<dashboardId> or <projectId>/<teamId>/<dashboardId>", req.ID),
		)
		return
	}

	var model dashboardModel
	model.ProjectID = types.StringValue(idParts[0])

	if len(idParts) == 2 {
		model.ID = types.StringValue(idParts[1])
		model.TeamID = types.StringNull()
	} else {
		model.TeamID = types.StringValue(idParts[1])
		model.ID = types.StringValue(idParts[2])
	}

	model.Name = types.StringUnknown()
	model.Description = types.StringUnknown()
	model.RefreshInterval = types.Int64Unknown()
	model.OwnerID = types.StringUnknown()

	found := r.readIntoModel(&model)
	if !found {
		resp.Diagnostics.AddError("Dashboard not found", fmt.Sprintf("Dashboard with ID %s was not found", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Internal read helper ──────────────────────────────────────────────────────

// readIntoModel fetches the dashboard by ID and populates model. Returns false
// if the resource was not found.
func (r *DashboardResource) readIntoModel(model *dashboardModel) bool {
	dashboardID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		return false
	}

	projectID := model.ProjectID.ValueString()
	params := azdodashboard.GetDashboardArgs{
		Project:     &projectID,
		DashboardId: &dashboardID,
	}

	if !model.TeamID.IsNull() && !model.TeamID.IsUnknown() && model.TeamID.ValueString() != "" {
		params.Team = converter.String(model.TeamID.ValueString())
	}

	apiResp, err := r.client.DashboardClient.GetDashboard(r.client.Ctx, params)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return false
		}
		return false
	}

	if apiResp == nil {
		return false
	}

	if apiResp.Name != nil {
		model.Name = types.StringValue(*apiResp.Name)
	}
	if apiResp.Description != nil {
		model.Description = types.StringValue(*apiResp.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if apiResp.RefreshInterval != nil {
		model.RefreshInterval = types.Int64Value(int64(*apiResp.RefreshInterval))
	} else {
		model.RefreshInterval = types.Int64Value(0)
	}
	if apiResp.OwnerId != nil {
		model.OwnerID = types.StringValue(apiResp.OwnerId.String())
	}

	// team_id: use the GroupId from the response if team_id was not already set.
	if (model.TeamID.IsNull() || model.TeamID.IsUnknown() || model.TeamID.ValueString() == "") && apiResp.GroupId != nil {
		model.TeamID = types.StringValue(apiResp.GroupId.String())
	}

	return true
}
