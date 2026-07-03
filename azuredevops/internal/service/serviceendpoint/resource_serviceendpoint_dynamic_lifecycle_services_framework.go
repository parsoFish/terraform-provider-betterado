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
	_ resource.Resource              = (*ServiceEndpointDynamicsLifecycleServicesResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointDynamicsLifecycleServicesResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seDynamicsLCSRequiresReplaceModifier struct{}

func seDynamicsLCSRequiresReplace() planmodifier.String {
	return seDynamicsLCSRequiresReplaceModifier{}
}

func (m seDynamicsLCSRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seDynamicsLCSRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seDynamicsLCSRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seDynamicsLCSUseStateForUnknownModifier struct{}

func seDynamicsLCSUseStateForUnknown() planmodifier.String {
	return seDynamicsLCSUseStateForUnknownModifier{}
}

func (m seDynamicsLCSUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seDynamicsLCSUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seDynamicsLCSUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seDynamicsLCSStringDefault struct{ value string }

func seDynamicsLCSDefaultString(v string) defaults.String { return seDynamicsLCSStringDefault{v} }

func (d seDynamicsLCSStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seDynamicsLCSStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seDynamicsLCSStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointDynamicsLifecycleServicesResource is the terraform-plugin-framework
// implementation of betterado_serviceendpoint_dynamics_lifecycle_services.
type ServiceEndpointDynamicsLifecycleServicesResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointDynamicsLifecycleServicesResource returns a new resource.Resource.
func NewServiceEndpointDynamicsLifecycleServicesResource() resource.Resource {
	return &ServiceEndpointDynamicsLifecycleServicesResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_dynamics_lifecycle_services"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Microsoft Dynamics Lifecycle Services Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seDynamicsLCSUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seDynamicsLCSRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seDynamicsLCSDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"authorization_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "The authorization endpoint URL.",
			},
			"lifecycle_services_api_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "The Lifecycle Services API endpoint URL.",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "The client ID (UUID) for the service principal.",
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
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme and parameters.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointDynamicsLCSModel struct {
	ID                           types.String `tfsdk:"id"`
	ProjectID                    types.String `tfsdk:"project_id"`
	ServiceEndpointName          types.String `tfsdk:"service_endpoint_name"`
	Description                  types.String `tfsdk:"description"`
	AuthorizationEndpoint        types.String `tfsdk:"authorization_endpoint"`
	LifecycleServicesAPIEndpoint types.String `tfsdk:"lifecycle_services_api_endpoint"`
	ClientID                     types.String `tfsdk:"client_id"`
	Username                     types.String `tfsdk:"username"`
	Password                     types.String `tfsdk:"password"`
	Authorization                types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointDynamicsLCSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := seDynamicsLCSBuildEndpoint(plan)
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(name),
			Description:      converter.String(description),
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
	plan.Authorization = seDynamicsLCSBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointDynamicsLCSModel
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
		state.AuthorizationEndpoint = types.StringValue(*ep.Url)
	}
	if ep.Data != nil {
		if v, ok := (*ep.Data)["apiurl"]; ok {
			state.LifecycleServicesAPIEndpoint = types.StringValue(v)
		}
	}
	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["clientid"]; ok {
			state.ClientID = types.StringValue(v)
		}
		if v, ok := params["username"]; ok {
			state.Username = types.StringValue(v)
		}
		// NOTE: API never returns password; preserve state value to avoid spurious diffs.
	}

	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seDynamicsLCSBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointDynamicsLCSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointDynamicsLCSModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	endpoint := seDynamicsLCSBuildEndpoint(plan)
	endpoint.Id = &endpointID
	endpoint.ServiceEndpointProjectReferences = &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{Id: &projectID},
			Name:             converter.String(name),
			Description:      converter.String(description),
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
	plan.Authorization = seDynamicsLCSBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointDynamicsLifecycleServicesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointDynamicsLCSModel
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

func seDynamicsLCSBuildEndpoint(plan serviceEndpointDynamicsLCSModel) *serviceendpoint.ServiceEndpoint {
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	return &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("lcsserviceendpoint"),
		Url:         converter.String(plan.AuthorizationEndpoint.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"clientid": plan.ClientID.ValueString(),
				"username": plan.Username.ValueString(),
				"password": plan.Password.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
		},
		Data: &map[string]string{
			"apiurl": plan.LifecycleServicesAPIEndpoint.ValueString(),
		},
	}
}

func seDynamicsLCSBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
