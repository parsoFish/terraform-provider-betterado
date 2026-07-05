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
	_ resource.Resource              = (*ServiceEndpointCheckMarxOneResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointCheckMarxOneResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seCheckMarxOneRequiresReplaceModifier struct{}

func seCheckMarxOneRequiresReplace() planmodifier.String {
	return seCheckMarxOneRequiresReplaceModifier{}
}

func (m seCheckMarxOneRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seCheckMarxOneRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seCheckMarxOneRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seCheckMarxOneUseStateForUnknownModifier struct{}

func seCheckMarxOneUseStateForUnknown() planmodifier.String {
	return seCheckMarxOneUseStateForUnknownModifier{}
}

func (m seCheckMarxOneUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seCheckMarxOneUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seCheckMarxOneUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seCheckMarxOneStringDefault struct{ value string }

func seCheckMarxOneDefaultString(v string) defaults.String { return seCheckMarxOneStringDefault{v} }
func (d seCheckMarxOneStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seCheckMarxOneStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seCheckMarxOneStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointCheckMarxOneResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_checkmarx_one.
type ServiceEndpointCheckMarxOneResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointCheckMarxOneResource returns a new resource.Resource.
func NewServiceEndpointCheckMarxOneResource() resource.Resource {
	return &ServiceEndpointCheckMarxOneResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_checkmarx_one"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Checkmarx One Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seCheckMarxOneUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seCheckMarxOneRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seCheckMarxOneDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"server_url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the Checkmarx One server.",
			},
			// API key auth
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API key for Checkmarx One authentication. Conflicts with client_id/client_secret.",
			},
			// OAuth client auth
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The client ID for OAuth authentication. Conflicts with api_key.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The client secret for OAuth authentication. Conflicts with api_key.",
			},
			"authorization_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seCheckMarxOneDefaultString(""),
				Description: "The authorization URL for OAuth authentication. Conflicts with api_key.",
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

func (r *ServiceEndpointCheckMarxOneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointCheckMarxOneModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	ServerURL           types.String `tfsdk:"server_url"`
	APIKey              types.String `tfsdk:"api_key"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	AuthorizationURL    types.String `tfsdk:"authorization_url"`
	Authorization       types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointCheckMarxOneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	authorization, scheme := seCheckMarxOneBuildAuth(plan)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:          converter.String(name),
		Owner:         converter.String("library"),
		Description:   converter.String(description),
		Type:          converter.String("CheckmarxASTService"),
		Url:           converter.String(plan.ServerURL.ValueString()),
		Authorization: authorization,
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
	plan.Authorization = seCheckMarxOneBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointCheckMarxOneModel
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

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
		if scheme == "UsernamePassword" && ep.Authorization.Parameters != nil {
			params := *ep.Authorization.Parameters
			if v, ok := params["authURL"]; ok {
				state.AuthorizationURL = types.StringValue(v)
			}
			if v, ok := params["username"]; ok {
				state.ClientID = types.StringValue(v)
			}
			// NOTE: API never returns client_secret / api_key; preserve state values.
		}
	}
	state.Authorization = seCheckMarxOneBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointCheckMarxOneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointCheckMarxOneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	authorization, scheme := seCheckMarxOneBuildAuth(plan)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:            &endpointID,
		Name:          converter.String(name),
		Owner:         converter.String("library"),
		Description:   converter.String(description),
		Type:          converter.String("CheckmarxASTService"),
		Url:           converter.String(plan.ServerURL.ValueString()),
		Authorization: authorization,
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
	plan.Authorization = seCheckMarxOneBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointCheckMarxOneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointCheckMarxOneModel
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

// seCheckMarxOneBuildAuth returns the EndpointAuthorization and scheme string
// based on the plan model (api_key takes priority over client_id).
func seCheckMarxOneBuildAuth(plan serviceEndpointCheckMarxOneModel) (*serviceendpoint.EndpointAuthorization, string) {
	if !plan.APIKey.IsNull() && !plan.APIKey.IsUnknown() && plan.APIKey.ValueString() != "" {
		return &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"apitoken": plan.APIKey.ValueString(),
			},
			Scheme: converter.String("Token"),
		}, "Token"
	}
	return &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{
			"authURL":  plan.AuthorizationURL.ValueString(),
			"username": plan.ClientID.ValueString(),
			"password": plan.ClientSecret.ValueString(),
		},
		Scheme: converter.String("UsernamePassword"),
	}, "UsernamePassword"
}

func seCheckMarxOneBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
