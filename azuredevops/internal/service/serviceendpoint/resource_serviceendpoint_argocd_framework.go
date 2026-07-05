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
	_ resource.Resource              = (*ServiceEndpointArgoCDResource)(nil)
	_ resource.ResourceWithConfigure = (*ServiceEndpointArgoCDResource)(nil)
)

// ── local plan modifier + default helpers ─────────────────────────────────────

type seArgoCDRequiresReplaceModifier struct{}

func seArgoCDRequiresReplace() planmodifier.String { return seArgoCDRequiresReplaceModifier{} }
func (m seArgoCDRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seArgoCDRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m seArgoCDRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type seArgoCDUseStateForUnknownModifier struct{}

func seArgoCDUseStateForUnknown() planmodifier.String {
	return seArgoCDUseStateForUnknownModifier{}
}

func (m seArgoCDUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seArgoCDUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m seArgoCDUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type seArgoCDStringDefault struct{ value string }

func seArgoCDDefaultString(v string) defaults.String { return seArgoCDStringDefault{v} }
func (d seArgoCDStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seArgoCDStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d seArgoCDStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Resource struct ───────────────────────────────────────────────────────────

// ServiceEndpointArgoCDResource is the terraform-plugin-framework implementation
// of betterado_serviceendpoint_argocd.
type ServiceEndpointArgoCDResource struct {
	client *client.AggregatedClient
}

// NewServiceEndpointArgoCDResource returns a new resource.Resource.
func NewServiceEndpointArgoCDResource() resource.Resource {
	return &ServiceEndpointArgoCDResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_argocd"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ArgoCD Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					seArgoCDUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					seArgoCDRequiresReplace(),
				},
			},
			"service_endpoint_name": schema.StringAttribute{
				Required:    true,
				Description: "The Service Endpoint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     seArgoCDDefaultString("Managed by Terraform"),
				Description: "The Service Endpoint description.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Url for the ArgoCD Server.",
			},
			"authentication_token": schema.ListNestedAttribute{
				Optional:    true,
				Description: "An authentication_token block as documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The ArgoCD access token.",
						},
					},
				},
			},
			"authentication_basic": schema.ListNestedAttribute{
				Optional:    true,
				Description: "An authentication_basic block as documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The ArgoCD user name.",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The ArgoCD password.",
						},
					},
				},
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

func (r *ServiceEndpointArgoCDResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type serviceEndpointArgoCDAuthTokenModel struct {
	Token types.String `tfsdk:"token"`
}

type serviceEndpointArgoCDAuthBasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type serviceEndpointArgoCDModel struct {
	ID                  types.String                          `tfsdk:"id"`
	ProjectID           types.String                          `tfsdk:"project_id"`
	ServiceEndpointName types.String                          `tfsdk:"service_endpoint_name"`
	Description         types.String                          `tfsdk:"description"`
	URL                 types.String                          `tfsdk:"url"`
	AuthenticationToken []serviceEndpointArgoCDAuthTokenModel `tfsdk:"authentication_token"`
	AuthenticationBasic []serviceEndpointArgoCDAuthBasicModel `tfsdk:"authentication_basic"`
	Authorization       types.Map                             `tfsdk:"authorization"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointArgoCDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	authScheme, authParams := seArgoCDExpandAuth(plan)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("argocd"),
		Url:         converter.String(plan.URL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &authParams,
			Scheme:     converter.String(authScheme),
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
	plan.Authorization = seArgoCDBuildAuthMap(ctx, authScheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointArgoCDModel
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
		state.URL = types.StringValue(*ep.Url)
	}

	// NOTE: API never returns token/password for argocd; preserve state values.
	scheme := ""
	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		scheme = *ep.Authorization.Scheme
	}
	state.Authorization = seArgoCDBuildAuthMap(ctx, scheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var plan serviceEndpointArgoCDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceEndpointArgoCDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID := uuid.MustParse(state.ID.ValueString())
	projectID := uuid.MustParse(plan.ProjectID.ValueString())
	name := plan.ServiceEndpointName.ValueString()
	description := plan.Description.ValueString()

	authScheme, authParams := seArgoCDExpandAuth(plan)

	endpoint := &serviceendpoint.ServiceEndpoint{
		Id:          &endpointID,
		Name:        converter.String(name),
		Owner:       converter.String("library"),
		Description: converter.String(description),
		Type:        converter.String("argocd"),
		Url:         converter.String(plan.URL.ValueString()),
		Authorization: &serviceendpoint.EndpointAuthorization{
			Parameters: &authParams,
			Scheme:     converter.String(authScheme),
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
	plan.Authorization = seArgoCDBuildAuthMap(ctx, authScheme)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServiceEndpointArgoCDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var state serviceEndpointArgoCDModel
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

func seArgoCDExpandAuth(plan serviceEndpointArgoCDModel) (string, map[string]string) {
	params := make(map[string]string)
	if len(plan.AuthenticationToken) > 0 {
		params["apitoken"] = plan.AuthenticationToken[0].Token.ValueString()
		return "Token", params
	}
	if len(plan.AuthenticationBasic) > 0 {
		params["username"] = plan.AuthenticationBasic[0].Username.ValueString()
		params["password"] = plan.AuthenticationBasic[0].Password.ValueString()
		return "UsernamePassword", params
	}
	return "Token", params
}

func seArgoCDBuildAuthMap(ctx context.Context, scheme string) types.Map {
	elems := map[string]string{"scheme": scheme}
	m, _ := types.MapValueFrom(ctx, types.StringType, elems)
	return m
}
