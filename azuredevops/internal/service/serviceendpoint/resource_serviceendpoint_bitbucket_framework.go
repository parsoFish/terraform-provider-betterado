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
	_ resource.Resource              = (*ServiceEndpointBitBucketResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointBitBucketResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seBitBucketRequiresReplaceModifier struct{}

func seBitBucketRequiresReplace() planmodifier.String { return seBitBucketRequiresReplaceModifier{} }

func (m seBitBucketRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seBitBucketRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seBitBucketRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seBitBucketUseStateForUnknownModifier struct{}

func seBitBucketUseStateForUnknown() planmodifier.String {
	return seBitBucketUseStateForUnknownModifier{}
}

func (m seBitBucketUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seBitBucketUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seBitBucketUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seBitBucketStringDefault struct{ value string }

func seBitBucketDefaultString(v string) defaults.String { return seBitBucketStringDefault{v} }
func (d seBitBucketStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seBitBucketStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seBitBucketStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointBitBucketResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_bitbucket.
type ServiceEndpointBitBucketResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointBitBucketResource returns a new resource.Resource.
func NewServiceEndpointBitBucketResource() resource.Resource {
	return &ServiceEndpointBitBucketResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_bitbucket"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seBitBucketUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seBitBucketRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seBitBucketDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The Bitbucket username.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The Bitbucket password.",
			},
			"authorization": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Specifies the authorization scheme.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointBitBucketModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	Description         types.String `tfsdk:"description"`
	Username            types.String `tfsdk:"username"`
	Password            types.String `tfsdk:"password"`
	Authorization       types.Map    `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointBitBucketModel
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
		Type:        converter.String("Bitbucket"),
		Url:         converter.String("https://api.bitbucket.org/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"username": plan.Username.ValueString(),
				"password": plan.Password.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
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
	plan.Authorization = seBitBucketBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointBitBucketModel
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
	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		if u, ok := (*ep.Authorization.Parameters)["username"]; ok {
			state.Username = types.StringValue(u)
		}
		// NOTE: password is never returned by API; preserve state value.
	}

	scheme := "UsernamePassword"
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seBitBucketBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointBitBucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointBitBucketModel
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
		Type:        converter.String("Bitbucket"),
		Url:         converter.String("https://api.bitbucket.org/"),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &map[string]string{
				"username": plan.Username.ValueString(),
				"password": plan.Password.ValueString(),
			},
			Scheme: converter.String("UsernamePassword"),
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
	plan.Authorization = seBitBucketBuildAuthMap(ctx, "UsernamePassword")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointBitBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointBitBucketModel
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

func seBitBucketBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
