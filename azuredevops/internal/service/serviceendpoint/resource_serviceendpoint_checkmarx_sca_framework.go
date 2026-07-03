package serviceendpoint

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*ServiceEndpointCheckMarxSCAResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointCheckMarxSCAResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seCheckMarxSCARequiresReplaceModifier struct{}

func seCheckMarxSCARequiresReplace() planmodifier.String {
	return seCheckMarxSCARequiresReplaceModifier{}
}

func (m seCheckMarxSCARequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seCheckMarxSCARequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seCheckMarxSCARequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seCheckMarxSCAUseStateForUnknownModifier struct{}

func seCheckMarxSCAUseStateForUnknown() planmodifier.String {
	return seCheckMarxSCAUseStateForUnknownModifier{}
}

func (m seCheckMarxSCAUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seCheckMarxSCAUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seCheckMarxSCAUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seCheckMarxSCAStringDefault struct{ value string }

func seCheckMarxSCADefaultString(v string) defaults.String { return seCheckMarxSCAStringDefault{v} }
func (d seCheckMarxSCAStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seCheckMarxSCAStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seCheckMarxSCAStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointCheckMarxSCAResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_checkmarx_sca.
type ServiceEndpointCheckMarxSCAResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointCheckMarxSCAResource returns a new resource.Resource.
func NewServiceEndpointCheckMarxSCAResource() resource.Resource {
	return &ServiceEndpointCheckMarxSCAResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_checkmarx_sca"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Checkmarx SCA Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seCheckMarxSCAUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seCheckMarxSCARequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seCheckMarxSCADefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"server_url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the Checkmarx SCA server.",
			},
			"access_control_url": schema.StringAttribute{
				Required:    true,
				Description: "The access control URL for Checkmarx SCA.",
			},
			"web_app_url": schema.StringAttribute{
				Required:    true,
				Description: "The web application URL for Checkmarx SCA.",
			},
			"account": schema.StringAttribute{
				Required:    true,
				Description: "The Checkmarx SCA account (tenant).",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username for authentication.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password for authentication.",
			},
			"team": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seCheckMarxSCADefaultString(""),
				Description: "The Checkmarx SCA team.",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointCheckMarxSCAModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	ServerURL           types.String `tfsdk:"server_url"`
	AccessControlURL    types.String `tfsdk:"access_control_url"`
	WebAppURL           types.String `tfsdk:"web_app_url"`
	Account             types.String `tfsdk:"account"`
	Username            types.String `tfsdk:"username"`
	Password            types.String `tfsdk:"password"`
	Team                types.String `tfsdk:"team"`
	Authorization       types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointCheckMarxSCAModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("SCA-Endpoint"),
		Url:         converter.String(plan.ServerURL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"username": plan.Username.ValueString(),
				"password": plan.Password.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
		},
		Data: &map[string]string{
			"dependencyAccessControlURL": plan.AccessControlURL.ValueString(),
			"dependencyTenant":           plan.Account.ValueString(),
			"dependencyWebAppURL":        plan.WebAppURL.ValueString(),
			"teams":                      plan.Team.ValueString(),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}

	created, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(ctx, serviceendpoint.CreateServiceEndpointArgs{
		Endpoint: endpoint,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service endpoint", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id.String())
	plan.Authorization = seCheckMarxSCABuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointCheckMarxSCAModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := state.ProjectID.ValueString()

	ep, err := r.client.ServiceEndpointClient.GetServiceEndpointDetails(ctx, serviceendpoint.GetServiceEndpointDetailsArgs{
		EndpointId: &endpointID,
		Project:    &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading service endpoint", err.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if ep.Name != nil {
		state.ServiceEndpointName = types.StringValue(*ep.Name)
	}
	if ep.Description != nil {
		state.Description = types.StringValue(*ep.Description)
	}
	if ep.Url != nil {
		state.ServerURL = types.StringValue(*ep.Url)
	}

	if ep.Data != nil {
		data := *ep.Data
		if v, ok := data["dependencyAccessControlURL"]; ok {
			state.AccessControlURL = types.StringValue(v)
		}
		if v, ok := data["dependencyWebAppURL"]; ok {
			state.WebAppURL = types.StringValue(v)
		}
		if v, ok := data["dependencyTenant"]; ok {
			state.Account = types.StringValue(v)
		}
		if v, ok := data["teams"]; ok {
			state.Team = types.StringValue(v)
		}
	}

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["username"]; ok {
			state.Username = types.StringValue(v)
		}
		// NOTE: API never returns password; preserve state value to avoid spurious diffs.
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seCheckMarxSCABuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointCheckMarxSCAModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointCheckMarxSCAModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:          &endpointID,
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("SCA-Endpoint"),
		Url:         converter.String(plan.ServerURL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"username": plan.Username.ValueString(),
				"password": plan.Password.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
		},
		Data: &map[string]string{
			"dependencyAccessControlURL": plan.AccessControlURL.ValueString(),
			"dependencyTenant":           plan.Account.ValueString(),
			"dependencyWebAppURL":        plan.WebAppURL.ValueString(),
			"teams":                      plan.Team.ValueString(),
		},
		ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
			{
				ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
				Name:             converter.String(name),
				Description:      converter.String(description),
			},
		},
	}

	_, err := r.client.ServiceEndpointClient.UpdateServiceEndpoint(ctx, serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   endpoint,
		EndpointId: &endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service endpoint", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Authorization = seCheckMarxSCABuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxSCAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointCheckMarxSCAModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := state.ProjectID.ValueString()
	projectIDs := []string{projectID}

	if err := r.client.ServiceEndpointClient.DeleteServiceEndpoint(ctx, serviceendpoint.DeleteServiceEndpointArgs{
		EndpointId: &endpointID,
		ProjectIds: &projectIDs,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting service endpoint", err.Error())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func seCheckMarxSCABuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
